package chat

import (
	"encoding/json"
	"time"
)

// MessageRole identifies who sent a chat message.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// ChatMessage represents a single message in the AI Chat Hub conversation.
type ChatMessage struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	Role           MessageRole     `json:"role"`
	Content        string          `json:"content"`
	DetectedIntent json.RawMessage `json:"detected_intent,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// ChatUsage tracks monthly AI Chat message consumption for a user.
type ChatUsage struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Month        time.Time `json:"month"`
	MessagesUsed int       `json:"messages_used"`
	Quota        int       `json:"quota"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Remaining returns how many messages the user may still send this month.
// Returns 0 when quota is exhausted; -1 when quota is unlimited (Quota == -1).
func (u *ChatUsage) Remaining() int {
	if u.Quota < 0 {
		return -1 // unlimited
	}
	r := u.Quota - u.MessagesUsed
	if r < 0 {
		return 0
	}
	return r
}
