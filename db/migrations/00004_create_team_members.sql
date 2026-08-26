-- +goose Up
CREATE TABLE IF NOT EXISTS team_members (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id      UUID NOT NULL REFERENCES teams(id),
    user_id      UUID NOT NULL REFERENCES users(id),
    role_in_team VARCHAR(50) NOT NULL DEFAULT 'member',
    is_primary   BOOLEAN NOT NULL DEFAULT false,
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    left_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS team_members;
