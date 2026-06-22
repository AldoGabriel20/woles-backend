-- +goose Up
CREATE TABLE user_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_token_id    UUID         NOT NULL REFERENCES refresh_tokens (id) ON DELETE CASCADE,
    device_name         VARCHAR(255),
    ip_address          VARCHAR(128),
    user_agent          VARCHAR(512),
    last_active_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_last_active ON user_sessions (user_id, last_active_at);

-- +goose Down
DROP TABLE IF EXISTS user_sessions;
