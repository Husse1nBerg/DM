-- +goose Up
-- +goose StatementBegin
ALTER TABLE marinas ADD COLUMN internal_announcement TEXT;
ALTER TABLE marinas ADD COLUMN external_announcement TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN IF EXISTS internal_announcement;
ALTER TABLE marinas DROP COLUMN IF EXISTS external_announcement;
-- +goose StatementEnd
