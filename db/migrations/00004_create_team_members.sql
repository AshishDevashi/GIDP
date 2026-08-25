-- +goose Up
CREATE TABLE IF NOT EXISTS team_members (
    team_id      UUID NOT NULL REFERENCES teams(id),
    user_id      UUID NOT NULL REFERENCES users(id),
    role_in_team VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    left_at      TIMESTAMPTZ,
    PRIMARY KEY (team_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS team_members;
