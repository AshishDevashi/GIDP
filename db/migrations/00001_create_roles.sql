-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO roles (name)
VALUES ('admin'), ('manager'), ('developer')
ON CONFLICT (name) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS roles;
