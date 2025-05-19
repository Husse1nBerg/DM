-- name: ListMessages :many
SELECT *
FROM messages
WHERE marina_id = $1
AND customer_id = $2
AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3
OFFSET $4;

-- name: ListMessagesAll :many
SELECT *
FROM messages
WHERE marina_id = $1
AND customer_id = $2
AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListMessagesByCustomer :many
SELECT *
FROM messages
WHERE marina_id = $1
AND customer_id = $2
AND type IN ('email', 'sms')
AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3
OFFSET $4;

-- name: ListMessagesByCustomerAll :many
SELECT *
FROM messages
WHERE marina_id = $1
AND customer_id = $2
AND type IN ('email', 'sms')
AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateMessage :one
INSERT INTO messages (
    marina_id,
    customer_id,
    type,
    direction,
    body,
    sender,
    recipient,
    contact,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetMessageByID :one
SELECT *
FROM messages
WHERE id = $1
AND marina_id = $2
AND customer_id = $3
AND deleted_at IS NULL;

-- name: UpdateMessageStatus :exec
UPDATE messages
SET status = $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2;

-- name: DeleteMessage :exec
UPDATE messages
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
AND marina_id = $2
AND customer_id = $3;
