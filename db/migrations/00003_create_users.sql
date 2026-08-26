-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username          VARCHAR(100) UNIQUE NOT NULL,
    email             VARCHAR(255) UNIQUE NOT NULL,
    password_hash     TEXT NOT NULL,
    full_name         VARCHAR(255),
    avatar_url        TEXT,

    -- Org structure
    manager_id        UUID REFERENCES users(id),
    role_id           UUID NOT NULL REFERENCES roles(id),

    -- Status
    is_active         BOOLEAN NOT NULL DEFAULT true,
    last_login_at     TIMESTAMPTZ,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS users;
