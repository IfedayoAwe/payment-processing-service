package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/config"
	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Set(key string, value any, duration ...time.Duration) error
	Get(key string, dest any) error
	Delete(key string) error
	Incr(key string) (int64, error)
	Expire(key string, duration time.Duration) error
	Exists(key string) (bool, error)
	HSet(key string, values map[string]any) error
	HGet(key, field string) (string, error)
	HGetAll(key string) (map[string]string, error)
	FetchKeys(pattern string) ([]string, error)
	SetNX(key string, value any, expiration time.Duration) (bool, error)
	TTL(key string) (time.Duration, error)
	Close() error
}

type Redis struct {
	Client    *redis.Client
	Ctx       context.Context
	CancelCtx context.CancelFunc
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
		log.Printf("Error closing Redis client: %v", err)
		return fmt.Errorf("failed to close Redis client: %w", err)
	}
	return nil
}

func (c *Redis) get(key string) ([]byte, error) {
	resp, err := c.Client.Get(c.Ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Redis) FetchKeys(pattern string) ([]string, error) {
	keys := []string{}
	iter := c.Client.Scan(c.Ctx, 0, pattern, 0).Iterator()
	for iter.Next(c.Ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate through Redis keys with pattern '%s': %w", pattern, err)
	}
	return keys, nil
}

func (c *Redis) Set(key string, value any, duration ...time.Duration) error {
	serialized, err := json.Marshal(value)
	if err != nil {
		return err
	}
	expiry := time.Hour * 24 * 30
	if len(duration) > 0 {
		expiry = duration[0]
	}
	_, err = c.Client.SetEX(c.Ctx, key, serialized, expiry).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Redis) Get(key string, dest any) error {
	data, err := c.get(key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return err
	}
	return nil
}

func (c *Redis) Delete(key string) error {
	del, err := c.Client.Del(c.Ctx, key).Result()
	if err != nil {
		return err
	}
	if del == 0 {
		return fmt.Errorf("key %s not found for deletion", key)
	}
	return nil
}

func (c *Redis) Incr(key string) (int64, error) {
	newValue, err := c.Client.Incr(c.Ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return newValue, nil
}

func (c *Redis) Expire(key string, duration time.Duration) error {
	if err := c.Client.Expire(c.Ctx, key, duration).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Redis) Exists(key string) (bool, error) {
	exists, err := c.Client.Exists(c.Ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (c *Redis) HSet(key string, values map[string]any) error {
	_, err := c.Client.HSet(c.Ctx, key, values).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Redis) HGet(key, field string) (string, error) {
	val, err := c.Client.HGet(c.Ctx, key, field).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *Redis) HGetAll(key string) (map[string]string, error) {
	vals, err := c.Client.HGetAll(c.Ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return vals, nil
}

// SetNX sets the value of a key, only if the key does not already exist.
func (c *Redis) SetNX(key string, value any, expiration time.Duration) (bool, error) {
	result, err := c.Client.SetNX(c.Ctx, key, value, expiration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// TTL returns the remaining time to live of a key that has a timeout.
func (c *Redis) TTL(key string) (time.Duration, error) {
	result, err := c.Client.TTL(c.Ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}
