-- name: CreateEsignSubmission :one
INSERT INTO esign_submissions (
    organization_id,
    marina_id,
    document_id,
    status,
    blob_url,
    blob_metadata,
    customer_id,
    email
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetEsignSubmissionByID :one
SELECT *
FROM esign_submissions
WHERE id = $1
    AND deleted_at IS NULL;

-- name: ListEsignSubmissionsByMarina :many
SELECT *
FROM esign_submissions
WHERE organization_id = $1
    AND marina_id = $2
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListEsignSubmissionsByDocument :many
SELECT *
FROM esign_submissions
WHERE document_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListEsignSubmissionsByStatus :many
SELECT *
FROM esign_submissions
WHERE organization_id = $1
    AND marina_id = $2
    AND status = $3
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: UpdateEsignSubmission :one
UPDATE esign_submissions
SET
    status = $2,
    blob_url = $3,
    blob_metadata = $4,
    customer_id = $5,
    email = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: UpdateEsignSubmissionStatus :one
UPDATE esign_submissions
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteEsignSubmission :exec
UPDATE esign_submissions
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteEsignSubmission :exec
DELETE FROM esign_submissions
WHERE id = $1;

-- name: CountEsignSubmissionsByMarina :one
SELECT COUNT(*)
FROM esign_submissions
WHERE organization_id = $1
    AND marina_id = $2
    AND deleted_at IS NULL;

-- name: CountEsignSubmissionsByDocument :one
SELECT COUNT(*)
FROM esign_submissions
WHERE document_id = $1
    AND deleted_at IS NULL;

-- name: CountEsignSubmissionsByStatus :one
SELECT COUNT(*)
FROM esign_submissions
WHERE organization_id = $1
    AND marina_id = $2
    AND status = $3
    AND deleted_at IS NULL; 