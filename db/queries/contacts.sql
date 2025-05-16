-- name: CreateContact :one
INSERT INTO contacts (
    marina_id,
    type,
    name,
    description,
    email,
    phone
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
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

-- name: UpdateContact :one
UPDATE contacts
SET type = $2,
    name = $3,
    description = $4,
    email = $5,
    phone = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: DeleteContact :exec
UPDATE contacts
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;
