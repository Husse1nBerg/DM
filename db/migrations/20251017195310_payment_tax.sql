-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS tax_configurations CASCADE;

CREATE TABLE tax_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    
    -- Convenience Fee Configuration
    convenience_fee NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    convenience_fee_type VARCHAR(20) NOT NULL DEFAULT 'percentage' CHECK (convenience_fee_type IN ('percentage', 'fixed')),
    convenience_fee_enabled BOOLEAN NOT NULL DEFAULT false,
    convenience_fee_description TEXT,
    
    -- Surcharge Configuration
    surcharge NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    surcharge_type VARCHAR(20) NOT NULL DEFAULT 'percentage' CHECK (surcharge_type IN ('percentage', 'fixed')),
    surcharge_enabled BOOLEAN NOT NULL DEFAULT false,
    surcharge_description TEXT,
    
    -- Payment Type Configuration
    -- CC: Credit Card, DB: Debit Card, CK: Check, ACH: ACH Transfer
    payment_type VARCHAR(10) NOT NULL CHECK (payment_type IN ('CC', 'DB', 'CK', 'ACH')),
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX tax_configurations_marina_id_idx ON tax_configurations(marina_id);
CREATE INDEX tax_configurations_payment_type_idx ON tax_configurations(payment_type);
CREATE INDEX tax_configurations_created_at_idx ON tax_configurations(created_at DESC);
-- Ensure only one configuration per marina per payment type
CREATE UNIQUE INDEX unique_marina_payment_type ON tax_configurations(marina_id, payment_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tax_configurations;
-- +goose StatementEnd
