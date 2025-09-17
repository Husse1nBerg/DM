-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    marina_id UUID REFERENCES marinas(id) ON DELETE CASCADE,
    api_key TEXT NOT NULL,
    merchant_account VARCHAR(255) NOT NULL,
    store_id VARCHAR(255),
    is_test BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(50),
    payment_type VARCHAR(50) NOT NULL,
    reference VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB,
    adyen_payment_id VARCHAR(255),
    adyen_merchant_reference VARCHAR(255),
    adyen_psp_reference VARCHAR(255),
    adyen_payment_method_details JSONB,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE payment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_payment_credentials_organization_id ON payment_credentials(organization_id);
CREATE INDEX idx_payment_credentials_marina_id ON payment_credentials(marina_id);
CREATE INDEX idx_payments_organization_id ON payments(organization_id);
CREATE INDEX idx_payments_marina_id ON payments(marina_id);
CREATE INDEX idx_payments_customer_id ON payments(customer_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payment_events_payment_id ON payment_events(payment_id);
CREATE INDEX idx_payment_events_event_type ON payment_events(event_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payment_events_event_type;
DROP INDEX IF EXISTS idx_payment_events_payment_id;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_customer_id;
DROP INDEX IF EXISTS idx_payments_marina_id;
DROP INDEX IF EXISTS idx_payments_organization_id;
DROP INDEX IF EXISTS idx_payment_credentials_marina_id;
DROP INDEX IF EXISTS idx_payment_credentials_organization_id;

DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS payment_credentials;
-- +goose StatementEnd
