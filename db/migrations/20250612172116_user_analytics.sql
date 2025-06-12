-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN user_analytics BOOLEAN DEFAULT true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN user_analytics;
-- +goose StatementEnd
