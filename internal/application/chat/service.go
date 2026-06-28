// Package chat implements the AI Chat application service.
package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	domainchat "github.com/woles/woles-backend/internal/domain/chat"
	domainidentity "github.com/woles/woles-backend/internal/domain/identity"
	"github.com/woles/woles-backend/internal/port/outbound/ai"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrQuotaExceeded = errors.New("monthly chat quota exceeded")
)

// ─── Response types ───────────────────────────────────────────────────────────

// SendMessageResult carries both the user and assistant messages produced by
// a single SendMessage call.
type SendMessageResult struct {
	UserMessage      *domainchat.ChatMessage `json:"user_message"`
	AssistantMessage *domainchat.ChatMessage `json:"assistant_message"`
}

// ChatUsageResponse is the public representation of monthly usage.
type ChatUsageResponse struct {
	MessagesUsed int    `json:"messages_used"`
	Quota        int    `json:"quota"`
	Remaining    int    `json:"remaining"`
	PlanName     string `json:"plan_name"`
}

// DetectedIntent is one aggregated intent category from chat history.
type DetectedIntent struct {
	Intent string `json:"intent"`
	Count  int    `json:"count"`
}

// freeQuota is the hard limit for PlanFree users (messages per month).
const freeQuota = 10

// ─── Service ──────────────────────────────────────────────────────────────────

// Service implements the AI Chat application service.
type Service struct {
	messages  database.ChatMessageRepository
	usage     database.ChatUsageRepository
	users     database.UserRepository
	extractor ai.IntentExtractor
}

// NewService constructs the chat service.
func NewService(
	messages database.ChatMessageRepository,
	usage database.ChatUsageRepository,
	users database.UserRepository,
	extractor ai.IntentExtractor,
) *Service {
	return &Service{
		messages:  messages,
		usage:     usage,
		users:     users,
		extractor: extractor,
	}
}

// ─── SendMessage ──────────────────────────────────────────────────────────────

// SendMessage stores a user message, calls the intent extractor, stores the
// assistant reply, and increments the monthly usage counter.
func (s *Service) SendMessage(ctx context.Context, userID, content string) (*SendMessageResult, error) {
	// Check quota.
	now := time.Now().UTC()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	u, err := s.usage.GetOrCreate(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("send message: check quota: %w", err)
	}

	// Determine effective quota: free plan = 10, otherwise unlimited (-1).
	if u.Quota == 0 {
		// Repo did not set a quota — look up the user's plan.
		user, uerr := s.users.FindByID(ctx, userID)
		if uerr != nil {
			return nil, fmt.Errorf("send message: find user: %w", uerr)
		}
		if user.Plan == domainidentity.PlanFree {
			u.Quota = freeQuota
		} else {
			u.Quota = -1 // unlimited
		}
	}

	if u.Quota > 0 && u.MessagesUsed >= u.Quota {
		return nil, ErrQuotaExceeded
	}

	// Sanitize content.
	content = sanitizeChat(content, 4000)

	// Store user message.
	userMsg := &domainchat.ChatMessage{
		ID:        uuid.NewString(),
		UserID:    userID,
		Role:      domainchat.RoleUser,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.messages.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("send message: store user message: %w", err)
	}

	// Call intent extractor.
	result, extractErr := s.extractor.Extract(ctx, content, "id")

	// Build assistant response content.
	var assistantContent string
	var detectedIntentJSON json.RawMessage
	if extractErr != nil {
		assistantContent = "Maaf, saya tidak dapat memproses permintaan Anda saat ini. Silakan coba lagi."
	} else {
		assistantContent = buildAssistantReply(result)
		if payloadBytes, merr := json.Marshal(result.Payload); merr == nil {
			detectedIntentJSON = payloadBytes
		}
	}

	// Store assistant message.
	assistantMsg := &domainchat.ChatMessage{
		ID:             uuid.NewString(),
		UserID:         userID,
		Role:           domainchat.RoleAssistant,
		Content:        assistantContent,
		DetectedIntent: detectedIntentJSON,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.messages.Create(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("send message: store assistant message: %w", err)
	}

	// Increment monthly usage counter.
	if incErr := s.usage.Increment(ctx, userID, month); incErr != nil {
		// Log but don't fail the request.
		_ = incErr
	}

	return &SendMessageResult{
		UserMessage:      userMsg,
		AssistantMessage: assistantMsg,
	}, nil
}

// ─── GetMessages ──────────────────────────────────────────────────────────────

// GetMessages returns paginated chat messages for the user, sorted by
// created_at ASC.
func (s *Service) GetMessages(ctx context.Context, userID string, page, perPage int) (*database.PaginatedResult[*domainchat.ChatMessage], error) {
	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "created_at",
		Order:   "asc",
	}
	result, err := s.messages.FindAllByUser(ctx, userID, p)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	return result, nil
}

