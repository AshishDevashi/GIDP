-- +goose Up
CREATE TABLE IF NOT EXISTS project_services (
    project_id UUID NOT NULL REFERENCES projects(id),
    service_id UUID NOT NULL REFERENCES services(id),
    PRIMARY KEY (project_id, service_id)
);

-- +goose Down
DROP TABLE IF EXISTS project_services;
