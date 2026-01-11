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
	db       *sql.DB
	queries  *gen.Queries
	config   *config.Config
	provider *providers.Processor
	queue    queue.Queue
}

func NewServices(db *sql.DB, queries *gen.Queries, cfg *config.Config, redisClient *redis.Client) *Services {
	processor := providers.SetupProcessor()

	var q queue.Queue
	if redisClient != nil {
		q = queue.NewRedisQueue(redisClient)
	}

	return &Services{
		db:       db,
		queries:  queries,
		config:   cfg,
		provider: processor,
		queue:    q,
	}
}

func (s *Services) Queue() queue.Queue {
	return s.queue
}

func (s *Services) StartWorkers(ctx context.Context) {
	workers := []struct {
		name   string
		worker func(context.Context) error
	}{
		{"payout", s.PayoutWorker().StartWorker},
		{"webhook", s.WebhookWorker().StartWorker},
	}

	for _, w := range workers {
		go func(name string, start func(context.Context) error) {
			if err := start(ctx); err != nil && err != context.Canceled {
				utils.Logger.Error().Err(err).Str("worker", name).Msg("worker error")
			}
		}(w.name, w.worker)
	}
}
