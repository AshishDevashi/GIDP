-- +goose Up
CREATE TABLE IF NOT EXISTS managed_databases (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    db_instance_id    UUID NOT NULL REFERENCES db_instances(id) ON DELETE CASCADE,

    name              VARCHAR(63) NOT NULL,
    username          VARCHAR(63) NOT NULL,
    password          VARCHAR(255) NOT NULL,

    allocated_mb      INTEGER NOT NULL,
    status            VARCHAR(30) NOT NULL DEFAULT 'active',
    -- active | deleting | deleted | failed

    connection_string TEXT NOT NULL DEFAULT '',

    created_by        UUID REFERENCES users(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

-- Partial index so database names are unique per db_instance among active databases.
CREATE UNIQUE INDEX IF NOT EXISTS uq_managed_database_instance_name
    ON managed_databases (db_instance_id, name)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_managed_databases_instance_id
    ON managed_databases (db_instance_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_managed_databases_status
    ON managed_databases (status);

-- +goose Down
DROP INDEX IF EXISTS idx_managed_databases_status;
DROP INDEX IF EXISTS idx_managed_databases_instance_id;
DROP INDEX IF EXISTS uq_managed_database_instance_name;
DROP TABLE IF EXISTS managed_databases;
