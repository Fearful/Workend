-- +goose Up
-- Stage 26: per-user disk quota. 5 GiB default; admin can adjust per row.
ALTER TABLE users ADD COLUMN quota_bytes BIGINT NOT NULL DEFAULT 5368709120;

-- +goose Down
ALTER TABLE users DROP COLUMN quota_bytes;
