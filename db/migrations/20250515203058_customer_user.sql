-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN customer_id TEXT;
ALTER TABLE users ADD COLUMN is_customer BOOLEAN DEFAULT FALSE;
ALTER TABLE user_marinas ADD COLUMN customer_id TEXT;
ALTER TABLE roles ADD COLUMN is_customer_role BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN customer_id;
ALTER TABLE users DROP COLUMN is_customer;
ALTER TABLE user_marinas DROP COLUMN customer_id;
ALTER TABLE roles DROP COLUMN is_customer_role;
-- +goose StatementEnd
