package services

import (
	"context"
	"database/sql"

	"github.com/IfedayoAwe/payment-processing-service/config"
	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/IfedayoAwe/payment-processing-service/queue"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/go-redis/redis/v8"
)

type Services struct {
	Payment          PaymentService
	Wallet           WalletService
	Ledger           LedgerService
	ExternalTransfer ExternalTransferService
	NameEnquiry      NameEnquiryService
	Webhook          WebhookService
	PayoutWorker     PayoutWorker
	WebhookWorker    WebhookWorker
	Queries          *gen.Queries
	Queue            queue.Queue
}

func NewServices(db *sql.DB, queries *gen.Queries, cfg *config.Config, redisClient *redis.Client) *Services {
	processor := providers.SetupProcessor()

	var q queue.Queue
	if redisClient != nil {
		q = queue.NewRedisQueue(redisClient)
	}

	ledgerService := newLedgerService(queries)
	walletService := newWalletService(queries, db)
	externalTransferService := newExternalTransferService(queries, db, walletService, ledgerService, q, processor)
	paymentService := newPaymentService(queries, db, walletService, ledgerService, externalTransferService, processor)
	nameEnquiryService := newNameEnquiryService(queries, processor)
	webhookService := newWebhookService(queries)
	payoutWorker := newPayoutWorker(queries, processor, q)
	webhookWorker := newWebhookWorker(queries, q)

	return &Services{
		Payment:          paymentService,
		Wallet:           walletService,
		Ledger:           ledgerService,
		ExternalTransfer: externalTransferService,
		NameEnquiry:      nameEnquiryService,
		Webhook:          webhookService,
		PayoutWorker:     payoutWorker,
		WebhookWorker:    webhookWorker,
		Queries:          queries,
		Queue:            q,
	}
}

func (s *Services) StartWorkers(ctx context.Context) {
	workers := []struct {
		name   string
		worker func(context.Context) error
	}{
		{"payout", s.PayoutWorker.StartWorker},
		{"webhook", s.WebhookWorker.StartWorker},
	}

	for _, w := range workers {
		go func(name string, start func(context.Context) error) {
			traceID := utils.TraceIDFromContext(ctx)
			utils.Logger.Info().Str("trace_id", traceID).Str("worker", name).Msg("starting worker")
			if err := start(ctx); err != nil && err != context.Canceled {
				utils.Logger.Error().Err(err).Str("trace_id", traceID).Str("worker", name).Msg("worker error")
			}
		}(w.name, w.worker)
	}
}
