-- +goose Up
-- +goose StatementBegin
CREATE TABLE esign_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    marina_id UUID REFERENCES marinas(id) ON DELETE CASCADE,
    document_id UUID REFERENCES esign_documents(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, signed, questions, sent
    blob_url TEXT NOT NULL,
    blob_metadata JSONB,
    customer_id VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_esign_submissions_organization_id ON esign_submissions(organization_id);
CREATE INDEX idx_esign_submissions_marina_id ON esign_submissions(marina_id);
CREATE INDEX idx_esign_submissions_document_id ON esign_submissions(document_id);
CREATE INDEX idx_esign_submissions_status ON esign_submissions(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_esign_submissions_status;
DROP INDEX IF EXISTS idx_esign_submissions_document_id;
DROP INDEX IF EXISTS idx_esign_submissions_marina_id;
DROP INDEX IF EXISTS idx_esign_submissions_organization_id;

DROP TABLE IF EXISTS esign_submissions;
-- +goose StatementEnd

