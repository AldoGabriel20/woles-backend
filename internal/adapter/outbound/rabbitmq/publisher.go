package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher implements port/outbound/message.EventPublisher using RabbitMQ.
type Publisher struct {
	ch *amqp.Channel
}

// NewPublisher creates a Publisher from an existing Client.
func NewPublisher(c *Client) *Publisher {
	return &Publisher{ch: c.ch}
}

// Publish serialises payload to JSON and sends it to woles.events exchange with
// the given routing key. All messages are persistent (delivery mode 2).
func (p *Publisher) Publish(_ context.Context, routingKey string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("publisher: marshal: %w", err)
	}
	return p.ch.Publish(
		"woles.events", // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // delivery mode 2
			Body:         body,
		},
	)
}
