-- name: AssignUserToMarina :exec
INSERT INTO user_marinas (user_id, marina_id, customer_id)
VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;
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
-- name: GetUserMarinasListPaginated :many
SELECT m.*
FROM marinas m
    JOIN user_marinas um ON m.id = um.marina_id
WHERE um.user_id = $1
    AND m.deleted_at IS NULL
ORDER BY m.created_at DESC
LIMIT $2 OFFSET $3;
-- name: GetMarinaUsersList :many
SELECT u.*, r.name as role_name
FROM users u
    JOIN user_marinas um ON u.id = um.user_id
    LEFT JOIN roles r ON u.role_id = r.id
WHERE um.marina_id = $1
    AND (u.is_customer = $2 OR $2 IS NULL)
    AND u.deleted_at IS NULL;
-- name: GetMarinaUsersListPaginated :many
SELECT u.*, r.name as role_name, CONCAT(u.first_name, ' ', u.last_name) as customer_name
FROM users u
    JOIN user_marinas um ON u.id = um.user_id
    LEFT JOIN roles r ON u.role_id = r.id
WHERE um.marina_id = $1
    AND (u.is_customer = $2 OR $2 IS NULL)
    AND u.deleted_at IS NULL
ORDER BY u.created_at DESC
LIMIT $3 OFFSET $4;
-- name: CustomerMarinaUser :one
SELECT u.*
FROM users u
    JOIN user_marinas um ON u.id = um.user_id
WHERE um.marina_id = $1
    AND um.customer_id = $2
    AND u.deleted_at IS NULL;
-- name: UserCanAccessMarina :one
SELECT EXISTS (
    SELECT 1
    FROM user_marinas um
    WHERE um.user_id = $1
        AND um.marina_id = $2
) AS can_access;