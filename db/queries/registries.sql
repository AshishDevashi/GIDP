-- name: CreateRegistry :one
INSERT INTO registries (
    name, description, provider_id, namespace, registry_url, visibility, status, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, 'active', $7
)
RETURNING *;

-- name: GetRegistryByID :one
SELECT * FROM registries
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetRegistryByName :one
SELECT * FROM registries
WHERE provider_id = $1 AND namespace = $2 AND name = $3 AND deleted_at IS NULL
LIMIT 1;

-- name: ListRegistries :many
SELECT * FROM registries
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateRegistry :one
UPDATE registries
SET description = $2,
    visibility = $3,
    status = $4,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteRegistry :execrows
UPDATE registries
SET deleted_at = now(), status = 'archived', updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
