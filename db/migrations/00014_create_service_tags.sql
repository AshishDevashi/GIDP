-- +goose Up
CREATE TABLE IF NOT EXISTS service_tags (
    service_id UUID NOT NULL REFERENCES services(id),
    tag        VARCHAR(100) NOT NULL,
    PRIMARY KEY (service_id, tag)
);

-- +goose Down
DROP TABLE IF EXISTS service_tags;
