-- name: CreateUser :one
INSERT INTO users (
        username,
        first_name,
        last_name,
        email,
        email_verified,
        phone,
        title,
        image,
        password_hash,
        last_login,
        failed_login_attempts,
        locked_until,
        last_password_reset,
        organization_id,
        marina_id,
        role_id,
        is_superuser,
        is_active,
        modules,
        permissions
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        $18,
        $19,
        $20
    )
RETURNING *;
-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
    AND deleted_at IS NULL;
-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1
    AND deleted_at IS NULL;
-- name: GetUserByUsernameAndOrg :one
SELECT *
FROM users
WHERE username = $1
    AND organization_id = $2
    AND deleted_at IS NULL;
-- name: GetUserByEmailAndOrg :one
SELECT *
FROM users
WHERE email = $1
    AND organization_id = $2
    AND deleted_at IS NULL;
-- name: GetAllUsers :many
SELECT *
FROM users
WHERE deleted_at IS NULL;
-- name: GetAllUsersPaginated :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
-- name: GetUsersByRole :many
SELECT *
FROM users
WHERE role_id = $1
    AND deleted_at IS NULL;
-- name: GetUsersByRolePaginated :many
SELECT *
FROM users
WHERE role_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: GetUsersByOrganization :many
SELECT *
FROM users
WHERE organization_id = $1
    AND deleted_at IS NULL;
-- name: GetUsersByMarina :many
SELECT *
FROM users
WHERE marina_id = $1
    AND deleted_at IS NULL;
-- name: GetUsersPaginated :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
-- name: GetUsersByOrganizationPaginated :many
SELECT *
FROM users
WHERE organization_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: GetUsersByMarinaPaginated :many
SELECT *
FROM users
WHERE marina_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: UpdateUser :one
UPDATE users
SET first_name = $2,
    last_name = $3,
    email = $4,
    email_verified = $5,
    phone = $6,
    title = $7,
    image = $8,
    password_hash = $9,
    last_login = $10,
    failed_login_attempts = $11,
    locked_until = $12,
    last_password_reset = $13,
    marina_id = $14,
    role_id = $15,
    is_superuser = $16,
    is_active = $17,
    modules = $18,
    permissions = $19,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;
-- name: CreateCustomerUser :one
INSERT INTO users (
        username,
        first_name,
        last_name,
        email,
        email_verified,
        phone,
        title,
        image,
        last_login,
        failed_login_attempts,
        locked_until,
        last_password_reset,
        organization_id,
        marina_id,
        role_id,
        customer_id,
        is_customer,
        is_superuser,
        is_active,
        modules,
        permissions
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        $18,
        $19,
        $20,
        $21
    )
RETURNING *;
-- name: ActivateUser :one
UPDATE users
SET is_active = TRUE
WHERE id = $1
RETURNING *;
-- name: DeactivateUser :one
UPDATE users
SET is_active = FALSE
WHERE id = $1
RETURNING *;
-- name: GetMarinaCustomerUsers :many
SELECT *
FROM users
WHERE marina_id = $1
    AND is_customer = TRUE
    AND deleted_at IS NULL;
-- name: GetMarinaCustomerUsersPaginated :many
SELECT *
FROM users
WHERE is_customer = TRUE
    AND marina_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: GetMarinaCustomerUsersByCustomerID :many
SELECT *
FROM users
WHERE is_customer = TRUE
    AND marina_id = $1
    AND customer_id = $2
    AND deleted_at IS NULL
ORDER BY created_at DESC;
-- name: GetMarinaCustomerUserByCustomerIDPaginated :many
SELECT *
FROM users
WHERE is_customer = TRUE
    AND marina_id = $1
    AND customer_id = $2
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
-- name: UpdateUserInvite :one
UPDATE users
SET password_hash = $2,
    last_password_reset = CURRENT_TIMESTAMP,
    failed_login_attempts = 0,
    joined_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;