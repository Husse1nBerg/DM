-- +goose Up
-- +goose StatementBegin
ALTER TABLE marinas ADD COLUMN storage_usage BIGINT DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN storage_usage;
-- +goose StatementEnd
