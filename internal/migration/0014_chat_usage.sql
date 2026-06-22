-- +goose Up
CREATE TABLE chat_usage (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    month         DATE        NOT NULL,
    messages_used INT         NOT NULL DEFAULT 0,
    quota         INT         NOT NULL DEFAULT 10,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_chat_usage_user_month ON chat_usage (user_id, month);

-- +goose Down
DROP TABLE IF EXISTS chat_usage;
