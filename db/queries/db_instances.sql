-- name: CreateDBInstance :one
INSERT INTO db_instances (
    name, description, engine, engine_version, provider, region, instance_type,
    storage_gb, workspace, ssh_key_name, admin_username, admin_secret_name,
    postgres_port, postgres_image, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: CountActiveDBInstances :one
SELECT count(*) FROM db_instances
WHERE deleted_at IS NULL;

-- name: GetDBInstanceByID :one
SELECT * FROM db_instances
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetDBInstanceByName :one
SELECT * FROM db_instances
WHERE name = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListDBInstances :many
SELECT * FROM db_instances
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: MarkDBInstanceProvisioned :one
UPDATE db_instances
SET status = 'running',
    status_message = NULL,
    provider_instance_id = $2,
    availability_zone = $3,
    public_ip = $4,
    private_ip = $5,
    security_group_id = $6,
    volume_id = $7,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: MarkDBInstanceStatus :execrows
UPDATE db_instances
SET status = $2,
    status_message = $3,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: MarkDBInstanceContainerStatus :execrows
UPDATE db_instances
SET container_status = $2,
    status_message = $3,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteDBInstance :execrows
UPDATE db_instances
SET deleted_at = now(), status = 'deleted', updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
