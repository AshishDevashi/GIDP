-- name: CreateService :one
INSERT INTO services (
    name, slug, description, service_type_id, lifecycle_id, tier_id,
    project_id, owner_team_id, tech_lead_id,
    repo_url, repo_provider_id, default_branch, language_id, framework,
    dockerfile_path, registry_image, ci_pipeline_url,
    gitops_repo_path, k8s_resource_kind,
    port, health_check_path, internal_url, external_url,
    api_spec_url, dashboard_url, runbook_url, slo_target
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12, $13, $14,
    $15, $16, $17,
    $18, $19,
    $20, $21, $22, $23,
    $24, $25, $26, $27
)
RETURNING *;

-- name: GetServiceByID :one
SELECT * FROM services
WHERE id = $1
LIMIT 1;

-- name: GetServiceBySlug :one
SELECT * FROM services
WHERE slug = $1
LIMIT 1;

-- name: ListServices :many
SELECT * FROM services
ORDER BY created_at;

-- name: ListServicesByProject :many
SELECT * FROM services
WHERE project_id = $1
ORDER BY created_at;
