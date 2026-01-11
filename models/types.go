package models

import (
	"time"

	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
)

type TransactionType string

const (
	TransactionTypeInternal TransactionType = "internal"
	TransactionTypeExternal TransactionType = "external"
)

type TransactionStatus string

const (
	TransactionStatusInitiated TransactionStatus = "initiated"
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

type User struct {
	ID        string
	Name      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Wallet struct {
	ID            string
	UserID        string
	BankAccountID *string
	Currency      string
	Balance       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Transaction struct {
	ID                string
	IdempotencyKey    string
	FromWalletID      *string
	ToWalletID        *string
	Type              TransactionType
	Amount            int64
	Currency          string
	Status            TransactionStatus
	ProviderName      *string
	ProviderReference *string
	ExchangeRate      *float64
	FailureReason     *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type LedgerEntry struct {
	ID            string
	WalletID      string
	TransactionID string
	Amount        int64
	Currency      string
	CreatedAt     time.Time
}

type IdempotencyKey struct {
	Key           string
	TransactionID string
	CreatedAt     time.Time
}

type BankAccount struct {
	ID            string
	UserID        string
	BankName      string
	BankCode      string
	AccountNumber string
	AccountName   *string
	Currency      string
	Provider      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type WebhookEvent struct {
	ID                string
	ProviderName      string
	EventType         string
	ProviderReference string
	TransactionID     *string
	Payload           []byte
	Processed         bool
	CreatedAt         time.Time
	ProcessedAt       *string
}

type NameEnquiryResponse struct {
	AccountName string `json:"account_name"`
	IsInternal  bool   `json:"is_internal"`
	Currency    string `json:"currency"`
}

type WalletWithBankAccount struct {
	ID            string    `json:"id"`
	Currency      string    `json:"currency"`
	Balance       int64     `json:"balance"`
	AccountNumber *string   `json:"account_number,omitempty"`
	BankName      *string   `json:"bank_name,omitempty"`
	BankCode      *string   `json:"bank_code,omitempty"`
	AccountName   *string   `json:"account_name,omitempty"`
	Provider      *string   `json:"provider,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TransactionHistoryResponse struct {
	Transactions []*Transaction `json:"transactions"`
	NextCursor   *string        `json:"next_cursor,omitempty"`
}

type TestUserData struct {
	UserID  string                   `json:"user_id"`
	Name    *string                  `json:"name"`
	PIN     string                   `json:"pin"`
	Wallets []*WalletWithBankAccount `json:"wallets"`
}

type TestUsersResponse struct {
	Users []*TestUserData `json:"users"`
}

// Response DTOs with amounts in major units (dollars/euros/pounds)
type TransactionResponse struct {
	ID                string    `json:"id"`
	IdempotencyKey    string    `json:"idempotency_key"`
	FromWalletID      *string   `json:"from_wallet_id,omitempty"`
	ToWalletID        *string   `json:"to_wallet_id,omitempty"`
	Type              string    `json:"type"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	ProviderName      *string   `json:"provider_name,omitempty"`
	ProviderReference *string   `json:"provider_reference,omitempty"`
	ExchangeRate      *float64  `json:"exchange_rate,omitempty"`
	FailureReason     *string   `json:"failure_reason,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func TransactionToResponse(tx *Transaction) *TransactionResponse {
	amount := money.ToMajorUnits(tx.Amount)
	return &TransactionResponse{
		ID:                tx.ID,
		IdempotencyKey:    tx.IdempotencyKey,
		FromWalletID:      tx.FromWalletID,
		ToWalletID:        tx.ToWalletID,
		Type:              string(tx.Type),
		Amount:            amount,
		Currency:          tx.Currency,
		Status:            string(tx.Status),
		ProviderName:      tx.ProviderName,
		ProviderReference: tx.ProviderReference,
		ExchangeRate:      tx.ExchangeRate,
		FailureReason:     tx.FailureReason,
		CreatedAt:         tx.CreatedAt,
		UpdatedAt:         tx.UpdatedAt,
	}
}

type WalletWithBankAccountResponse struct {
	ID            string    `json:"id"`
	Currency      string    `json:"currency"`
	Balance       float64   `json:"balance"`
	AccountNumber *string   `json:"account_number,omitempty"`
	BankName      *string   `json:"bank_name,omitempty"`
	BankCode      *string   `json:"bank_code,omitempty"`
	AccountName   *string   `json:"account_name,omitempty"`
	Provider      *string   `json:"provider,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func WalletWithBankAccountToResponse(w *WalletWithBankAccount) *WalletWithBankAccountResponse {
	balance := money.ToMajorUnits(w.Balance)
	return &WalletWithBankAccountResponse{
		ID:            w.ID,
		Currency:      w.Currency,
		Balance:       balance,
		AccountNumber: w.AccountNumber,
		BankName:      w.BankName,
		BankCode:      w.BankCode,
		AccountName:   w.AccountName,
		Provider:      w.Provider,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}

type TransactionHistoryResponseDTO struct {
	Transactions []*TransactionResponse `json:"transactions"`
	NextCursor   *string                `json:"next_cursor,omitempty"`
}

type TestUserDataResponse struct {
	UserID  string                           `json:"user_id"`
	Name    *string                          `json:"name"`
	PIN     string                           `json:"pin"`
	Wallets []*WalletWithBankAccountResponse `json:"wallets"`
}

type TestUsersResponseDTO struct {
	Users []*TestUserDataResponse `json:"users"`
}
