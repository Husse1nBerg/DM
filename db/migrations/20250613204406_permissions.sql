-- +goose Up
-- +goose StatementBegin
ALTER TABLE marinas ADD COLUMN modules JSONB;
ALTER TABLE users DROP COLUMN IF EXISTS permissions;
ALTER TABLE users DROP COLUMN IF EXISTS modules;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN IF EXISTS modules;
ALTER TABLE users ADD COLUMN permissions JSONB;
ALTER TABLE users ADD COLUMN modules JSONB;
-- +goose StatementEnd
