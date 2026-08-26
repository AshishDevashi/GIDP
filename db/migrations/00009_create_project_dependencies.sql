-- +goose Up
CREATE TABLE IF NOT EXISTS project_dependencies (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id            UUID NOT NULL REFERENCES projects(id),
    depends_on_project_id UUID NOT NULL REFERENCES projects(id),
    dependency_type       VARCHAR(50) NOT NULL DEFAULT 'runtime',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, depends_on_project_id)
);

-- +goose Down
DROP TABLE IF EXISTS project_dependencies;
