-- +goose Up
-- Stage 32: per-run comments + @mention inbox.
CREATE TABLE run_comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id     UUID        NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX run_comments_run_idx ON run_comments (run_id, created_at);
CREATE INDEX run_comments_user_idx ON run_comments (user_id, created_at DESC);

-- One row per @mention parsed from a comment body. Lets the recipient see a
-- chronological inbox without scanning every comment they could possibly be
-- in.
CREATE TABLE mentions (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    run_id      UUID        NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    comment_id  UUID        NOT NULL REFERENCES run_comments(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at     TIMESTAMPTZ
);
CREATE INDEX mentions_user_unread_idx ON mentions (user_id, created_at DESC) WHERE read_at IS NULL;
CREATE INDEX mentions_user_idx ON mentions (user_id, created_at DESC);

-- +goose Down
DROP TABLE mentions;
DROP TABLE run_comments;
