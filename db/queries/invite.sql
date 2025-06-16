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