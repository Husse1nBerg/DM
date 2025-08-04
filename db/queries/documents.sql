-- name: GetDocumentByID :one
SELECT * FROM documents
WHERE id = $1;

-- name: CreateDocument :one
INSERT INTO documents (
    marina_id,
    entity_type,
    entity_id,
    file_name,
    file_type,
    file_path,
    file_size
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListDocumentsByEntity :many
SELECT *
FROM documents
WHERE marina_id = $1
AND entity_type = $2
AND entity_id = $3
ORDER BY created_at DESC;

-- name: ListDocumentsByMarina :many
SELECT *
FROM documents
WHERE marina_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteDocument :exec
DELETE FROM documents
WHERE id = $1;

-- name: UpdateDocument :one
UPDATE documents
SET
    file_name = $2,
    file_type = $3,
    file_path = $4,
    file_size = $5,
    public = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *; 