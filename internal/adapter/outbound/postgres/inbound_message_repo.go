package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// InboundMessageRepo implements database.InboundMessageRepository.
type InboundMessageRepo struct {
	pool *pgxpool.Pool
}

// NewInboundMessageRepo creates a new InboundMessageRepo.
func NewInboundMessageRepo(pool *pgxpool.Pool) *InboundMessageRepo {
	return &InboundMessageRepo{pool: pool}
}

const inboundMessageSelectCols = `id, user_id, channel, provider_message_id, from_phone,
       raw_text, parsed_intent, processing_status, created_at`

func scanInboundMessage(row interface {
	Scan(dest ...any) error
}) (*database.InboundMessage, error) {
	m := &database.InboundMessage{}
	err := row.Scan(
		&m.ID, &m.UserID, &m.Channel, &m.ProviderMessageID,
		&m.FromPhone, &m.RawText, &m.ParsedIntent, &m.ProcessingStatus, &m.CreatedAt,
	)
	return m, err
}

// Create inserts a new inbound message row.
func (r *InboundMessageRepo) Create(ctx context.Context, m *database.InboundMessage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO inbound_messages
		  (id, user_id, channel, provider_message_id, from_phone,
		   raw_text, parsed_intent, processing_status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.UserID, m.Channel, m.ProviderMessageID,
		m.FromPhone, m.RawText, m.ParsedIntent, m.ProcessingStatus, m.CreatedAt,
	)
	return err
}

// FindByProviderMessageID returns the message with the given provider-side ID, or ErrNotFound.
func (r *InboundMessageRepo) FindByProviderMessageID(ctx context.Context, providerMessageID string) (*database.InboundMessage, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT "+inboundMessageSelectCols+" FROM inbound_messages WHERE provider_message_id = $1",
		providerMessageID,
	)
	m, err := scanInboundMessage(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return m, err
}

// UpdateStatus changes the processing_status for the given message.
func (r *InboundMessageRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE inbound_messages SET processing_status = $1 WHERE id = $2`,
		status, id,
	)
	return err
}
