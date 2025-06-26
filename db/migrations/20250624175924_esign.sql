-- +goose Up
-- +goose StatementBegin
CREATE TABLE esign_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    marina_id UUID REFERENCES marinas(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) NOT NULL DEFAULT 'template',
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- draft, active, archived
    blob_url TEXT NOT NULL,
    blob_metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE esign_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID REFERENCES esign_templates(id) ON DELETE SET NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL DEFAULT 'document',
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- draft, signed, questions, sent
    blob_url TEXT NOT NULL,
    blob_metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_esign_templates_organization_id ON esign_templates(organization_id);
CREATE INDEX idx_esign_templates_marina_id ON esign_templates(marina_id);
CREATE INDEX idx_esign_templates_type ON esign_templates(type);
CREATE INDEX idx_esign_templates_status ON esign_templates(status);

CREATE INDEX idx_esign_documents_template_id ON esign_documents(template_id);
CREATE INDEX idx_esign_documents_organization_id ON esign_documents(organization_id);
CREATE INDEX idx_esign_documents_marina_id ON esign_documents(marina_id);
CREATE INDEX idx_esign_documents_type ON esign_documents(type);
CREATE INDEX idx_esign_documents_status ON esign_documents(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_esign_documents_status;
DROP INDEX IF EXISTS idx_esign_documents_type;
DROP INDEX IF EXISTS idx_esign_documents_marina_id;
DROP INDEX IF EXISTS idx_esign_documents_organization_id;
DROP INDEX IF EXISTS idx_esign_documents_template_id;

DROP INDEX IF EXISTS idx_esign_templates_status;
DROP INDEX IF EXISTS idx_esign_templates_type;
DROP INDEX IF EXISTS idx_esign_templates_marina_id;
DROP INDEX IF EXISTS idx_esign_templates_organization_id;

DROP TABLE IF EXISTS esign_documents;
DROP TABLE IF EXISTS esign_templates;
-- +goose StatementEnd
