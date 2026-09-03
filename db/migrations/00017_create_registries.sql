-- +goose Up
CREATE TABLE IF NOT EXISTS registry_providers (
    id    SMALLINT PRIMARY KEY,
    code  VARCHAR(50) UNIQUE NOT NULL,
    label VARCHAR(100)
);

INSERT INTO registry_providers (id, code, label) VALUES
    (1, 'dockerhub', 'Docker Hub')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS registries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name          VARCHAR(255) NOT NULL,
    description   VARCHAR(255) NOT NULL DEFAULT '',

    provider_id   SMALLINT NOT NULL REFERENCES registry_providers(id),
    namespace     VARCHAR(255) NOT NULL,
    registry_url  VARCHAR(512),

    visibility    VARCHAR(20) NOT NULL DEFAULT 'private',
    -- public | private

    status        VARCHAR(20) NOT NULL DEFAULT 'active',
    -- active | failed | archived

    created_by    UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- Partial index so a soft-deleted registry name can be reused.
CREATE UNIQUE INDEX IF NOT EXISTS uq_registry_provider_namespace_name
    ON registries (provider_id, namespace, name)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_registry_provider_namespace_name;
DROP TABLE IF EXISTS registries;
DROP TABLE IF EXISTS registry_providers;
