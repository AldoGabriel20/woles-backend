// Package message defines outbound port interfaces for message broker publishing.
package message

import "context"

// EventPublisher publishes domain events to a message broker (e.g. RabbitMQ).
type EventPublisher interface {
	// Publish sends payload to the given routing key.
	Publish(ctx context.Context, routingKey string, payload interface{}) error
}
