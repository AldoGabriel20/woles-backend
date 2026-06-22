-- +goose Up
CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email               VARCHAR(255),
    phone               VARCHAR(32),
    password_hash       VARCHAR(255),
    name                VARCHAR(255),
    avatar_url          VARCHAR(1024),
    timezone            VARCHAR(64)  NOT NULL DEFAULT 'Asia/Jakarta',
    plan                VARCHAR(32)  NOT NULL DEFAULT 'free'
                            CHECK (plan IN ('free', 'premium', 'advanced')),
    account_status      VARCHAR(32)  NOT NULL DEFAULT 'active'
                            CHECK (account_status IN ('active', 'suspended', 'deleted')),
    failed_login_count  INT          NOT NULL DEFAULT 0,
    locked_until        TIMESTAMPTZ,
    totp_secret         VARCHAR(512),
    totp_enabled        BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_email ON users (email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX idx_users_phone ON users (phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_account_status ON users (account_status);
CREATE INDEX idx_users_locked_until   ON users (locked_until) WHERE locked_until IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS users;
