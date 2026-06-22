package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/chat"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ChatMessageRepo implements database.ChatMessageRepository.
type ChatMessageRepo struct {
	pool *pgxpool.Pool
}

// NewChatMessageRepo creates a new ChatMessageRepo.
func NewChatMessageRepo(pool *pgxpool.Pool) *ChatMessageRepo {
	return &ChatMessageRepo{pool: pool}
}

var chatMessageSortAllowlist = map[string]struct{}{
	"id": {}, "created_at": {},
}

func scanChatMessage(row interface {
	Scan(dest ...any) error
}) (*chat.ChatMessage, error) {
	m := &chat.ChatMessage{}
	var role string
	var payload []byte
	err := row.Scan(&m.ID, &m.UserID, &role, &m.Content, &payload, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	m.Role = chat.MessageRole(role)
	if len(payload) > 0 {
		m.DetectedIntent = json.RawMessage(payload)
	}
	return m, nil
}

// Create inserts a new chat message.
func (r *ChatMessageRepo) Create(ctx context.Context, m *chat.ChatMessage) error {
	var payload []byte
	if len(m.DetectedIntent) > 0 {
		payload = m.DetectedIntent
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO chat_messages (id, user_id, role, content, detected_intent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.UserID, string(m.Role), m.Content, payload, m.CreatedAt,
	)
	return err
}

// FindAllByUser returns a paginated list of chat messages, sorted by created_at ASC.
func (r *ChatMessageRepo) FindAllByUser(
	ctx context.Context, userID string, p database.PaginationParams,
) (*database.PaginatedResult[*chat.ChatMessage], error) {
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM chat_messages WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, err
	}

	offset := pageOffset(p.Page, p.PerPage)
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, role, content, detected_intent, created_at
		FROM chat_messages
		WHERE user_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`,
		userID, p.PerPage, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*chat.ChatMessage
	for rows.Next() {
		m, err := scanChatMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &database.PaginatedResult[*chat.ChatMessage]{
		Items: items, Total: total, Page: p.Page,
		PerPage: p.PerPage, TotalPages: totalPages(total, p.PerPage),
	}, nil
}

// DeleteAllByUser hard-deletes all chat messages for the user.
func (r *ChatMessageRepo) DeleteAllByUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM chat_messages WHERE user_id = $1`, userID)
	return err
}
