-- Get all notes and messages plans
-- name: ListNotesMessagesPlans :many
SELECT * FROM notes_messages_plans ORDER BY monthly_price;

-- Get all notes and messages plans with pagination
-- name: ListNotesMessagesPlansPaginated :many
SELECT * FROM notes_messages_plans 
ORDER BY monthly_price
LIMIT $1 OFFSET $2;

-- Get a specific plan by ID
-- name: GetNotesMessagesPlanByID :one
SELECT * FROM notes_messages_plans WHERE id = $1;

-- Get the most popular plan
-- name: GetMostPopularNotesMessagesPlan :one
SELECT * FROM notes_messages_plans WHERE is_most_popular = true LIMIT 1;

-- Get plan by name
-- name: GetNotesMessagesPlanByName :one
SELECT * FROM notes_messages_plans WHERE name = $1;

-- Get marina's current plan
-- name: GetMarinaNotesMessagesPlan :one
SELECT nmp.* 
FROM notes_messages_plans nmp
JOIN marinas m ON m.notes_messages_plan_id = nmp.id
WHERE m.id = $1;

-- Get plans with usage limits
-- name: GetNotesMessagesPlansWithUsageLimits :many
SELECT * FROM notes_messages_plans 
WHERE text_limit != 'Unlimited Texts'
   OR email_limit != 'Unlimited Emails'
ORDER BY monthly_price;

-- Get plans with user limits
-- name: GetNotesMessagesPlansWithUserLimits :many
SELECT * FROM notes_messages_plans 
WHERE user_limit != 'Unlimited Users'
ORDER BY monthly_price;

-- Create a new notes and messages plan
-- name: CreateNotesMessagesPlan :one
INSERT INTO notes_messages_plans (
    name,
    monthly_price,
    text_limit,
    email_limit,
    user_limit,
    is_most_popular
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- Update an existing notes and messages plan
-- name: UpdateNotesMessagesPlan :one
UPDATE notes_messages_plans
SET name = $2,
    monthly_price = $3,
    text_limit = $4,
    email_limit = $5,
    user_limit = $6,
    is_most_popular = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
