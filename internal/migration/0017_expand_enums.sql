-- +goose Up

-- Expand reminder category check constraint
ALTER TABLE reminders DROP CONSTRAINT IF EXISTS reminders_category_check;
ALTER TABLE reminders ADD CONSTRAINT reminders_category_check
    CHECK (category IN (
        'bill', 'vehicle', 'document', 'custom',
        'health', 'insurance', 'subscription', 'tax',
        'personal', 'work', 'family'
    ));

-- Expand document vault_category check constraint
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_vault_category_check;
ALTER TABLE documents ADD CONSTRAINT documents_vault_category_check
    CHECK (vault_category IN (
        'identity', 'vehicles', 'insurance', 'financials', 'other',
        'property', 'financial', 'health', 'education', 'legal'
    ));

-- +goose Down
ALTER TABLE reminders DROP CONSTRAINT IF EXISTS reminders_category_check;
ALTER TABLE reminders ADD CONSTRAINT reminders_category_check
    CHECK (category IN ('bill', 'vehicle', 'document', 'custom'));

ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_vault_category_check;
ALTER TABLE documents ADD CONSTRAINT documents_vault_category_check
    CHECK (vault_category IN ('vehicles', 'identity', 'insurance', 'financials', 'other'));
