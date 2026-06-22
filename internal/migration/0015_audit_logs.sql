-- +goose Up
CREATE TABLE audit_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        REFERENCES users (id) ON DELETE SET NULL,
    actor_type  VARCHAR(32) NOT NULL DEFAULT 'user'
                    CHECK (actor_type IN ('user', 'system', 'admin')),
    action      VARCHAR(128) NOT NULL,
    entity_type VARCHAR(64),
    entity_id   UUID,
    ip_address  VARCHAR(128),
    user_agent  VARCHAR(512),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_created_at    ON audit_logs (user_id, created_at DESC);
CREATE INDEX idx_audit_logs_entity             ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_action             ON audit_logs (action);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
