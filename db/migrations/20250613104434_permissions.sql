-- +goose Up
-- +goose StatementBegin
ALTER TABLE marinas ADD COLUMN modules JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN IF EXISTS modules;
-- +goose StatementEnd
