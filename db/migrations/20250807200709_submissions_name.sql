-- +goose Up
-- +goose StatementBegin
ALTER TABLE esign_submissions ADD COLUMN name TEXT;
ALTER TABLE esign_submissions ADD COLUMN attachment_required BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE esign_submissions DROP COLUMN name;
ALTER TABLE esign_submissions DROP COLUMN attachment_required;
-- +goose StatementEnd
