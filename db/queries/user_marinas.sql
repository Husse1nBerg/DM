-- name: AssignUserToMarina :exec
INSERT INTO user_marinas (user_id, marina_id, customer_id, role_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, marina_id) DO UPDATE SET role_id = EXCLUDED.role_id;
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

-- name: ListUserMarinasAssignmentsPaginatedAsc :many
SELECT 
    um.*, 
    u.*, 
    r.name AS role_name
FROM user_marinas um
JOIN users u ON u.id = um.user_id
LEFT JOIN roles r ON um.role_id = r.id
WHERE um.marina_id = $1
  -- customer_id filter (true = must exist, false = must not exist, null = ignore)
  AND ($2::boolean IS NULL OR ($2 = TRUE AND um.customer_id IS NOT NULL) OR ($2 = FALSE AND um.customer_id IS NULL))
  -- search filter across user fields
  AND (
    $3 = '' 
    OR u.username ILIKE '%' || $3 || '%'
    OR u.first_name ILIKE '%' || $3 || '%'
    OR u.last_name ILIKE '%' || $3 || '%'
    OR u.email ILIKE '%' || $3 || '%'
    OR u.phone ILIKE '%' || $3 || '%'
    OR u.title ILIKE '%' || $3 || '%'
  )
  -- optional role filter
  AND ($4 = '' OR um.role_id = $4::uuid)
  -- optional is_active filter
  AND ($5 = '' OR u.is_active = $5::boolean)
  -- fixed rules
  AND u.is_superuser = FALSE
  AND u.deleted_at IS NULL
ORDER BY
  (CASE WHEN $6 = 'username'             THEN u.username END) ASC,
  (CASE WHEN $6 = 'first_name'           THEN u.first_name END) ASC,
  (CASE WHEN $6 = 'last_name'            THEN u.last_name END) ASC,
  (CASE WHEN $6 = 'email'                THEN u.email END) ASC,
  (CASE WHEN $6 = 'phone'                THEN u.phone END) ASC,
  (CASE WHEN $6 = 'title'                THEN u.title END) ASC,
  (CASE WHEN $6 = 'last_login'           THEN u.last_login END) ASC,
  (CASE WHEN $6 = 'failed_login_attempts'THEN u.failed_login_attempts END) ASC,
  (CASE WHEN $6 = 'locked_until'         THEN u.locked_until END) ASC,
  (CASE WHEN $6 = 'last_password_reset'  THEN u.last_password_reset END) ASC,
  (CASE WHEN $6 = 'created_at'           THEN u.created_at END) ASC,
  (CASE WHEN $6 = 'updated_at'           THEN u.updated_at END) ASC,
  (CASE WHEN $6 = 'role_name'            THEN r.name END) ASC
LIMIT $7 OFFSET $8;

-- name: ListUserMarinasAssignmentsPaginatedDesc :many
SELECT 
    um.*, 
    u.*, 
    r.name AS role_name
FROM user_marinas um
JOIN users u ON u.id = um.user_id
LEFT JOIN roles r ON um.role_id = r.id
WHERE um.marina_id = $1
  -- customer_id filter (true = must exist, false = must not exist, null = ignore)
  AND ($2::boolean IS NULL OR ($2 = TRUE AND um.customer_id IS NOT NULL) OR ($2 = FALSE AND um.customer_id IS NULL))
  -- search filter across user fields
  AND (
    $3 = '' 
    OR u.username ILIKE '%' || $3 || '%'
    OR u.first_name ILIKE '%' || $3 || '%'
    OR u.last_name ILIKE '%' || $3 || '%'
    OR u.email ILIKE '%' || $3 || '%'
    OR u.phone ILIKE '%' || $3 || '%'
    OR u.title ILIKE '%' || $3 || '%'
  )
  -- optional role filter
  AND ($4 = '' OR um.role_id = $4::uuid)
  -- optional is_active filter
  AND ($5 = '' OR u.is_active = $5::boolean)
  -- fixed rules
  AND u.is_superuser = FALSE
  AND u.deleted_at IS NULL
ORDER BY
  (CASE WHEN $6 = 'username'             THEN u.username END) DESC,
  (CASE WHEN $6 = 'first_name'           THEN u.first_name END) DESC,
  (CASE WHEN $6 = 'last_name'            THEN u.last_name END) DESC,
  (CASE WHEN $6 = 'email'                THEN u.email END) DESC,
  (CASE WHEN $6 = 'phone'                THEN u.phone END) DESC,
  (CASE WHEN $6 = 'title'                THEN u.title END) DESC,
  (CASE WHEN $6 = 'last_login'           THEN u.last_login END) DESC,
  (CASE WHEN $6 = 'failed_login_attempts'THEN u.failed_login_attempts END) DESC,
  (CASE WHEN $6 = 'locked_until'         THEN u.locked_until END) DESC,
  (CASE WHEN $6 = 'last_password_reset'  THEN u.last_password_reset END) DESC,
  (CASE WHEN $6 = 'created_at'           THEN u.created_at END) DESC,
  (CASE WHEN $6 = 'updated_at'           THEN u.updated_at END) DESC,
  (CASE WHEN $6 = 'role_name'            THEN r.name END) DESC
LIMIT $7 OFFSET $8;

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

-- name: CountUserMarinasAssignmentsWithFilters :one
SELECT COUNT(*)
FROM user_marinas um
JOIN users u ON u.id = um.user_id
WHERE um.marina_id = $1
  -- customer_id filter (true = must exist, false = must not exist, null = ignore)
  AND ($2 = '' OR ($2 = 'true' AND um.customer_id IS NOT NULL) OR ($2 = 'false' AND um.customer_id IS NULL))
  -- search filter across user fields
  AND (
    $3 = '' 
    OR u.username ILIKE '%' || $3 || '%'
    OR u.first_name ILIKE '%' || $3 || '%'
    OR u.last_name ILIKE '%' || $3 || '%'
    OR u.email ILIKE '%' || $3 || '%'
    OR u.phone ILIKE '%' || $3 || '%'
    OR u.title ILIKE '%' || $3 || '%'
  )
  -- optional role filter
  AND ($4 = '' OR um.role_id = $4::uuid)
  -- optional is_active filter
  AND ($5 = '' OR u.is_active = $5::boolean)
  -- fixed rules
  AND u.is_superuser = FALSE
  AND u.deleted_at IS NULL;
-- name: GetUserMarinaAssignmentByUserAndMarina :one
SELECT role_id, customer_id
FROM user_marinas
WHERE user_id = $1 AND marina_id = $2;
-- name: UpdateUserMarinaRole :exec
UPDATE user_marinas
SET role_id = $3
WHERE user_id = $1 AND marina_id = $2;
-- name: CountUsersByRoleID :one
SELECT COUNT(*)
FROM user_marinas um
JOIN roles r ON r.id = um.role_id
WHERE r.id = $1
  AND r.deleted_at IS NULL;