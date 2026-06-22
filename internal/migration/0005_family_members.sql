-- +goose Up
CREATE TABLE family_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id   UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    role            VARCHAR(32)  NOT NULL
                        CHECK (role IN ('primary', 'spouse', 'parent', 'child', 'other')),
    relation_label  VARCHAR(128),
    avatar_url      VARCHAR(1024),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_family_members_owner_user_id ON family_members (owner_user_id);

-- +goose Down
DROP TABLE IF EXISTS family_members;
