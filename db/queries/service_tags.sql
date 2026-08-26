-- name: AddServiceTag :exec
INSERT INTO service_tags (service_id, tag)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveServiceTag :exec
DELETE FROM service_tags
WHERE service_id = $1 AND tag = $2;

-- name: ListServiceTags :many
SELECT * FROM service_tags
WHERE service_id = $1
ORDER BY tag;
