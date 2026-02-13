package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"event-listener/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queueName     = "webhook_queue"
	dlxName       = "webhook_dlx"
	dlqName       = "webhook_queue_dlq"
	dlqRoutingKey = "webhook_failed"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare dead letter exchange
	err = ch.ExchangeDeclare(
		dlxName,
		"direct",
		false, // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLX: %w", err)
	}

	// Declare dead letter queue
	_, err = ch.QueueDeclare(
		dlqName,
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Bind DLQ to DLX
	err = ch.QueueBind(
		dlqName,
		dlqRoutingKey,
		dlxName,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind DLQ: %w", err)
	}

	// Declare main queue with DLX configuration
	_, err = ch.QueueDeclare(
		queueName,
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    dlxName,
			"x-dead-letter-routing-key": dlqRoutingKey,
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

func (p *Publisher) Publish(event models.WebhookEvent) error {
	qMessage := models.QMessage{
		Content:   event,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(qMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published event: %s (webhook_id: %s)", event.EventName, event.WebhookID)
	return nil
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
