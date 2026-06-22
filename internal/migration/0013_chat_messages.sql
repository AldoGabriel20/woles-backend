-- +goose Up
CREATE TABLE chat_messages (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role            VARCHAR(16) NOT NULL CHECK (role IN ('user', 'assistant')),
    content         TEXT        NOT NULL,
    detected_intent JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_user_created_at ON chat_messages (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS chat_messages;
