-- name: AddTeamMember :one
INSERT INTO team_members (team_id, user_id, role_in_team, is_primary)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetActiveTeamMember :one
SELECT * FROM team_members
WHERE team_id = $1 AND user_id = $2 AND left_at IS NULL
LIMIT 1;

-- name: ListTeamMembers :many
SELECT * FROM team_members
WHERE team_id = $1 AND left_at IS NULL
ORDER BY joined_at;

-- name: RemoveTeamMember :exec
UPDATE team_members
SET left_at = now(), updated_at = now()
WHERE id = $1;
