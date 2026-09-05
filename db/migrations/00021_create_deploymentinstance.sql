-- +goose Up
CREATE TABLE IF NOT EXISTS deploymentinstance (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(255) NOT NULL,
    ec2_instance_id   VARCHAR(64),
    public_ip         VARCHAR(45),
    private_ip        VARCHAR(45),
    api_server_url    VARCHAR(255),
    auth_type         VARCHAR(20) NOT NULL DEFAULT 'kubeconfig',
    credentials_ref   VARCHAR(255),
    max_deployments   INTEGER NOT NULL DEFAULT 3,
    status            VARCHAR(20) NOT NULL DEFAULT 'provisioning',
    -- provisioning | active | stopped | terminated | failed
    last_error        TEXT,
    workspace         VARCHAR(255) NOT NULL,
    ssh_key_name      VARCHAR(255) NOT NULL DEFAULT '',
    security_group_id VARCHAR(100),
    created_by        UUID NOT NULL REFERENCES users(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_deploymentinstance_name
    ON deploymentinstance (name)
    WHERE status <> 'terminated';

CREATE UNIQUE INDEX IF NOT EXISTS uq_deploymentinstance_workspace
    ON deploymentinstance (workspace);

CREATE UNIQUE INDEX IF NOT EXISTS uq_single_live_deploymentinstance
    ON deploymentinstance ((true))
    WHERE status <> 'terminated';

-- +goose Down
DROP INDEX IF EXISTS uq_single_live_deploymentinstance;
DROP INDEX IF EXISTS uq_deploymentinstance_workspace;
DROP INDEX IF EXISTS uq_deploymentinstance_name;
DROP TABLE IF EXISTS deploymentinstance;