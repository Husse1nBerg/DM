-- name: CreateInvite :one
INSERT INTO invites (
    user_id,
    email,
    token,
    expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetInviteByToken :one
SELECT * FROM invites
WHERE token = $1 AND used = FALSE AND expires_at > CURRENT_TIMESTAMP;

-- name: MarkInviteAsUsed :one
UPDATE invites
SET used = TRUE
WHERE token = $1 AND used = FALSE AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: GetInviteByUserID :one
SELECT * FROM invites
WHERE user_id = $1 AND used = FALSE AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC
LIMIT 1;

-- name: GetInvitesByEmail :many
SELECT * FROM invites
WHERE email = $1 AND used = FALSE AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC;

-- name: ExpireInvitesByUserID :exec
UPDATE invites
SET expires_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND used = FALSE;

-- name: ExpireInvitesByEmail :exec
UPDATE invites
SET expires_at = CURRENT_TIMESTAMP
WHERE email = $1 AND used = FALSE;