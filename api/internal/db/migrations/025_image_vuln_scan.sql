-- +goose Up
-- Stage 36: Trivy-based vulnerability scan for images built by the
-- Dockerfile task source. Stored as a JSONB summary alongside each image.
ALTER TABLE images
    ADD COLUMN scan_status      TEXT,        -- 'pending' | 'ok' | 'error'
    ADD COLUMN scan_completed_at TIMESTAMPTZ,
    ADD COLUMN vuln_summary      JSONB;     -- {critical, high, medium, low, top: [{id, pkg, severity}]}

-- +goose Down
ALTER TABLE images DROP COLUMN vuln_summary;
ALTER TABLE images DROP COLUMN scan_completed_at;
ALTER TABLE images DROP COLUMN scan_status;
