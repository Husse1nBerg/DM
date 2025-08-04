-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS roles_name_marina_id_key ON roles(name, marina_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS roles_name_marina_id_key;
ALTER TABLE roles ADD CONSTRAINT roles_name_key UNIQUE (name);
-- +goose StatementEnd

