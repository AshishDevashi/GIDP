-- +goose Up
CREATE TABLE IF NOT EXISTS service_environments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id        UUID NOT NULL REFERENCES services(id),
    environment       VARCHAR(50) NOT NULL,
    cluster_name      VARCHAR(100),
    namespace         VARCHAR(100),
    replicas_min      INT NOT NULL DEFAULT 1,
    replicas_max      INT NOT NULL DEFAULT 1,
    cpu_request       VARCHAR(20),
    memory_request    VARCHAR(20),
    current_image_tag VARCHAR(255),
    url               TEXT,
    UNIQUE (service_id, environment)
);

-- +goose Down
DROP TABLE IF EXISTS service_environments;
