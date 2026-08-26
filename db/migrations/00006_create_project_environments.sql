-- +goose Up
CREATE TABLE IF NOT EXISTS project_environments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID NOT NULL REFERENCES projects(id),
    environment  VARCHAR(50) NOT NULL,
    cluster_name VARCHAR(100),
    namespace    VARCHAR(100),
    url          TEXT,
    replicas     INT NOT NULL DEFAULT 1,
    UNIQUE (project_id, environment)
);

-- +goose Down
DROP TABLE IF EXISTS project_environments;
