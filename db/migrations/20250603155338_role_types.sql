-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles ADD COLUMN type TEXT NOT NULL DEFAULT 'marina';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles DROP COLUMN type;
-- +goose StatementEnd
