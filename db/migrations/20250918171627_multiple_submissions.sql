-- +goose Up
-- +goose StatementBegin

-- Create esign_submission_signers table to support multiple signers per submission
CREATE TABLE esign_submission_signers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES esign_submissions(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    sign_order INTEGER NOT NULL DEFAULT 1, -- Order in which signers should sign
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, signed, declined, expired
    signed_at TIMESTAMP,
    declined_at TIMESTAMP,
    declined_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_esign_submission_signers_submission_id ON esign_submission_signers(submission_id);
CREATE INDEX idx_esign_submission_signers_email ON esign_submission_signers(email);
CREATE INDEX idx_esign_submission_signers_status ON esign_submission_signers(status);
CREATE INDEX idx_esign_submission_signers_sign_order ON esign_submission_signers(sign_order);

-- Add a unique constraint to prevent duplicate signers per submission
CREATE UNIQUE INDEX idx_esign_submission_signers_unique ON esign_submission_signers(submission_id, email) WHERE deleted_at IS NULL;

-- Add a column to track if this is a multiple signature submission
ALTER TABLE esign_submissions ADD COLUMN is_multiple_signature BOOLEAN NOT NULL DEFAULT FALSE;

-- Add index for the new column
CREATE INDEX idx_esign_submissions_is_multiple_signature ON esign_submissions(is_multiple_signature);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the new column from esign_submissions
DROP INDEX IF EXISTS idx_esign_submissions_is_multiple_signature;
ALTER TABLE esign_submissions DROP COLUMN IF EXISTS is_multiple_signature;

-- Drop the esign_submission_signers table and its indexes
DROP INDEX IF EXISTS idx_esign_submission_signers_unique;
DROP INDEX IF EXISTS idx_esign_submission_signers_sign_order;
DROP INDEX IF EXISTS idx_esign_submission_signers_status;
DROP INDEX IF EXISTS idx_esign_submission_signers_email;
DROP INDEX IF EXISTS idx_esign_submission_signers_submission_id;

DROP TABLE IF EXISTS esign_submission_signers;

-- +goose StatementEnd
