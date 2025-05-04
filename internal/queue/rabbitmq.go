// internal/queue/rabbitmq.go
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Nishithcs/bank-info/internal/domain"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// Queue names
	AccountCreationQueue = "account_creation"
	TransactionQueue     = "transaction"
)

// Task types
const (
	CreateAccountTask   = "create_account"
	ProcessTransactionTask = "process_transaction"
)

// Task represents a task to be processed
type Task struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"created_at"`
}

// AccountCreationPayload represents the payload for account creation
type AccountCreationPayload struct {
	Name          string  `json:"name"`
	InitialAmount float64 `json:"initial_amount"`
}

// TransactionPayload represents the payload for a transaction
type TransactionPayload struct {
	AccountID   string                `json:"account_id"`
	Amount      float64               `json:"amount"`
	Type        domain.TransactionType `json:"type"`
	ReferenceID string                `json:"reference_id"`
}

// QueueService handles message publishing and consuming
type QueueService interface {
	PublishAccountCreation(ctx context.Context, payload AccountCreationPayload) error
	PublishTransaction(ctx context.Context, payload TransactionPayload) (string, error)
	Consume(queueName string, handler func([]byte) error) error
	Close() error
}

type rabbitMQService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQService creates a new RabbitMQ service
func NewRabbitMQService(url string) (QueueService, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare queues
	queues := []string{AccountCreationQueue, TransactionQueue}
	for _, queue := range queues {
		_, err = ch.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue %s: %w", queue, err)
		}
	}

	return &rabbitMQService{
		conn:    conn,
		channel: ch,
	}, nil
}

// PublishAccountCreation publishes an account creation task
func (s *rabbitMQService) PublishAccountCreation(ctx context.Context, payload AccountCreationPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := Task{
		ID:        uuid.New().String(),
		Type:      CreateAccountTask,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	return s.publishTask(ctx, AccountCreationQueue, task)
}

// PublishTransaction publishes a transaction task
func (s *rabbitMQService) PublishTransaction(ctx context.Context, payload TransactionPayload) (string, error) {
	if payload.ReferenceID == "" {
		payload.ReferenceID = uuid.New().String()
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := Task{
		ID:        uuid.New().String(),
		Type:      ProcessTransactionTask,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	err = s.publishTask(ctx, TransactionQueue, task)
	if err != nil {
		return "", err
	}

	return payload.ReferenceID, nil
}

func (s *rabbitMQService) publishTask(ctx context.Context, queue string, task Task) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	return s.channel.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         taskBytes,
			DeliveryMode: amqp.Persistent,
		},
	)
}

// Consume starts consuming messages from a queue
func (s *rabbitMQService) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := s.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			err := handler(msg.Body)
			if err != nil {
				// Log error and nack the message to requeue
				fmt.Printf("Error processing message: %v\n", err)
				_ = msg.Nack(false, true)
			} else {
				_ = msg.Ack(false)
			}
		}
	}()

	return nil
}

// Close closes the connection
func (s *rabbitMQService) Close() error {
	if err := s.channel.Close(); err != nil {
		return err
	}
	return s.conn.Close()
}