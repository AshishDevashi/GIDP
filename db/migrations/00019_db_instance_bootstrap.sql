-- +goose Up
ALTER TABLE db_instances
    ADD COLUMN IF NOT EXISTS admin_username    VARCHAR(63) NOT NULL DEFAULT 'gidp',
    ADD COLUMN IF NOT EXISTS admin_secret_name VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS postgres_port     INTEGER NOT NULL DEFAULT 5432,
    ADD COLUMN IF NOT EXISTS postgres_image    VARCHAR(255) NOT NULL DEFAULT '';

-- Only a single live DB instance may exist at a time. The constant index
-- expression makes the whole "not soft-deleted" set collapse to one key.
CREATE UNIQUE INDEX IF NOT EXISTS uq_single_active_db_instance
    ON db_instances ((true))
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_single_active_db_instance;

ALTER TABLE db_instances
    DROP COLUMN IF EXISTS postgres_image,
    DROP COLUMN IF EXISTS postgres_port,
    DROP COLUMN IF EXISTS admin_secret_name,
    DROP COLUMN IF EXISTS admin_username;
