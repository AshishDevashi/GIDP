-- +goose Up
CREATE TABLE IF NOT EXISTS services (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(150) NOT NULL,
    slug              VARCHAR(150) UNIQUE NOT NULL,
    description       TEXT,

    -- Classification
    service_type_id   SMALLINT NOT NULL REFERENCES service_types(id) DEFAULT 1,
    lifecycle_id      SMALLINT REFERENCES lifecycles(id) DEFAULT 2,
    tier_id           SMALLINT REFERENCES tiers(id),

    -- Ownership
    project_id        UUID REFERENCES projects(id),
    owner_team_id     UUID NOT NULL REFERENCES teams(id),
    tech_lead_id      UUID REFERENCES users(id),

    -- Source
    repo_url          TEXT NOT NULL,
    repo_provider_id  SMALLINT REFERENCES repo_providers(id),
    default_branch    VARCHAR(100) NOT NULL DEFAULT 'main',
    language_id       SMALLINT REFERENCES languages(id),
    framework         VARCHAR(100),

    -- Build / registry
    dockerfile_path   TEXT NOT NULL DEFAULT 'Dockerfile',
    registry_image    TEXT,
    ci_pipeline_url   TEXT,

    -- Deployment target
    gitops_repo_path  TEXT,
    k8s_resource_kind VARCHAR(50) NOT NULL DEFAULT 'Deployment',

    -- Networking / discovery
    port              INT,
    health_check_path VARCHAR(255) NOT NULL DEFAULT '/healthz',
    internal_url      TEXT,
    external_url      TEXT,

    -- Docs & observability
    api_spec_url      TEXT,
    dashboard_url     TEXT,
    runbook_url       TEXT,
    slo_target        NUMERIC(5,2),

    is_active         BOOLEAN NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS services;
