package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockQueue struct {
	mock.Mock
}

func (m *mockQueue) Enqueue(ctx context.Context, jobType queue.JobType, payload interface{}) error {
	args := m.Called(ctx, jobType, payload)
	return args.Error(0)
}

func (m *mockQueue) Dequeue(ctx context.Context, jobType queue.JobType, timeout time.Duration) (*queue.Job, error) {
	args := m.Called(ctx, jobType, timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*queue.Job), args.Error(1)
}

func (m *mockQueue) Process(ctx context.Context, jobType queue.JobType, handler queue.JobHandler, timeout time.Duration) error {
	args := m.Called(ctx, jobType, handler, timeout)
	return args.Error(0)
}

func (m *mockQueue) Retry(ctx context.Context, job *queue.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func TestOutboxWorker_ProcessOutboxEntry_Payout(t *testing.T) {
	mockQueries := &gen.Queries{}
	mockQueue := new(mockQueue)
	mockDB := &sql.DB{}

	worker := &outboxWorker{
		queries: mockQueries,
		db:      mockDB,
		queue:   mockQueue,
	}

	payload := queue.PayoutJobPayload{
		TransactionID: "tx_123",
		TraceID:       "trace_123",
		Amount:        10000,
		Currency:      "USD",
		AccountNumber: "1234567890",
		BankCode:      "044",
	}
	payloadJSON, _ := json.Marshal(payload)

	entry := gen.Outbox{
		ID:         "outbox_1",
		JobType:    "payout",
		Payload:    payloadJSON,
		Processed:  false,
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}

	mockQueue.On("Enqueue", mock.Anything, queue.JobTypePayout, payload).Return(nil)

	err := worker.processOutboxEntry(context.Background(), entry)
	assert.NoError(t, err)
	mockQueue.AssertExpectations(t)
}

func TestOutboxWorker_ProcessOutboxEntry_Webhook(t *testing.T) {
	mockQueries := &gen.Queries{}
	mockQueue := new(mockQueue)
	mockDB := &sql.DB{}

	worker := &outboxWorker{
		queries: mockQueries,
		db:      mockDB,
		queue:   mockQueue,
	}

	payload := queue.WebhookJobPayload{
		ProviderName:      "currencycloud",
		EventType:         "payout.completed",
		ProviderReference: "ref_123",
		TransactionID:     stringPtr("tx_123"),
		Payload:           []byte(`{"status":"completed"}`),
	}
	payloadJSON, _ := json.Marshal(payload)

	entry := gen.Outbox{
		ID:         "outbox_1",
		JobType:    "webhook",
		Payload:    payloadJSON,
		Processed:  false,
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}

	mockQueue.On("Enqueue", mock.Anything, queue.JobTypeWebhook, payload).Return(nil)

	err := worker.processOutboxEntry(context.Background(), entry)
	assert.NoError(t, err)
	mockQueue.AssertExpectations(t)
}

func TestOutboxWorker_ProcessOutboxEntry_UnknownJobType(t *testing.T) {
	mockQueries := &gen.Queries{}
	mockQueue := new(mockQueue)
	mockDB := &sql.DB{}

	worker := &outboxWorker{
		queries: mockQueries,
		db:      mockDB,
		queue:   mockQueue,
	}

	entry := gen.Outbox{
		ID:         "outbox_1",
		JobType:    "unknown",
		Payload:    []byte(`{}`),
		Processed:  false,
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}

	err := worker.processOutboxEntry(context.Background(), entry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown job type")
}

func TestOutboxWorker_ProcessOutboxEntry_InvalidPayload(t *testing.T) {
	mockQueries := &gen.Queries{}
	mockQueue := new(mockQueue)
	mockDB := &sql.DB{}

	worker := &outboxWorker{
		queries: mockQueries,
		db:      mockDB,
		queue:   mockQueue,
	}

	entry := gen.Outbox{
		ID:         "outbox_1",
		JobType:    "payout",
		Payload:    []byte(`invalid json`),
		Processed:  false,
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}

	err := worker.processOutboxEntry(context.Background(), entry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestOutboxWorker_ProcessOutboxEntry_EnqueueFailure(t *testing.T) {
	mockQueries := &gen.Queries{}
	mockQueue := new(mockQueue)
	mockDB := &sql.DB{}

	worker := &outboxWorker{
		queries: mockQueries,
		db:      mockDB,
		queue:   mockQueue,
	}

	payload := queue.PayoutJobPayload{
		TransactionID: "tx_123",
		TraceID:       "trace_123",
		Amount:        10000,
		Currency:      "USD",
		AccountNumber: "1234567890",
		BankCode:      "044",
	}
	payloadJSON, _ := json.Marshal(payload)

	entry := gen.Outbox{
		ID:         "outbox_1",
		JobType:    "payout",
		Payload:    payloadJSON,
		Processed:  false,
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}

	mockQueue.On("Enqueue", mock.Anything, queue.JobTypePayout, payload).Return(assert.AnError)

	err := worker.processOutboxEntry(context.Background(), entry)
	assert.Error(t, err)
	mockQueue.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
