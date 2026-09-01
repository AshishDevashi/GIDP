-- name: CreateRepo :one
INSERT INTO repos (
    name, owner, provider_id, visibility, created_by, status
) VALUES (
    $1, $2, $3, $4, $5, 'pending'
)
RETURNING *;

-- name: MarkRepoCreating :one
UPDATE repos
SET status = 'creating', error_message = NULL, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ActivateRepo :one
UPDATE repos
SET full_name = $2,
    owner = $3,
    external_id = $4,
    url = $5,
    clone_url_ssh = $6,
    clone_url_https = $7,
    default_branch = $8,
    visibility = $9,
    status = 'active',
    error_message = NULL,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: FailRepo :one
UPDATE repos
SET status = 'failed', error_message = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetRepoByID :one
SELECT * FROM repos
WHERE id = $1
LIMIT 1;

-- name: ListRepos :many
SELECT * FROM repos
ORDER BY created_at DESC;

-- name: UpdateRepo :one
UPDATE repos
SET name = $2,
    default_branch = $3,
    visibility = $4,
    template_used = $5,
    status = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteRepo :execrows
DELETE FROM repos
WHERE id = $1;