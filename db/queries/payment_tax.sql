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
    payment_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetTaxConfigurationByID :one
SELECT * FROM tax_configurations
WHERE id = $1;

-- name: GetTaxConfigurationByMarinaAndPaymentType :one
SELECT * FROM tax_configurations
WHERE marina_id = $1 
    AND payment_type = $2
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
    payment_type = COALESCE(sqlc.narg('payment_type'), payment_type),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
RETURNING *;


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
    payment_type
FROM tax_configurations
WHERE id = $1;

-- name: CountTaxConfigurationsByMarinas :one
SELECT COUNT(*) FROM tax_configurations
WHERE marina_id = ANY($1::uuid[]);

