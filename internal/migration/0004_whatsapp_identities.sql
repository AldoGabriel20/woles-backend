-- +goose Up
CREATE TABLE whatsapp_identities (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    phone               VARCHAR(32)  NOT NULL,
    provider            VARCHAR(64)  NOT NULL,
    provider_contact_id VARCHAR(255),
    status              VARCHAR(32)  NOT NULL DEFAULT 'active'
                            CHECK (status IN ('active', 'blocked', 'disconnected')),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_whatsapp_identities_phone_provider ON whatsapp_identities (phone, provider);
CREATE INDEX idx_whatsapp_identities_user_id ON whatsapp_identities (user_id);

-- +goose Down
DROP TABLE IF EXISTS whatsapp_identities;
