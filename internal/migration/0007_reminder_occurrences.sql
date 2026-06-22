-- +goose Up
CREATE TABLE reminder_occurrences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reminder_id     UUID         NOT NULL REFERENCES reminders (id) ON DELETE CASCADE,
    user_id         UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    scheduled_at    TIMESTAMPTZ  NOT NULL,
    completed_at    TIMESTAMPTZ,
    status          VARCHAR(32)  NOT NULL DEFAULT 'scheduled'
                        CHECK (status IN ('scheduled', 'sent', 'done', 'skipped', 'failed')),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reminder_occurrences_reminder_id      ON reminder_occurrences (reminder_id);
CREATE INDEX idx_reminder_occurrences_scheduled_status ON reminder_occurrences (scheduled_at, status);
CREATE INDEX idx_reminder_occurrences_user_status      ON reminder_occurrences (user_id, status);

-- +goose Down
DROP TABLE IF EXISTS reminder_occurrences;
