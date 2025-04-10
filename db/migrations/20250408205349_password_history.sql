-- +goose Up
-- +goose StatementBegin
CREATE TABLE password_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_password UNIQUE(user_id, password_hash)
);

CREATE INDEX idx_password_history_user_id ON password_history(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_password_history_user_id;
DROP TABLE IF EXISTS password_history;
-- +goose StatementEnd
