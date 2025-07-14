-- name: CreateDocumentPlan :one
INSERT INTO document_plans (
    name, monthly_price, document_limit, user_limit, is_most_popular
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateDocumentPlan :one
UPDATE document_plans
SET
    name = $2,
    monthly_price = $3,
    document_limit = $4,
    user_limit = $5,
    is_most_popular = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDocumentPlan :exec
DELETE FROM document_plans WHERE id = $1;

-- name: GetDocumentPlanByID :one
SELECT * FROM document_plans WHERE id = $1;

-- name: GetAllDocumentPlans :many
SELECT * FROM document_plans ORDER BY monthly_price ASC, name ASC;

-- name: GetDocumentPlanByName :one
SELECT * FROM document_plans WHERE name = $1;
