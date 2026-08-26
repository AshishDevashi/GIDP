-- name: CreateDeployment :one
INSERT INTO deployments (
    service_id, environment, image_tag, previous_image_tag, git_commit_sha, git_branch,
    triggered_by_user_id, trigger_type, ci_run_url, deploy_strategy, gitops_commit_sha,
    status, started_at, is_rollback, rolled_back_from_deployment_id
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11,
    $12, $13, $14, $15
)
RETURNING *;

-- name: UpdateDeploymentStatus :one
-- The only mutation ever allowed on a deployment row after creation.
UPDATE deployments
SET status = $2, completed_at = $3, failure_reason = $4
WHERE id = $1
RETURNING *;

-- name: GetDeploymentByID :one
SELECT * FROM deployments
WHERE id = $1
LIMIT 1;

-- name: ListDeploymentsByService :many
SELECT * FROM deployments
WHERE service_id = $1
ORDER BY created_at DESC;

-- name: ListDeploymentsByServiceEnvironment :many
SELECT * FROM deployments
WHERE service_id = $1 AND environment = $2
ORDER BY created_at DESC;

-- name: GetCurrentDeployment :one
-- Latest successful deployment for a service+environment, i.e. what's actually running.
SELECT * FROM deployments
WHERE service_id = $1 AND environment = $2 AND status = 'succeeded'
ORDER BY completed_at DESC
LIMIT 1;
