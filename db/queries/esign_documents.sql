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

-- name: ListEsignDocumentsByMarinaStatus :many
SELECT *
FROM esign_documents
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountEsignDocumentsByMarinaStatus :one
SELECT COUNT(*)
FROM esign_documents
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3);

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

-- name: ListEsignDocumentsWithFilters :many
SELECT *
FROM esign_documents
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR type = $4)
  AND ($5 = '' OR (
    CAST(id AS TEXT) ILIKE '%' || $5 || '%' OR
    (blob_metadata->>'customerId') ILIKE '%' || $5 || '%'
  ))
ORDER BY 
  CASE 
    WHEN $6 = 'type' AND $7 = 'asc' THEN type
    WHEN $6 = 'status' AND $7 = 'asc' THEN status
  END ASC,
  CASE 
    WHEN $6 = 'type' AND $7 = 'desc' THEN type
    WHEN $6 = 'status' AND $7 = 'desc' THEN status
  END DESC,
  CASE 
    WHEN $6 = 'created_at' AND $7 = 'asc' THEN created_at
    WHEN $6 = 'updated_at' AND $7 = 'asc' THEN updated_at
  END ASC,
  CASE 
    WHEN $6 = 'created_at' AND $7 = 'desc' THEN created_at
    WHEN $6 = 'updated_at' AND $7 = 'desc' THEN updated_at
    ELSE created_at
  END DESC
LIMIT $8 OFFSET $9;

-- name: CountEsignDocumentsWithFilters :one
SELECT COUNT(*)
FROM esign_documents
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR type = $4)
  AND ($5 = '' OR (
    LOWER(type) LIKE LOWER('%' || $5 || '%') OR
    LOWER(status) LIKE LOWER('%' || $5 || '%') OR
    CAST(id AS TEXT) ILIKE '%' || $5 || '%' OR
    (blob_metadata->>'customerId') ILIKE '%' || $5 || '%'
  ));

-- name: ListEsignDocumentsByTemplateFiltered :many
SELECT *
FROM esign_documents
WHERE template_id = $1
  AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR LOWER(type) LIKE LOWER('%' || $3 || '%'))
ORDER BY 
  CASE 
    WHEN $4 = 'type' AND $5 = 'asc' THEN type
    WHEN $4 = 'status' AND $5 = 'asc' THEN status
  END ASC,
  CASE 
    WHEN $4 = 'type' AND $5 = 'desc' THEN type
    WHEN $4 = 'status' AND $5 = 'desc' THEN status
  END DESC,
  CASE 
    WHEN $4 = 'created_at' AND $5 = 'asc' THEN created_at
    WHEN $4 = 'updated_at' AND $5 = 'asc' THEN updated_at
  END ASC,
  CASE 
    WHEN $4 = 'created_at' AND $5 = 'desc' THEN created_at
    WHEN $4 = 'updated_at' AND $5 = 'desc' THEN updated_at
    ELSE created_at
  END DESC
LIMIT $6 OFFSET $7;

-- name: CountEsignDocumentsByTemplateFiltered :one
SELECT COUNT(*)
FROM esign_documents
WHERE template_id = $1
  AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR LOWER(type) LIKE LOWER('%' || $3 || '%'));