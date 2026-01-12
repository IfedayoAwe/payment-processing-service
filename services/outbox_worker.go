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

type OutboxWorker interface {
	StartWorker(ctx context.Context) error
}

type outboxWorker struct {
	queries *gen.Queries
	db      *sql.DB
	queue   queue.Queue
}

func newOutboxWorker(queries *gen.Queries, db *sql.DB, queue queue.Queue) OutboxWorker {
	return &outboxWorker{
		queries: queries,
		db:      db,
		queue:   queue,
	}
}

func (ow *outboxWorker) StartWorker(ctx context.Context) error {
	utils.Logger.Info().Msg("outbox worker started")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info().Msg("outbox worker stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := ow.processOutboxBatch(ctx); err != nil {
				traceID := utils.TraceIDFromContext(ctx)
				utils.Logger.Error().Err(err).Str("trace_id", traceID).Msg("error processing outbox batch")
			}
		}
	}
}

func (ow *outboxWorker) processOutboxBatch(ctx context.Context) error {
	batchSize := int32(10)

	tx, err := ow.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	queries := ow.queries.WithTx(tx)

	entries, err := queries.GetUnprocessedOutboxEntries(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("get unprocessed outbox entries: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	traceID := utils.TraceIDFromContext(ctx)
	utils.Logger.Info().
		Str("trace_id", traceID).
		Int("batch_size", len(entries)).
		Msg("processing outbox batch")

	var processedEntries []string
	for _, entry := range entries {
		if entry.RetryCount >= 5 {
			utils.Logger.Warn().
				Str("trace_id", traceID).
				Str("outbox_id", entry.ID).
				Str("job_type", entry.JobType).
				Int32("retry_count", entry.RetryCount).
				Msg("outbox entry exceeded max retries, skipping")
			continue
		}

		if err := ow.processOutboxEntry(ctx, entry); err != nil {
			utils.Logger.Error().
				Err(err).
				Str("trace_id", traceID).
				Str("outbox_id", entry.ID).
				Str("job_type", entry.JobType).
				Int32("retry_count", entry.RetryCount).
				Msg("error processing outbox entry")

			if entry.RetryCount < 5 {
				if retryErr := queries.IncrementOutboxRetryCount(ctx, entry.ID); retryErr != nil {
					utils.Logger.Error().
						Err(retryErr).
						Str("trace_id", traceID).
						Str("outbox_id", entry.ID).
						Msg("error incrementing retry count")
				}
			}
			continue
		}

		processedEntries = append(processedEntries, entry.ID)
	}

	for _, entryID := range processedEntries {
		if err := queries.MarkOutboxEntryProcessed(ctx, entryID); err != nil {
			utils.Logger.Error().
				Err(err).
				Str("trace_id", traceID).
				Str("outbox_id", entryID).
				Msg("error marking outbox entry as processed")
			return fmt.Errorf("mark outbox entry processed: %w", err)
		}

		utils.Logger.Info().
			Str("trace_id", traceID).
			Str("outbox_id", entryID).
			Msg("outbox entry processed successfully")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (ow *outboxWorker) processOutboxEntry(ctx context.Context, entry gen.Outbox) error {
	jobType := queue.JobType(entry.JobType)

	switch jobType {
	case queue.JobTypePayout:
		var payload queue.PayoutJobPayload
		if err := json.Unmarshal(entry.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal payout payload: %w", err)
		}
		return ow.queue.Enqueue(ctx, queue.JobTypePayout, payload)

	case queue.JobTypeWebhook:
		var payload queue.WebhookJobPayload
		if err := json.Unmarshal(entry.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal webhook payload: %w", err)
		}
		return ow.queue.Enqueue(ctx, queue.JobTypeWebhook, payload)

	default:
		return fmt.Errorf("unknown job type: %s", entry.JobType)
	}
}
