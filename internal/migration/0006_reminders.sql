-- +goose Up
CREATE TABLE reminders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title           VARCHAR(200) NOT NULL,
    category        VARCHAR(32)  NOT NULL DEFAULT 'custom'
                        CHECK (category IN ('bill', 'vehicle', 'document', 'custom')),
    recurrence_type VARCHAR(32)  NOT NULL
                        CHECK (recurrence_type IN ('one_time', 'daily', 'weekly', 'monthly', 'yearly', 'custom_interval')),
    recurrence_rule JSONB,
    next_run_at     TIMESTAMPTZ  NOT NULL,
    timezone        VARCHAR(64)  NOT NULL DEFAULT 'Asia/Jakarta',
    status          VARCHAR(32)  NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'paused', 'archived')),
    source          VARCHAR(32)  NOT NULL DEFAULT 'web'
                        CHECK (source IN ('whatsapp', 'web', 'system')),
    original_text   TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reminders_user_status  ON reminders (user_id, status);
CREATE INDEX idx_reminders_next_run_at  ON reminders (next_run_at) WHERE status = 'active';

-- +goose Down
DROP TABLE IF EXISTS reminders;
