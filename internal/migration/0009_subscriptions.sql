-- +goose Up
CREATE TABLE subscriptions (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name            VARCHAR(200) NOT NULL,
    amount          NUMERIC(14, 2) NOT NULL DEFAULT 0,
    currency        VARCHAR(8)   NOT NULL DEFAULT 'IDR',
    billing_cycle   VARCHAR(32)  NOT NULL
                        CHECK (billing_cycle IN ('monthly', 'yearly', 'custom')),
    next_billing_at TIMESTAMPTZ  NOT NULL,
    category        VARCHAR(32)  NOT NULL DEFAULT 'other'
                        CHECK (category IN ('entertainment', 'productivity', 'bill', 'other')),
    status          VARCHAR(32)  NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'archived', 'canceled')),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_status      ON subscriptions (user_id, status);
CREATE INDEX idx_subscriptions_next_billing_at  ON subscriptions (next_billing_at) WHERE status = 'active';

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
