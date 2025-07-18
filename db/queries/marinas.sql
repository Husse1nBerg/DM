-- name: CreateMarina :one
INSERT INTO marinas (
        organization_id,
        name,
        email,
        location,
        phone,
        country,
        currency,
        working_hours,
        website,
        image,
        max_users,
        is_active,
        is_test,
        address_id,
        system_id,
        storage_usage,
        email_usage,
        text_usage,
        notes_messages_plan_id,
        storage_plan_id,
        document_plan_id,
        modules,
        document_usage
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        0,
        0,
        0,
        $16,
        $17,
        $18,
        $19,
        0
    )
RETURNING *;
-- name: GetMarinaByID :one
SELECT *
FROM marinas
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetMarinaByEmail :one
SELECT *
FROM marinas
WHERE email = $1
    AND deleted_at IS NULL;
-- name: GetAllMarinas :many
SELECT *
FROM marinas
WHERE deleted_at IS NULL;
-- name: GetMarinasByOrganization :many
SELECT *
FROM marinas
WHERE organization_id = $1
    AND deleted_at IS NULL;
-- name: GetMarinasPaginated :many
SELECT *
FROM marinas
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
-- name: GetMarinasByOrganizationPaginated :many
SELECT *
FROM marinas
WHERE organization_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
-- name: UpdateMarina :one
UPDATE marinas
SET name = $2,
    email = $3,
    location = $4,
    phone = $5,
    country = $6,
    currency = $7,
    working_hours = $8,
    website = $9,
    image = $10,
    max_users = $11,
    is_active = $12,
    is_test = $13,
    updated_at = CURRENT_TIMESTAMP,
    address_id = $14,
    system_id = $15,
    email_usage = COALESCE($16, email_usage),
    text_usage = COALESCE($17, text_usage),
    notes_messages_plan_id = $18,
    storage_plan_id = $19,
    document_plan_id = $20,
    modules = $21,
    document_usage = COALESCE($22, document_usage),
    internal_announcement = $23,
    external_announcement = $24
WHERE id = $1
RETURNING *;
-- name: SoftDeleteMarina :exec
UPDATE marinas
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;
-- name: UpdateMarinaSystemID :one
UPDATE marinas
SET system_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: IncrementMarinaStorageUsage :one
UPDATE marinas
SET storage_usage = storage_usage + $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: DecrementMarinaStorageUsage :one
UPDATE marinas
SET storage_usage = GREATEST(storage_usage - $2, 0),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: GetMarinaStorageUsage :one
SELECT storage_usage
FROM marinas
WHERE id = $1
    AND deleted_at IS NULL;
-- name: IncrementMarinaEmailUsage :one
UPDATE marinas
SET email_usage = COALESCE(email_usage, 0)::smallint + $2::smallint,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: DecrementMarinaEmailUsage :one
UPDATE marinas
SET email_usage = GREATEST(COALESCE(email_usage, 0)::smallint - $2::smallint, 0)::smallint,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: GetMarinaEmailUsage :one
SELECT COALESCE(email_usage, 0)::smallint
FROM marinas
WHERE id = $1
    AND deleted_at IS NULL;
-- name: IncrementMarinaTextUsage :one
UPDATE marinas
SET text_usage = COALESCE(text_usage, 0)::smallint + $2::smallint,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: DecrementMarinaTextUsage :one
UPDATE marinas
SET text_usage = GREATEST(COALESCE(text_usage, 0)::smallint - $2::smallint, 0)::smallint,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: GetMarinaTextUsage :one
SELECT COALESCE(text_usage, 0)::smallint
FROM marinas
WHERE id = $1
    AND deleted_at IS NULL;
-- name: IncrementMarinaDocumentUsage :one
UPDATE marinas
SET document_usage = COALESCE(document_usage, 0)::bigint + $2::bigint,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: DecrementMarinaDocumentUsage :one
UPDATE marinas
SET document_usage = GREATEST(COALESCE(document_usage, 0)::bigint - $2::bigint, 0)::bigint,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
-- name: GetMarinaDocumentUsage :one
SELECT COALESCE(document_usage, 0)::bigint
FROM marinas
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetMarinasOverCurrentLimit :many
SELECT m.*
FROM marinas m
LEFT JOIN storage_plans sp ON m.storage_plan_id = sp.id
LEFT JOIN notes_messages_plans nmp ON m.notes_messages_plan_id = nmp.id
LEFT JOIN document_plans dp ON m.document_plan_id = dp.id
LEFT JOIN (
    SELECT um.marina_id, COUNT(*) AS user_count
    FROM user_marinas um
    GROUP BY um.marina_id
) uc ON m.id = uc.marina_id
WHERE m.deleted_at IS NULL
  AND (
    (sp.storage_limit_gb IS NOT NULL AND sp.storage_limit_gb != 'Unlimited Storage' AND m.storage_usage > sp.storage_limit_gb * 1024 * 1024 * 1024)
    OR (nmp.text_limit IS NOT NULL AND nmp.text_limit != 'Unlimited Texts' AND m.text_usage > nmp.text_limit)
    OR (nmp.email_limit IS NOT NULL AND nmp.email_limit != 'Unlimited Emails' AND m.email_usage > nmp.email_limit)
    OR (dp.document_limit IS NOT NULL AND dp.document_limit != 'Unlimited Documents' AND m.document_usage > dp.document_limit)
    OR (
      COALESCE(sp.user_limit, 1000000) != 'Unlimited Users'
      AND uc.user_count > COALESCE(NULLIF(sp.user_limit, 'Unlimited Users')::int, 1000000)
    )
    OR (
      COALESCE(nmp.user_limit, 1000000) != 'Unlimited Users'
      AND uc.user_count > COALESCE(NULLIF(nmp.user_limit, 'Unlimited Users')::int, 1000000)
    )
    OR (
      COALESCE(dp.user_limit, 1000000) != 'Unlimited Users'
      AND uc.user_count > COALESCE(NULLIF(dp.user_limit, 'Unlimited Users')::int, 1000000)
    )
  );