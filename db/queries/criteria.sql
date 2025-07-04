-- name: CreateCriteria :one
INSERT INTO criteria (
    marina_id,
    name,
    description,
    criteria
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetCriteriaByID :one
SELECT *
FROM criteria
WHERE id = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: ListCriteria :many
SELECT *
FROM criteria
WHERE marina_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListCriteriaPaginated :many
SELECT *
FROM criteria
WHERE marina_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetCriteriaCount :one
SELECT COUNT(*)
FROM criteria
WHERE marina_id = $1
    AND deleted_at IS NULL;

-- name: UpdateCriteria :one
UPDATE criteria
SET name = $2,
    description = $3,
    criteria = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: DeleteCriteria :exec
UPDATE criteria
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: SearchCriteriaByName :many
SELECT *
FROM criteria
WHERE marina_id = $1
    AND name ILIKE '%' || $2 || '%'
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: SearchCriteriaByNamePaginated :many
SELECT *
FROM criteria
WHERE marina_id = $1
    AND name ILIKE '%' || $2 || '%'
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4; 