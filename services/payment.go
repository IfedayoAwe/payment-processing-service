package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/lib/pq"
)

type PaymentService interface {
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency money.Currency) (float64, error)
	CreateInternalTransfer(ctx context.Context, fromUserID string, toAccountNumber string, toBankCode string, fromCurrency money.Currency, toAmount money.Money, idempotencyKey string) (*models.Transaction, error)
	CreateExternalTransfer(ctx context.Context, userID string, toAccountNumber string, toBankCode string, fromCurrency money.Currency, toAmount money.Money, idempotencyKey string) (*models.Transaction, error)
	ConfirmTransaction(ctx context.Context, transactionID string, userID string, pin string) (*models.Transaction, error)
	GetTransactionByID(ctx context.Context, transactionID string) (*models.Transaction, error)
	GetTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Transaction, error)
	GetTransactionHistory(ctx context.Context, userID string, cursor string, limit int32) (*models.TransactionHistoryResponse, error)
}

type paymentService struct {
	queries          gen.Querier
	db               *sql.DB
	wallet           WalletService
	ledger           LedgerService
	externalTransfer ExternalTransferService
	provider         *providers.Processor
}

func newPaymentService(queries gen.Querier, db *sql.DB, wallet WalletService, ledger LedgerService, externalTransfer ExternalTransferService, provider *providers.Processor) PaymentService {
	return &paymentService{
		queries:          queries,
		db:               db,
		wallet:           wallet,
		ledger:           ledger,
		externalTransfer: externalTransfer,
		provider:         provider,
	}
}

func (ps *paymentService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency money.Currency) (float64, error) {
	rateResp, err := ps.provider.GetExchangeRate(ctx, providers.ExchangeRateRequest{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
	})
	if err != nil {
		return 0, utils.ServerErr(fmt.Errorf("get exchange rate: %w", err))
	}
	return rateResp.Rate, nil
}

func (ps *paymentService) CreateInternalTransfer(ctx context.Context, fromUserID string, toAccountNumber string, toBankCode string, fromCurrency money.Currency, toAmount money.Money, idempotencyKey string) (*models.Transaction, error) {
	if !toAmount.IsPositive() {
		return nil, utils.BadRequestErr("amount must be positive")
	}

	existing, err := ps.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, utils.ServerErr(fmt.Errorf("check idempotency: %w", err))
	}

	exchangeRate, err := ps.GetExchangeRate(ctx, fromCurrency, toAmount.Currency)
	if err != nil {
		return nil, err
	}

	fromWallet, err := ps.wallet.GetWalletByUserAndCurrency(ctx, fromUserID, fromCurrency)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("sender wallet not found")
		}
		return nil, err
	}

	bankAccount, err := ps.queries.GetBankAccountByAccountAndBankCode(ctx, gen.GetBankAccountByAccountAndBankCodeParams{
		AccountNumber: toAccountNumber,
		BankCode:      toBankCode,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("recipient account not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("get recipient bank account: %w", err))
	}

	if bankAccount.Currency != toAmount.Currency.String() {
		return nil, utils.BadRequestErr("recipient account currency mismatch")
	}

	toWallet, err := ps.wallet.GetWalletByBankAccount(ctx, bankAccount.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("recipient wallet not found for bank account")
		}
		return nil, err
	}

	if fromWallet.ID == toWallet.ID {
		return nil, utils.BadRequestErr("cannot transfer to same wallet")
	}

	if fromWallet.UserID == toWallet.UserID {
		return ps.processInternalTransferImmediate(ctx, fromWallet, toWallet, fromCurrency, toAmount, exchangeRate, idempotencyKey)
	}

	return ps.createInitiatedInternalTransfer(ctx, fromWallet, toWallet, toAmount, exchangeRate, idempotencyKey)
}

