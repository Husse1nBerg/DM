-- name: GetPaymentCredentials :one
SELECT * FROM payment_credentials
WHERE organization_id = $1 AND marina_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: CreatePayment :one
INSERT INTO payments (
    organization_id,
    marina_id,
    customer_id,
    amount,
    currency,
    status,
    payment_method,
    payment_type,
    reference,
    description,
    metadata,
    adyen_payment_id,
    adyen_merchant_reference,
    adyen_psp_reference,
    adyen_payment_method_details
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetPayment :one
SELECT * FROM payments
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPaymentByReference :one
SELECT * FROM payments
WHERE reference = $1 AND deleted_at IS NULL;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET 
    status = $2,
    adyen_psp_reference = COALESCE($3, adyen_psp_reference),
    adyen_payment_method_details = COALESCE($4, adyen_payment_method_details),
    error_message = COALESCE($5, error_message),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListPayments :many
SELECT * FROM payments
WHERE organization_id = $1
    AND marina_id = $2
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CreatePaymentEvent :one
INSERT INTO payment_events (
    payment_id,
    event_type,
    event_data
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetPaymentEvents :many
SELECT * FROM payment_events
WHERE payment_id = $1
ORDER BY created_at DESC;