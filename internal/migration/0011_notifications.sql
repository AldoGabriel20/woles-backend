-- +goose Up
CREATE TABLE notifications (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    entity_type         VARCHAR(32)  NOT NULL
                            CHECK (entity_type IN ('reminder', 'document', 'subscription', 'goal')),
    entity_id           UUID         NOT NULL,
    occurrence_id       UUID,
    channel             VARCHAR(32)  NOT NULL DEFAULT 'whatsapp'
                            CHECK (channel IN ('whatsapp', 'email', 'web_push')),
    scheduled_at        TIMESTAMPTZ  NOT NULL,
    sent_at             TIMESTAMPTZ,
    status              VARCHAR(32)  NOT NULL DEFAULT 'scheduled'
                            CHECK (status IN ('scheduled', 'sending', 'sent', 'failed', 'canceled')),
    idempotency_key     VARCHAR(512) NOT NULL,
    provider_message_id VARCHAR(255),
    failure_reason      TEXT,
    retry_count         INT          NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_notifications_idempotency_key    ON notifications (idempotency_key);
CREATE INDEX        idx_notifications_user_status        ON notifications (user_id, status);
CREATE INDEX        idx_notifications_scheduled_status   ON notifications (scheduled_at, status);

-- +goose Down
DROP TABLE IF EXISTS notifications;
