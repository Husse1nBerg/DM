-- name: CreateNotification :one
INSERT INTO notifications (
    user_id,
    organization_id,
    marina_id,
    type,
    title,
    content,
    data,
    priority
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetNotificationByID :one
SELECT * FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
AND organization_id = $2
AND ($3::uuid IS NULL OR marina_id = $3)
ORDER BY 
    CASE WHEN priority = 'urgent' THEN 1
         WHEN priority = 'high' THEN 2
         WHEN priority = 'normal' THEN 3
         WHEN priority = 'low' THEN 4
         ELSE 5 END,
    created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListUnreadNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
AND organization_id = $2
AND ($3::uuid IS NULL OR marina_id = $3)
AND read = FALSE
ORDER BY 
    CASE WHEN priority = 'urgent' THEN 1
         WHEN priority = 'high' THEN 2
         WHEN priority = 'normal' THEN 3
         WHEN priority = 'low' THEN 4
         ELSE 5 END,
    created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*) as count FROM notifications
WHERE user_id = $1
AND organization_id = $2
AND ($3::uuid IS NULL OR marina_id = $3)
AND read = FALSE;

-- name: MarkNotificationAsRead :one
UPDATE notifications
SET read = TRUE, read_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET read = TRUE, read_at = CURRENT_TIMESTAMP
WHERE user_id = $1
AND organization_id = $2
AND ($3::uuid IS NULL OR marina_id = $3)
AND read = FALSE;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: DeleteOldNotifications :exec
DELETE FROM notifications
WHERE created_at < $1;

-- name: GetNotificationsByType :many
SELECT * FROM notifications
WHERE user_id = $1
AND organization_id = $2
AND type = $3
AND ($4::uuid IS NULL OR marina_id = $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListNotificationsWithFilters :many
SELECT * FROM notifications
WHERE user_id = $1
  AND organization_id = $2
  AND ($3 = '00000000-0000-0000-0000-000000000000'::uuid OR marina_id = $3)
  AND ($4 = '' OR (
    LOWER(title) LIKE LOWER('%' || $4 || '%') OR
    LOWER(content) LIKE LOWER('%' || $4 || '%')
  ))
  AND ($5::text = '' OR read = $5::boolean)
  AND ($6 = '' OR type = $6)
ORDER BY 
  -- Mantener prioridad como orden principal si no se especifica sort
  CASE 
    WHEN $7 = '' THEN (
      CASE WHEN priority = 'urgent' THEN 1
           WHEN priority = 'high' THEN 2
           WHEN priority = 'normal' THEN 3
           WHEN priority = 'low' THEN 4
           ELSE 5 END
    )
  END,
  -- Sorting dinámico opcional
  CASE 
    WHEN $7 = 'type' AND $8 = 'asc' THEN type
    WHEN $7 = 'title' AND $8 = 'asc' THEN title
  END ASC,
  CASE 
    WHEN $7 = 'type' AND $8 = 'desc' THEN type
    WHEN $7 = 'title' AND $8 = 'desc' THEN title
  END DESC,
  CASE 
    WHEN $7 = 'read' AND $8 = 'asc' THEN read
  END ASC,
  CASE 
    WHEN $7 = 'read' AND $8 = 'desc' THEN read
  END DESC,
  CASE 
    WHEN $7 = 'created_at' AND $8 = 'asc' THEN created_at
  END ASC,
  CASE 
    WHEN $7 = 'created_at' AND $8 = 'desc' THEN created_at
    ELSE created_at
  END DESC
LIMIT $9 OFFSET $10;

-- name: CountNotificationsWithFilters :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1
  AND organization_id = $2
  AND ($3 = '00000000-0000-0000-0000-000000000000'::uuid OR marina_id = $3)
  AND ($4 = '' OR (
    LOWER(title) LIKE LOWER('%' || $4 || '%') OR
    LOWER(content) LIKE LOWER('%' || $4 || '%')
  ))
  AND ($5::text = '' OR read = $5::boolean)
  AND ($6 = '' OR type = $6);
