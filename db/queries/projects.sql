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

-- name: UpdateProject :one
UPDATE projects
SET name = $2,
    slug = $3,
    description = $4,
    project_type = $5,
    architecture = $6,
    owner_team_id = $7,
    tech_lead_id = $8,
    lifecycle_id = $9,
    tier_id = $10,
    docs_url = $11,
    dashboard_url = $12,
    runbook_url = $13,
    parent_project_id = $14,
    is_active = $15,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteProject :execrows
DELETE FROM projects
WHERE id = $1;
