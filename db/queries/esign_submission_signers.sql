-- name: CreateEsignSubmissionSigner :one
INSERT INTO esign_submission_signers (
    submission_id,
    email,
    name,
    sign_order,
    status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetEsignSubmissionSignerByID :one
SELECT *
FROM esign_submission_signers
WHERE id = $1
    AND deleted_at IS NULL;

-- name: ListEsignSubmissionSignersBySubmissionID :many
SELECT *
FROM esign_submission_signers
WHERE submission_id = $1
    AND deleted_at IS NULL
ORDER BY sign_order ASC;

-- name: GetNextSignerForSubmission :one
SELECT *
FROM esign_submission_signers
WHERE submission_id = $1
    AND status = 'pending'
    AND deleted_at IS NULL
ORDER BY sign_order ASC
LIMIT 1;

-- name: UpdateEsignSubmissionSignerStatus :one
UPDATE esign_submission_signers
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: UpdateEsignSubmissionSignerToSigned :one
UPDATE esign_submission_signers
SET
    status = 'signed',
    signed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: UpdateEsignSubmissionSignerToDeclined :one
UPDATE esign_submission_signers
SET
    status = 'declined',
    declined_at = CURRENT_TIMESTAMP,
    declined_reason = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: UpdateEsignSubmissionSigner :one
UPDATE esign_submission_signers
SET
    email = $2,
    name = $3,
    sign_order = $4,
    status = $5,
    signed_at = CASE WHEN $5 = 'signed' THEN CURRENT_TIMESTAMP ELSE signed_at END,
    declined_at = CASE WHEN $5 = 'declined' THEN CURRENT_TIMESTAMP ELSE declined_at END,
    declined_reason = CASE WHEN $5 = 'declined' THEN $6 ELSE declined_reason END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteEsignSubmissionSigner :exec
UPDATE esign_submission_signers
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteEsignSubmissionSigner :exec
DELETE FROM esign_submission_signers
WHERE id = $1;

-- name: CountEsignSubmissionSignersBySubmissionID :one
SELECT COUNT(*)
FROM esign_submission_signers
WHERE submission_id = $1
    AND deleted_at IS NULL;

-- name: CountEsignSubmissionSignersByStatus :one
SELECT COUNT(*)
FROM esign_submission_signers
WHERE submission_id = $1
    AND status = $2
    AND deleted_at IS NULL;

-- name: GetEsignSubmissionSignerByEmailAndSubmissionID :one
SELECT *
FROM esign_submission_signers
WHERE submission_id = $1
    AND email = $2
    AND deleted_at IS NULL;

-- name: ListEsignSubmissionSignersByStatus :many
SELECT *
FROM esign_submission_signers
WHERE submission_id = $1
    AND status = $2
    AND deleted_at IS NULL
ORDER BY sign_order ASC;

-- name: UpdateEsignSubmissionSignerOrder :exec
UPDATE esign_submission_signers
SET
    sign_order = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL;

-- name: GetEsignSubmissionSignersWithSubmissionDetails :many
SELECT 
    ess.*,
    es.organization_id,
    es.marina_id,
    es.document_id,
    es.blob_url,
    es.customer_id,
    es.is_multiple_signature
FROM esign_submission_signers ess
JOIN esign_submissions es ON ess.submission_id = es.id
WHERE ess.submission_id = $1
    AND ess.deleted_at IS NULL
    AND es.deleted_at IS NULL
ORDER BY ess.sign_order ASC;

-- name: ListEsignSubmissionSignersBySubmissionIDs :many
SELECT *
FROM esign_submission_signers
WHERE submission_id = ANY($1::uuid[])
    AND deleted_at IS NULL
ORDER BY submission_id, sign_order ASC;
