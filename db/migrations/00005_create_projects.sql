-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(150) NOT NULL,
    slug              VARCHAR(150) UNIQUE NOT NULL,

    description       TEXT,

    -- Type & architecture
    project_type      VARCHAR(50) NOT NULL DEFAULT 'service',
    architecture      VARCHAR(50),

    -- Ownership
    owner_team_id     UUID NOT NULL REFERENCES teams(id),
    tech_lead_id      UUID REFERENCES users(id),

    -- Source & CI/CD linkage
    repo_url          TEXT,
    repo_provider     VARCHAR(50),
    default_branch    VARCHAR(100) NOT NULL DEFAULT 'main',
    ci_pipeline_url   TEXT,
    gitops_path       TEXT,

    -- Catalog metadata
    lifecycle         VARCHAR(50) NOT NULL DEFAULT 'production',
    tier              VARCHAR(20),
    language          VARCHAR(50),
    framework         VARCHAR(100),

    -- Docs & observability links
    docs_url          TEXT,
    dashboard_url     TEXT,
    runbook_url       TEXT,

    -- Grouping (for microservices under one project/system)
    parent_project_id UUID REFERENCES projects(id),

    is_active         BOOLEAN NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS projects;
