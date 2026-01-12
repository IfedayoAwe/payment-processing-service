package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/utils"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQQueue struct {
	url         string
	conn        *amqp.Connection
	channel     *amqp.Channel
	connMutex   sync.RWMutex
	queries     gen.Querier
	notifyClose chan *amqp.Error
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewRabbitMQQueue(url string, queries gen.Querier) (*RabbitMQQueue, error) {
	ctx, cancel := context.WithCancel(context.Background())

	queue := &RabbitMQQueue{
		url:         url,
		queries:     queries,
		notifyClose: make(chan *amqp.Error),
		ctx:         ctx,
		cancel:      cancel,
	}

	if err := queue.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	if err := queue.setupQueues(); err != nil {
		cancel()
		queue.close()
		return nil, fmt.Errorf("setup queues: %w", err)
	}

	go queue.reconnect()

	return queue, nil
}

func (q *RabbitMQQueue) connect() error {
	q.connMutex.Lock()
	defer q.connMutex.Unlock()

	conn, err := amqp.Dial(q.url)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	if q.conn != nil {
		q.conn.Close()
	}
	if q.channel != nil {
		q.channel.Close()
	}

	q.conn = conn
	q.channel = channel
	q.notifyClose = make(chan *amqp.Error)
	q.conn.NotifyClose(q.notifyClose)

	utils.Logger.Info().Msg("connected to rabbitmq")
	return nil
}

func (q *RabbitMQQueue) setupQueues() error {
	q.connMutex.RLock()
	ch := q.channel
	q.connMutex.RUnlock()

	if ch == nil {
		return fmt.Errorf("channel not available")
	}

	exchangeName := "payment_processing"
	if err := ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	queues := []string{"payout", "webhook"}
	for _, queueName := range queues {
		queue, err := ch.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			amqp.Table{
				"x-dead-letter-exchange":    "",
				"x-dead-letter-routing-key": queueName + "_dlq",
			},
		)
		if err != nil {
			return fmt.Errorf("declare queue %s: %w", queueName, err)
		}

		dlq, err := ch.QueueDeclare(
			queueName+"_dlq",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("declare dead letter queue %s: %w", queueName, err)
		}

		if err := ch.QueueBind(
			queue.Name,
			queueName,
			exchangeName,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("bind queue %s: %w", queueName, err)
		}

		if err := ch.QueueBind(
			dlq.Name,
			queueName+"_dlq",
			exchangeName,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("bind dead letter queue %s: %w", queueName, err)
		}
	}

	return nil
}

func (q *RabbitMQQueue) reconnect() {
	for {
		select {
		case <-q.ctx.Done():
			return
		case err := <-q.notifyClose:
			if err != nil {
				utils.Logger.Error().Err(err).Msg("rabbitmq connection closed, reconnecting")
				q.reconnectWithBackoff()
			}
		}
	}
}

func (q *RabbitMQQueue) reconnectWithBackoff() {
	backoff := reconnectDelay
	maxBackoff := maxReconnectTime

	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			time.Sleep(backoff)

			if err := q.connect(); err != nil {
				utils.Logger.Error().Err(err).Dur("backoff", backoff).Msg("reconnection failed, retrying")
				backoff = min(backoff*2, maxBackoff)
				continue
			}

			if err := q.setupQueues(); err != nil {
				utils.Logger.Error().Err(err).Msg("failed to setup queues after reconnection")
				backoff = min(backoff*2, maxBackoff)
				continue
			}

			utils.Logger.Info().Msg("rabbitmq reconnected successfully")
			backoff = reconnectDelay
			return
		}
	}
}

func (q *RabbitMQQueue) getChannel() (*amqp.Channel, error) {
	q.connMutex.RLock()
	ch := q.channel
	q.connMutex.RUnlock()

	if ch == nil || ch.IsClosed() {
		return nil, fmt.Errorf("channel not available")
	}

	return ch, nil
}

