-- +goose Up
-- +goose StatementBegin
ALTER TABLE marina_gallery ADD COLUMN public BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE vessel_gallery ADD COLUMN public BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE documents ADD COLUMN public BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE dme_attachment_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    s3_path TEXT NOT NULL UNIQUE,
    public BOOLEAN NOT NULL DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marina_gallery DROP COLUMN public;
ALTER TABLE vessel_gallery DROP COLUMN public;
ALTER TABLE documents DROP COLUMN public;
DROP TABLE dme_attachment_metadata;
-- +goose StatementEnd