// Package rabbitmq provides a RabbitMQ connection helper shared by the
// publisher and consumer adapters.
package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client holds an AMQP connection and a channel pool (single channel for MVP).
type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// New dials RabbitMQ using the RABBITMQ_URL environment variable, declares the
// exchanges required by the application, and returns a ready Client.
func New(_ context.Context) (*Client, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		return nil, errors.New("RABBITMQ_URL is not set")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq: open channel: %w", err)
	}
	c := &Client{conn: conn, ch: ch}
	if err := c.declareTopology(); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// declareTopology idempotently declares exchanges and queues.
func (c *Client) declareTopology() error {
	// Primary events exchange.
	if err := c.ch.ExchangeDeclare(
		"woles.events", // name
		"topic",        // kind
		true,           // durable
		false,          // auto-delete
		false,          // internal
		false,          // no-wait
		nil,
	); err != nil {
		return fmt.Errorf("rabbitmq: declare exchange woles.events: %w", err)
	}

	// Dead-letter exchange.
	if err := c.ch.ExchangeDeclare(
		"woles.dlx",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("rabbitmq: declare exchange woles.dlx: %w", err)
	}

	// Dead-letter queue.
	if _, err := c.ch.QueueDeclare(
		"woles.dlq",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	); err != nil {
		return fmt.Errorf("rabbitmq: declare woles.dlq: %w", err)
	}
	if err := c.ch.QueueBind("woles.dlq", "#", "woles.dlx", false, nil); err != nil {
		return fmt.Errorf("rabbitmq: bind woles.dlq: %w", err)
	}

	// Worker queues with DLX routing.
	dlxArgs := amqp.Table{
		"x-dead-letter-exchange": "woles.dlx",
	}

	queues := []struct {
		name        string
		bindingKeys []string
	}{
		{"intent_worker_queue", []string{"whatsapp.message_received"}},
		{"notification_send_queue", []string{"notification.send_requested"}},
		{"notification_result_queue", []string{"notification.sent", "notification.failed"}},
	}
	for _, q := range queues {
		if _, err := c.ch.QueueDeclare(q.name, true, false, false, false, dlxArgs); err != nil {
			return fmt.Errorf("rabbitmq: declare queue %s: %w", q.name, err)
		}
		for _, key := range q.bindingKeys {
			if err := c.ch.QueueBind(q.name, key, "woles.events", false, nil); err != nil {
				return fmt.Errorf("rabbitmq: bind %s → %s: %w", q.name, key, err)
			}
		}
	}
	return nil
}

// Channel returns the underlying amqp.Channel for use by publisher/consumer.
func (c *Client) Channel() *amqp.Channel { return c.ch }

// Close releases the channel and connection.
func (c *Client) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
