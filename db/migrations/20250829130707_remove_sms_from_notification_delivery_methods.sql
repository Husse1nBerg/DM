-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

-- Update existing records that have 'sms' delivery method to use 'system' instead
UPDATE notification_preferences 
SET delivery_method = 'system' 
WHERE delivery_method = 'sms';

-- Update existing records that have 'push' delivery method to use 'system' instead
UPDATE notification_preferences 
SET delivery_method = 'system' 
WHERE delivery_method = 'push';

-- Add a check constraint to prevent 'sms' and 'push' from being used in the future
ALTER TABLE notification_preferences 
ADD CONSTRAINT chk_delivery_method 
CHECK (delivery_method IN ('system', 'email', 'all'));

-- Update the default value for new records
ALTER TABLE notification_preferences 
ALTER COLUMN delivery_method SET DEFAULT 'system';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

-- Remove the check constraint
ALTER TABLE notification_preferences 
DROP CONSTRAINT IF EXISTS chk_delivery_method;

-- Revert the default value
ALTER TABLE notification_preferences 
ALTER COLUMN delivery_method SET DEFAULT 'system';

-- +goose StatementEnd
