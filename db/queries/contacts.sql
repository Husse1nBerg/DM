-- name: CreateContact :one
INSERT INTO contacts (
    marina_id,
    type,
    name,
    description,
    email,
    phone,
    is_cp_contact
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: GetContactByID :one
SELECT *
FROM contacts
WHERE id = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: ListContacts :many
SELECT *
FROM contacts
WHERE marina_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListContactsByType :many
SELECT *
FROM contacts
WHERE marina_id = $1
    AND type = $2
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListCPContacts :many
SELECT *
FROM contacts
WHERE marina_id = $1
    AND is_cp_contact = TRUE
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateContact :one
UPDATE contacts
SET type = $2,
    name = $3,
    description = $4,
    email = $5,
    phone = $6,
    is_cp_contact = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: UnsetCPContact :exec
UPDATE contacts
SET is_cp_contact = FALSE
WHERE marina_id = $1
    AND deleted_at IS NULL;

-- name: DeleteContact :exec
UPDATE contacts
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: ListContactsWithFiltersAsc :many
SELECT *
FROM contacts
WHERE marina_id = $1
  AND deleted_at IS NULL
  AND (
    ($2 = '' OR name ILIKE '%' || $2 || '%'
      OR email ILIKE '%' || $2 || '%'
      OR phone ILIKE '%' || $2 || '%')
  )
  AND ($3 = '' OR type = $3)
ORDER BY
  (CASE WHEN $4 = 'name' THEN name END) ASC,
  (CASE WHEN $4 = 'email' THEN email END) ASC,
  (CASE WHEN $4 = 'phone' THEN phone END) ASC,
  (CASE WHEN $4 = 'type' THEN type END) ASC,
  (CASE WHEN $4 = 'created_at' THEN created_at END) ASC,
  (CASE WHEN $4 = 'updated_at' THEN updated_at END) ASC
LIMIT $5 OFFSET $6;

-- name: ListContactsWithFiltersDesc :many
SELECT *
FROM contacts
WHERE marina_id = $1
  AND deleted_at IS NULL
  AND (
    ($2 = '' OR name ILIKE '%' || $2 || '%'
      OR email ILIKE '%' || $2 || '%'
      OR phone ILIKE '%' || $2 || '%')
  )
  AND ($3 = '' OR type = $3)
ORDER BY
  (CASE WHEN $4 = 'name' THEN name END) DESC,
  (CASE WHEN $4 = 'email' THEN email END) DESC,
  (CASE WHEN $4 = 'phone' THEN phone END) DESC,
  (CASE WHEN $4 = 'type' THEN type END) DESC,
  (CASE WHEN $4 = 'created_at' THEN created_at END) DESC,
  (CASE WHEN $4 = 'updated_at' THEN updated_at END) DESC
LIMIT $5 OFFSET $6;

-- name: CountContactsWithFilters :one
SELECT COUNT(*)
FROM contacts
WHERE marina_id = $1
  AND deleted_at IS NULL
  AND (
    ($2 = '' OR name ILIKE '%' || $2 || '%'
      OR email ILIKE '%' || $2 || '%'
      OR phone ILIKE '%' || $2 || '%')
  )
  AND ($3 = '' OR type = $3);
