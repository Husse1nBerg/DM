-- +goose Up
-- +goose StatementBegin
CREATE TABLE document_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    monthly_price NUMERIC(10, 2),
    document_limit INTEGER,
    user_limit TEXT DEFAULT 'Unlimited Users',
    is_most_popular BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

ALTER TABLE marinas ADD COLUMN document_plan_id UUID REFERENCES document_plans(id);
ALTER TABLE marinas ADD COLUMN document_usage BIGINT DEFAULT 0;
ALTER TABLE marina_usage_history ADD COLUMN document_usage BIGINT DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN document_plan_id;
ALTER TABLE marinas DROP COLUMN document_usage;
ALTER TABLE marina_usage_history DROP COLUMN document_usage;
DROP TABLE IF EXISTS document_plans;
-- +goose StatementEnd