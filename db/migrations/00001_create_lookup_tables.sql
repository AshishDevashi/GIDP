-- +goose Up
-- All small, static reference/lookup tables live here (roles + project catalog metadata).

CREATE TABLE IF NOT EXISTS roles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO roles (name)
VALUES ('admin'), ('manager'), ('developer')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS lifecycles (
    id    SMALLINT PRIMARY KEY,
    code  VARCHAR(50) UNIQUE NOT NULL,
    label VARCHAR(100)
);

INSERT INTO lifecycles (id, code, label) VALUES
    (1, 'experimental', 'Experimental'),
    (2, 'production', 'Production'),
    (3, 'deprecated', 'Deprecated'),
    (4, 'retired', 'Retired')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS tiers (
    id            SMALLINT PRIMARY KEY,
    code          VARCHAR(20) UNIQUE NOT NULL,
    description   TEXT,
    paging_policy VARCHAR(50)
);

INSERT INTO tiers (id, code, description, paging_policy) VALUES
    (1, 'tier-1', 'Business critical', 'immediate'),
    (2, 'tier-2', 'Important', 'business_hours'),
    (3, 'tier-3', 'Non critical', 'none')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS service_types (
    id   SMALLINT PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL
);

INSERT INTO service_types (id, code) VALUES
    (1, 'backend'),
    (2, 'frontend'),
    (3, 'worker'),
    (4, 'cron_job'),
    (5, 'library')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS repo_providers (
    id   SMALLINT PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL
);

INSERT INTO repo_providers (id, code) VALUES
    (1, 'github')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS languages (
    id    SMALLINT PRIMARY KEY,
    code  VARCHAR(50) UNIQUE NOT NULL,
    label VARCHAR(100)
);

INSERT INTO languages (id, code, label) VALUES
    (1, 'go', 'Go'),
    (2, 'java', 'Java'),
    (3, 'python', 'Python'),
    (4, 'node', 'Node.js')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS repo_templates (
    id            SMALLINT PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    slug          VARCHAR(50)  NOT NULL UNIQUE,

    template_owner VARCHAR(150) NOT NULL,
    template_repo  VARCHAR(150) NOT NULL,

    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_template_owner_repo UNIQUE (template_owner, template_repo)
);

INSERT INTO repo_templates (id, name, slug, template_owner, template_repo) VALUES
    (1, 'Go App',     'go-app',     'AshishDevashi', 'go-app-template-idp'),
    (2, 'Python App', 'python-app', 'AshishDevashi', 'python-app-template-idp'),
    (3, 'Java App',   'java-app',   'AshishDevashi', 'java-app-template-idp'),
    (4, 'Node App',   'node-app',   'AshishDevashi', 'node-app-template-idp')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS repo_templates;
DROP TABLE IF EXISTS languages;
DROP TABLE IF EXISTS repo_providers;
DROP TABLE IF EXISTS service_types;
DROP TABLE IF EXISTS tiers;
DROP TABLE IF EXISTS lifecycles;
DROP TABLE IF EXISTS roles;
