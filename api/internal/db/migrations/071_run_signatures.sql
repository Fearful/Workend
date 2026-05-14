-- +goose Up
CREATE TABLE signing_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    public_key   TEXT NOT NULL,
    key_hash     TEXT NOT NULL,
    created_by   UUID NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at   TIMESTAMPTZ,
    UNIQUE (workspace_id, key_hash)
);

CREATE TABLE run_signatures (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id      UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    key_id      UUID NOT NULL REFERENCES signing_keys(id),
    signature   TEXT NOT NULL,
    digest      TEXT NOT NULL,
    signed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    provenance  JSONB NOT NULL DEFAULT '{}',
    UNIQUE (run_id)
);
CREATE INDEX run_signatures_run_idx ON run_signatures (run_id);

-- +goose Down
DROP TABLE run_signatures;
DROP TABLE signing_keys;
