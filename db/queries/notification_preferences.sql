-- name: CreateNotificationPreference :one
INSERT INTO notification_preferences (
    user_id,
    notification_type,
    enabled,
    delivery_method
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetNotificationPreferences :many
SELECT u.email, np.* 
FROM notification_preferences np
INNER JOIN users u on np.user_id = u.id
WHERE np.user_id = $1
ORDER BY notification_type;

-- name: GetNotificationPreference :one
SELECT * FROM notification_preferences
WHERE user_id = $1 AND notification_type = $2;

-- name: UpdateNotificationPreference :one
UPDATE notification_preferences
SET enabled = $3, delivery_method = $4, updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND notification_type = $2
RETURNING *;

-- name: UpsertNotificationPreference :one
INSERT INTO notification_preferences (
    user_id,
    notification_type,
    enabled,
    delivery_method
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (user_id, notification_type) 
DO UPDATE SET
    enabled = EXCLUDED.enabled,
    delivery_method = EXCLUDED.delivery_method,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: DeleteNotificationPreference :exec
DELETE FROM notification_preferences
WHERE user_id = $1 AND notification_type = $2;

-- name: GetEnabledNotificationPreferences :many
SELECT * FROM notification_preferences
WHERE user_id = $1 AND enabled = TRUE; 