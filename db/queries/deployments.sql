-- name: CreateDeployment :one
INSERT INTO deployment (
    deployment_instance_id, repo_id, registry_id, image_name, image_tag, name,
    namespace, replicas, resources, env_vars, secret_refs, expose, status,
    current_revision, k8s_deployment_name, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'pending', 1, $13, $14
)
RETURNING *;

-- name: CreateDeploymentRevision :one
INSERT INTO deploymentrevision (
    deployment_id, revision_no, image_tag, config_snapshot, status, triggered_by
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: CountActiveDeploymentsByInstanceID :one
SELECT count(*) FROM deployment
WHERE deployment_instance_id = $1
  AND status NOT IN ('stopped', 'failed');

-- name: GetDeploymentByID :one
SELECT * FROM deployment
WHERE id = $1
LIMIT 1;

-- name: ListDeployments :many
SELECT * FROM deployment
ORDER BY created_at DESC;

-- name: GetLatestDeploymentRevision :one
SELECT * FROM deploymentrevision
WHERE deployment_id = $1
ORDER BY revision_no DESC
LIMIT 1;

-- name: MarkDeploymentStatus :execrows
UPDATE deployment
SET status = $2,
    last_error = $3,
    updated_at = now()
WHERE id = $1;

-- name: MarkDeploymentDeploying :one
UPDATE deployment
SET status = 'deploying',
    last_error = NULL,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDeployment :execrows
DELETE FROM deployment
WHERE id = $1;