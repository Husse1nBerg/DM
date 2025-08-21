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
-- name: GetOrganizationsWithFiltersAsc :many
SELECT *
FROM organizations o
WHERE o.deleted_at IS NULL
  -- free text search
  AND (
    $1 = '' 
    OR o.name ILIKE '%' || $1 || '%'
    OR o.email ILIKE '%' || $1 || '%'
    OR o.website ILIKE '%' || $1 || '%'
    OR o.country ILIKE '%' || $1 || '%'
    OR o.phone ILIKE '%' || $1 || '%'
  )
  -- optional filters
  AND ($2 = '' OR o.is_active = $2::boolean)
  AND ($3 = '' OR o.is_test = $3::boolean)
ORDER BY
  (CASE WHEN $4 = 'name'       THEN o.name END) ASC,
  (CASE WHEN $4 = 'email'      THEN o.email END) ASC,
  (CASE WHEN $4 = 'website'    THEN o.website END) ASC,
  (CASE WHEN $4 = 'country'    THEN o.country END) ASC,
  (CASE WHEN $4 = 'phone'      THEN o.phone END) ASC,
  (CASE WHEN $4 = 'is_active'  THEN o.is_active END) ASC,
  (CASE WHEN $4 = 'is_test'    THEN o.is_test END) ASC,
  (CASE WHEN $4 = 'created_at' THEN o.created_at END) ASC,
  (CASE WHEN $4 = 'updated_at' THEN o.updated_at END) ASC
LIMIT $5 OFFSET $6;
-- name: GetOrganizationsWithFiltersDesc :many
SELECT *
FROM organizations o
WHERE o.deleted_at IS NULL
  -- free text search
  AND (
    $1 = '' 
    OR o.name ILIKE '%' || $1 || '%'
    OR o.email ILIKE '%' || $1 || '%'
    OR o.website ILIKE '%' || $1 || '%'
    OR o.country ILIKE '%' || $1 || '%'
    OR o.phone ILIKE '%' || $1 || '%'
  )
  -- optional filters
  AND ($2 = '' OR o.is_active = $2::boolean)
  AND ($3 = '' OR o.is_test = $3::boolean)
ORDER BY
  (CASE WHEN $4 = 'name'       THEN o.name END) DESC,
  (CASE WHEN $4 = 'email'      THEN o.email END) DESC,
  (CASE WHEN $4 = 'website'    THEN o.website END) DESC,
  (CASE WHEN $4 = 'country'    THEN o.country END) DESC,
  (CASE WHEN $4 = 'phone'      THEN o.phone END) DESC,
  (CASE WHEN $4 = 'is_active'  THEN o.is_active END) DESC,
  (CASE WHEN $4 = 'is_test'    THEN o.is_test END) DESC,
  (CASE WHEN $4 = 'created_at' THEN o.created_at END) DESC,
  (CASE WHEN $4 = 'updated_at' THEN o.updated_at END) DESC
LIMIT $5 OFFSET $6;
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