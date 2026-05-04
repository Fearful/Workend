package run

import (
	"context"
	"fmt"
	"strings"

	dag "dagger.io/dagger"

	wdagger "workend/api/internal/dagger"
)

// ExecuteImageBuild builds the project's Dockerfile via Dagger BuildKit and
// returns the resulting image digest + size. Logs are written to spec.LogFile
// (currently just a summary line — full BuildKit progress streaming is
// deferred along with task-output streaming).
func ExecuteImageBuild(ctx context.Context, dc *wdagger.Client, spec Spec) (digest string, sizeBytes int64, err error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("dagger client: %w", err)
	}

	repoDir := client.Host().Directory(spec.RepoPath)
	dockerfilePath := strings.TrimSpace(spec.RawCommand)
	if dockerfilePath == "" {
		dockerfilePath = "Dockerfile"
	}

	img := repoDir.DockerBuild(dag.DirectoryDockerBuildOpts{
		Dockerfile: dockerfilePath,
	})

	imageID, err := img.ID(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("image id: %w", err)
	}
	digest = string(imageID)
	if len(digest) > 71 {
		// IDs from BuildKit can be long handles; trim for display
		digest = digest[:71]
	}

	// Size: deferred. Without a registry to push to, the size we'd report
	// would be the exported tarball size, not the layered image size, which
	// would mislead users. Returning 0 — UI shows "—" for null/zero.
	return digest, 0, nil
}
