-- +goose Up
-- +goose StatementBegin
ALTER TABLE customer_settings DROP COLUMN customer_user_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customer_settings ADD COLUMN customer_user_id UUID REFERENCES users(id) ON DELETE SET NULL;
-- +goose StatementEnd
