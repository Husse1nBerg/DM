-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments
ADD COLUMN first_name VARCHAR(100),
ADD COLUMN last_name VARCHAR(100),
ADD COLUMN primary_email VARCHAR(255),
ADD COLUMN card_summary VARCHAR(10);

CREATE INDEX payments_first_name_idx ON payments(first_name);
CREATE INDEX payments_last_name_idx ON payments(last_name);
CREATE INDEX payments_primary_email_idx ON payments(primary_email);
CREATE INDEX payments_card_summary_idx ON payments(card_summary);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS payments_card_summary_idx;
DROP INDEX IF EXISTS payments_primary_email_idx;
DROP INDEX IF EXISTS payments_last_name_idx;
DROP INDEX IF EXISTS payments_first_name_idx;

ALTER TABLE payments
DROP COLUMN IF EXISTS card_summary,
DROP COLUMN IF EXISTS primary_email,
DROP COLUMN IF EXISTS last_name,
DROP COLUMN IF EXISTS first_name;
-- +goose StatementEnd

