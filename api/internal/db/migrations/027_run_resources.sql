-- +goose Up
-- Stage 38: per-run resource usage. Best-effort capture from /proc/self
-- inside the Dagger container; columns are nullable so older runs (and
-- runs whose base image lacks /proc) don't break.
ALTER TABLE runs
    ADD COLUMN cpu_ms          BIGINT,
    ADD COLUMN mem_peak_bytes  BIGINT,
    ADD COLUMN net_rx_bytes    BIGINT,
    ADD COLUMN net_tx_bytes    BIGINT;

-- +goose Down
ALTER TABLE runs DROP COLUMN net_tx_bytes;
ALTER TABLE runs DROP COLUMN net_rx_bytes;
ALTER TABLE runs DROP COLUMN mem_peak_bytes;
ALTER TABLE runs DROP COLUMN cpu_ms;
