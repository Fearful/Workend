-- +goose Up
ALTER TABLE audit_log ADD COLUMN retained_until TIMESTAMPTZ;
CREATE INDEX audit_log_retained_idx ON audit_log (retained_until) WHERE retained_until IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS audit_log_retained_idx;
ALTER TABLE audit_log DROP COLUMN retained_until;
