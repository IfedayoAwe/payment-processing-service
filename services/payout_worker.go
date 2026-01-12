package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/IfedayoAwe/payment-processing-service/queue"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type PayoutWorker interface {
	ProcessPayoutJob(ctx context.Context, job *queue.Job) error
	StartWorker(ctx context.Context) error
}

type payoutWorker struct {
	queries  *gen.Queries
	provider *providers.Processor
	queue    queue.Queue
}

func newPayoutWorker(queries *gen.Queries, provider *providers.Processor, queue queue.Queue) PayoutWorker {
	return &payoutWorker{
		queries:  queries,
		provider: provider,
		queue:    queue,
	}
}

func (pw *payoutWorker) ProcessPayoutJob(ctx context.Context, job *queue.Job) error {
	var payload queue.PayoutJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payout job payload: %w", err)
	}

	ctx = utils.WithTraceID(ctx, payload.TraceID)

	utils.Logger.Info().
		Str("trace_id", payload.TraceID).
		Str("transaction_id", payload.TransactionID).
		Str("account_number", payload.AccountNumber).
		Str("bank_code", payload.BankCode).
		Int64("amount", payload.Amount).
		Str("currency", payload.Currency).
		Msg("processing payout job")

	transaction, err := pw.queries.GetTransactionByID(ctx, payload.TransactionID)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}

	if transaction.Status != string(models.TransactionStatusPending) {
		return nil
	}

	currency, err := money.ParseCurrency(payload.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	providerRef := pw.generateProviderReference(payload.TransactionID)

	err = pw.queries.UpdateTransactionWithProvider(ctx, gen.UpdateTransactionWithProviderParams{
		ProviderName:      sql.NullString{Valid: false},
		ProviderReference: sql.NullString{String: string(providerRef), Valid: true},
		Status:            transaction.Status,
		ID:                payload.TransactionID,
	})
	if err != nil {
		return fmt.Errorf("save provider reference: %w", err)
	}

	payoutReq := providers.PayoutRequest{
		Amount: money.NewMoney(payload.Amount, currency),
		Destination: providers.BankAccount{
			BankName:      "",
			BankCode:      payload.BankCode,
			AccountNumber: payload.AccountNumber,
			AccountName:   "",
			Currency:      currency,
		},
		Metadata: map[string]string{
			"transaction_id": payload.TransactionID,
			"trace_id":       payload.TraceID,
		},
		ProviderRef: &providerRef,
	}

	payoutResp, err := pw.provider.SendPayout(ctx, payoutReq)
	if err != nil {
		failErr := pw.failTransaction(ctx, payload.TransactionID, fmt.Sprintf("provider payout failed: %v", err))
		if failErr != nil {
			return fmt.Errorf("provider payout failed: %w (also failed to mark transaction failed: %v)", err, failErr)
		}
		return err
	}

	if payoutResp.ProviderRef != "" {
		providerRef = payoutResp.ProviderRef
	}

	err = pw.queries.UpdateTransactionWithProvider(ctx, gen.UpdateTransactionWithProviderParams{
		ProviderName:      sql.NullString{String: payoutResp.ProviderName, Valid: true},
		ProviderReference: sql.NullString{String: string(providerRef), Valid: true},
		Status:            string(models.TransactionStatusCompleted),
		ID:                payload.TransactionID,
	})
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}

	webhookPayload, err := json.Marshal(map[string]interface{}{
		"transaction_id":     payload.TransactionID,
		"provider_reference": string(providerRef),
		"status":             "completed",
		"amount":             payload.Amount,
		"currency":           payload.Currency,
	})
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	webhookJobPayload := queue.WebhookJobPayload{
		ProviderName:      payoutResp.ProviderName,
		EventType:         "payout.completed",
		ProviderReference: string(providerRef),
		TransactionID:     &payload.TransactionID,
		Payload:           webhookPayload,
	}

	if err := pw.queue.Enqueue(ctx, queue.JobTypeWebhook, webhookJobPayload); err != nil {
		return fmt.Errorf("enqueue webhook job: %w", err)
	}

	return nil
}

func (pw *payoutWorker) StartWorker(ctx context.Context) error {
	utils.Logger.Info().Msg("payout worker started")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info().Msg("payout worker stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := pw.queue.Process(ctx, queue.JobTypePayout, pw.ProcessPayoutJob, 5*time.Second); err != nil {
				traceID := utils.TraceIDFromContext(ctx)
				utils.Logger.Error().Err(err).Str("trace_id", traceID).Str("job_type", string(queue.JobTypePayout)).Msg("error processing payout job")
			}
		}
	}
}

func (pw *payoutWorker) failTransaction(ctx context.Context, transactionID string, reason string) error {
	err := pw.queries.UpdateTransactionFailure(ctx, gen.UpdateTransactionFailureParams{
		Status:        string(models.TransactionStatusFailed),
		FailureReason: sql.NullString{String: reason, Valid: true},
		ID:            transactionID,
	})
	return err
}

func (pw *payoutWorker) generateProviderReference(transactionID string) providers.ProviderReference {
	prefix := transactionID
	if len(transactionID) > 8 {
		prefix = transactionID[:8]
	}
	return providers.ProviderReference(fmt.Sprintf("TXN-%s-%d", prefix, time.Now().UnixNano()))
}
