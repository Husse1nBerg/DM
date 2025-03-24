-- +goose Up
-- +goose StatementBegin
CREATE TABLE addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    street TEXT,
    city VARCHAR(255),
    state VARCHAR(255),
    postal_code VARCHAR(50),
    country VARCHAR(255),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
ALTER TABLE organizations
ADD COLUMN address_id UUID REFERENCES addresses(id) ON DELETE
SET NULL;
ALTER TABLE marinas
ADD COLUMN address_id UUID REFERENCES addresses(id) ON DELETE
SET NULL;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN IF EXISTS address_id;
ALTER TABLE organizations DROP COLUMN IF EXISTS address_id;
DROP TABLE IF EXISTS addresses;
-- +goose StatementEnd