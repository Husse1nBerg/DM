-- name: CreateEsignTemplate :one
INSERT INTO esign_templates (
    organization_id,
    marina_id,
    name,
    description,
    type,
    status,
    blob_url,
    blob_metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
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