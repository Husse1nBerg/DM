-- name: CreatePasswordRecoveryToken :one
INSERT INTO password_recovery (user_id, email, token, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPasswordRecoveryToken :one
SELECT * FROM password_recovery
WHERE token = $1 AND email = $2 AND expires_at > NOW() AND used = FALSE
LIMIT 1;

-- name: MarkTokenAsUsed :exec
UPDATE password_recovery
SET used = TRUE
WHERE token = $1 AND email = $2;

-- name: DeleteExpiredTokens :exec
DELETE FROM password_recovery
WHERE expires_at < NOW(); 