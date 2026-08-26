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

    -- Catalog metadata
    lifecycle_id      SMALLINT NOT NULL REFERENCES lifecycles(id) DEFAULT 2,
    tier_id           SMALLINT REFERENCES tiers(id),

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
