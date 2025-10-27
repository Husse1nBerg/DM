-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS payments CASCADE;

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Entity relationship (invoice, boat, customer, work_order, etc.)
    entity_type VARCHAR(50),
    entity_id TEXT,
    
    -- Payment details
    amount NUMERIC(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50),
    reference_number VARCHAR(100) NOT NULL,
    
    -- Status tracking (2-step process)
    -- pending -> authorized -> completed (or failed at any step)
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    authorization_status VARCHAR(50),  -- tracks Adyen webhook status
    batch_status VARCHAR(50),          -- tracks DME batch submission status
    
    -- External references
    adyen_psp_reference VARCHAR(100),
    adyen_session_id VARCHAR(100),
    batch_id VARCHAR(100),
    batch_payment_id UUID REFERENCES batch_payments(id),
    
    -- Payment timestamps
    payment_date TIMESTAMPTZ,
    authorized_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    
    -- Additional metadata
    customer_id VARCHAR(100),
    location_code VARCHAR(50),
    transaction_id VARCHAR(100),
    auth_code VARCHAR(100),
    
    -- Payload tracking (for debugging and audit)
    adyen_webhook_payload TEXT,
    dme_batch_request TEXT,
    dme_batch_response TEXT,
    
    -- Error handling
    error_message TEXT,
    error_code VARCHAR(50),
    
    -- Notes and comments
    internal_notes TEXT,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX payments_marina_id_idx ON payments(marina_id);
CREATE INDEX payments_organization_id_idx ON payments(organization_id);
CREATE INDEX payments_entity_type_entity_id_idx ON payments(entity_type, entity_id);
CREATE INDEX payments_status_idx ON payments(status);
CREATE INDEX payments_reference_number_idx ON payments(reference_number);
CREATE INDEX payments_adyen_psp_reference_idx ON payments(adyen_psp_reference);
CREATE INDEX payments_batch_id_idx ON payments(batch_id);
CREATE INDEX payments_payment_date_idx ON payments(payment_date);
CREATE INDEX payments_created_at_idx ON payments(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payments;
-- +goose StatementEnd
