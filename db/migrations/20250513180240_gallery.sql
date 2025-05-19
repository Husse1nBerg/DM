-- +goose Up
-- +goose StatementBegin
CREATE TABLE vessel_gallery (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    customer_id TEXT NOT NULL,
    vessel_id TEXT NOT NULL,
    image_url TEXT NOT NULL,
    description TEXT,
    main BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_vessel_gallery_marina_id ON vessel_gallery(marina_id);
CREATE INDEX idx_vessel_gallery_customer_id ON vessel_gallery(customer_id);
CREATE INDEX idx_vessel_gallery_vessel_id ON vessel_gallery(vessel_id);

CREATE TABLE marina_gallery (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_marina_gallery_marina_id ON marina_gallery(marina_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_vessel_gallery_marina_id;
DROP INDEX IF EXISTS idx_vessel_gallery_customer_id;
DROP INDEX IF EXISTS idx_vessel_gallery_vessel_id;
DROP TABLE IF EXISTS vessel_gallery;

DROP INDEX IF EXISTS idx_marina_gallery_marina_id;
DROP TABLE IF EXISTS marina_gallery;
-- +goose StatementEnd
