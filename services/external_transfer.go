package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/IfedayoAwe/payment-processing-service/queue"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/lib/pq"
)

type ExternalTransferService interface {
	CreateExternalTransfer(ctx context.Context, userID string, bankAccountID string, fromCurrency money.Currency, toAmount money.Money, exchangeRate float64, idempotencyKey string) (*models.Transaction, error)
}

type externalTransferService struct {
	queries  *gen.Queries
	db       *sql.DB
	wallet   WalletService
	ledger   LedgerService
	queue    queue.Queue
	provider *providers.Processor
}

func (s *Services) ExternalTransfer() ExternalTransferService {
	return &externalTransferService{
		queries:  s.queries,
		db:       s.db,
		wallet:   s.Wallet(),
		ledger:   s.Ledger(),
		queue:    s.queue,
		provider: s.provider,
	}
}

func (ets *externalTransferService) CreateExternalTransfer(ctx context.Context, userID string, bankAccountID string, fromCurrency money.Currency, toAmount money.Money, exchangeRate float64, idempotencyKey string) (*models.Transaction, error) {
	if !toAmount.IsPositive() {
		return nil, utils.BadRequestErr("amount must be positive")
	}

	existing, err := ets.getTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, utils.ServerErr(fmt.Errorf("check idempotency: %w", err))
	}

	bankAccount, err := ets.getBankAccount(ctx, bankAccountID)
	if err != nil {
		return nil, err
	}

	if bankAccount.Currency != toAmount.Currency.String() {
		return nil, utils.BadRequestErr("bank account currency mismatch")
	}

	if bankAccount.UserID != userID {
		return nil, utils.BadRequestErr("bank account does not belong to user")
	}

	fromWallet, err := ets.wallet.GetWalletByUserAndCurrency(ctx, userID, fromCurrency)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("sender wallet not found")
		}
		return nil, err
	}

	fromAmount := int64(float64(toAmount.Amount) / exchangeRate)
	fromBalance, err := ets.ledger.GetWalletBalance(ctx, nil, fromWallet.ID, fromCurrency)
	if err != nil {
		return nil, err
	}

	if fromBalance < fromAmount {
		return nil, utils.BadRequestErr("insufficient funds")
	}

	exchangeRateStr := strconv.FormatFloat(exchangeRate, 'f', 8, 64)
	transaction, err := ets.queries.CreateTransaction(ctx, gen.CreateTransactionParams{
		IdempotencyKey: idempotencyKey,
		FromWalletID:   fromWallet.ID,
		ToWalletID:     sql.NullString{Valid: false},
		Type:           string(models.TransactionTypeExternal),
		Amount:         toAmount.Amount,
		Currency:       toAmount.Currency.String(),
		Status:         string(models.TransactionStatusInitiated),
		ExchangeRate:   sql.NullString{String: exchangeRateStr, Valid: true},
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			existing, err := ets.getTransactionByIdempotencyKey(ctx, idempotencyKey)
			if err != nil {
				return nil, utils.DuplicateKeyErr("transaction with this idempotency key already exists")
			}
			return existing, nil
		}
		return nil, utils.ServerErr(fmt.Errorf("create transaction: %w", err))
	}

	err = ets.queries.UpdateTransactionWithProvider(ctx, gen.UpdateTransactionWithProviderParams{
		ProviderName:      sql.NullString{Valid: false},
		ProviderReference: sql.NullString{String: bankAccountID, Valid: true},
		Status:            string(models.TransactionStatusInitiated),
		ID:                transaction.ID,
	})
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("store bank account ID: %w", err))
	}

	return &models.Transaction{
		ID:             transaction.ID,
		IdempotencyKey: transaction.IdempotencyKey,
		FromWalletID:   &fromWallet.ID,
		Type:           models.TransactionTypeExternal,
		Amount:         transaction.Amount,
		Currency:       transaction.Currency,
		Status:         models.TransactionStatusInitiated,
		CreatedAt:      transaction.CreatedAt,
		UpdatedAt:      transaction.UpdatedAt,
	}, nil
}

