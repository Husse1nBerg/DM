-- +goose Up
-- +goose StatementBegin
ALTER TABLE marinas ADD COLUMN ai_form_detection_usage SMALLINT DEFAULT 0;
ALTER TABLE marinas ADD COLUMN ai_compose_message_usage SMALLINT DEFAULT 0;
ALTER TABLE marina_usage_history ADD COLUMN ai_form_detection_usage SMALLINT DEFAULT 0;
ALTER TABLE marina_usage_history ADD COLUMN ai_compose_message_usage SMALLINT DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE marinas DROP COLUMN ai_form_detection_usage;
ALTER TABLE marinas DROP COLUMN ai_compose_message_usage;
ALTER TABLE marina_usage_history DROP COLUMN ai_form_detection_usage;
ALTER TABLE marina_usage_history DROP COLUMN ai_compose_message_usage;
-- +goose StatementEnd
