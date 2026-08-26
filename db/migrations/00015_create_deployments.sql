-- +goose Up
-- Append-only deploy history/audit log. Rows are never deleted; only `status`,
-- `started_at`, `completed_at`, and `failure_reason` are ever updated after insert.
CREATE TABLE IF NOT EXISTS deployments (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- What was deployed
    service_id         UUID NOT NULL REFERENCES services(id),
    environment        VARCHAR(50) NOT NULL,

    -- Version info
    image_tag          VARCHAR(255) NOT NULL,
    previous_image_tag VARCHAR(255),
    git_commit_sha     VARCHAR(100),
    git_branch         VARCHAR(150),

    -- Trigger
    triggered_by_user_id UUID REFERENCES users(id),
    trigger_type       VARCHAR(50) NOT NULL,
    ci_run_url         TEXT,

    -- Deployment mechanics
    deploy_strategy    VARCHAR(50) NOT NULL DEFAULT 'rolling',
    gitops_commit_sha  VARCHAR(100),

    -- Status
    status             VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at         TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ,
    failure_reason     TEXT,

    -- Rollback linkage
    is_rollback        BOOLEAN NOT NULL DEFAULT false,
    rolled_back_from_deployment_id UUID REFERENCES deployments(id),

    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS deployments;
