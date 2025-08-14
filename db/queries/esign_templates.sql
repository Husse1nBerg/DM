-- name: CreateEsignTemplate :one
INSERT INTO esign_templates (
    organization_id,
    marina_id,
    name,
    description,
    type,
    status,
    blob_url,
    blob_metadata,
    json_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetEsignTemplateByID :one
SELECT *
FROM esign_templates
WHERE id = $1
    AND deleted_at IS NULL;

-- name: ListEsignTemplatesByMarina :many
SELECT *
FROM esign_templates
WHERE organization_id = $1
    AND marina_id = $2
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateEsignTemplate :one
UPDATE esign_templates
SET
    name = $2,
    description = $3,
    type = $4,
    status = $5,
    blob_url = $6,
    blob_metadata = $7,
    json_data = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: UpdateEsignTemplateStatus :one
UPDATE esign_templates
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteEsignTemplate :exec
UPDATE esign_templates
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteEsignTemplate :exec
DELETE FROM esign_templates
WHERE id = $1;

-- name: CountEsignTemplatesByMarina :one
SELECT COUNT(*)
FROM esign_templates
WHERE organization_id = $1
    AND marina_id = $2
    AND deleted_at IS NULL;

-- name: ListEsignTemplatesByMarinaStatus :many
SELECT *
FROM esign_templates
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountEsignTemplatesByMarinaStatus :one
SELECT COUNT(*)
FROM esign_templates
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3);

-- name: ListEsignTemplatesFiltered :many
SELECT *
FROM esign_templates
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR LOWER(name) LIKE LOWER('%' || $4 || '%'))
  AND ($5 = '' OR LOWER(type) LIKE LOWER('%' || $5 || '%'))
  AND ($6 = '' OR LOWER(description) LIKE LOWER('%' || $6 || '%'))
ORDER BY 
  CASE 
    WHEN $7 = 'name' AND $8 = 'asc' THEN name
    WHEN $7 = 'type' AND $8 = 'asc' THEN type
    WHEN $7 = 'status' AND $8 = 'asc' THEN status
  END ASC,
  CASE 
    WHEN $7 = 'name' AND $8 = 'desc' THEN name
    WHEN $7 = 'type' AND $8 = 'desc' THEN type
    WHEN $7 = 'status' AND $8 = 'desc' THEN status
  END DESC,
  CASE 
    WHEN $7 = 'created_at' AND $8 = 'asc' THEN created_at
    WHEN $7 = 'updated_at' AND $8 = 'asc' THEN updated_at
  END ASC,
  CASE 
    WHEN $7 = 'created_at' AND $8 = 'desc' THEN created_at
    WHEN $7 = 'updated_at' AND $8 = 'desc' THEN updated_at
    ELSE created_at
  END DESC
LIMIT $9 OFFSET $10;

-- name: CountEsignTemplatesFiltered :one
SELECT COUNT(*)
FROM esign_templates
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR LOWER(name) LIKE LOWER('%' || $4 || '%'))
  AND ($5 = '' OR LOWER(type) LIKE LOWER('%' || $5 || '%'))
  AND ($6 = '' OR LOWER(description) LIKE LOWER('%' || $6 || '%')); 