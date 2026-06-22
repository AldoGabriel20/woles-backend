-- +goose Up

-- ── Plans ─────────────────────────────────────────────────────────────────────
CREATE TABLE plans (
    id                  UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(64)    NOT NULL,
    price_idr           NUMERIC(12, 2) NOT NULL DEFAULT 0,
    reminder_limit      INT            NOT NULL DEFAULT -1,   -- -1 = unlimited
    document_limit      INT            NOT NULL DEFAULT -1,
    subscription_limit  INT            NOT NULL DEFAULT -1,
    goal_tracker        BOOLEAN        NOT NULL DEFAULT FALSE,
    timeline            BOOLEAN        NOT NULL DEFAULT FALSE,
    family_account      BOOLEAN        NOT NULL DEFAULT FALSE,
    ai_chat             BOOLEAN        NOT NULL DEFAULT FALSE,
    ai_chat_quota       INT            NOT NULL DEFAULT 0,    -- -1 = unlimited
    ocr                 BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_plans_name ON plans (name);

-- ── Usage limits (per user) ───────────────────────────────────────────────────
CREATE TABLE usage_limits (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reminders_used      INT         NOT NULL DEFAULT 0,
    documents_used      INT         NOT NULL DEFAULT 0,
    subscriptions_used  INT         NOT NULL DEFAULT 0,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_usage_limits_user_id ON usage_limits (user_id);

-- ── Seed plan data ────────────────────────────────────────────────────────────
-- +goose StatementBegin
INSERT INTO plans (name, price_idr, reminder_limit, document_limit, subscription_limit,
                   goal_tracker, timeline, family_account, ai_chat, ai_chat_quota, ocr)
VALUES
    ('free',     0,     20,  5,  5,  FALSE, FALSE, FALSE, FALSE,  0,  FALSE),
    ('premium',  39000, -1, -1, -1,  TRUE,  TRUE,  FALSE, FALSE,  0,  FALSE),
    ('advanced', 99000, -1, -1, -1,  TRUE,  TRUE,  TRUE,  TRUE,  -1,  TRUE);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS usage_limits;
DROP TABLE IF EXISTS plans;
