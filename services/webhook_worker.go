package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/queue"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type WebhookWorker interface {
	ProcessWebhookJob(ctx context.Context, job *queue.Job) error
	StartWorker(ctx context.Context) error
}

type webhookWorker struct {
	queries *gen.Queries
	queue   queue.Queue
}

func (s *Services) WebhookWorker() WebhookWorker {
	return &webhookWorker{
		queries: s.queries,
		queue:   s.queue,
	}
}

func (ww *webhookWorker) ProcessWebhookJob(ctx context.Context, job *queue.Job) error {
	var payload queue.WebhookJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal webhook job payload: %w", err)
	}

	var txID sql.NullString
	if payload.TransactionID != nil {
		txID = sql.NullString{String: *payload.TransactionID, Valid: true}
	}

	_, err := ww.queries.CreateWebhookEvent(ctx, gen.CreateWebhookEventParams{
		ProviderName:      payload.ProviderName,
		EventType:         payload.EventType,
		ProviderReference: payload.ProviderReference,
		TransactionID:     txID,
		Payload:           json.RawMessage(payload.Payload),
	})
	if err != nil {
		return fmt.Errorf("create webhook event: %w", err)
	}

	return nil
}

func (ww *webhookWorker) StartWorker(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := ww.queue.Process(ctx, queue.JobTypeWebhook, ww.ProcessWebhookJob, 5*time.Second); err != nil {
				utils.Logger.Error().Err(err).Str("job_type", string(queue.JobTypeWebhook)).Msg("error processing webhook job")
			}
		}
	}
}
