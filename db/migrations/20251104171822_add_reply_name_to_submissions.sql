-- +goose Up
-- +goose StatementBegin
ALTER TABLE esign_submissions
    ADD COLUMN reply_name VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE esign_submissions
    DROP COLUMN IF EXISTS reply_name;
-- +goose StatementEnd
