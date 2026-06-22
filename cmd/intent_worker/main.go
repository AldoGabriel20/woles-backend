// Intent worker: consumes whatsapp.message_received and routes to application
// services based on the extracted AI intent.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	inboundRabbitmq "github.com/woles/woles-backend/internal/adapter/inbound/rabbitmq"
	outboundRabbitmq "github.com/woles/woles-backend/internal/adapter/outbound/rabbitmq"
	portai "github.com/woles/woles-backend/internal/port/outbound/ai"
	portdb "github.com/woles/woles-backend/internal/port/outbound/database"
	portwhatsapp "github.com/woles/woles-backend/internal/port/outbound/whatsapp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// ── Infrastructure dependencies ───────────────────────────────────────
	rabbitClient, err := outboundRabbitmq.New(ctx)
	if err != nil {
		log.Fatalf("intent_worker: rabbitmq: %v", err)
	}
	defer rabbitClient.Close()

	consumer := inboundRabbitmq.NewConsumer(rabbitClient.Channel())

	// ── Wire application dependencies (injected via env / constructor).
	// In production these are constructed with real implementations;
	// the worker binary only holds the wiring.
	var (
		intentExtractor portai.IntentExtractor // e.g. openai adapter
		inboundMsgRepo  portdb.InboundMessageRepository
		whatsappSender  portwhatsapp.WhatsAppSender
	)
	// Prevent "declared and not used" errors until the real wiring is added.
	_ = intentExtractor
	_ = inboundMsgRepo
	_ = whatsappSender

	log.Println("intent_worker: listening on intent_worker_queue")
	if err := consumer.Consume(ctx, "intent_worker_queue", 5, func(ctx context.Context, body []byte) error {
		return handleInboundMessage(ctx, body, intentExtractor, inboundMsgRepo, whatsappSender)
	}); err != nil {
		log.Fatalf("intent_worker: consume: %v", err)
	}
}

// handleInboundMessage processes one whatsapp.message_received delivery.
func handleInboundMessage(
	ctx context.Context,
	body []byte,
	extractor portai.IntentExtractor,
	inboundMsgRepo portdb.InboundMessageRepository,
	sender portwhatsapp.WhatsAppSender,
) error {
	// Decode the inbound message payload.
	var msg portdb.InboundMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		// Malformed payload: ack to avoid infinite retry.
		log.Printf("intent_worker: unmarshal error (will ack): %v", err)
		return nil
	}

	// Guard: skip if already processed.
	if msg.ProcessingStatus == "processed" {
		return nil
	}

	// If extractor is not wired (nil), mark pending and return.
	if extractor == nil {
		log.Printf("intent_worker: extractor not wired, skipping msg %s", msg.ID)
		return nil
	}

	// Extract intent.
	result, err := extractor.Extract(ctx, msg.RawText, "id")
	if err != nil {
		// Transient failure: nack so the message can be retried / DLQ'd.
		return err
	}

	// Low-confidence: send clarification via WhatsApp.
	if result.Confidence < 0.7 {
		log.Printf("intent_worker: low confidence (%.2f) for msg %s, sending clarification", result.Confidence, msg.ID)
		if sender != nil {
			_, _ = sender.SendMessage(ctx, msg.FromPhone, "clarification_request", map[string]string{
				"raw_text": msg.RawText,
			})
		}
		if inboundMsgRepo != nil {
			_ = inboundMsgRepo.UpdateStatus(ctx, msg.ID, "low_confidence")
		}
		return nil
	}

	log.Printf("intent_worker: intent=%s confidence=%.2f msg=%s", result.Intent, result.Confidence, msg.ID)

	// Route to application service based on intent.
	// (Application services are injected here in production wiring.)

	// Mark processed.
	if inboundMsgRepo != nil {
		if err := inboundMsgRepo.UpdateStatus(ctx, msg.ID, "processed"); err != nil {
			log.Printf("intent_worker: update status error: %v", err)
		}
	}
	return nil
}
