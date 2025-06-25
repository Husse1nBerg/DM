-- name: CreateEsignDocument :one
INSERT INTO esign_documents (
    organization_id,
    marina_id,
    type,
    status,
    blob_url,
    blob_metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: CreateEsignDocumentWithTemplate :one
INSERT INTO esign_documents (
    template_id,
    organization_id,
    marina_id,
    type,
    status,
    blob_url,
    blob_metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetEsignDocumentByID :one
SELECT *
FROM esign_documents
WHERE id = $1
    AND deleted_at IS NULL;

-- name: ListEsignDocumentsByMarina :many
SELECT *
FROM esign_documents
WHERE organization_id = $1
    AND marina_id = $2
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListEsignDocumentsByTemplate :many
SELECT *
FROM esign_documents
WHERE template_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- name: UpdateEsignDocument :one
UPDATE esign_documents
SET
    type = $2,
    status = $3,
    blob_url = $4,
    blob_metadata = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: UpdateEsignDocumentStatus :one
UPDATE esign_documents
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteEsignDocument :exec
UPDATE esign_documents
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteEsignDocument :exec
DELETE FROM esign_documents
WHERE id = $1;

-- name: CountEsignDocumentsByMarina :one
SELECT COUNT(*)
FROM esign_documents
WHERE organization_id = $1
    AND marina_id = $2
    AND deleted_at IS NULL;