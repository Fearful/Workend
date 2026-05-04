-- +goose Up
-- Per-task concurrency control. max_concurrency caps how many runs of the
-- same task can be in non-terminal states at once (queued + running). 0 = no
-- limit. supersede_policy controls what happens when a new run starts while
-- the task is already at its cap:
--   'queue'      — let the new run wait (current behavior)
--   'cancel-old' — cancel the oldest running run of this task and proceed
--   'reject'     — fail the new run synchronously with a 'busy' status note
ALTER TABLE tasks
    ADD COLUMN max_concurrency  INT  NOT NULL DEFAULT 0,
    ADD COLUMN supersede_policy TEXT NOT NULL DEFAULT 'queue';

-- Sanity check: only known policy values.
ALTER TABLE tasks
    ADD CONSTRAINT tasks_supersede_policy_chk
    CHECK (supersede_policy IN ('queue','cancel-old','reject'));

-- Index used by the scheduler to count non-terminal runs of a task fast.
CREATE INDEX runs_task_active_idx
    ON runs (task_id)
    WHERE status IN ('queued','running','pending_approval');

-- +goose Down
DROP INDEX IF EXISTS runs_task_active_idx;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_supersede_policy_chk;
ALTER TABLE tasks DROP COLUMN IF EXISTS supersede_policy;
ALTER TABLE tasks DROP COLUMN IF EXISTS max_concurrency;
