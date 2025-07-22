-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles ADD COLUMN marina_id UUID REFERENCES marinas(id) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles DROP COLUMN marina_id;
-- +goose StatementEnd
