-- name: CreateVesselGalleryItem :one
INSERT INTO vessel_gallery (
    marina_id,
    customer_id,
    vessel_id,
    image_url,
    description,
    main,
    public
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

-- name: GetVesselGalleryItemByID :one
SELECT *
FROM vessel_gallery
WHERE id = $1
    AND deleted_at IS NULL;

-- name: GetVesselGallery :many
SELECT *
FROM vessel_gallery
WHERE vessel_id = $1
    AND customer_id = $2
    AND marina_id = $3
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetVesselGalleryWithVisibility :many
SELECT *
FROM vessel_gallery
WHERE vessel_id = $1
    AND customer_id = $2
    AND marina_id = $3
    AND deleted_at IS NULL
    AND ($4 = true OR public = true)
ORDER BY created_at DESC;

-- name: UpdateVesselGalleryItem :one
UPDATE vessel_gallery
SET image_url = $2,
    description = $3,
    main = $4,
    public = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: RemoveMainVesselImage :exec
UPDATE vessel_gallery
SET main = false,
    updated_at = CURRENT_TIMESTAMP
WHERE vessel_id = $1
    AND customer_id = $2
    AND marina_id = $3
    AND deleted_at IS NULL;

-- name: SoftDeleteVesselGalleryItem :exec
UPDATE vessel_gallery
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteVesselGalleryItem :exec
DELETE FROM vessel_gallery
WHERE id = $1;

-- name: CreateMarinaGalleryItem :one
INSERT INTO marina_gallery (
    marina_id,
    image_url,
    description
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetMarinaGalleryItemByID :one
SELECT *
FROM marina_gallery
WHERE id = $1
    AND deleted_at IS NULL;

-- name: GetMarinaGallery :many
SELECT *
FROM marina_gallery
WHERE marina_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateMarinaGalleryItem :one
UPDATE marina_gallery
SET image_url = $2,
    description = $3,
    public = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteMarinaGalleryItem :exec
UPDATE marina_gallery
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteMarinaGalleryItem :exec
DELETE FROM marina_gallery
WHERE id = $1;
