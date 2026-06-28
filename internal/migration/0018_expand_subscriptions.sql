-- +goose Up

-- Expand subscription category check constraint
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_category_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_category_check
    CHECK (category IN (
        'entertainment', 'productivity', 'bill', 'other',
        'health', 'education', 'finance', 'utilities'
    ));

-- Expand subscription billing_cycle check constraint
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_billing_cycle_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_billing_cycle_check
    CHECK (billing_cycle IN ('weekly', 'monthly', 'quarterly', 'yearly', 'custom'));

-- +goose Down
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_category_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_category_check
    CHECK (category IN ('entertainment', 'productivity', 'bill', 'other'));

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_billing_cycle_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_billing_cycle_check
    CHECK (billing_cycle IN ('monthly', 'yearly', 'custom'));
