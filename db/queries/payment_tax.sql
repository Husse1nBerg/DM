-- name: CreateTaxConfiguration :one
INSERT INTO tax_configurations (
    marina_id,
    convenience_fee,
    convenience_fee_type,
    convenience_fee_enabled,
    convenience_fee_description,
    surcharge,
    surcharge_type,
    surcharge_enabled,
    surcharge_description,
    tax_rate,
    tax_enabled,
    tax_description,
    is_active,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetTaxConfigurationByID :one
SELECT * FROM tax_configurations
WHERE id = $1;

-- name: GetActiveTaxConfigurationByMarinaID :one
SELECT * FROM tax_configurations
WHERE marina_id = $1 
    AND is_active = true
LIMIT 1;

-- name: GetAllTaxConfigurationsByMarinaID :many
SELECT * FROM tax_configurations
WHERE marina_id = $1
ORDER BY created_at DESC;

-- name: ListTaxConfigurations :many
SELECT * FROM tax_configurations
WHERE marina_id = ANY($1::uuid[])
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateTaxConfiguration :one
UPDATE tax_configurations
SET 
    convenience_fee = COALESCE(sqlc.narg('convenience_fee'), convenience_fee),
    convenience_fee_type = COALESCE(sqlc.narg('convenience_fee_type'), convenience_fee_type),
    convenience_fee_enabled = COALESCE(sqlc.narg('convenience_fee_enabled'), convenience_fee_enabled),
    convenience_fee_description = COALESCE(sqlc.narg('convenience_fee_description'), convenience_fee_description),
    surcharge = COALESCE(sqlc.narg('surcharge'), surcharge),
    surcharge_type = COALESCE(sqlc.narg('surcharge_type'), surcharge_type),
    surcharge_enabled = COALESCE(sqlc.narg('surcharge_enabled'), surcharge_enabled),
    surcharge_description = COALESCE(sqlc.narg('surcharge_description'), surcharge_description),
    tax_rate = COALESCE(sqlc.narg('tax_rate'), tax_rate),
    tax_enabled = COALESCE(sqlc.narg('tax_enabled'), tax_enabled),
    tax_description = COALESCE(sqlc.narg('tax_description'), tax_description),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeactivateTaxConfiguration :exec
UPDATE tax_configurations
SET 
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeactivateAllMarinaConfigurations :exec
UPDATE tax_configurations
SET 
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE marina_id = $1 AND is_active = true;

-- name: DeleteTaxConfiguration :exec
DELETE FROM tax_configurations
WHERE id = $1;

-- name: GetTaxConfigurationForCalculation :one
SELECT 
    convenience_fee,
    convenience_fee_type,
    convenience_fee_enabled,
    surcharge,
    surcharge_type,
    surcharge_enabled,
    tax_rate,
    tax_enabled
FROM tax_configurations
WHERE id = $1 AND is_active = true;

-- name: CountTaxConfigurationsByMarinas :one
SELECT COUNT(*) FROM tax_configurations
WHERE marina_id = ANY($1::uuid[]);

