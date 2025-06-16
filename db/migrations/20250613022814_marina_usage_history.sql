-- +goose Up
-- +goose StatementBegin
CREATE TABLE marina_usage_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    marina_id UUID NOT NULL REFERENCES marinas(id),
    storage_usage BIGINT NOT NULL,
    email_usage SMALLINT NOT NULL,
    text_usage SMALLINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    month_date DATE NOT NULL
);

-- Create an index on marina_id and created_at for efficient querying
CREATE INDEX idx_marina_usage_history_marina_id_created_at ON marina_usage_history(marina_id, created_at);

-- Add a unique constraint to ensure only one record per marina per month
CREATE UNIQUE INDEX idx_marina_usage_history_marina_month ON marina_usage_history(marina_id, month_date);

-- Create a function to update month_date
CREATE OR REPLACE FUNCTION update_marina_usage_history_month_date()
RETURNS TRIGGER AS $$
BEGIN
    NEW.month_date := DATE_TRUNC('month', NEW.created_at)::date;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to automatically update month_date
CREATE TRIGGER set_marina_usage_history_month_date
    BEFORE INSERT OR UPDATE ON marina_usage_history
    FOR EACH ROW
    EXECUTE FUNCTION update_marina_usage_history_month_date();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS set_marina_usage_history_month_date ON marina_usage_history;
DROP FUNCTION IF EXISTS update_marina_usage_history_month_date();
DROP TABLE IF EXISTS marina_usage_history;
-- +goose StatementEnd
