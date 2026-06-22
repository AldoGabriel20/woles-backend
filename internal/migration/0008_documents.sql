-- +goose Up
CREATE TABLE documents (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    family_member_id  UUID         REFERENCES family_members (id) ON DELETE SET NULL,
    document_type     VARCHAR(64)  NOT NULL
                          CHECK (document_type IN (
                              'stnk', 'bpkb', 'vehicle_insurance',
                              'sim', 'passport', 'visa', 'ktp',
                              'health_insurance', 'life_insurance',
                              'tax', 'investment', 'other'
                          )),
    vault_category    VARCHAR(32)  NOT NULL
                          CHECK (vault_category IN ('vehicles', 'identity', 'insurance', 'financials', 'other')),
    title             VARCHAR(200) NOT NULL,
    expiry_date       DATE,
    reminder_offsets  INT[]        NOT NULL DEFAULT '{30,7,1}',
    notes             TEXT,
    storage_type      VARCHAR(32)  NOT NULL DEFAULT 'physical'
                          CHECK (storage_type IN ('physical', 'digital', 'scan_verified')),
    file_url          VARCHAR(1024),
    file_size_bytes   BIGINT,
    file_mime_type    VARCHAR(128),
    status            VARCHAR(32)  NOT NULL DEFAULT 'active'
                          CHECK (status IN ('active', 'archived')),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_user_vault_category ON documents (user_id, vault_category);
CREATE INDEX idx_documents_expiry_date         ON documents (expiry_date) WHERE status = 'active';
CREATE INDEX idx_documents_family_member_id    ON documents (family_member_id) WHERE family_member_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS documents;
