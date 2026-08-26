-- name: AddProjectEnvironment :one
INSERT INTO project_environments (project_id, environment, cluster_name, namespace, url, replicas)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProjectEnvironment :one
SELECT * FROM project_environments
WHERE project_id = $1 AND environment = $2
LIMIT 1;

-- name: ListProjectEnvironments :many
SELECT * FROM project_environments
WHERE project_id = $1
ORDER BY environment;
