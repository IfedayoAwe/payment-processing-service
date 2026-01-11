package models

import "time"

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

type UserType string

const (
	UserTypeInternal UserType = "internal"
	UserTypeExternal UserType = "external"
)

type User struct {
	ID        string
	Name      *string
	Type      UserType
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
	AccountName   *string   `json:"account_name,omitempty"`
	Provider      *string   `json:"provider,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TransactionHistoryResponse struct {
	Transactions []*Transaction `json:"transactions"`
	NextCursor   *string        `json:"next_cursor,omitempty"`
}
