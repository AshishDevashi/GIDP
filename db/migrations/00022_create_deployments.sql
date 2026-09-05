-- +goose Up
CREATE TABLE IF NOT EXISTS deployment (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_instance_id UUID NOT NULL REFERENCES deploymentinstance(id),
    repo_id                UUID REFERENCES repos(id),
    registry_id            UUID NOT NULL REFERENCES registries(id),
    image_name             VARCHAR(255) NOT NULL,
    image_tag              VARCHAR(255) NOT NULL,
    name                   VARCHAR(255) NOT NULL,
    namespace              VARCHAR(255) NOT NULL DEFAULT 'default',
    replicas               INTEGER NOT NULL DEFAULT 1,
    resources              JSONB NOT NULL DEFAULT '{"cpu":"250m","memory":"256Mi"}'::jsonb,
    env_vars               JSONB NOT NULL DEFAULT '{}'::jsonb,
    secret_refs            JSONB NOT NULL DEFAULT '{}'::jsonb,
    expose                 JSONB NOT NULL DEFAULT '{"type":"ClusterIP","port":80}'::jsonb,
    status                 VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending | deploying | running | failed | stopped
    current_revision       INTEGER NOT NULL DEFAULT 1,
    k8s_deployment_name    VARCHAR(255),
    last_error             TEXT,
    created_by             UUID NOT NULL REFERENCES users(id),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_deployment_status
    ON deployment (status);

CREATE INDEX IF NOT EXISTS idx_deployment_registry
    ON deployment (registry_id);

CREATE INDEX IF NOT EXISTS idx_deployment_instance
    ON deployment (deployment_instance_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_live_deployment_instance_namespace_name
    ON deployment (deployment_instance_id, namespace, name)
    WHERE status NOT IN ('stopped', 'failed');

CREATE TABLE IF NOT EXISTS deploymentrevision (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id   UUID NOT NULL REFERENCES deployment(id) ON DELETE CASCADE,
    revision_no     INTEGER NOT NULL,
    image_tag       VARCHAR(255) NOT NULL,
    config_snapshot JSONB NOT NULL,
    status          VARCHAR(20) NOT NULL,
    triggered_by    UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_deployment_revision_no
    ON deploymentrevision (deployment_id, revision_no);

-- +goose Down
DROP INDEX IF EXISTS uq_deployment_revision_no;
DROP TABLE IF EXISTS deploymentrevision;
DROP INDEX IF EXISTS uq_live_deployment_instance_namespace_name;
DROP INDEX IF EXISTS idx_deployment_instance;
DROP INDEX IF EXISTS idx_deployment_registry;
DROP INDEX IF EXISTS idx_deployment_status;
DROP TABLE IF EXISTS deployment;