package run

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	wdagger "workend/api/internal/dagger"
)

// GenerateSBOMAsync runs syft against the project's working tree and stores
// the SPDX-JSON document on the images row. Lives next to the Trivy scan;
// both run after a successful image build, so we get a full provenance pair
// (vuln summary + SBOM) per image with one Dockerfile task.
func GenerateSBOMAsync(pool *pgxpool.Pool, dc *wdagger.Client, imageID uuid.UUID, repoPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if _, err := pool.Exec(ctx,
		`UPDATE images SET sbom_status = 'pending' WHERE id = $1`, imageID); err != nil {
		return
	}

	client, err := dc.Get(ctx)
	if err != nil {
		_, _ = pool.Exec(ctx, `UPDATE images SET sbom_status = 'error' WHERE id = $1`, imageID)
		return
	}

	repoDir := client.Host().Directory(repoPath)
	out, err := client.Container().
		From("anchore/syft:v1.16.0").
		WithMountedDirectory("/scan", repoDir).
		WithExec([]string{"syft", "/scan", "-o", "spdx-json", "-q"}).
		Stdout(ctx)
	if err != nil || out == "" {
		_, _ = pool.Exec(ctx, `UPDATE images SET sbom_status = 'error', sbom_generated_at = now() WHERE id = $1`, imageID)
		return
	}
	_, _ = pool.Exec(ctx, `
		UPDATE images
		SET sbom_status = 'ok', sbom_generated_at = now(),
		    sbom_format = 'spdx-json', sbom_data = $1::jsonb
		WHERE id = $2
	`, out, imageID)
}
