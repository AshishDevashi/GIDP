-- name: CreateDeploymentInstance :one
INSERT INTO deploymentinstance (
    name, workspace, ssh_key_name, max_deployments, created_by
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetLiveDeploymentInstance :one
SELECT * FROM deploymentinstance
WHERE status <> 'terminated'
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkDeploymentInstanceProvisioned :one
UPDATE deploymentinstance
SET status = 'active',
    last_error = NULL,
    ec2_instance_id = $2,
    public_ip = $3,
    private_ip = $4,
    api_server_url = $5,
    credentials_ref = $6,
    security_group_id = $7,
    updated_at = now()
WHERE id = $1 AND status <> 'terminated'
RETURNING *;

-- name: MarkDeploymentInstanceStatus :execrows
UPDATE deploymentinstance
SET status = $2,
    last_error = $3,
    updated_at = now()
WHERE id = $1 AND status <> 'terminated';

-- name: MarkDeploymentInstanceTerminated :execrows
UPDATE deploymentinstance
SET status = 'terminated',
    credentials_ref = NULL,
    updated_at = now()
WHERE id = $1 AND status <> 'terminated';