-- +goose Up
-- +goose StatementBegin
CREATE TABLE criteria (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    marina_id UUID NOT NULL REFERENCES marinas(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    criteria JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_criteria_marina_id ON criteria(marina_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_criteria_marina_id;
DROP TABLE IF EXISTS criteria;
-- +goose StatementEnd
