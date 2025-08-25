-- +goose Up
-- +goose StatementBegin
ALTER TABLE esign_submissions
    ADD COLUMN reply_to VARCHAR(255),
    ADD COLUMN custom_message TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE esign_submissions
    DROP COLUMN IF EXISTS reply_to,
    DROP COLUMN IF EXISTS custom_message;
-- +goose StatementEnd
