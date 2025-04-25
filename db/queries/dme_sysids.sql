-- name: CreateDMESysID :one
INSERT INTO dme_sysids (
        organization_id,
        marina_id,
        name,
        description,
        system_id,
        is_active
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6
    )
RETURNING *;
-- name: CreateDMESysIDWithoutMarinaID :one
INSERT INTO dme_sysids (
    organization_id,
    name,
    description,
    system_id,
    is_active
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
-- name: GetDMESysIDByID :one
SELECT *
FROM dme_sysids
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetDMESysIDByIDWithDeleted :one
SELECT *
FROM dme_sysids
WHERE id = $1;
-- name: GetDMESysIDBySystemID :one
SELECT *
FROM dme_sysids
WHERE system_id = $1
    AND deleted_at IS NULL;
-- name: GetDMESysIDByOrgAndSystemID :one
SELECT *
FROM dme_sysids
WHERE organization_id = $1
    AND system_id = $2
    AND deleted_at IS NULL;
-- name: GetDMESysIDsByOrgID :many
SELECT *
FROM dme_sysids
WHERE organization_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC;
-- name: GetDMESysIDByMarinaID :one
SELECT *
FROM dme_sysids
WHERE marina_id = $1
    AND deleted_at IS NULL;
-- name: ListDMESysIDs :many
SELECT *
FROM dme_sysids
WHERE deleted_at IS NULL
ORDER BY created_at DESC;
-- name: UpdateDMESysID :one
UPDATE dme_sysids
SET organization_id = $2,
    marina_id = $3,
    name = $4,
    description = $5,
    system_id = $6,
    is_active = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: UpsertDMESysID :one
INSERT INTO dme_sysids (
    organization_id,
    marina_id,
    name,
    description,
    system_id,
    is_active
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
ON CONFLICT (organization_id, system_id) 
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;
-- name: LinkDMESysIDToMarina :one
UPDATE dme_sysids
SET marina_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: UnlinkDMESysIDFromMarina :one
UPDATE dme_sysids
SET marina_id = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: DeleteDMESysID :exec
UPDATE dme_sysids
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;
-- name: HardDeleteDMESysID :exec
DELETE FROM dme_sysids
WHERE id = $1;