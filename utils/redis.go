package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/config"
	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Client    *redis.Client
	Ctx       context.Context
	CancelCtx context.CancelFunc
}

func (r *Redis) GetClient() *redis.Client {
	return r.Client
}

func NewRedis(cfg *config.Config) (*Redis, error) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	redisCtx, cancelRedis := context.WithCancel(context.Background())

	pingCtx, cancelPing := context.WithTimeout(redisCtx, 5*time.Second)
	defer cancelPing()

	if err := client.Ping(pingCtx).Err(); err != nil {
		cancelRedis()
		return nil, err
	}

	return &Redis{
		Client:    client,
		Ctx:       redisCtx,
		CancelCtx: cancelRedis,
	}, nil
}

func (c *Redis) Close() error {
	c.CancelCtx()
	return closeRedisClient(c.Client)
}

func closeRedisClient(client *redis.Client) error {
	if err := client.Close(); err != nil {
		Logger.Error().Err(err).Msg("error closing Redis client")
		return fmt.Errorf("failed to close Redis client: %w", err)
	}
	return nil
}
