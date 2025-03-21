-- name: AssignUserToMarina :exec
INSERT INTO user_marinas (user_id, marina_id)
VALUES ($1, $2) ON CONFLICT DO NOTHING;
-- name: UnassignUserFromMarina :exec
DELETE FROM user_marinas
WHERE user_id = $1
    AND marina_id = $2;
-- name: GetUserMarinasList :many
SELECT m.*
FROM marinas m
    JOIN user_marinas um ON m.id = um.marina_id
WHERE um.user_id = $1
    AND m.deleted_at IS NULL;
-- name: GetMarinaUsersList :many
SELECT u.*
FROM users u
    JOIN user_marinas um ON u.id = um.user_id
WHERE um.marina_id = $1
    AND u.deleted_at IS NULL;