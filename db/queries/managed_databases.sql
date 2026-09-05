-- name: CreateManagedDatabase :one
INSERT INTO managed_databases (
    db_instance_id, name, username, password, allocated_mb,
    status, connection_string, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetManagedDatabaseByID :one
SELECT * FROM managed_databases
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetManagedDatabaseByName :one
SELECT * FROM managed_databases
WHERE db_instance_id = $1 AND name = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: ListManagedDatabases :many
SELECT * FROM managed_databases
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListManagedDatabasesByInstanceID :many
SELECT * FROM managed_databases
WHERE db_instance_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListActiveManagedDatabasesByInstanceID :many
SELECT * FROM managed_databases
WHERE db_instance_id = $1 AND status = 'active' AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: GetTotalAllocatedMB :one
SELECT COALESCE(SUM(allocated_mb), 0)::BIGINT AS total_allocated_mb
FROM managed_databases
WHERE status = 'active' AND deleted_at IS NULL;

-- name: GetTotalAllocatedMBByInstanceID :one
SELECT COALESCE(SUM(allocated_mb), 0)::BIGINT AS total_allocated_mb
FROM managed_databases
WHERE db_instance_id = $1 AND status = 'active' AND deleted_at IS NULL;

-- name: MarkManagedDatabaseStatus :execrows
UPDATE managed_databases
SET status = $2,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteManagedDatabase :execrows
UPDATE managed_databases
SET deleted_at = now(),
    status = 'deleted',
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteManagedDatabasesByInstanceID :execrows
UPDATE managed_databases
SET deleted_at = now(),
    status = 'deleted',
    updated_at = now()
WHERE db_instance_id = $1 AND deleted_at IS NULL;
