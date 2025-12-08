-- name: CreatePaymentLink :one
INSERT INTO payment_links (
    token,
    organization_id,
    marina_id,
    customer_id,
    scope,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetValidPaymentLinkByToken :one
SELECT * FROM payment_links
WHERE token = $1 AND used = FALSE AND revoked = FALSE AND expires_at > CURRENT_TIMESTAMP;

-- name: MarkPaymentLinkUsed :one
UPDATE payment_links
SET used = TRUE, used_at = CURRENT_TIMESTAMP
WHERE token = $1 AND used = FALSE AND revoked = FALSE AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: RevokePaymentLink :one
UPDATE payment_links
SET revoked = TRUE, revoked_at = CURRENT_TIMESTAMP
WHERE token = $1 AND revoked = FALSE
RETURNING *;

-- name: ListActivePaymentLinksByCustomer :many
SELECT * FROM payment_links
WHERE customer_id = $1 AND marina_id = $2 AND revoked = FALSE AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC;

