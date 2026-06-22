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
	ID             string
	UserID         string
	Role           MessageRole
	Content        string
	DetectedIntent json.RawMessage
	CreatedAt      time.Time
}

// ChatUsage tracks monthly AI Chat message consumption for a user.
type ChatUsage struct {
	ID           string
	UserID       string
	Month        time.Time // first day of the billing month
	MessagesUsed int
	Quota        int
	UpdatedAt    time.Time
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
