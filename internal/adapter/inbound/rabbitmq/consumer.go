// Package rabbitmq provides a RabbitMQ consumer for the inbound adapter.
package rabbitmq

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// HandlerFunc is the callback invoked for each delivered message.
// Return nil to ack; return any error to nack without requeue (moves to DLQ).
type HandlerFunc func(ctx context.Context, body []byte) error

// Consumer wraps an amqp.Channel and registers per-queue handlers.
type Consumer struct {
	ch *amqp.Channel
}

// NewConsumer creates a Consumer from an amqp.Channel (e.g. from outbound Client.Channel()).
func NewConsumer(ch *amqp.Channel) *Consumer {
	return &Consumer{ch: ch}
}

// Consume starts consuming messages from queueName and calls handler for each.
// It runs until ctx is cancelled. prefetch sets the QoS prefetch count.
func (c *Consumer) Consume(ctx context.Context, queueName string, prefetch int, handler HandlerFunc) error {
	if err := c.ch.Qos(prefetch, 0, false); err != nil {
		return fmt.Errorf("consumer: set qos: %w", err)
	}
	deliveries, err := c.ch.Consume(
		queueName,
		"",    // consumer tag (auto-generated)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consumer: consume %s: %w", queueName, err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("consumer: channel closed for queue %s", queueName)
			}
			if err := handler(ctx, d.Body); err != nil {
				// Permanent failure: nack without requeue → DLQ via policy.
				log.Printf("[consumer] nack message from %s: %v", queueName, err)
				if nackErr := d.Nack(false, false); nackErr != nil {
					log.Printf("[consumer] nack error: %v", nackErr)
				}
			} else {
				if ackErr := d.Ack(false); ackErr != nil {
					log.Printf("[consumer] ack error: %v", ackErr)
				}
			}
		}
	}
}
