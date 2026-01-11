package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type WalletService interface {
	GetWalletByID(ctx context.Context, walletID string) (*models.Wallet, error)
	GetWalletByUserAndCurrency(ctx context.Context, userID string, currency money.Currency) (*models.Wallet, error)
	GetWalletByBankAccount(ctx context.Context, bankAccountID string) (*models.Wallet, error)
	GetUserWallets(ctx context.Context, userID string) ([]*models.WalletWithBankAccount, error)
	LockWalletForUpdate(ctx context.Context, tx *sql.Tx, walletID string) (*models.Wallet, error)
	LockWalletByUserAndCurrency(ctx context.Context, tx *sql.Tx, userID string, currency money.Currency) (*models.Wallet, error)
}

type walletService struct {
	queries *gen.Queries
	db      *sql.DB
}

func (s *Services) Wallet() WalletService {
	return &walletService{
		queries: s.queries,
		db:      s.db,
	}
}

func (ws *walletService) GetWalletByID(ctx context.Context, walletID string) (*models.Wallet, error) {
	wallet, err := ws.queries.GetWalletByID(ctx, walletID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("wallet not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("get wallet: %w", err))
	}

	return &models.Wallet{
		ID:        wallet.ID,
		UserID:    wallet.UserID,
		Currency:  wallet.Currency,
		Balance:   wallet.Balance,
		CreatedAt: wallet.CreatedAt,
		UpdatedAt: wallet.UpdatedAt,
	}, nil
}

func (ws *walletService) GetWalletByUserAndCurrency(ctx context.Context, userID string, currency money.Currency) (*models.Wallet, error) {
	wallet, err := ws.queries.GetWalletByUserAndCurrency(ctx, gen.GetWalletByUserAndCurrencyParams{
		UserID:   userID,
		Currency: currency.String(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, utils.ServerErr(fmt.Errorf("get wallet: %w", err))
	}

	var bankAccountID *string
	if wallet.BankAccountID.Valid {
		bankAccountID = &wallet.BankAccountID.String
	}

	return &models.Wallet{
		ID:            wallet.ID,
		UserID:        wallet.UserID,
		BankAccountID: bankAccountID,
		Currency:      wallet.Currency,
		Balance:       wallet.Balance,
		CreatedAt:     wallet.CreatedAt,
		UpdatedAt:     wallet.UpdatedAt,
	}, nil
}

func (ws *walletService) GetWalletByBankAccount(ctx context.Context, bankAccountID string) (*models.Wallet, error) {
	wallet, err := ws.queries.GetWalletByBankAccount(ctx, sql.NullString{String: bankAccountID, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("wallet not found for bank account")
		}
		return nil, utils.ServerErr(fmt.Errorf("get wallet: %w", err))
	}

	var bankAccountIDPtr *string
	if wallet.BankAccountID.Valid {
		bankAccountIDPtr = &wallet.BankAccountID.String
	}

	return &models.Wallet{
		ID:            wallet.ID,
		UserID:        wallet.UserID,
		BankAccountID: bankAccountIDPtr,
		Currency:      wallet.Currency,
		Balance:       wallet.Balance,
		CreatedAt:     wallet.CreatedAt,
		UpdatedAt:     wallet.UpdatedAt,
	}, nil
}

func (ws *walletService) GetUserWallets(ctx context.Context, userID string) ([]*models.WalletWithBankAccount, error) {
	rows, err := ws.queries.GetUserWalletsWithBankAccounts(ctx, userID)
	if err != nil {
		return nil, utils.ServerErr(fmt.Errorf("get user wallets: %w", err))
	}

	wallets := make([]*models.WalletWithBankAccount, 0, len(rows))
	for _, row := range rows {
		var accountNumber, bankName, accountName, provider *string

		if row.AccountNumber.Valid {
			accountNumber = &row.AccountNumber.String
		}
		if row.BankName.Valid {
			bankName = &row.BankName.String
		}
		if row.AccountName.Valid {
			accountName = &row.AccountName.String
		}
		if row.Provider.Valid {
			provider = &row.Provider.String
		}

		wallets = append(wallets, &models.WalletWithBankAccount{
			ID:            row.ID,
			Currency:      row.Currency,
			Balance:       row.Balance,
			AccountNumber: accountNumber,
			BankName:      bankName,
			AccountName:   accountName,
			Provider:      provider,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}

	return wallets, nil
}

func (ws *walletService) LockWalletForUpdate(ctx context.Context, tx *sql.Tx, walletID string) (*models.Wallet, error) {
	queries := ws.queries.WithTx(tx)
	wallet, err := queries.GetWalletByIDForUpdate(ctx, walletID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("wallet not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("lock wallet: %w", err))
	}

	var bankAccountID *string
	if wallet.BankAccountID.Valid {
		bankAccountID = &wallet.BankAccountID.String
	}

	return &models.Wallet{
		ID:            wallet.ID,
		UserID:        wallet.UserID,
		BankAccountID: bankAccountID,
		Currency:      wallet.Currency,
		Balance:       wallet.Balance,
		CreatedAt:     wallet.CreatedAt,
		UpdatedAt:     wallet.UpdatedAt,
	}, nil
}

func (ws *walletService) LockWalletByUserAndCurrency(ctx context.Context, tx *sql.Tx, userID string, currency money.Currency) (*models.Wallet, error) {
	queries := ws.queries.WithTx(tx)
	wallet, err := queries.GetWalletByUserAndCurrencyForUpdate(ctx, gen.GetWalletByUserAndCurrencyForUpdateParams{
		UserID:   userID,
		Currency: currency.String(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("wallet not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("lock wallet: %w", err))
	}

	var bankAccountID *string
	if wallet.BankAccountID.Valid {
		bankAccountID = &wallet.BankAccountID.String
	}

	return &models.Wallet{
		ID:            wallet.ID,
		UserID:        wallet.UserID,
		BankAccountID: bankAccountID,
		Currency:      wallet.Currency,
		Balance:       wallet.Balance,
		CreatedAt:     wallet.CreatedAt,
		UpdatedAt:     wallet.UpdatedAt,
	}, nil
}
