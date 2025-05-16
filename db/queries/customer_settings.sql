-- name: CreateCustomerSettings :one
INSERT INTO customer_settings (
    marina_id,
    customer_id,
    enable_portal
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetCustomerSettings :one
SELECT *
FROM customer_settings
WHERE marina_id = $1 AND customer_id = $2;

-- name: UpdateCustomerSettings :one
UPDATE customer_settings
SET enable_portal = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE marina_id = $1 AND customer_id = $2
RETURNING *;

-- name: UpsertCustomerSettings :one
INSERT INTO customer_settings (
    marina_id,
    customer_id,
    customer_user_id,
    enable_portal
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
ON CONFLICT (marina_id, customer_id) DO UPDATE
SET enable_portal = EXCLUDED.enable_portal,
    customer_user_id = EXCLUDED.customer_user_id,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;
-- name: UpdateCustomerSettingsCustomerUserID :one
UPDATE customer_settings
SET customer_user_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE marina_id = $1 AND customer_id = $2
RETURNING *;