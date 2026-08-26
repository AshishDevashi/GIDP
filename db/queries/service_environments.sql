-- name: AddServiceEnvironment :one
INSERT INTO service_environments (
    service_id, environment, cluster_name, namespace,
    replicas_min, replicas_max, cpu_request, memory_request, current_image_tag, url
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetServiceEnvironment :one
SELECT * FROM service_environments
WHERE service_id = $1 AND environment = $2
LIMIT 1;

-- name: ListServiceEnvironments :many
SELECT * FROM service_environments
WHERE service_id = $1
ORDER BY environment;
