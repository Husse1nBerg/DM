-- name: CreateMarinaUsageHistory :one
INSERT INTO marina_usage_history (
    marina_id,
    storage_usage,
    email_usage,
    text_usage,
    document_usage
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetMarinaUsageHistoryByID :one
SELECT * FROM marina_usage_history
WHERE id = $1;

-- name: GetMarinaUsageHistoryByMarinaID :many
SELECT * FROM marina_usage_history
WHERE marina_id = $1
ORDER BY created_at DESC;

-- name: GetMarinaUsageHistoryByDateRange :many
SELECT * FROM marina_usage_history
WHERE marina_id = $1
AND created_at >= $2
AND created_at <= $3
ORDER BY created_at DESC;

-- name: GetLatestMarinaUsageHistory :one
SELECT * FROM marina_usage_history
WHERE marina_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateMarinaUsageHistory :one
UPDATE marina_usage_history
SET
    storage_usage = $2,
    email_usage = $3,
    text_usage = $4,
    document_usage = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteMarinaUsageHistory :exec
DELETE FROM marina_usage_history
WHERE id = $1;

-- name: GetMarinaUsageHistoryByMonth :one
SELECT * FROM marina_usage_history
WHERE marina_id = $1
AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', $2::timestamp)
ORDER BY created_at DESC
LIMIT 1;

-- name: GetAllMarinaUsageHistory :many
SELECT * FROM marina_usage_history
ORDER BY created_at DESC;

-- name: GetAllMarinaUsageHistoryByDateRange :many
SELECT * FROM marina_usage_history
WHERE created_at >= $1
  AND created_at <= $2
ORDER BY created_at DESC;