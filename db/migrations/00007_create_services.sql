-- +goose Up
-- Minimal placeholder for the future services/components catalog, needed as an
-- FK target for project_services. Expand into a full module when that catalog
-- is built out.
CREATE TABLE IF NOT EXISTS services (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(150) NOT NULL,
    slug       VARCHAR(150) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS services;
