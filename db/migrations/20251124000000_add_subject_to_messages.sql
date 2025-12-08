-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages
    ADD COLUMN subject TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages
    DROP COLUMN IF EXISTS subject;
-- +goose StatementEnd

