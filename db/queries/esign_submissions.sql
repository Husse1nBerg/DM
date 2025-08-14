-- name: CreateEsignSubmission :one
INSERT INTO esign_submissions (
    organization_id,
    marina_id,
    document_id,
    status,
    blob_url,
    blob_metadata,
    customer_id,
    email,
    name,
    attachment_required
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
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

-- name: ListEsignSubmissionsByCustomerID :many
SELECT *
FROM esign_submissions
WHERE customer_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateEsignSubmission :one
UPDATE esign_submissions
SET
    status = $2,
    blob_url = $3,
    blob_metadata = $4,
    customer_id = $5,
    email = $6,
    name = $7,
    attachment_required = $8,
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

-- name: ListEsignSubmissionsByMarinaFiltered :many
SELECT *
FROM esign_submissions
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR customer_id = $3)
  AND ($4 = '' OR status = $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountEsignSubmissionsByMarinaFiltered :one
SELECT COUNT(*)
FROM esign_submissions
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR customer_id = $3)
  AND ($4 = '' OR status = $4);

-- name: ListEsignSubmissionsFilteredByMarina :many
SELECT *
FROM esign_submissions
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR customer_id = $4)
  AND ($5 = '' OR LOWER(email) LIKE LOWER('%' || $5 || '%'))
  AND ($6 = '' OR LOWER(name) LIKE LOWER('%' || $6 || '%'))
  AND ($7 = '' OR document_id = $7::uuid)
ORDER BY 
  CASE 
    WHEN $8 = 'email' AND $9 = 'asc' THEN email
    WHEN $8 = 'name' AND $9 = 'asc' THEN name
    WHEN $8 = 'status' AND $9 = 'asc' THEN status
    WHEN $8 = 'customer_id' AND $9 = 'asc' THEN customer_id
  END ASC,
  CASE 
    WHEN $8 = 'email' AND $9 = 'desc' THEN email
    WHEN $8 = 'name' AND $9 = 'desc' THEN name
    WHEN $8 = 'status' AND $9 = 'desc' THEN status
    WHEN $8 = 'customer_id' AND $9 = 'desc' THEN customer_id
  END DESC,
  CASE 
    WHEN $8 = 'created_at' AND $9 = 'asc' THEN created_at
    WHEN $8 = 'updated_at' AND $9 = 'asc' THEN updated_at
  END ASC,
  CASE 
    WHEN $8 = 'created_at' AND $9 = 'desc' THEN created_at
    WHEN $8 = 'updated_at' AND $9 = 'desc' THEN updated_at
    ELSE created_at
  END DESC
LIMIT $10 OFFSET $11;

-- name: CountEsignSubmissionsFilteredByMarina :one
SELECT COUNT(*)
FROM esign_submissions
WHERE organization_id = $1
  AND marina_id = $2
  AND deleted_at IS NULL
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR customer_id = $4)
  AND ($5 = '' OR LOWER(email) LIKE LOWER('%' || $5 || '%'))
  AND ($6 = '' OR LOWER(name) LIKE LOWER('%' || $6 || '%'))
  AND ($7 = '' OR document_id = $7::uuid);

-- name: ListEsignSubmissionsFilteredByDocument :many
SELECT *
FROM esign_submissions
WHERE document_id = $1
  AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR customer_id = $3)
  AND ($4 = '' OR LOWER(email) LIKE LOWER('%' || $4 || '%'))
  AND ($5 = '' OR LOWER(name) LIKE LOWER('%' || $5 || '%'))
ORDER BY 
  CASE 
    WHEN $6 = 'email' AND $7 = 'asc' THEN email
    WHEN $6 = 'name' AND $7 = 'asc' THEN name
    WHEN $6 = 'status' AND $7 = 'asc' THEN status
    WHEN $6 = 'customer_id' AND $7 = 'asc' THEN customer_id
  END ASC,
  CASE 
    WHEN $6 = 'email' AND $7 = 'desc' THEN email
    WHEN $6 = 'name' AND $7 = 'desc' THEN name
    WHEN $6 = 'status' AND $7 = 'desc' THEN status
    WHEN $6 = 'customer_id' AND $7 = 'desc' THEN customer_id
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

-- name: CountEsignSubmissionsFilteredByDocument :one
SELECT COUNT(*)
FROM esign_submissions
WHERE document_id = $1
  AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR customer_id = $3)
  AND ($4 = '' OR LOWER(email) LIKE LOWER('%' || $4 || '%'))
  AND ($5 = '' OR LOWER(name) LIKE LOWER('%' || $5 || '%'));

-- name: ListEsignSubmissionsFilteredByStatus :many
SELECT *
FROM esign_submissions
WHERE organization_id = $1
  AND marina_id = $2
  AND status = $3
  AND deleted_at IS NULL
  AND ($4 = '' OR customer_id = $4)
  AND ($5 = '' OR LOWER(email) LIKE LOWER('%' || $5 || '%'))
  AND ($6 = '' OR LOWER(name) LIKE LOWER('%' || $6 || '%'))
  AND ($7 = '' OR document_id = $7::uuid)
ORDER BY 
  CASE 
    WHEN $8 = 'email' AND $9 = 'asc' THEN email
    WHEN $8 = 'name' AND $9 = 'asc' THEN name
    WHEN $8 = 'customer_id' AND $9 = 'asc' THEN customer_id
  END ASC,
  CASE 
    WHEN $8 = 'email' AND $9 = 'desc' THEN email
    WHEN $8 = 'name' AND $9 = 'desc' THEN name
    WHEN $8 = 'customer_id' AND $9 = 'desc' THEN customer_id
  END DESC,
  CASE 
    WHEN $8 = 'created_at' AND $9 = 'asc' THEN created_at
    WHEN $8 = 'updated_at' AND $9 = 'asc' THEN updated_at
  END ASC,
  CASE 
    WHEN $8 = 'created_at' AND $9 = 'desc' THEN created_at
    WHEN $8 = 'updated_at' AND $9 = 'desc' THEN updated_at
    ELSE created_at
  END DESC
LIMIT $10 OFFSET $11;

-- name: CountEsignSubmissionsFilteredByStatus :one
SELECT COUNT(*)
FROM esign_submissions
WHERE organization_id = $1
  AND marina_id = $2
  AND status = $3
  AND deleted_at IS NULL
  AND ($4 = '' OR customer_id = $4)
  AND ($5 = '' OR LOWER(email) LIKE LOWER('%' || $5 || '%'))
  AND ($6 = '' OR LOWER(name) LIKE LOWER('%' || $6 || '%'))
  AND ($7 = '' OR document_id = $7::uuid);