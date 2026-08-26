-- name: LinkProjectService :exec
INSERT INTO project_services (project_id, service_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnlinkProjectService :exec
DELETE FROM project_services
WHERE project_id = $1 AND service_id = $2;

-- name: ListProjectServices :many
SELECT * FROM project_services
WHERE project_id = $1;
