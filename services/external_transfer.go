package services

import (
	"context"
	"database/sql"
	"encoding/json"
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
	CreateExternalTransfer(ctx context.Context, userID string, toAccountNumber string, toBankCode string, fromCurrency money.Currency, toAmount money.Money, exchangeRate float64, idempotencyKey string) (*models.Transaction, error)
	confirmExternalTransfer(ctx context.Context, transaction gen.Transaction) (*models.Transaction, error)
}

type externalTransferService struct {
	queries  gen.Querier
	db       *sql.DB
	wallet   WalletService
	ledger   LedgerService
	queue    queue.Queue
	provider *providers.Processor
}

func newExternalTransferService(queries gen.Querier, db *sql.DB, wallet WalletService, ledger LedgerService, queue queue.Queue, provider *providers.Processor) ExternalTransferService {
	return &externalTransferService{
		queries:  queries,
		db:       db,
		wallet:   wallet,
		ledger:   ledger,
		queue:    queue,
		provider: provider,
	}
}

func (ets *externalTransferService) CreateExternalTransfer(ctx context.Context, userID string, toAccountNumber string, toBankCode string, fromCurrency money.Currency, toAmount money.Money, exchangeRate float64, idempotencyKey string) (*models.Transaction, error) {
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

	fromWallet, err := ets.wallet.GetWalletByUserAndCurrency(ctx, userID, fromCurrency)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("sender wallet not found")
		}
		return nil, err
	}

	fromAmount := int64(float64(toAmount.Amount) / exchangeRate)

	if fromWallet.Balance < fromAmount {
		return nil, utils.BadRequestErr("insufficient funds")
	}

	exchangeRateStr := strconv.FormatFloat(exchangeRate, 'f', 8, 64)
	traceID := utils.TraceIDFromContext(ctx)
	transaction, err := ets.queries.CreateTransaction(ctx, gen.CreateTransactionParams{
		IdempotencyKey: idempotencyKey,
		TraceID:        sql.NullString{String: traceID, Valid: traceID != ""},
		FromWalletID:   sql.NullString{String: fromWallet.ID, Valid: true},
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

	recipientDetails := map[string]string{
		"account_number": toAccountNumber,
		"bank_code":      toBankCode,
	}
	recipientJSON, err := json.Marshal(recipientDetails)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("marshal recipient details: %w", err))
	}

	err = ets.queries.UpdateTransactionWithProvider(ctx, gen.UpdateTransactionWithProviderParams{
		ProviderName:      sql.NullString{Valid: false},
		ProviderReference: sql.NullString{String: string(recipientJSON), Valid: true},
		Status:            string(models.TransactionStatusInitiated),
		ID:                transaction.ID,
	})
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("store recipient details: %w", err))
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

	if !transaction.FromWalletID.Valid {
		return nil, utils.ServerErr(fmt.Errorf("from wallet ID not found in transaction"))
	}

	fromWallet, err := ets.wallet.GetWalletByID(ctx, transaction.FromWalletID.String)
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

	var queries gen.Querier
	if q, ok := ets.queries.(*gen.Queries); ok {
		queries = q.WithTx(tx)
	} else {
		queries = ets.queries
	}

	lockedWallet, err := ets.wallet.LockWalletForUpdate(ctx, tx, fromWallet.ID)
	if err != nil {
		return nil, err
	}

	if lockedWallet.Balance < fromAmount {
		return nil, utils.BadRequestErr("insufficient funds")
	}

	if err := ets.ledger.CreateDebitEntry(ctx, tx, lockedWallet.ID, transaction.ID, -fromAmount, fromCurrency); err != nil {
		return nil, err
	}

	toCurrency, err := money.ParseCurrency(transaction.Currency)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("parse transaction currency: %w", err))
	}

	if err := ets.ledger.CreateExternalSystemCreditEntry(ctx, tx, transaction.ID, transaction.Amount, toCurrency); err != nil {
		return nil, err
	}

	walletMoney := money.NewMoney(lockedWallet.Balance, fromCurrency)
	fromAmountMoney := money.NewMoney(fromAmount, fromCurrency)
	newBalanceMoney, err := walletMoney.Subtract(fromAmountMoney)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("calculate new wallet balance: %w", err))
	}

	if err := queries.UpdateWalletBalance(ctx, gen.UpdateWalletBalanceParams{
		Balance: newBalanceMoney.Amount,
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

	if !transaction.ProviderReference.Valid {
		return nil, utils.ServerErr(fmt.Errorf("recipient details not found in transaction"))
	}

	var recipientDetails map[string]string
	if err := json.Unmarshal([]byte(transaction.ProviderReference.String), &recipientDetails); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("unmarshal recipient details: %w", err))
	}

	accountNumber, ok := recipientDetails["account_number"]
	if !ok {
		return nil, utils.ServerErr(fmt.Errorf("account_number not found in recipient details"))
	}

	bankCode, ok := recipientDetails["bank_code"]
	if !ok {
		return nil, utils.ServerErr(fmt.Errorf("bank_code not found in recipient details"))
	}

	traceID := utils.TraceIDFromContext(ctx)
	if transaction.TraceID.Valid {
		traceID = transaction.TraceID.String
	}

	payload := queue.PayoutJobPayload{
		TransactionID: transaction.ID,
		TraceID:       traceID,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		AccountNumber: accountNumber,
		BankCode:      bankCode,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("marshal payout payload: %w", err))
	}

	_, err = queries.CreateOutboxEntry(ctx, gen.CreateOutboxEntryParams{
		JobType: string(queue.JobTypePayout),
		Payload: payloadJSON,
	})
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("create outbox entry: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.ServerErr(fmt.Errorf("commit transaction: %w", err))
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
