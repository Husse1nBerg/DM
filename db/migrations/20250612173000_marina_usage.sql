-- +goose Up
-- +goose StatementBegin
ALTER TABLE marinas ADD COLUMN storage_usage BIGINT DEFAULT 0;
ALTER TABLE marinas ADD COLUMN email_usage SMALLINT DEFAULT 0;
ALTER TABLE marinas ADD COLUMN text_usage SMALLINT DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN storage_usage;
ALTER TABLE marinas DROP COLUMN email_usage;
ALTER TABLE marinas DROP COLUMN text_usage;
-- +goose StatementEnd