func (ets *externalTransferService) confirmExternalTransfer(ctx context.Context, transaction gen.Transaction) (*models.Transaction, error) {
	if !transaction.ExchangeRate.Valid {
		return nil, utils.ServerErr(fmt.Errorf("exchange rate not found in transaction"))
	}

	exchangeRate, err := strconv.ParseFloat(transaction.ExchangeRate.String, 64)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("invalid exchange rate: %w", err))
	}

	fromWallet, err := ets.wallet.GetWalletByID(ctx, transaction.FromWalletID)
	if err != nil {
		return nil, err
	}

	fromCurrency, err := money.ParseCurrency(fromWallet.Currency)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("parse from currency: %w", err))
	}

	fromAmount := int64(float64(transaction.Amount) / exchangeRate)

	tx, err := ets.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("begin transaction: %w", err))
	}
	defer func() { _ = tx.Rollback() }()

	queries := ets.queries.WithTx(tx)

	lockedWallet, err := ets.wallet.LockWalletForUpdate(ctx, tx, fromWallet.ID)
	if err != nil {
		return nil, err
	}

	balance, err := ets.ledger.GetWalletBalance(ctx, tx, lockedWallet.ID, fromCurrency)
	if err != nil {
		return nil, err
	}

	if balance < fromAmount {
		return nil, utils.BadRequestErr("insufficient funds")
	}

	if err := ets.ledger.CreateDebitEntry(ctx, tx, lockedWallet.ID, transaction.ID, -fromAmount, fromCurrency); err != nil {
		return nil, err
	}

	newBalance, err := ets.ledger.GetWalletBalance(ctx, tx, lockedWallet.ID, fromCurrency)
	if err != nil {
		return nil, err
	}

	if err := queries.UpdateWalletBalance(ctx, gen.UpdateWalletBalanceParams{
		Balance: newBalance,
		ID:      lockedWallet.ID,
	}); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("update wallet balance: %w", err))
	}

	if err := queries.UpdateTransactionStatus(ctx, gen.UpdateTransactionStatusParams{
		Status: string(models.TransactionStatusPending),
		ID:     transaction.ID,
	}); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("update transaction status: %w", err))
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

	if !transaction.ProviderReference.Valid {
		return nil, utils.ServerErr(fmt.Errorf("bank account ID not found in transaction"))
	}

	bankAccountID := transaction.ProviderReference.String

	payload := queue.PayoutJobPayload{
		TransactionID: transaction.ID,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		BankAccountID: bankAccountID,
	}

	if err := ets.queue.Enqueue(ctx, queue.JobTypePayout, payload); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("enqueue payout job: %w", err))
	}

	updatedTransaction, err := ets.queries.GetTransactionByID(ctx, transaction.ID)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("get updated transaction: %w", err))
	}

	return mapTransaction(updatedTransaction), nil
}

func (ets *externalTransferService) getTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Transaction, error) {
	transaction, err := ets.queries.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, utils.ServerErr(fmt.Errorf("get transaction: %w", err))
	}

	return mapTransaction(transaction), nil
}

func (ets *externalTransferService) getBankAccount(ctx context.Context, bankAccountID string) (*models.BankAccount, error) {
	bankAccount, err := ets.queries.GetBankAccountByID(ctx, bankAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("bank account not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("get bank account: %w", err))
	}

	var accountName *string
	if bankAccount.AccountName.Valid {
		accountName = &bankAccount.AccountName.String
	}

	return &models.BankAccount{
		ID:            bankAccount.ID,
		UserID:        bankAccount.UserID,
		BankName:      bankAccount.BankName,
		BankCode:      bankAccount.BankCode,
		AccountNumber: bankAccount.AccountNumber,
		AccountName:   accountName,
		Currency:      bankAccount.Currency,
		Provider:      bankAccount.Provider,
		CreatedAt:     bankAccount.CreatedAt,
		UpdatedAt:     bankAccount.UpdatedAt,
	}, nil
}
