-- name: AddTeamMember :one
INSERT INTO team_members (team_id, user_id, role_in_team)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeamMember :one
SELECT * FROM team_members
WHERE team_id = $1 AND user_id = $2
LIMIT 1;

-- name: ListTeamMembers :many
SELECT * FROM team_members
WHERE team_id = $1 AND left_at IS NULL
ORDER BY joined_at;

-- name: RemoveTeamMember :exec
UPDATE team_members
SET left_at = now()
WHERE team_id = $1 AND user_id = $2;
