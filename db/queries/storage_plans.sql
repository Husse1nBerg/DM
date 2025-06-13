-- Get all storage plans
-- name: ListStoragePlans :many
SELECT id, name, monthly_price, storage_limit_gb, user_limit, is_most_popular, created_at, updated_at FROM storage_plans ORDER BY monthly_price;

-- Get all storage plans with pagination
-- name: ListStoragePlansPaginated :many
SELECT id, name, monthly_price, storage_limit_gb, user_limit, is_most_popular, created_at, updated_at FROM storage_plans 
ORDER BY monthly_price
LIMIT $1 OFFSET $2;

-- Get a specific plan by ID
-- name: GetStoragePlanByID :one
SELECT id, name, monthly_price, storage_limit_gb, user_limit, is_most_popular, created_at, updated_at FROM storage_plans WHERE id = $1;

-- Get the most popular plan
-- name: GetMostPopularStoragePlan :one
SELECT * FROM storage_plans WHERE is_most_popular = true LIMIT 1;

-- Get plan by name
-- name: GetStoragePlanByName :one
SELECT id, name, monthly_price, storage_limit_gb, user_limit, is_most_popular, created_at, updated_at FROM storage_plans WHERE name = $1;

-- Get marina's current plan
-- name: GetMarinaStoragePlan :one
SELECT sp.* 
FROM storage_plans sp
JOIN marinas m ON m.storage_plan_id = sp.id
WHERE m.id = $1;

-- Get plans with storage limits
-- name: GetStoragePlansWithStorageLimits :many
SELECT * FROM storage_plans 
WHERE storage_limit_gb != 'Unlimited Storage'
ORDER BY monthly_price;

-- Get plans with user limits
-- name: GetStoragePlansWithUserLimits :many
SELECT * FROM storage_plans 
WHERE user_limit != 'Unlimited Users'
ORDER BY monthly_price;

-- Get plans within a specific price range
-- name: GetStoragePlansInPriceRange :many
SELECT * FROM storage_plans 
WHERE monthly_price BETWEEN $1 AND $2
ORDER BY monthly_price;

-- Create a new storage plan
-- name: CreateStoragePlan :one
INSERT INTO storage_plans (
    name,
    monthly_price,
    storage_limit_gb,
    user_limit,
    is_most_popular
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, name, monthly_price, storage_limit_gb, user_limit, is_most_popular, created_at, updated_at;

-- Update an existing storage plan
-- name: UpdateStoragePlan :one
UPDATE storage_plans
SET name = $2,
    monthly_price = $3,
    storage_limit_gb = $4,
    user_limit = $5,
    is_most_popular = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, monthly_price, storage_limit_gb, user_limit, is_most_popular, created_at, updated_at;

-- Get plans with usage limits
-- name: GetStoragePlansWithUsageLimits :many
SELECT id, name, monthly_price, storage_limit_gb, user_limit, is_most_popular, created_at, updated_at FROM storage_plans 
WHERE storage_limit_gb IS NOT NULL
ORDER BY monthly_price;
