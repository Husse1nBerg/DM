-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN modules JSONB,
    ADD COLUMN permissions JSONB;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS permissions,
    DROP COLUMN IF EXISTS modules;
-- +goose StatementEnd