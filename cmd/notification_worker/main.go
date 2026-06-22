// Notification worker: consumes notification.send_requested, sends via
// WhatsApp, and updates notification status.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	inboundRabbitmq "github.com/woles/woles-backend/internal/adapter/inbound/rabbitmq"
	outboundRabbitmq "github.com/woles/woles-backend/internal/adapter/outbound/rabbitmq"
	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
	portdb "github.com/woles/woles-backend/internal/port/outbound/database"
	portmessage "github.com/woles/woles-backend/internal/port/outbound/message"
	portwhatsapp "github.com/woles/woles-backend/internal/port/outbound/whatsapp"
)

const maxRetries = 3

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rabbitClient, err := outboundRabbitmq.New(ctx)
	if err != nil {
		log.Fatalf("notification_worker: rabbitmq: %v", err)
	}
	defer rabbitClient.Close()

	consumer := inboundRabbitmq.NewConsumer(rabbitClient.Channel())
	publisher := outboundRabbitmq.NewPublisher(rabbitClient)

	// In production, inject real implementations via DI / wire.
	var (
		notifRepo portdb.NotificationRepository
		sender    portwhatsapp.WhatsAppSender
	)

	log.Println("notification_worker: listening on notification_send_queue")
	if err := consumer.Consume(ctx, "notification_send_queue", 10, func(ctx context.Context, body []byte) error {
		return handleSendRequested(ctx, body, notifRepo, sender, publisher)
	}); err != nil {
		log.Fatalf("notification_worker: consume: %v", err)
	}
}

// sendRequestedPayload is the message body for notification.send_requested.
type sendRequestedPayload struct {
	NotificationID string `json:"notification_id"`
}

func handleSendRequested(
	ctx context.Context,
	body []byte,
	notifRepo portdb.NotificationRepository,
	sender portwhatsapp.WhatsAppSender,
	publisher portmessage.EventPublisher,
) error {
	var payload sendRequestedPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("notification_worker: unmarshal error (ack): %v", err)
		return nil // permanent failure → ack
	}

	if notifRepo == nil || sender == nil {
		log.Printf("notification_worker: dependencies not wired, skipping %s", payload.NotificationID)
		return nil
	}

	// Look up notification.
	n, err := notifRepo.FindByID(ctx, payload.NotificationID)
	if err != nil {
		log.Printf("notification_worker: find notification %s: %v (ack)", payload.NotificationID, err)
		return nil // not found → ack to avoid DLQ loop
	}

	// Double-check status to prevent duplicate sends.
	if n.Status != domainnotification.StatusSending {
		log.Printf("notification_worker: notification %s status=%s, skipping", n.ID, n.Status)
		return nil
	}

	// Attempt WhatsApp send.
	providerMsgID, sendErr := sender.SendMessage(ctx, n.UserID, "reminder_alert", map[string]string{
		"notification_id": n.ID,
	})

	if sendErr == nil {
		// Success path.
		now := time.Now()
		n.SentAt = &now
		n.ProviderMessageID = &providerMsgID
		if err := notifRepo.UpdateStatus(ctx, n.ID, domainnotification.StatusSent); err != nil {
			log.Printf("notification_worker: update sent status for %s: %v", n.ID, err)
		}
		_ = publisher.Publish(ctx, "notification.sent", map[string]string{"notification_id": n.ID})
		return nil
	}

	// Failure path.
	log.Printf("notification_worker: send failed for %s: %v", n.ID, sendErr)
	if err := notifRepo.IncrementRetry(ctx, n.ID); err != nil {
		log.Printf("notification_worker: increment retry for %s: %v", n.ID, err)
	}

	// Re-fetch to get updated retry count.
	n, err = notifRepo.FindByID(ctx, n.ID)
	if err != nil {
		return nil
	}

	if n.RetryCount < maxRetries {
		// Reset to scheduled so the scheduler picks it up again.
		_ = notifRepo.UpdateStatus(ctx, n.ID, domainnotification.StatusScheduled)
	} else {
		// Max retries exceeded → mark failed.
		reason := sendErr.Error()
		n.FailureReason = &reason
		_ = notifRepo.UpdateStatus(ctx, n.ID, domainnotification.StatusFailed)
		_ = publisher.Publish(ctx, "notification.failed", map[string]string{
			"notification_id": n.ID,
			"reason":          reason,
		})
	}
	return nil
}
