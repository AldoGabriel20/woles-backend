package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	httpfiber "github.com/woles/woles-backend/internal/adapter/inbound/http_fiber"
	"github.com/woles/woles-backend/internal/adapter/inbound/http_fiber/middleware"
	"github.com/woles/woles-backend/internal/adapter/outbound/postgres"
	"github.com/woles/woles-backend/internal/adapter/outbound/rabbitmq"
	redisadapter "github.com/woles/woles-backend/internal/adapter/outbound/redis"
	storageprovider "github.com/woles/woles-backend/internal/adapter/outbound/storage_provider"
	appchat "github.com/woles/woles-backend/internal/application/chat"
	appdocument "github.com/woles/woles-backend/internal/application/document"
	appfamily "github.com/woles/woles-backend/internal/application/family"
	appgoal "github.com/woles/woles-backend/internal/application/goal"
	appidentity "github.com/woles/woles-backend/internal/application/identity"
	appnotification "github.com/woles/woles-backend/internal/application/notification"
	appreminder "github.com/woles/woles-backend/internal/application/reminder"
	appsubscription "github.com/woles/woles-backend/internal/application/subscription"
	apptimeline "github.com/woles/woles-backend/internal/application/timeline"
	aiport "github.com/woles/woles-backend/internal/port/outbound/ai"
	"github.com/woles/woles-backend/internal/port/outbound/storage"
	whatsappport "github.com/woles/woles-backend/internal/port/outbound/whatsapp"
)

// ─── Stubs for optional providers not yet implemented ─────────────────────────

type noopWhatsAppSender struct{}

func (noopWhatsAppSender) SendMessage(_ context.Context, _, _ string, _ map[string]string) (string, error) {
	return "", nil
}

var _ whatsappport.WhatsAppSender = noopWhatsAppSender{}

type noopIntentExtractor struct{}

