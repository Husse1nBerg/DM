-- name: CreateDMECredentials :one
INSERT INTO dme_credentials (
        organization_id,
        username,
        password,
        is_old_api,
        access_token,
        refresh_token,
        expiry_date
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7
    )
RETURNING *;
-- name: GetDMECredentialsByOrgID :one
SELECT *
FROM dme_credentials
WHERE organization_id = $1
    AND deleted_at IS NULL;
-- name: GetDMECredentialsByOrgIDWithDeleted :one
SELECT *
FROM dme_credentials
WHERE organization_id = $1;
-- name: UpdateDMECredentials :one
UPDATE dme_credentials
SET username = $2,
    password = $3,
    is_old_api = $4,
    access_token = $5,
    refresh_token = $6,
    expiry_date = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
RETURNING *;
-- name: UpdateDMEToken :one
UPDATE dme_credentials
SET access_token = $2,
    refresh_token = $3,
    expiry_date = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1
RETURNING *;
-- name: DeleteDMECredentials :exec
UPDATE dme_credentials
SET deleted_at = CURRENT_TIMESTAMP
WHERE organization_id = $1;
-- name: HardDeleteDMECredentials :exec
DELETE FROM dme_credentials
WHERE organization_id = $1;