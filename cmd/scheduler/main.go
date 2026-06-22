// Scheduler: claims due notifications every 60 seconds and publishes
// notification.send_requested for each.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	outboundRabbitmq "github.com/woles/woles-backend/internal/adapter/outbound/rabbitmq"
	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
	portdb "github.com/woles/woles-backend/internal/port/outbound/database"
	portmessage "github.com/woles/woles-backend/internal/port/outbound/message"
)

const batchSize = 50

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rabbitClient, err := outboundRabbitmq.New(ctx)
	if err != nil {
		log.Fatalf("scheduler: rabbitmq: %v", err)
	}
	defer rabbitClient.Close()

	publisher := outboundRabbitmq.NewPublisher(rabbitClient)

	// In production, inject a real NotificationRepository via DI / wire.
	var notifRepo portdb.NotificationRepository

	log.Println("scheduler: starting, tick interval = 60s")
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Run immediately on start, then on every tick.
	runOnce(ctx, notifRepo, publisher)

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler: shutting down")
			return
		case <-ticker.C:
			runOnce(ctx, notifRepo, publisher)
		}
	}
}

func runOnce(ctx context.Context, notifRepo portdb.NotificationRepository, publisher portmessage.EventPublisher) {
	if notifRepo == nil {
		log.Println("scheduler: notifRepo not wired, skipping")
		return
	}

	notifications, err := notifRepo.ClaimDue(ctx, batchSize)
	if err != nil {
		log.Printf("scheduler: ClaimDue error: %v", err)
		return
	}
	if len(notifications) == 0 {
		return
	}
	log.Printf("scheduler: claimed %d notification(s)", len(notifications))

	for _, n := range notifications {
		// Mark as sending.
		if err := notifRepo.UpdateStatus(ctx, n.ID, domainnotification.StatusSending); err != nil {
			log.Printf("scheduler: update sending status for %s: %v", n.ID, err)
			continue
		}
		// Publish send request.
		if err := publisher.Publish(ctx, "notification.send_requested", map[string]string{
			"notification_id": n.ID,
		}); err != nil {
			log.Printf("scheduler: publish send_requested for %s: %v", n.ID, err)
		}
	}
}
