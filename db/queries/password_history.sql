-- name: AddPasswordToHistory :one
INSERT INTO password_history (user_id, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetPasswordHistoryByUser :many
SELECT password_hash
FROM password_history
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: CleanupOldPasswords :exec
DELETE FROM password_history
WHERE id IN (
    SELECT ph.id
    FROM password_history ph
    WHERE ph.user_id = $1
    ORDER BY ph.created_at DESC
    OFFSET $2
); 