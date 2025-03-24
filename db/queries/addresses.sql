-- name: CreateAddress :one
INSERT INTO addresses (
        street,
        city,
        state,
        postal_code,
        country,
        latitude,
        longitude
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: GetAddressByID :one
SELECT *
FROM addresses
WHERE id = $1;
-- name: UpdateAddress :one
UPDATE addresses
SET street = $2,
    city = $3,
    state = $4,
    postal_code = $5,
    country = $6,
    latitude = $7,
    longitude = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;