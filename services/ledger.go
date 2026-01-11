package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type LedgerService interface {
	CreateDebitEntry(ctx context.Context, tx *sql.Tx, walletID string, transactionID string, amount int64, currency money.Currency) error
	CreateCreditEntry(ctx context.Context, tx *sql.Tx, walletID string, transactionID string, amount int64, currency money.Currency) error
	GetWalletBalance(ctx context.Context, tx *sql.Tx, walletID string, currency money.Currency) (int64, error)
}

type ledgerService struct {
	queries gen.Querier
}

func (s *Services) Ledger() LedgerService {
	return &ledgerService{
		queries: s.queries,
	}
}

func (ls *ledgerService) CreateDebitEntry(ctx context.Context, tx *sql.Tx, walletID string, transactionID string, amount int64, currency money.Currency) error {
	if amount >= 0 {
		return utils.BadRequestErr("debit amount must be negative")
	}

	var queries gen.Querier
	if q, ok := ls.queries.(*gen.Queries); ok {
		queries = q.WithTx(tx)
	} else {
		queries = ls.queries
	}
	_, err := queries.CreateLedgerEntry(ctx, gen.CreateLedgerEntryParams{
		WalletID:      walletID,
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      currency.String(),
	})
	if err != nil {
		return utils.ServerErr(fmt.Errorf("create debit entry: %w", err))
	}

	return nil
}

func (ls *ledgerService) CreateCreditEntry(ctx context.Context, tx *sql.Tx, walletID string, transactionID string, amount int64, currency money.Currency) error {
	if amount <= 0 {
		return utils.BadRequestErr("credit amount must be positive")
	}

	var queries gen.Querier
	if q, ok := ls.queries.(*gen.Queries); ok {
		queries = q.WithTx(tx)
	} else {
		queries = ls.queries
	}
	_, err := queries.CreateLedgerEntry(ctx, gen.CreateLedgerEntryParams{
		WalletID:      walletID,
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      currency.String(),
	})
	if err != nil {
		return utils.ServerErr(fmt.Errorf("create credit entry: %w", err))
	}

	return nil
}

func (ls *ledgerService) GetWalletBalance(ctx context.Context, tx *sql.Tx, walletID string, currency money.Currency) (int64, error) {
	var queries gen.Querier
	if tx != nil {
		if q, ok := ls.queries.(*gen.Queries); ok {
			queries = q.WithTx(tx)
		} else {
			queries = ls.queries
		}
	} else {
		queries = ls.queries
	}

	balance, err := queries.GetWalletBalance(ctx, gen.GetWalletBalanceParams{
		WalletID: walletID,
		Currency: currency.String(),
	})
	if err != nil {
		return 0, utils.ServerErr(fmt.Errorf("get wallet balance: %w", err))
	}

	return balance, nil
}