func (ps *paymentService) processInternalTransferImmediate(ctx context.Context, fromWallet *models.Wallet, toWallet *models.Wallet, fromCurrency money.Currency, toAmount money.Money, exchangeRate float64, idempotencyKey string) (*models.Transaction, error) {
	fromAmount := int64(float64(toAmount.Amount) / exchangeRate)

	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("begin transaction: %w", err))
	}
	defer func() { _ = tx.Rollback() }()

	var queries gen.Querier
	if q, ok := ps.queries.(*gen.Queries); ok {
		queries = q.WithTx(tx)
	} else {
		queries = ps.queries
	}

	lockedFromWallet, err := ps.wallet.LockWalletForUpdate(ctx, tx, fromWallet.ID)
	if err != nil {
		return nil, err
	}

	lockedToWallet, err := ps.wallet.LockWalletForUpdate(ctx, tx, toWallet.ID)
	if err != nil {
		return nil, err
	}

	if lockedFromWallet.Balance < fromAmount {
		return nil, utils.BadRequestErr("insufficient funds")
	}

	exchangeRateStr := strconv.FormatFloat(exchangeRate, 'f', 8, 64)
	traceID := utils.TraceIDFromContext(ctx)
	transaction, err := queries.CreateTransaction(ctx, gen.CreateTransactionParams{
		IdempotencyKey: idempotencyKey,
		TraceID:        sql.NullString{String: traceID, Valid: traceID != ""},
		FromWalletID:   lockedFromWallet.ID,
		ToWalletID:     sql.NullString{String: lockedToWallet.ID, Valid: true},
		Type:           string(models.TransactionTypeInternal),
		Amount:         toAmount.Amount,
		Currency:       toAmount.Currency.String(),
		Status:         string(models.TransactionStatusPending),
		ExchangeRate:   sql.NullString{String: exchangeRateStr, Valid: true},
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			_ = tx.Rollback()
			existing, err := ps.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
			if err != nil {
				return nil, utils.DuplicateKeyErr("transaction with this idempotency key already exists")
			}
			return existing, nil
		}
		return nil, utils.ServerErr(fmt.Errorf("create transaction: %w", err))
	}

	if err := ps.ledger.CreateDebitEntry(ctx, tx, lockedFromWallet.ID, transaction.ID, -fromAmount, fromCurrency); err != nil {
		return nil, err
	}

	if err := ps.ledger.CreateCreditEntry(ctx, tx, lockedToWallet.ID, transaction.ID, toAmount.Amount, toAmount.Currency); err != nil {
		return nil, err
	}

	if err := queries.UpdateTransactionStatus(ctx, gen.UpdateTransactionStatusParams{
		Status: string(models.TransactionStatusCompleted),
		ID:     transaction.ID,
	}); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("update transaction status: %w", err))
	}

	fromWalletMoney := money.NewMoney(lockedFromWallet.Balance, fromCurrency)
	fromAmountMoney := money.NewMoney(fromAmount, fromCurrency)
	newFromBalanceMoney, err := fromWalletMoney.Subtract(fromAmountMoney)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("calculate new from balance: %w", err))
	}

	toWalletMoney := money.NewMoney(lockedToWallet.Balance, toAmount.Currency)
	newToBalanceMoney, err := toWalletMoney.Add(toAmount)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("calculate new to balance: %w", err))
	}

	if err := queries.UpdateWalletBalance(ctx, gen.UpdateWalletBalanceParams{
		Balance: newFromBalanceMoney.Amount,
		ID:      lockedFromWallet.ID,
	}); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("update from wallet balance: %w", err))
	}

	if err := queries.UpdateWalletBalance(ctx, gen.UpdateWalletBalanceParams{
		Balance: newToBalanceMoney.Amount,
		ID:      lockedToWallet.ID,
	}); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("update to wallet balance: %w", err))
	}

	_, err = queries.CreateIdempotencyKey(ctx, gen.CreateIdempotencyKeyParams{
		Key:           idempotencyKey,
		TransactionID: transaction.ID,
	})
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("create idempotency key: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("commit transaction: %w", err))
	}

	return &models.Transaction{
		ID:             transaction.ID,
		IdempotencyKey: transaction.IdempotencyKey,
		FromWalletID:   &lockedFromWallet.ID,
		ToWalletID:     &lockedToWallet.ID,
		Type:           models.TransactionTypeInternal,
		Amount:         transaction.Amount,
		Currency:       transaction.Currency,
		Status:         models.TransactionStatusCompleted,
		CreatedAt:      transaction.CreatedAt,
		UpdatedAt:      transaction.UpdatedAt,
	}, nil
}

