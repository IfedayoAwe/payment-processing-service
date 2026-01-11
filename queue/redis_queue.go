package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	maxRetries = 3
	retryDelay = 5 * time.Second
)

type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{
		client: client,
		ctx:    context.Background(),
	}
}

func (q *RedisQueue) Enqueue(ctx context.Context, jobType JobType, payload interface{}) error {
	job := Job{
		ID:        uuid.New().String(),
		Type:      jobType,
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	job.Payload = payloadBytes

	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	queueKey := fmt.Sprintf("queue:%s", jobType)
	if err := q.client.LPush(ctx, queueKey, jobBytes).Err(); err != nil {
		return fmt.Errorf("enqueue job: %w", err)
	}

	return nil
}

func (q *RedisQueue) Dequeue(ctx context.Context, jobType JobType, timeout time.Duration) (*Job, error) {
	queueKey := fmt.Sprintf("queue:%s", jobType)

	result, err := q.client.BRPop(ctx, timeout, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid brpop result")
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}

	return &job, nil
}

func (q *RedisQueue) Process(ctx context.Context, jobType JobType, handler JobHandler, timeout time.Duration) error {
	job, err := q.Dequeue(ctx, jobType, timeout)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	if err := handler(ctx, job); err != nil {
		job.Attempts++
		if job.Attempts < maxRetries {
			time.Sleep(retryDelay)
			return q.Retry(ctx, job)
		}
		return fmt.Errorf("job failed after %d attempts: %w", maxRetries, err)
	}

	return nil
}

func (q *RedisQueue) Retry(ctx context.Context, job *Job) error {
	return q.Enqueue(ctx, job.Type, job.Payload)
}
