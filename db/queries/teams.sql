-- name: CreateTeam :one
INSERT INTO teams (name, slug, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams
WHERE id = $1
LIMIT 1;

-- name: GetTeamBySlug :one
SELECT * FROM teams
WHERE slug = $1
LIMIT 1;

-- name: ListTeams :many
SELECT * FROM teams
ORDER BY created_at;
