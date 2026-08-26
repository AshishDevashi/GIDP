-- +goose Up
CREATE TABLE IF NOT EXISTS service_dependencies (
    service_id            UUID NOT NULL REFERENCES services(id),
    depends_on_service_id UUID NOT NULL REFERENCES services(id),
    dependency_type       VARCHAR(50),
    is_critical           BOOLEAN NOT NULL DEFAULT true,
    PRIMARY KEY (service_id, depends_on_service_id)
);

-- +goose Down
DROP TABLE IF EXISTS service_dependencies;
