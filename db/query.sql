-- name: GetRoleById :one
SELECT * FROM roles WHERE id = $1 LIMIT 1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1 LIMIT 1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: CreateRole :one
INSERT INTO roles (name) 
VALUES ($1) 
RETURNING *;

-- name: UpdateRole :exec
UPDATE roles 
SET name = $2
WHERE id = $1;

-- name: DeleteRole :exec
DELETE FROM roles 
WHERE id = $1;

-- name: GetSubscriptionPlanById :one
SELECT * FROM subscription_plans WHERE id = $1 LIMIT 1;

-- name: GetSubscriptionPlanByTitle :one
SELECT * FROM subscription_plans WHERE title = $1 LIMIT 1;

-- name: ListSubscriptionPlans :many
SELECT * FROM subscription_plans ORDER BY title;

-- name: CreateSubscriptionPlan :one
INSERT INTO subscription_plans (
    title, description, options, monthly_price, annually_price, is_active, 
    is_default, number_of_users, trial_period
) 
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateSubscriptionPlan :exec
UPDATE subscription_plans
SET title = $2, 
    description = $3, 
    options = $4, 
    monthly_price = $5, 
    annually_price = $6, 
    is_active = $7,
    is_default = $8,
    number_of_users = $9,
    trial_period = $10,
    updated = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteSubscriptionPlan :exec
DELETE FROM subscription_plans 
WHERE id = $1;

-- name: GetAddressById :one
SELECT * FROM addresses WHERE id = $1 LIMIT 1;

-- name: ListAddresses :many
SELECT * FROM addresses ORDER BY city, street;

-- name: CreateAddress :one
INSERT INTO addresses (
    street, city, state, zip_code, country
) 
VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateAddress :exec
UPDATE addresses
SET street = $2, 
    city = $3, 
    state = $4, 
    zip_code = $5, 
    country = $6
WHERE id = $1;

-- name: DeleteAddress :exec
DELETE FROM addresses 
WHERE id = $1;

-- name: GetCompanyById :one
SELECT * FROM companies WHERE id = $1 LIMIT 1;

-- name: ListCompanies :many
SELECT * FROM companies ORDER BY name;

-- name: CreateCompany :one
INSERT INTO companies (
    name, website, email, phone, company_size, logo, plan_id, address
) 
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateCompany :exec
UPDATE companies
SET name = $2, 
    website = $3, 
    email = $4, 
    phone = $5, 
    company_size = $6, 
    logo = $7, 
    plan_id = $8, 
    address = $9,
    updated = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteCompany :exec
DELETE FROM companies 
WHERE id = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY first_name;

-- name: CreateUser :one
INSERT INTO users (
    email, password, first_name, last_name, phone, role_id, company_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
    SET first_name = $2, 
    last_name = $3, 
    phone = $4, 
    updated = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users 
WHERE id = $1;