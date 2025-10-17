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
    
    -- Tax Configuration (optional, for general tax)
    tax_rate NUMERIC(5, 2) DEFAULT 0.00,
    tax_enabled BOOLEAN NOT NULL DEFAULT false,
    tax_description TEXT,
    
    -- Configuration metadata
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    created_by UUID
);

-- Indexes for performance
CREATE INDEX tax_configurations_marina_id_idx ON tax_configurations(marina_id);
CREATE INDEX tax_configurations_active_idx ON tax_configurations(is_active) WHERE is_active = true;
-- Ensure only one active configuration per marina at a time (partial unique index)
CREATE UNIQUE INDEX unique_active_marina_config ON tax_configurations(marina_id) WHERE is_active = true;
CREATE INDEX tax_configurations_created_at_idx ON tax_configurations(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tax_configurations;
-- +goose StatementEnd
