-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments
	ADD COLUMN IF NOT EXISTS adyen_payment_response TEXT,
	ADD COLUMN IF NOT EXISTS adyen_payment_payload TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments
	DROP COLUMN IF EXISTS adyen_payment_response,
	DROP COLUMN IF EXISTS adyen_payment_payload;
-- +goose StatementEnd


