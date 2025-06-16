-- name: ListMessages :many
SELECT *
FROM messages
WHERE marina_id = $1
AND customer_id = $2
AND deleted_at IS NULL
ORDER BY pinned DESC, created_at DESC
LIMIT $3
OFFSET $4;

-- name: ListMessagesAll :many
SELECT *
FROM messages
WHERE marina_id = $1
AND customer_id = $2
AND deleted_at IS NULL
ORDER BY pinned DESC, created_at DESC;

-- name: ListMessagesByCustomer :many
SELECT *
FROM messages
WHERE marina_id = $1
AND customer_id = $2
AND type IN ('email', 'sms')
AND deleted_at IS NULL
ORDER BY pinned DESC, created_at DESC
LIMIT $3
OFFSET $4;

-- name: ListMessagesByCustomerAll :many
SELECT *
FROM messages
WHERE marina_id = $1
AND customer_id = $2
AND type IN ('email', 'sms')
AND deleted_at IS NULL
ORDER BY pinned DESC, created_at DESC;

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
    status,
    pinned
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
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

-- name: UpdateOtherMessagesPinnedStatus :exec
UPDATE messages
SET pinned = false,
    updated_at = CURRENT_TIMESTAMP
WHERE marina_id = $1
AND customer_id = $2
AND id != $3
AND deleted_at IS NULL;

-- name: UpdateMessage :one
UPDATE messages
SET body = $1,
    pinned = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $3
AND marina_id = $4
AND customer_id = $5
AND deleted_at IS NULL
RETURNING *;
