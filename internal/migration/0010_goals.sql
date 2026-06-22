-- +goose Up
CREATE TABLE goals (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID           NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title           VARCHAR(200)   NOT NULL,
    icon            VARCHAR(32)
                        CHECK (icon IN ('love', 'emergency', 'vehicle', 'home', 'travel', 'other')),
    target_amount   NUMERIC(18, 2) NOT NULL,
    current_amount  NUMERIC(18, 2) NOT NULL DEFAULT 0,
    monthly_target  NUMERIC(18, 2),
    currency        VARCHAR(8)     NOT NULL DEFAULT 'IDR',
    target_date     DATE,
    status          VARCHAR(32)    NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'completed', 'archived')),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_goals_user_status ON goals (user_id, status);

-- +goose Down
DROP TABLE IF EXISTS goals;
