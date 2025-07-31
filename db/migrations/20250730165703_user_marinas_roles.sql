-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_marinas ADD COLUMN role_id UUID REFERENCES roles(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_marinas DROP COLUMN role_id;
-- +goose StatementEnd
