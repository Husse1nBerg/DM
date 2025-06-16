-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    customer_id TEXT NOT NULL,
    type TEXT NOT NULL, -- email, sms, internal
    direction TEXT NOT NULL, -- to_customer, to_marina, internal
    body TEXT NOT NULL,
    sender TEXT NOT NULL, -- User Name
    recipient TEXT NOT NULL, -- Receiver Name or Internal
    contact TEXT NOT NULL, -- the phone number or the email address or internal
    status TEXT NOT NULL, -- pending, sent, failed
    pinned BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE messages;
-- +goose StatementEnd
