-- +goose Up
-- +goose StatementBegin
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS is_cp_contact BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE contacts DROP COLUMN IF EXISTS is_cp_contact;
-- +goose StatementEnd
