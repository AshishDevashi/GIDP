-- name: AddServiceDependency :one
INSERT INTO service_dependencies (service_id, depends_on_service_id, dependency_type, is_critical)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListServiceDependencies :many
-- Services that the given service depends on.
SELECT * FROM service_dependencies
WHERE service_id = $1;

-- name: ListDependentServices :many
-- Services that depend on the given service (impact check before deletion).
SELECT * FROM service_dependencies
WHERE depends_on_service_id = $1;

-- name: RemoveServiceDependency :exec
DELETE FROM service_dependencies
WHERE service_id = $1 AND depends_on_service_id = $2;
