-- +goose Up
-- +goose StatementBegin
CREATE TABLE batch_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    marina_id UUID NOT NULL REFERENCES marinas(id) ON DELETE CASCADE,
    location_code VARCHAR(50) NOT NULL,
    batch_id VARCHAR(100) NOT NULL,
    post_batch BOOLEAN NOT NULL DEFAULT false,
    total_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    receipt_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    post_result TEXT,
    reference_ids TEXT[], -- Array of reference IDs returned from DME
    submitted_by VARCHAR(100) NOT NULL,
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE batch_payment_receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_payment_id UUID NOT NULL REFERENCES batch_payments(id) ON DELETE CASCADE,
    customer_id VARCHAR(50) NOT NULL,
    invoice_id VARCHAR(50) NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    reference VARCHAR(100) NOT NULL,
    description TEXT,
    payment_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_batch_payments_organization_marina ON batch_payments(organization_id, marina_id);
CREATE INDEX idx_batch_payments_batch_id ON batch_payments(batch_id);
CREATE INDEX idx_batch_payments_status ON batch_payments(status);
CREATE INDEX idx_batch_payments_submitted_at ON batch_payments(submitted_at);
CREATE INDEX idx_batch_payment_receipts_batch_id ON batch_payment_receipts(batch_payment_id);
CREATE INDEX idx_batch_payment_receipts_customer_id ON batch_payment_receipts(customer_id);
CREATE INDEX idx_batch_payment_receipts_invoice_id ON batch_payment_receipts(invoice_id);

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_batch_payments_updated_at BEFORE UPDATE ON batch_payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_batch_payment_receipts_updated_at BEFORE UPDATE ON batch_payment_receipts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_batch_payment_receipts_updated_at ON batch_payment_receipts;
DROP TRIGGER IF EXISTS update_batch_payments_updated_at ON batch_payments;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_batch_payment_receipts_invoice_id;
DROP INDEX IF EXISTS idx_batch_payment_receipts_customer_id;
DROP INDEX IF EXISTS idx_batch_payment_receipts_batch_id;
DROP INDEX IF EXISTS idx_batch_payments_submitted_at;
DROP INDEX IF EXISTS idx_batch_payments_status;
DROP INDEX IF EXISTS idx_batch_payments_batch_id;
DROP INDEX IF EXISTS idx_batch_payments_organization_marina;

DROP TABLE IF EXISTS batch_payment_receipts;
DROP TABLE IF EXISTS batch_payments;
-- +goose StatementEnd