-- name: CreateMarina :one
INSERT INTO marinas (
        organization_id,
        name,
        email,
        location,
        phone,
        country,
        currency,
        working_hours,
        website,
        image,
        max_users,
        is_active,
        is_test,
        address_id,
        system_id,
        storage_usage,
        email_text_usage
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
        0,
        0
    )
RETURNING *;
-- name: GetMarinaByID :one
SELECT *
FROM marinas
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetMarinaByEmail :one
SELECT *
FROM marinas
WHERE email = $1
    AND deleted_at IS NULL;
-- name: GetAllMarinas :many
SELECT *
FROM marinas
WHERE deleted_at IS NULL;
-- name: GetMarinasByOrganization :many
SELECT *
FROM marinas
WHERE organization_id = $1
    AND deleted_at IS NULL;
-- name: GetMarinasPaginated :many
SELECT *
FROM marinas
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
-- name: GetMarinasByOrganizationPaginated :many
SELECT *
FROM marinas
WHERE organization_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: UpdateMarina :one
UPDATE marinas
SET name = $2,
    email = $3,
    location = $4,
    phone = $5,
    country = $6,
    currency = $7,
    working_hours = $8,
    website = $9,
    image = $10,
    max_users = $11,
    is_active = $12,
    is_test = $13,
    updated_at = CURRENT_TIMESTAMP,
    address_id = $14,
    system_id = $15,
    email_text_usage = COALESCE($16, email_text_usage)
WHERE id = $1
RETURNING *;
-- name: SoftDeleteMarina :exec
UPDATE marinas
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;
-- name: UpdateMarinaSystemID :one
UPDATE marinas
SET system_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: IncrementMarinaStorageUsage :one
UPDATE marinas
SET storage_usage = storage_usage + $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: DecrementMarinaStorageUsage :one
UPDATE marinas
SET storage_usage = GREATEST(storage_usage - $2, 0),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: GetMarinaStorageUsage :one
SELECT storage_usage
FROM marinas
WHERE id = $1
    AND deleted_at IS NULL;
-- name: IncrementMarinaEmailTextUsage :one
UPDATE marinas
SET email_text_usage = COALESCE(email_text_usage, 0)::smallint + $2::smallint,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: DecrementMarinaEmailTextUsage :one
UPDATE marinas
SET email_text_usage = GREATEST(COALESCE(email_text_usage, 0)::smallint - $2::smallint, 0)::smallint,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: GetMarinaEmailTextUsage :one
SELECT COALESCE(email_text_usage, 0)::smallint
FROM marinas
WHERE id = $1
    AND deleted_at IS NULL;