// ─── DeleteAllMessages ────────────────────────────────────────────────────────

// DeleteAllMessages hard-deletes all chat messages for the user.
func (s *Service) DeleteAllMessages(ctx context.Context, userID string) error {
	if err := s.messages.DeleteAllByUser(ctx, userID); err != nil {
		return fmt.Errorf("delete all messages: %w", err)
	}
	return nil
}

// ─── GetUsage ─────────────────────────────────────────────────────────────────

// GetUsage returns the current month's chat usage counters and the user's plan.
func (s *Service) GetUsage(ctx context.Context, userID string) (*ChatUsageResponse, error) {
	now := time.Now().UTC()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	u, err := s.usage.GetOrCreate(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("get usage: %w", err)
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get usage: find user: %w", err)
	}

	quota := u.Quota
	if quota == 0 {
		if user.Plan == domainidentity.PlanFree {
			quota = freeQuota
		} else {
			quota = -1
		}
	}

	remaining := quota - u.MessagesUsed
	if quota < 0 {
		remaining = -1
	} else if remaining < 0 {
		remaining = 0
	}

	return &ChatUsageResponse{
		MessagesUsed: u.MessagesUsed,
		Quota:        quota,
		Remaining:    remaining,
		PlanName:     string(user.Plan),
	}, nil
}

// ─── GetDetectedIntents ───────────────────────────────────────────────────────

// GetDetectedIntents returns the top 5 unique intent types from the user's
// recent assistant messages, ordered by frequency descending.
func (s *Service) GetDetectedIntents(ctx context.Context, userID string) ([]*DetectedIntent, error) {
	p := database.PaginationParams{
		Page:    1,
		PerPage: 200, // scan last 200 messages
		Sort:    "created_at",
		Order:   "desc",
	}
	result, err := s.messages.FindAllByUser(ctx, userID, p)
	if err != nil {
		return nil, fmt.Errorf("get detected intents: %w", err)
	}

	counts := map[string]int{}
	for _, msg := range result.Items {
		if msg.Role != domainchat.RoleAssistant || len(msg.DetectedIntent) == 0 {
			continue
		}
		// DetectedIntent is a JSON object; extract the "intent" key.
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.DetectedIntent, &payload); err != nil {
			continue
		}
		if intent, ok := payload["intent"].(string); ok && intent != "" {
			counts[intent]++
		}
	}

	// Convert to slice and sort by count descending.
	intents := make([]*DetectedIntent, 0, len(counts))
	for intent, count := range counts {
		intents = append(intents, &DetectedIntent{Intent: intent, Count: count})
	}
	sort.Slice(intents, func(i, j int) bool {
		return intents[i].Count > intents[j].Count
	})

	// Return top 5.
	if len(intents) > 5 {
		intents = intents[:5]
	}
	return intents, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// buildAssistantReply constructs a short human-readable acknowledgement
// based on the detected intent. A richer response generation would call an
// LLM; this scaffold provides intent-aware fallback text.
func buildAssistantReply(r *ai.IntentResult) string {
	switch r.Intent {
	case "create_reminder":
		return "Baik, saya akan membuat pengingat untuk Anda."
	case "create_subscription":
		return "Saya catat langganan baru Anda."
	case "create_goal":
		return "Tujuan keuangan Anda telah dicatat."
	case "create_document":
		return "Dokumen Anda siap disimpan."
	case "query_timeline":
		return "Berikut adalah jadwal mendatang Anda."
	case "general_query":
		return "Saya siap membantu! Anda bisa meminta saya untuk membuat pengingat, mencatat langganan, menyimpan dokumen, atau menetapkan target keuangan. Ada yang bisa saya bantu?"
	default:
		return "Ada yang bisa saya bantu? Contohnya: \"Ingatkan bayar listrik tiap tanggal 10\" atau \"Catat langganan Netflix Rp 186.000 per bulan\"."
	}
}

// sanitizeChat strips leading/trailing whitespace and truncates to maxRunes.
func sanitizeChat(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > maxRunes {
		runes := []rune(s)
		s = string(runes[:maxRunes])
	}
	return s
}
