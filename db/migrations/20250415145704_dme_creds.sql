-- +goose Up
-- +goose StatementBegin
CREATE TABLE dme_credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255),
    is_old_api BOOLEAN DEFAULT FALSE,
    access_token TEXT,
    refresh_token TEXT,
    expiry_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_org_dme_creds UNIQUE (organization_id)
);
CREATE TABLE dme_sysids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    marina_id UUID REFERENCES marinas(id) ON DELETE
    SET NULL,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        system_id VARCHAR(255) NOT NULL UNIQUE,
        is_active BOOLEAN DEFAULT TRUE,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP,
        deleted_at TIMESTAMP
);
ALTER TABLE marinas
ADD COLUMN system_id VARCHAR(255);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN IF EXISTS system_id;
DROP TABLE IF EXISTS dme_sysids;
DROP TABLE IF EXISTS dme_credentials;
-- +goose StatementEnd