-- name: CreateOrganization :one
INSERT INTO organizations (
        email,
        name,
        image,
        website,
        country,
        phone,
        is_active,
        is_test,
        address_id
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;
-- name: GetOrganizationByID :one
SELECT *
FROM organizations
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetOrganizationByEmail :one
SELECT *
FROM organizations
WHERE email = $1
    AND deleted_at IS NULL;
-- name: GetAllOrganizations :many
SELECT *
FROM organizations
WHERE deleted_at IS NULL
ORDER BY created_at DESC;
-- name: GetOrganizationsPaginated :many
SELECT *
FROM organizations
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
-- name: UpdateOrganization :one
UPDATE organizations
SET email = $2,
    name = $3,
    image = $4,
    website = $5,
    country = $6,
    phone = $7,
    is_active = $8,
    is_test = $9,
    updated_at = CURRENT_TIMESTAMP,
    address_id = $10
WHERE id = $1
RETURNING *;
-- name: SoftDeleteOrganization :exec
UPDATE organizations
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;