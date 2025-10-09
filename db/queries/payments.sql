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

-- name: GetBatchPayment :one
SELECT * FROM batch_payments
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetBatchPaymentByBatchID :one
SELECT * FROM batch_payments
WHERE batch_id = $1 AND deleted_at IS NULL;

-- name: UpdateBatchPaymentStatus :one
UPDATE batch_payments
SET 
    status = $2,
    post_result = COALESCE($3, post_result),
    reference_ids = COALESCE($4, reference_ids),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListBatchPayments :many
SELECT * FROM batch_payments
WHERE organization_id = $1
    AND marina_id = $2
    AND deleted_at IS NULL
ORDER BY submitted_at DESC
LIMIT $3 OFFSET $4;

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

-- name: GetBatchPaymentReceipts :many
SELECT * FROM batch_payment_receipts
WHERE batch_payment_id = $1
ORDER BY created_at ASC;