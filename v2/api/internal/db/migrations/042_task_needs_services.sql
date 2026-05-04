-- +goose Up
-- Tasks can declare a list of compose service names they need running. The
-- run handler ensures the project's compose stack is started before the
-- task's command executes; compose teardown remains a manual operation
-- (Stop button in the UI) so we don't tear down between back-to-back runs.
ALTER TABLE tasks
    ADD COLUMN needs_services TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[];

-- +goose Down
ALTER TABLE tasks DROP COLUMN IF EXISTS needs_services;
