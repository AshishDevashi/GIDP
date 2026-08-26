-- name: CreateProject :one
INSERT INTO projects (
    name, slug, description, project_type, architecture, owner_team_id, tech_lead_id,
    lifecycle_id, tier_id, docs_url, dashboard_url, runbook_url,
    parent_project_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12,
    $13
)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1
LIMIT 1;

-- name: GetProjectBySlug :one
SELECT * FROM projects
WHERE slug = $1
LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects
ORDER BY created_at;

-- name: ListChildProjects :many
SELECT * FROM projects
WHERE parent_project_id = $1
ORDER BY created_at;
