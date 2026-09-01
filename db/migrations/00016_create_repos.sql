-- +goose Up
CREATE TABLE IF NOT EXISTS repos (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name              VARCHAR(150) NOT NULL,
    full_name         VARCHAR(255),
    owner             VARCHAR(150) NOT NULL,

    provider_id       SMALLINT NOT NULL REFERENCES repo_providers(id),
    external_id       VARCHAR(100),

    url               TEXT,
    clone_url_ssh     TEXT,
    clone_url_https   TEXT,
    default_branch    VARCHAR(100) NOT NULL DEFAULT 'main',
    visibility        VARCHAR(20) NOT NULL DEFAULT 'private',

    template_used     VARCHAR(150),

    status            VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending | creating | active | failed | archived

    error_message     TEXT,

    created_by        UUID REFERENCES users(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_repo_provider_owner_name
        UNIQUE (provider_id, owner, name)
);

-- +goose Down
DROP TABLE IF EXISTS repos;