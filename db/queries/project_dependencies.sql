-- name: AddProjectDependency :one
INSERT INTO project_dependencies (project_id, depends_on_project_id, dependency_type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListProjectDependencies :many
-- Projects that the given project depends on.
SELECT * FROM project_dependencies
WHERE project_id = $1
ORDER BY created_at;

-- name: ListDependentProjects :many
-- Projects that depend on the given project (impact check before deletion).
SELECT * FROM project_dependencies
WHERE depends_on_project_id = $1
ORDER BY created_at;

-- name: RemoveProjectDependency :exec
DELETE FROM project_dependencies
WHERE project_id = $1 AND depends_on_project_id = $2;
