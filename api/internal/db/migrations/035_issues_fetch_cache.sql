-- +goose Up
ALTER TABLE projects ADD COLUMN last_issues_fetch_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE projects DROP COLUMN last_issues_fetch_at;
