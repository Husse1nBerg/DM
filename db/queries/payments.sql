-- name: CreatePayment :one
INSERT INTO payments (
    marina_id,
    organization_id,
    entity_type,
    entity_id,
    amount,
    currency,
    payment_method,
    reference_number,
    status,
    authorization_status,
    adyen_session_id,
    customer_id,
    location_code,
    payment_date,
    internal_notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: UpdatePaymentAuthorized :one
UPDATE payments
SET 
    status = 'authorized',
    authorization_status = $2,
    adyen_psp_reference = $3,
    auth_code = $4,
    transaction_id = $5,
    authorized_at = $6,
    adyen_webhook_payload = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentCompleted :one
UPDATE payments
SET 
    status = 'completed',
    batch_status = 'submitted',
    batch_id = $2,
    batch_payment_id = $3,
    dme_batch_request = $4,
    dme_batch_response = $5,
    completed_at = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentFailed :one
UPDATE payments
SET 
    status = 'failed',
    error_message = $2,
    error_code = $3,
    failed_at = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentBatchStatus :one
UPDATE payments
SET 
    batch_status = $2,
    batch_id = $3,
    batch_payment_id = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetPaymentByID :one
SELECT * FROM payments
WHERE id = $1;

-- name: GetPaymentByReferenceNumber :one
SELECT * FROM payments
WHERE reference_number = $1;

-- name: GetPaymentByAdyenPSPReference :one
SELECT * FROM payments
WHERE adyen_psp_reference = $1;

-- name: ListPaymentsByMarina :many
SELECT * FROM payments
WHERE marina_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPaymentsByEntity :many
SELECT * FROM payments
WHERE marina_id = $1
AND entity_type = $2
AND entity_id = $3
ORDER BY created_at DESC;

-- name: ListPaymentsByStatus :many
SELECT * FROM payments
WHERE marina_id = $1
AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListPaymentsByDateRange :many
SELECT * FROM payments
WHERE marina_id = $1
AND payment_date >= $2
AND payment_date <= $3
ORDER BY payment_date DESC;

-- name: GetPaymentStats :one
SELECT 
    COUNT(*) as total_count,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_count,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_count,
    SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_count,
    SUM(CASE WHEN status = 'authorized' THEN 1 ELSE 0 END) as authorized_count,
    COALESCE(SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END), 0) as total_completed_amount
FROM payments
WHERE marina_id = $1
AND payment_date >= $2
AND payment_date <= $3;

-- name: CountPaymentsByMarina :one
SELECT COUNT(*) FROM payments
WHERE marina_id = $1;

-- name: ListPaymentsWithFilters :many
SELECT * FROM payments
WHERE marina_id = $1
AND ($2::text IS NULL OR status = $2)
AND ($3::text IS NULL OR entity_type = $3)
AND ($4::text IS NULL OR entity_id = $4)
AND ($5::text IS NULL OR customer_id = $5)
AND ($6::timestamptz IS NULL OR payment_date >= $6)
AND ($7::timestamptz IS NULL OR payment_date <= $7)
ORDER BY created_at DESC
LIMIT $8 OFFSET $9;