func (q *RabbitMQQueue) Enqueue(ctx context.Context, jobType JobType, payload interface{}) error {
	ch, err := q.getChannel()
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}

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

	exchangeName := "payment_processing"
	routingKey := string(jobType)

	err = ch.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jobBytes,
			DeliveryMode: amqp.Persistent,
			MessageId:    job.ID,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("publish job: %w", err)
	}

	utils.Logger.Info().
		Str("job_id", job.ID).
		Str("job_type", string(jobType)).
		Msg("job enqueued to rabbitmq")

	return nil
}

func (q *RabbitMQQueue) Dequeue(ctx context.Context, jobType JobType, timeout time.Duration) (*Job, error) {
	ch, err := q.getChannel()
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	queueName := string(jobType)
	consumerTag := fmt.Sprintf("consumer-%s-%s", queueName, uuid.New().String()[:8])
	msgs, err := ch.Consume(
		queueName,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("consume queue: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer func() {
		cancel()
		ch.Cancel(consumerTag, false)
	}()

	select {
	case <-ctx.Done():
		return nil, nil
	case msg, ok := <-msgs:
		if !ok {
			return nil, nil
		}

		jobID := msg.MessageId
		if jobID == "" {
			var jobData Job
			if err := json.Unmarshal(msg.Body, &jobData); err == nil && jobData.ID != "" {
				jobID = jobData.ID
			} else {
				jobID = uuid.New().String()
			}
		}

		alreadyProcessed, err := q.queries.IsJobProcessed(ctx, jobID)
		if err != nil {
			utils.Logger.Error().
				Err(err).
				Str("job_id", jobID).
				Msg("error checking if job is processed, rejecting to be safe")
			msg.Nack(false, false)
			return nil, fmt.Errorf("check job processed: %w", err)
		}

		if alreadyProcessed {
			msg.Nack(false, false)
			utils.Logger.Warn().
				Str("job_id", jobID).
				Str("job_type", string(jobType)).
				Msg("duplicate job detected, rejecting")
			return nil, nil
		}

		var job Job
		if err := json.Unmarshal(msg.Body, &job); err != nil {
			msg.Nack(false, false)
			return nil, fmt.Errorf("unmarshal job: %w", err)
		}

		if job.ID == "" {
			job.ID = jobID
		}

		msgRef := msg
		jobIDForAck := jobID
		job.ack = func() {
			msgRef.Ack(false)
			q.markProcessed(context.Background(), jobIDForAck)
		}
		job.nack = func(requeue bool) {
			msgRef.Nack(false, requeue)
		}

		return &job, nil
	}
}

func (q *RabbitMQQueue) markProcessed(ctx context.Context, jobID string) {
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err := q.queries.MarkJobProcessed(ctx, gen.MarkJobProcessedParams{
		JobID:     jobID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		utils.Logger.Error().Err(err).Str("job_id", jobID).Msg("failed to mark job as processed")
	}
}

func (q *RabbitMQQueue) Process(ctx context.Context, jobType JobType, handler JobHandler, timeout time.Duration) error {
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
			if job.nack != nil {
				job.nack(true)
			}
			time.Sleep(retryDelay)
			return nil
		}

		if job.nack != nil {
			job.nack(false)
		}
		return fmt.Errorf("job failed after %d attempts: %w", maxRetries, err)
	}

	if job.ack != nil {
		job.ack()
	}

	return nil
}

func (q *RabbitMQQueue) Retry(ctx context.Context, job *Job) error {
	return q.Enqueue(ctx, job.Type, job.Payload)
}

func (q *RabbitMQQueue) Close() error {
	q.cancel()
	return q.close()
}

func (q *RabbitMQQueue) close() error {
	q.connMutex.Lock()
	defer q.connMutex.Unlock()

	var errs []error
	if q.channel != nil {
		if err := q.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close channel: %w", err))
		}
	}
	if q.conn != nil {
		if err := q.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing rabbitmq: %v", errs)
	}

	return nil
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
