-- +goose Up
CREATE TABLE service_dependencies (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    target_project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    dep_type          TEXT NOT NULL, -- 'dockerfile_from', 'compose_service', 'compose_image', 'dagger_module', 'config_reference'
    reference         TEXT NOT NULL, -- the raw string found (e.g. "myapp:latest", "./other-service")
    file_path         TEXT NOT NULL DEFAULT '', -- relative path where dependency was found
    discovered_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_project_id, dep_type, reference)
);
CREATE INDEX service_deps_source_idx ON service_dependencies (source_project_id);
CREATE INDEX service_deps_target_idx ON service_dependencies (target_project_id) WHERE target_project_id IS NOT NULL;

-- +goose Down
DROP TABLE service_dependencies;