func (ps *paymentService) createInitiatedInternalTransfer(ctx context.Context, fromWallet *models.Wallet, toWallet *models.Wallet, toAmount money.Money, exchangeRate float64, idempotencyKey string) (*models.Transaction, error) {
	fromAmount := int64(float64(toAmount.Amount) / exchangeRate)
	if fromWallet.Balance < fromAmount {
		return nil, utils.BadRequestErr("insufficient funds")
	}

	exchangeRateStr := strconv.FormatFloat(exchangeRate, 'f', 8, 64)
	traceID := utils.TraceIDFromContext(ctx)
	transaction, err := ps.queries.CreateTransaction(ctx, gen.CreateTransactionParams{
		IdempotencyKey: idempotencyKey,
		TraceID:        sql.NullString{String: traceID, Valid: traceID != ""},
		FromWalletID:   fromWallet.ID,
		ToWalletID:     sql.NullString{String: toWallet.ID, Valid: true},
		Type:           string(models.TransactionTypeInternal),
		Amount:         toAmount.Amount,
		Currency:       toAmount.Currency.String(),
		Status:         string(models.TransactionStatusInitiated),
		ExchangeRate:   sql.NullString{String: exchangeRateStr, Valid: true},
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			existing, err := ps.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
			if err != nil {
				return nil, utils.DuplicateKeyErr("transaction with this idempotency key already exists")
			}
			return existing, nil
		}
		return nil, utils.ServerErr(fmt.Errorf("create transaction: %w", err))
	}

	return &models.Transaction{
		ID:             transaction.ID,
		IdempotencyKey: transaction.IdempotencyKey,
		FromWalletID:   &fromWallet.ID,
		ToWalletID:     &toWallet.ID,
		Type:           models.TransactionTypeInternal,
		Amount:         transaction.Amount,
		Currency:       transaction.Currency,
		Status:         models.TransactionStatusInitiated,
		CreatedAt:      transaction.CreatedAt,
		UpdatedAt:      transaction.UpdatedAt,
	}, nil
}

func (ps *paymentService) GetTransactionByID(ctx context.Context, transactionID string) (*models.Transaction, error) {
	transaction, err := ps.queries.GetTransactionByID(ctx, transactionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("transaction not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("get transaction: %w", err))
	}

	return mapTransaction(transaction), nil
}

func (ps *paymentService) GetTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Transaction, error) {
	transaction, err := ps.queries.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, utils.ServerErr(fmt.Errorf("get transaction: %w", err))
	}

	return mapTransaction(transaction), nil
}

func (ps *paymentService) CreateExternalTransfer(ctx context.Context, userID string, toAccountNumber string, toBankCode string, fromCurrency money.Currency, toAmount money.Money, idempotencyKey string) (*models.Transaction, error) {
	exchangeRate, err := ps.GetExchangeRate(ctx, fromCurrency, toAmount.Currency)
	if err != nil {
		return nil, err
	}
	return ps.externalTransfer.CreateExternalTransfer(ctx, userID, toAccountNumber, toBankCode, fromCurrency, toAmount, exchangeRate, idempotencyKey)
}

