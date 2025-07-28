-- +goose Up
-- +goose StatementBegin
ALTER TABLE esign_templates ADD COLUMN json_data JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE esign_templates DROP COLUMN json_data;
-- +goose StatementEnd
