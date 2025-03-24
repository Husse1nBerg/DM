-- +goose Up
-- +goose StatementBegin
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified TIMESTAMP,
    phone VARCHAR(255),
    title VARCHAR(255),
    image TEXT,
    password_hash TEXT NOT NULL,
    last_login TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    last_password_reset TIMESTAMP,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    marina_id UUID REFERENCES marinas(id) ON DELETE
    SET NULL,
        role_id UUID NOT NULL REFERENCES roles(id),
        is_superuser BOOLEAN DEFAULT FALSE,
        is_active BOOLEAN DEFAULT TRUE,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP,
        deleted_at TIMESTAMP
);
CREATE TABLE user_marinas (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, marina_id)
);
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_marina_id ON users(marina_id);
CREATE INDEX idx_user_marinas_marina_id ON user_marinas(marina_id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_marinas_marina_id;
DROP INDEX IF EXISTS idx_users_marina_id;
DROP INDEX IF EXISTS idx_users_organization_id;
DROP TABLE IF EXISTS user_marinas;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd