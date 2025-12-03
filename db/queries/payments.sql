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
    internal_notes,
    adyen_payment_payload
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
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
    adyen_payment_response = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListPaymentsFilteredSortedAsc :many
SELECT *
FROM payments
WHERE marina_id = $1
  AND (
    $2 = '' OR
    reference_number ILIKE '%' || $2 || '%' OR
    adyen_psp_reference ILIKE '%' || $2 || '%' OR
    customer_id ILIKE '%' || $2 || '%' OR
    COALESCE(payment_method, '') ILIKE '%' || $2 || '%'
  )
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR customer_id = $4)
  AND ($5 = '' OR entity_type = $5)
  AND ($6 = '' OR entity_id = $6)
  AND ($7 = '' OR payment_method = $7)
  AND ($8 = '' OR currency = $8)
  AND ($9::timestamptz IS NULL OR payment_date >= $9::timestamptz)
  AND ($10::timestamptz IS NULL OR payment_date <= $10::timestamptz)
ORDER BY
  (CASE WHEN $11 = 'reference_number' THEN reference_number END) ASC,
  (CASE WHEN $11 = 'payment_method' THEN payment_method END) ASC,
  (CASE WHEN $11 = 'currency' THEN currency END) ASC,
  (CASE WHEN $11 = 'payment_date' THEN payment_date END) ASC,
  (CASE WHEN $11 = 'created_at' THEN created_at END) ASC,
  (CASE WHEN $11 = 'amount' THEN amount END) ASC,
  (CASE WHEN $11 = 'status' THEN
    CASE status
      WHEN 'pending' THEN 1
      WHEN 'authorized' THEN 2
      WHEN 'completed' THEN 3
      WHEN 'failed' THEN 4
      ELSE 5
    END
  END) ASC
LIMIT $12 OFFSET $13;

-- name: ListPaymentsFilteredSortedDesc :many
SELECT *
FROM payments
WHERE marina_id = $1
  AND (
    $2 = '' OR
    reference_number ILIKE '%' || $2 || '%' OR
    adyen_psp_reference ILIKE '%' || $2 || '%' OR
    customer_id ILIKE '%' || $2 || '%' OR
    COALESCE(payment_method, '') ILIKE '%' || $2 || '%'
  )
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR customer_id = $4)
  AND ($5 = '' OR entity_type = $5)
  AND ($6 = '' OR entity_id = $6)
  AND ($7 = '' OR payment_method = $7)
  AND ($8 = '' OR currency = $8)
  AND ($9::timestamptz IS NULL OR payment_date >= $9::timestamptz)
  AND ($10::timestamptz IS NULL OR payment_date <= $10::timestamptz)
ORDER BY
  (CASE WHEN $11 = 'reference_number' THEN reference_number END) DESC,
  (CASE WHEN $11 = 'payment_method' THEN payment_method END) DESC,
  (CASE WHEN $11 = 'currency' THEN currency END) DESC,
  (CASE WHEN $11 = 'payment_date' THEN payment_date END) DESC,
  (CASE WHEN $11 = 'created_at' THEN created_at END) DESC,
  (CASE WHEN $11 = 'amount' THEN amount END) DESC,
  (CASE WHEN $11 = 'status' THEN
    CASE status
      WHEN 'pending' THEN 1
      WHEN 'authorized' THEN 2
      WHEN 'completed' THEN 3
      WHEN 'failed' THEN 4
      ELSE 5
    END
  END) DESC
LIMIT $12 OFFSET $13;

-- name: CountPaymentsWithFilters :one
SELECT COUNT(*)
FROM payments
WHERE marina_id = $1
  AND (
    $2 = '' OR
    reference_number ILIKE '%' || $2 || '%' OR
    adyen_psp_reference ILIKE '%' || $2 || '%' OR
    customer_id ILIKE '%' || $2 || '%' OR
    COALESCE(payment_method, '') ILIKE '%' || $2 || '%'
  )
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR customer_id = $4)
  AND ($5 = '' OR entity_type = $5)
  AND ($6 = '' OR entity_id = $6)
  AND ($7 = '' OR payment_method = $7)
  AND ($8 = '' OR currency = $8)
  AND ($9::timestamptz IS NULL OR payment_date >= $9::timestamptz)
  AND ($10::timestamptz IS NULL OR payment_date <= $10::timestamptz);

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

-- name: GetPaymentByAdyenSessionID :one
SELECT * FROM payments
WHERE adyen_session_id = $1;

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

-- name: CountPaymentsByStatus :one
SELECT COUNT(*) FROM payments
WHERE marina_id = $1
AND status = $2;

-- name: UpdatePaymentAdyenPayloadsByID :one
UPDATE payments
SET 
    adyen_payment_payload = $2,
    adyen_payment_response = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentReferenceNumberByID :one
UPDATE payments
SET 
    reference_number = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentAdyenPayloadsByReferenceNumber :one
UPDATE payments
SET 
    adyen_payment_payload = $2,
    adyen_payment_response = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE reference_number = $1
RETURNING *;