func (noopIntentExtractor) Extract(_ context.Context, text, _ string) (*aiport.IntentResult, error) {
	lower := strings.ToLower(text)
	switch {
	case containsAny(lower, "ingatkan", "reminder", "pengingat", "bayar", "tagihan", "jatuh tempo", "perpanjang"):
		return &aiport.IntentResult{Intent: "create_reminder", Confidence: 0.75, Payload: map[string]interface{}{"text": text}}, nil
	case containsAny(lower, "langganan", "subscribe", "subscription", "berlangganan", "netflix", "spotify", "disney"):
		return &aiport.IntentResult{Intent: "create_subscription", Confidence: 0.75, Payload: map[string]interface{}{"text": text}}, nil
	case containsAny(lower, "target", "nabung", "tabungan", "tujuan", "goal", "simpan", "dana"):
		return &aiport.IntentResult{Intent: "create_goal", Confidence: 0.75, Payload: map[string]interface{}{"text": text}}, nil
	case containsAny(lower, "dokumen", "document", "sim", "stnk", "passport", "paspor", "ktp", "bpkb"):
		return &aiport.IntentResult{Intent: "create_document", Confidence: 0.75, Payload: map[string]interface{}{"text": text}}, nil
	case containsAny(lower, "jadwal", "timeline", "schedule", "kapan", "apa saja"):
		return &aiport.IntentResult{Intent: "query_timeline", Confidence: 0.70, Payload: map[string]interface{}{"text": text}}, nil
	default:
		return &aiport.IntentResult{Intent: "general_query", Confidence: 0.5, Payload: map[string]interface{}{"text": text}}, nil
	}
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

var _ aiport.IntentExtractor = noopIntentExtractor{}

// ─── main ─────────────────────────────────────────────────────────────────────

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// ── Load RSA public key for JWT verification ───────────────────────────────
	if err := middleware.LoadJWTPublicKey(); err != nil {
		log.Fatalf("jwt public key: %v", err)
	}
	log.Println("jwt: public key loaded")

	// ── Infrastructure ────────────────────────────────────────────────────────
	db, err := postgres.New(ctx)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Pool.Close()
	log.Println("postgres: connected")

	redisClient, err := redisadapter.New(ctx)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	log.Println("redis: connected")

	rabbitClient, err := rabbitmq.New(ctx)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer rabbitClient.Close()
	log.Println("rabbitmq: connected")

	var fileStore storage.FileStore
	minioStore, err := storageprovider.New(ctx)
	if err != nil {
		log.Printf("minio: warning — %v (file upload endpoints unavailable)", err)
	} else {
		fileStore = minioStore
		log.Println("minio: connected")
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	pool := db.Pool
	userRepo := postgres.NewUserRepo(pool)
	refreshTokenRepo := postgres.NewRefreshTokenRepo(pool)
	sessionRepo := postgres.NewUserSessionRepo(pool)
	usageLimitRepo := postgres.NewUsageLimitRepo(pool)
	auditLogRepo := postgres.NewAuditLogRepo(pool)
	reminderRepo := postgres.NewReminderRepo(pool)
	occurrenceRepo := postgres.NewReminderOccurrenceRepo(pool)
	notificationRepo := postgres.NewNotificationRepo(pool)
	documentRepo := postgres.NewDocumentRepo(pool)
	subscriptionRepo := postgres.NewSubscriptionRepo(pool)
	goalRepo := postgres.NewGoalRepo(pool)
	familyMemberRepo := postgres.NewFamilyMemberRepo(pool)
	chatMessageRepo := postgres.NewChatMessageRepo(pool)
	chatUsageRepo := postgres.NewChatUsageRepo(pool)
	timelineRepo := postgres.NewTimelineRepo(pool)

	// ── Message publisher ─────────────────────────────────────────────────────
	publisher := rabbitmq.NewPublisher(rabbitClient)

	// ── Application services ──────────────────────────────────────────────────
	identitySvc, err := appidentity.NewService(
		userRepo, refreshTokenRepo, sessionRepo,
		usageLimitRepo, auditLogRepo,
		redisadapter.NewOTPStore(redisClient),
		noopWhatsAppSender{},
		os.Getenv("JWT_PRIVATE_KEY_PATH"),
		[]byte(os.Getenv("APP_SECRET")),
	)
	if err != nil {
		log.Fatalf("identity service: %v", err)
	}

	reminderSvc := appreminder.NewService(
		reminderRepo, occurrenceRepo, notificationRepo,
		usageLimitRepo, auditLogRepo, publisher,
	)

	documentSvc := appdocument.NewService(
		documentRepo, notificationRepo, usageLimitRepo, auditLogRepo, fileStore,
	)

	subscriptionSvc := appsubscription.NewService(
		subscriptionRepo, notificationRepo, usageLimitRepo, auditLogRepo,
	)

	goalSvc := appgoal.NewService(goalRepo, userRepo, auditLogRepo)

	timelineSvc := apptimeline.NewService(timelineRepo)

	notificationSvc := appnotification.NewService(notificationRepo)

	familySvc := appfamily.NewService(familyMemberRepo, userRepo, reminderRepo, auditLogRepo)

	chatSvc := appchat.NewService(chatMessageRepo, chatUsageRepo, userRepo, noopIntentExtractor{})

	svc := &httpfiber.Services{
		Identity:     identitySvc,
		Reminder:     reminderSvc,
		Document:     documentSvc,
		Subscription: subscriptionSvc,
		Goal:         goalSvc,
		Timeline:     timelineSvc,
		Notification: notificationSvc,
		Family:       familySvc,
		Chat:         chatSvc,
		Users:        userRepo,
		UsageLimits:  usageLimitRepo,
		FileStore:    fileStore,
		RateLimiter:  redisadapter.NewRateLimiter(redisClient),
	}

	// ── HTTP server ───────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Woles API v1",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		BodyLimit:    11 * 1024 * 1024, // 11 MB (accounts for 10 MB file + metadata)
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   "internal_server_error",
				"message": err.Error(),
			})
		},
	})

	// Global middleware.
	app.Use(recover.New())
	app.Use(middleware.SecurityHeadersMiddleware())
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.CSRFMiddleware())

	// Health check — no auth, no rate limit.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "woles-backend"})
	})

	// Register all API and webhook routes.
	httpfiber.RegisterRoutes(app, svc)

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Woles backend listening on :%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("bye")
}
