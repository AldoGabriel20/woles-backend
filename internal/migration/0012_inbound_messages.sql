-- +goose Up
CREATE TABLE inbound_messages (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        REFERENCES users (id) ON DELETE SET NULL,
    channel             VARCHAR(32) NOT NULL DEFAULT 'whatsapp'
                            CHECK (channel IN ('whatsapp')),
    provider_message_id VARCHAR(255) NOT NULL,
    from_phone          VARCHAR(32)  NOT NULL,
    raw_text            TEXT         NOT NULL,
    parsed_intent       JSONB,
    processing_status   VARCHAR(32)  NOT NULL DEFAULT 'received'
                            CHECK (processing_status IN ('received', 'parsed', 'handled', 'failed')),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_inbound_messages_provider_message_id ON inbound_messages (provider_message_id);
CREATE INDEX        idx_inbound_messages_user_status         ON inbound_messages (user_id, processing_status);

-- +goose Down
DROP TABLE IF EXISTS inbound_messages;
