-- name: CreateRole :one
INSERT INTO roles (name, description, permissions, is_active, is_customer_role, type, marina_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: GetRoleByID :one
SELECT *
FROM roles
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetRoleByName :one
SELECT *
FROM roles
WHERE name = $1
    AND deleted_at IS NULL
    AND marina_id IS NULL;
-- name: GetRoleByNameAndMarina :one
SELECT *
FROM roles
WHERE name = $1
    AND marina_id = $2
    AND deleted_at IS NULL;
-- name: GetAllRoles :many
SELECT *
FROM roles
WHERE deleted_at IS NULL;
-- name: GetRolesPaginated :many
SELECT *
FROM roles
WHERE deleted_at IS NULL
    AND (marina_id IS NULL OR marina_id = $1)
    AND type != 'internal'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: UpdateRole :one
UPDATE roles
SET name = $2,
    description = $3,
    permissions = $4,
    is_active = $5,
    is_customer_role = $6,
    type = $7,
    marina_id = NULLIF($8::uuid, '00000000-0000-0000-0000-000000000000'),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: SoftDeleteRole :exec
UPDATE roles
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;
-- name: GetAllRolesByType :many
SELECT *
FROM roles
WHERE deleted_at IS NULL
    AND type = $1;
-- name: GetAllRolesByTypePaginated :many
SELECT *
FROM roles
WHERE deleted_at IS NULL
    AND type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: GetAllRolesByTypes :many
SELECT *
FROM roles
WHERE deleted_at IS NULL
    AND type = ANY($1::text[])
ORDER BY created_at DESC;
-- name: GetAllRolesByTypesPaginated :many
SELECT *
FROM roles
WHERE deleted_at IS NULL
    AND type = ANY($1::text[])
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: CountRolesByMarina :one
SELECT COUNT(*)
FROM roles
WHERE deleted_at IS NULL
    AND (marina_id IS NULL OR marina_id = $1)
    AND type != 'internal';