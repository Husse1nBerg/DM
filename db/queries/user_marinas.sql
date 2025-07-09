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
SELECT u.*, r.name as role_name
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
-- name: GetUserRoleInMarina :one
SELECT u.role_id
FROM users u
    JOIN user_marinas um ON u.id = um.user_id
WHERE u.id = $1
    AND um.marina_id = $2
    AND u.deleted_at IS NULL;
-- name: GetCustomerMarinaUsersPaginated :many
SELECT u.*
FROM users u
    JOIN user_marinas um ON u.id = um.user_id
WHERE um.marina_id = $1
    AND um.customer_id = $2
    AND u.deleted_at IS NULL
ORDER BY u.created_at DESC
LIMIT $3 OFFSET $4;
-- name: CountCustomerMarinaUsers :one
SELECT COUNT(*)
FROM users u
    JOIN user_marinas um ON u.id = um.user_id
WHERE um.marina_id = $1
    AND um.customer_id = $2
    AND u.deleted_at IS NULL;
-- name: GetUsersNotAssignedToMarinaPaginated :many
SELECT u.*
FROM users u
WHERE u.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM user_marinas um
    WHERE um.user_id = u.id
      AND u.is_superuser = FALSE
      AND um.marina_id = $1
  )
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;
-- name: GetUsersNotAssignedToMarinaPaginatedAdmin :many
SELECT u.*
FROM users u
WHERE u.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM user_marinas um
    WHERE um.user_id = u.id
      AND um.marina_id = $1
  )
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;
-- name: CountUsersNotAssignedToMarina :one
SELECT COUNT(*)
FROM users u
WHERE u.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM user_marinas um
    WHERE um.user_id = u.id
      AND u.is_superuser = FALSE
      AND um.marina_id = $1
  );

-- name: CountUsersNotAssignedToMarinaAdmin :one
SELECT COUNT(*)
FROM users u
WHERE u.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM user_marinas um
    WHERE um.user_id = u.id
      AND um.marina_id = $1
  );

-- name: ListUserMarinasAssignmentsPaginated :many
SELECT um.*, u.*, r.name as role_name
FROM user_marinas um
JOIN users u ON u.id = um.user_id
LEFT JOIN roles r ON u.role_id = r.id
WHERE um.marina_id = $1
  AND ($2::bool IS NULL OR ($2 = TRUE AND um.customer_id IS NOT NULL) OR ($2 = FALSE AND um.customer_id IS NULL))
  AND u.is_superuser = FALSE
  AND u.deleted_at IS NULL
ORDER BY u.created_at DESC
LIMIT $3 OFFSET $4;
-- name: ListUserMarinasAssignmentsPaginatedAdmin :many
SELECT um.*, u.*, r.name as role_name
FROM user_marinas um
JOIN users u ON u.id = um.user_id
LEFT JOIN roles r ON u.role_id = r.id
WHERE um.marina_id = $1
  AND ($2::bool IS NULL OR ($2 = TRUE AND um.customer_id IS NOT NULL) OR ($2 = FALSE AND um.customer_id IS NULL))
  AND u.deleted_at IS NULL
ORDER BY u.created_at DESC
LIMIT $3 OFFSET $4;
-- name: ListUserMarinasAssignmentsPaginatedAdminOnly :many
SELECT um.*, u.*, r.name as role_name
FROM user_marinas um
JOIN users u ON u.id = um.user_id
LEFT JOIN roles r ON u.role_id = r.id
WHERE um.marina_id = $1
  AND u.is_superuser = TRUE 
  AND u.deleted_at IS NULL
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;
-- name: CountUserMarinasAssignmentsPaginatedAdminOnly :one
SELECT COUNT(*)
FROM user_marinas um
JOIN users u ON u.id = um.user_id
WHERE um.marina_id = $1
  AND u.is_superuser = TRUE 
  AND u.deleted_at IS NULL;