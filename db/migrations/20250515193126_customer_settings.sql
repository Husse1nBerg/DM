-- +goose Up
-- +goose StatementBegin
CREATE TABLE customer_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    customer_id TEXT NOT NULL,
    enable_portal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE (marina_id, customer_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE customer_settings;
-- +goose StatementEnd
