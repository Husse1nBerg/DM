-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    image TEXT,
    website TEXT,
    country VARCHAR(255),
    phone VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    is_test BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE TABLE marinas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    location TEXT,
    phone VARCHAR(255),
    country VARCHAR(255),
    currency VARCHAR(255),
    working_hours JSONB,
    website TEXT,
    image TEXT,
    max_users INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT TRUE,
    is_test BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP TABLE IF EXISTS marinas;
DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd