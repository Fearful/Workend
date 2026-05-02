package detect

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// builtinDetectors are the always-on detectors compiled in to the binary.
// The Dagger detector needs a wdagger.Client; it's added via InitDetectors.
func builtinDetectors() []Detector {
	return []Detector{
		NPM{},
		Just{},
		Dockerfile{},
	}
}

// active is the registry of detectors actually used by Run().
// Populated by InitDetectors at startup with builtins + any user-defined
// custom detectors loaded from WORKEND_DETECTORS_FILE.
var active = builtinDetectors()

// InitDetectors swaps the active list to include user-defined custom
// detectors after the dagger client is available.
func InitDetectors(extras []Detector) {
	active = append(builtinDetectors(), extras...)
}

// AllDetectors returns the active set.
func AllDetectors() []Detector {
	return active
}

// Run runs every detector against the project's local clone, then atomically
// replaces the project's task rows. Old tasks not present in the new scan
// are deleted; new tasks are inserted; persistent task IDs are preserved
// (because we keep them keyed by project_id+source+name).
func Run(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger, projectID uuid.UUID, repoPath string) error {
	all := []Task{}
	for _, d := range AllDetectors() {
		tasks, err := d.Detect(ctx, repoPath)
		if err != nil {
			logger.Warn("detector failed", "detector", d.Name(), "project", projectID, "err", err)
			continue
		}
		all = append(all, tasks...)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete tasks that no longer exist
	if _, err := tx.Exec(ctx, `DELETE FROM tasks WHERE project_id = $1`, projectID); err != nil {
		return fmt.Errorf("delete old tasks: %w", err)
	}

	for _, t := range all {
		_, err := tx.Exec(ctx, `
			INSERT INTO tasks (project_id, source, name, raw_command)
			VALUES ($1, $2, $3, $4)
		`, projectID, t.Source, t.Name, t.RawCommand)
		if err != nil {
			return fmt.Errorf("insert task %s/%s: %w", t.Source, t.Name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	logger.Info("detection complete", "project", projectID, "count", len(all))
	return nil
}
