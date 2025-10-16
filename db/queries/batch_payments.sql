-- name: CreateBatchPayment :one
INSERT INTO batch_payments (
    organization_id,
    marina_id,
    location_code,
    batch_id,
    post_batch,
    total_amount,
    receipt_count,
    status,
    submitted_by,
    submitted_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: CreateBatchPaymentReceipt :one
INSERT INTO batch_payment_receipts (
    batch_payment_id,
    customer_id,
    invoice_id,
    amount,
    payment_method,
    reference,
    description,
    payment_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateBatchPaymentStatus :one
UPDATE batch_payments
SET 
    status = $2,
    post_result = $3,
    reference_ids = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetBatchPaymentByID :one
SELECT * FROM batch_payments
WHERE id = $1;

-- name: GetBatchPaymentByBatchID :one
SELECT * FROM batch_payments
WHERE batch_id = $1;

-- name: ListBatchPaymentsByMarina :many
SELECT * FROM batch_payments
WHERE marina_id = $1
ORDER BY submitted_at DESC
LIMIT $2 OFFSET $3;

-- name: ListBatchPaymentReceipts :many
SELECT * FROM batch_payment_receipts
WHERE batch_payment_id = $1
ORDER BY created_at DESC;

