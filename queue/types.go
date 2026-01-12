package queue

import (
	"context"
	"encoding/json"
	"time"
)

type JobType string

const (
	JobTypePayout  JobType = "payout"
	JobTypeWebhook JobType = "webhook"
)

const (
	maxRetries       = 3
	retryDelay       = 5 * time.Second
	reconnectDelay   = 5 * time.Second
	maxReconnectTime = 30 * time.Second
)

type Job struct {
	ID        string
	Type      JobType
	Payload   json.RawMessage
	Attempts  int
	CreatedAt time.Time
	ack       func()
	nack      func(requeue bool)
}

type JobHandler func(ctx context.Context, job *Job) error

type Queue interface {
	Enqueue(ctx context.Context, jobType JobType, payload interface{}) error
	Dequeue(ctx context.Context, jobType JobType, timeout time.Duration) (*Job, error)
	Process(ctx context.Context, jobType JobType, handler JobHandler, timeout time.Duration) error
	Retry(ctx context.Context, job *Job) error
	Close() error
}

type PayoutJobPayload struct {
	TransactionID string `json:"transaction_id"`
	TraceID       string `json:"trace_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
}

type WebhookJobPayload struct {
	ProviderName      string  `json:"provider_name"`
	EventType         string  `json:"event_type"`
	ProviderReference string  `json:"provider_reference"`
	TransactionID     *string `json:"transaction_id,omitempty"`
	Payload           []byte  `json:"payload"`
}
