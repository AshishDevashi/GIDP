-- +goose Up
CREATE TABLE IF NOT EXISTS db_instances (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name                 VARCHAR(255) NOT NULL,
    description          VARCHAR(255) NOT NULL DEFAULT '',

    engine               VARCHAR(30) NOT NULL DEFAULT 'postgres',
    engine_version       VARCHAR(20) NOT NULL DEFAULT '16',

    provider             VARCHAR(30) NOT NULL DEFAULT 'aws',
    region               VARCHAR(30) NOT NULL,
    instance_type        VARCHAR(30) NOT NULL,
    storage_gb           INTEGER NOT NULL DEFAULT 20,

    status               VARCHAR(30) NOT NULL DEFAULT 'provisioning',
    -- provisioning | running | failed | deleting | deleted
    container_status     VARCHAR(30) NOT NULL DEFAULT 'pending',
    -- pending | running | failed
    status_message       TEXT,

    workspace            VARCHAR(255) NOT NULL,
    ssh_key_name         VARCHAR(255) NOT NULL DEFAULT '',

    provider_instance_id VARCHAR(100),
    availability_zone    VARCHAR(30),
    public_ip            VARCHAR(45),
    private_ip           VARCHAR(45),
    security_group_id    VARCHAR(100),
    volume_id            VARCHAR(100),

    created_by           UUID REFERENCES users(id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);

-- Partial index so a soft-deleted db instance name can be reused.
CREATE UNIQUE INDEX IF NOT EXISTS uq_db_instance_name
    ON db_instances (name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_db_instance_workspace
    ON db_instances (workspace);

-- +goose Down
DROP INDEX IF EXISTS uq_db_instance_workspace;
DROP INDEX IF EXISTS uq_db_instance_name;
DROP TABLE IF EXISTS db_instances;