func (ps *paymentService) ConfirmTransaction(ctx context.Context, transactionID string, userID string, pin string) (*models.Transaction, error) {
	transaction, err := ps.queries.GetTransactionByID(ctx, transactionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("transaction not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("get transaction: %w", err))
	}

	if transaction.Status != string(models.TransactionStatusInitiated) {
		return nil, utils.BadRequestErr("transaction is not in initiated status")
	}

	if time.Since(transaction.CreatedAt) > 10*time.Minute {
		return nil, utils.BadRequestErr("transaction has expired")
	}

	fromWallet, err := ps.wallet.GetWalletByID(ctx, transaction.FromWalletID)
	if err != nil {
		return nil, err
	}

	if fromWallet.UserID != userID {
		return nil, utils.BadRequestErr("transaction does not belong to user")
	}

	user, err := ps.queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("user not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("get user: %w", err))
	}

	if !user.PinHash.Valid {
		return nil, utils.BadRequestErr("PIN not set for user")
	}

	if err := utils.VerifyPIN(user.PinHash.String, pin); err != nil {
		return nil, utils.BadRequestErr("invalid PIN")
	}

	if transaction.Type == string(models.TransactionTypeInternal) {
		return ps.confirmInternalTransfer(ctx, transaction)
	}

	if transaction.Type == string(models.TransactionTypeExternal) {
		return ps.externalTransfer.confirmExternalTransfer(ctx, transaction)
	}

	return nil, utils.BadRequestErr("invalid transaction type")
}

func (ps *paymentService) confirmInternalTransfer(ctx context.Context, transaction gen.Transaction) (*models.Transaction, error) {
	if !transaction.ExchangeRate.Valid {
		return nil, utils.ServerErr(fmt.Errorf("exchange rate not found in transaction"))
	}

	exchangeRate, err := strconv.ParseFloat(transaction.ExchangeRate.String, 64)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("invalid exchange rate: %w", err))
	}

	toCurrency, err := money.ParseCurrency(transaction.Currency)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("parse to currency: %w", err))
	}

	toWalletID := transaction.ToWalletID.String
	toWallet, err := ps.wallet.GetWalletByID(ctx, toWalletID)
	if err != nil {
		return nil, err
	}

	fromWallet, err := ps.wallet.GetWalletByID(ctx, transaction.FromWalletID)
	if err != nil {
		return nil, err
	}

	fromCurrency, err := money.ParseCurrency(fromWallet.Currency)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("parse from currency: %w", err))
	}

	fromAmount := int64(float64(transaction.Amount) / exchangeRate)

	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("begin transaction: %w", err))
	}
	defer func() { _ = tx.Rollback() }()

	var queries gen.Querier
	if q, ok := ps.queries.(*gen.Queries); ok {
		queries = q.WithTx(tx)
	} else {
		queries = ps.queries
	}

	lockedFromWallet, err := ps.wallet.LockWalletForUpdate(ctx, tx, fromWallet.ID)
	if err != nil {
		return nil, err
	}

	lockedToWallet, err := ps.wallet.LockWalletForUpdate(ctx, tx, toWallet.ID)
	if err != nil {
		return nil, err
	}

	if lockedFromWallet.Balance < fromAmount {
		return nil, utils.BadRequestErr("insufficient funds")
	}

	if err := ps.ledger.CreateDebitEntry(ctx, tx, lockedFromWallet.ID, transaction.ID, -fromAmount, fromCurrency); err != nil {
		return nil, err
	}

	if err := ps.ledger.CreateCreditEntry(ctx, tx, lockedToWallet.ID, transaction.ID, transaction.Amount, toCurrency); err != nil {
		return nil, err
	}

	if err := queries.UpdateTransactionStatus(ctx, gen.UpdateTransactionStatusParams{
		Status: string(models.TransactionStatusCompleted),
		ID:     transaction.ID,
	}); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("update transaction status: %w", err))
	}

	fromWalletMoney := money.NewMoney(lockedFromWallet.Balance, fromCurrency)
	fromAmountMoney := money.NewMoney(fromAmount, fromCurrency)
	newFromBalanceMoney, err := fromWalletMoney.Subtract(fromAmountMoney)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("calculate new from balance: %w", err))
	}

	toWalletMoney := money.NewMoney(lockedToWallet.Balance, toCurrency)
	toAmountMoney := money.NewMoney(transaction.Amount, toCurrency)
	newToBalanceMoney, err := toWalletMoney.Add(toAmountMoney)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("calculate new to balance: %w", err))
	}

	if err := queries.UpdateWalletBalance(ctx, gen.UpdateWalletBalanceParams{
		Balance: newFromBalanceMoney.Amount,
		ID:      lockedFromWallet.ID,
	}); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("update from wallet balance: %w", err))
	}

	if err := queries.UpdateWalletBalance(ctx, gen.UpdateWalletBalanceParams{
		Balance: newToBalanceMoney.Amount,
		ID:      lockedToWallet.ID,
	}); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("update to wallet balance: %w", err))
	}

	_, err = queries.CreateIdempotencyKey(ctx, gen.CreateIdempotencyKeyParams{
		Key:           transaction.IdempotencyKey,
		TransactionID: transaction.ID,
	})
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("create idempotency key: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("commit transaction: %w", err))
	}

	updatedTransaction, err := ps.queries.GetTransactionByID(ctx, transaction.ID)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("get updated transaction: %w", err))
	}

	return mapTransaction(updatedTransaction), nil
}

func (ps *paymentService) GetTransactionHistory(ctx context.Context, userID string, cursor string, limit int32) (*models.TransactionHistoryResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	cursorTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	var cursorID string
	if cursor != "" {
		decodedCursor, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, utils.BadRequestErr("invalid cursor")
		}
		if decodedCursor != nil {
			cursorTime = decodedCursor.CreatedAt
			cursorID = decodedCursor.ID
		}
	}

	transactions, err := ps.queries.ListTransactionsByUser(ctx, gen.ListTransactionsByUserParams{
		UserID:  userID,
		Column2: cursorTime,
		ID:      cursorID,
		Limit:   limit + 1,
	})
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("list transactions: %w", err))
	}

	hasMore := len(transactions) > int(limit)
	if hasMore {
		transactions = transactions[:limit]
	}

	result := make([]*models.Transaction, 0, len(transactions))
	for _, t := range transactions {
		result = append(result, mapTransaction(t))
	}

	var nextCursor *string
	if hasMore && len(result) > 0 {
		lastTx := result[len(result)-1]
		encoded := utils.EncodeCursor(lastTx.CreatedAt, lastTx.ID)
		nextCursor = &encoded
	}

	return &models.TransactionHistoryResponse{
		Transactions: result,
		NextCursor:   nextCursor,
	}, nil
}
