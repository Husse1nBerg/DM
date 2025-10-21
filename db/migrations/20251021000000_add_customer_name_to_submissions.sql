-- +goose Up
-- +goose StatementBegin
ALTER TABLE esign_submissions ADD COLUMN customer_name TEXT;
CREATE INDEX idx_esign_submissions_customer_name ON esign_submissions(customer_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_esign_submissions_customer_name;
ALTER TABLE esign_submissions DROP COLUMN customer_name;
-- +goose StatementEnd
