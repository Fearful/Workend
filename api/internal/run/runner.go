// Package run executes detected tasks via Dagger pipelines.
//
// Each Run is a single invocation of a Task. The runner constructs a Dagger
// container from the source-appropriate base image, mounts the project's
// cloned tree, and executes the raw command. Combined stdout+stderr is
// captured to a log file on the logs volume; the exit code goes back into
// the runs table.
package run

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	dag "dagger.io/dagger"

	wdagger "workend/api/internal/dagger"
)

// Spec captures everything the runner needs to execute a single task. Caller
// resolves these fields by joining the runs / tasks / projects tables.
type Spec struct {
	Source     string // 'npm' | 'pnpm' | 'yarn' | 'bun' | 'just' | 'dagger' | 'dockerfile'
	Name       string // task name
	RawCommand string // verbatim command to run inside the container
	RepoPath   string // absolute path to the cloned repo on the api volume
	LogFile    string // absolute path to write combined stdout+stderr to
	TimeoutSec int    // hard deadline for this run; 0 means caller's default

	// Per-run user inputs. Env keys/values must match [A-Z_][A-Z0-9_]* / safe
	// strings; ExtraArgs are appended to the command verbatim. Validated at
	// the HTTP boundary, not here.
	Env       map[string]string
	ExtraArgs []string

	// Pin-to-commit. When CommitSHA is non-empty, the runner ignores RepoPath
	// and instead fetches the tree at that exact commit via the Dagger Git
	// API. GitURL is the repo URL (token-injected by the caller if private).
	CommitSHA string
	GitURL    string

	// Artifact patterns to capture from RepoPath after the run finishes.
	// Empty = skip capture. See artifact.Capture for the supported globs.
	ArtifactPatterns []string

	// BaseImage overrides the default container image for this task when
	// non-empty. Set via the task's base_image column.
	BaseImage string
}

// Result is what Execute returns on completion.
type Result struct {
	ExitCode  int
	Resources Resources
}

// Resources is best-effort resource accounting for a run. Captured by
// reading /proc inside the Dagger container after the task's main command
// finishes. Zero values are perfectly acceptable when the base image lacks
// /proc or when the wrapper script can't run.
type Resources struct {
	CPUMs       int64 // total CPU time in milliseconds
	MemPeakBytes int64
	NetRxBytes   int64
	NetTxBytes   int64
}

// Execute runs the task's command in a Dagger-managed container. Returns
// (Result, nil) on container completion regardless of exit code; returns
// (zero, error) only on infrastructure failures (engine unreachable,
// log write failed, etc.).
func Execute(ctx context.Context, dc *wdagger.Client, spec Spec) (Result, error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("dagger client: %w", err)
	}

	base := baseImage(spec.Source, spec.RepoPath)
	if spec.BaseImage != "" {
		base = spec.BaseImage
	}
	setup := setupSteps(spec.Source)

	container := client.Container().From(base)

	for _, step := range setup {
		container = container.WithExec(step)
	}

	// Source the working tree: either the locally-cloned (HEAD) directory or
	// — for re-runs pinned to a specific commit — a freshly-resolved tree at
	// that commit, fetched by the Dagger Git API.
	var workTree *dag.Directory
	if spec.CommitSHA != "" && spec.GitURL != "" {
		workTree = client.Git(spec.GitURL).Commit(spec.CommitSHA).Tree()
	} else {
		workTree = client.Host().Directory(spec.RepoPath)
	}

	container = container.
		WithMountedDirectory("/work", workTree).
		WithWorkdir("/work").
		WithEnvVariable("CACHE_BUSTER", fmt.Sprintf("%d", time.Now().UnixNano()))

	container = mountLanguageCaches(client, container, spec)

	// Apply user-provided env vars (validated by caller).
	for k, v := range spec.Env {
		container = container.WithEnvVariable(k, v)
	}

	rawCmd := spec.RawCommand
	if len(spec.ExtraArgs) > 0 {
		rawCmd = rawCmd + " " + shellQuoteArgs(spec.ExtraArgs)
	}

	// Auto-inject install step for Node.js projects because monorepo node_modules
	// are scattered and cannot be fully restored by a single root CacheVolume.
	cmd := rawCmd
	needsInstall := !strings.Contains(rawCmd, " install") && !strings.Contains(rawCmd, " i ")
	if needsInstall {
		switch spec.Source {
		case "npm":
			cmd = "npm install && " + rawCmd
		case "pnpm":
			cmd = "pnpm install && " + rawCmd
		case "yarn":
			cmd = "yarn install && " + rawCmd
		case "bun":
			cmd = "bun install && " + rawCmd
		}
	}

	// Stage 23 incremental log streaming: write the live output to a Dagger
	// cache volume that's also visible to a sidecar polling container. The
	// SSE handler (run.Stream) already tails spec.LogFile by-offset, so all
	// we need to do is keep that file growing as the run progresses.
	streamID := filepath.Base(spec.LogFile) // unique-per-run already
	logCache := client.CacheVolume("workend-runlog-" + streamID)

	wrapped := fmt.Sprintf(`set -o pipefail
echo "STARTING WRAPPER SCRIPT" 2>&1 | tee -a /shared/out.log
(%s) 2>&1 | tee -a /shared/out.log
__rc=$?
{
  awk 'NR==1{printf "cpu_ms=%%d\n", ($14+$15)*10/1}' /proc/self/stat 2>/dev/null
  awk '/VmHWM:/{printf "mem_peak_kb=%%d\n", $2}' /proc/self/status 2>/dev/null
  awk '/^[ \t]*(eth|en|wl)[a-z0-9]+:/{rx+=$2; tx+=$10} END{printf "net_rx=%%d\nnet_tx=%%d\n", rx, tx}' /proc/net/dev 2>/dev/null
} > /tmp/workend.metrics 2>/dev/null
echo "FINISHING WRAPPER SCRIPT" 2>&1 | tee -a /shared/out.log
echo __workend_done__ >> /shared/out.log
exit $__rc
`, cmd)

	container = container.
		WithMountedCache("/shared", logCache, dag.ContainerWithMountedCacheOpts{
			Sharing: dag.CacheSharingModeShared,
		}).
		WithExec(
			[]string{"sh", "-c", wrapped},
			dag.ContainerWithExecOpts{Expect: dag.ReturnTypeAny},
		)

	// Run the producer in a goroutine so the polling loop can read the cache
	// volume mid-execution.
	doneCh := make(chan prodResult, 1)
	go func() {
		exitCode, err := container.ExitCode(ctx)
		if err != nil {
			doneCh <- prodResult{err: fmt.Errorf("read exit code: %w", err)}
			return
		}

		stderr, _ := container.Stderr(ctx)
		stdout, _ := container.Stdout(ctx)

		var res Resources
		if metricsRaw, err := container.File("/tmp/workend.metrics").Contents(ctx); err == nil {
			res = parseMetrics(metricsRaw)
		}
		doneCh <- prodResult{exit: exitCode, res: res, stderr: stderr, stdout: stdout}
	}()

	// Polling loop: pulls /shared/out.log from a separate exec against the
	// same cache volume every ~1s, slices off the bytes we've already
	// flushed to spec.LogFile, and appends new content. The producer writes
	// `__workend_done__\n` as its very last action so we can stop polling
	// without racing the exit-code read.
	resPtr, err := streamLog(ctx, dc, client, logCache, spec.LogFile, doneCh)
	if err != nil {
		// The producer may still finish; honor its result.
		select {
		case res := <-doneCh:
			if res.err != nil {
				return Result{}, res.err
			}
			return Result{ExitCode: res.exit, Resources: res.res}, nil
		case <-ctx.Done():
			return Result{}, ctx.Err()
		}
	}

	var res prodResult
	if resPtr != nil {
		res = *resPtr
	} else {
		select {
		case res = <-doneCh:
		case <-ctx.Done():
			return Result{}, ctx.Err()
		}
	}

	// FALLBACK: If streaming failed or we bypassed it, write the stdout directly.
	if res.stdout != "" {
		if f, err := os.OpenFile(spec.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			_, _ = f.WriteString("\n--- DIRECT STDOUT FALLBACK ---\n" + res.stdout)
			f.Close()
		}
	}

	if res.err != nil {
		return Result{}, res.err
	}
	return Result{ExitCode: res.exit, Resources: res.res}, nil
}

// prodResult is what the producer goroutine reports back to Execute. The
// streamer also receives this on its `done` channel so it can drain one
// last time before returning.
type prodResult struct {
	exit   int
	res    Resources
	err    error
	stderr string
	stdout string
}

// streamLog polls the producer's shared cache volume and appends incremental
// content to spec.LogFile. Returns when it sees the `__workend_done__` marker
// the wrapper writes after the user command exits, or when the producer's
// done channel signals (whichever comes first).
func streamLog(ctx context.Context, dc *wdagger.Client, client *dag.Client, cache *dag.CacheVolume, logFile string, done <-chan prodResult) (*prodResult, error) {
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	var (
		offset    int
		seenDone  bool
		marker    = "__workend_done__\n"
	)

	poller := client.Container().
		From("alpine:3.20").
		WithMountedCache("/shared", cache, dag.ContainerWithMountedCacheOpts{
			Sharing: dag.CacheSharingModeShared,
		})

	read := func() error {
		out, err := poller.
			WithEnvVariable("CACHE_BUSTER", time.Now().String()).
			WithExec([]string{"sh", "-c", "[ -f /shared/out.log ] && cat /shared/out.log || true"}).
			Stdout(ctx)
		if err != nil {
			return err
		}
		if len(out) <= offset {
			return nil
		}
		chunk := out[offset:]
		// Strip the done-marker if present, but record that we saw it so
		// the loop can break after one more drain.
		if i := strings.Index(chunk, marker); i >= 0 {
			seenDone = true
			chunk = chunk[:i] + chunk[i+len(marker):]
		}
		if _, err := f.WriteString(chunk); err != nil {
			return err
		}
		offset = len(out)
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-done:
			// Producer finished; read the final logs using a fresh container to guarantee we
			// bypass any internal Dagger cache or state issues from the persistent poller object.
			finalOut, err := client.Container().
				From("alpine:3.20").
				WithMountedCache("/shared", cache).
				WithEnvVariable("CACHE_BUSTER", time.Now().String()).
				WithExec([]string{"sh", "-c", "[ -f /shared/out.log ] && cat /shared/out.log || true"}).
				Stdout(ctx)
			
			fmt.Printf("DEBUG FINAL: err=%v, offset=%d, finalOut_len=%d, exitCode=%d, stderr=%q, stdout_len=%d\n", err, offset, len(finalOut), res.exit, res.stderr, len(res.stdout))
			if len(res.stdout) > 0 {
				fmt.Printf("DEBUG PRODUCER STDOUT: %s\n", res.stdout[:min(len(res.stdout), 100)])
			}
			if len(finalOut) > 0 {
				fmt.Printf("DEBUG FINAL OUT (first 100 chars): %s\n", finalOut[:min(len(finalOut), 100)])
			}

			if err == nil {
				if len(finalOut) > offset {
					chunk := finalOut[offset:]
					if i := strings.Index(chunk, marker); i >= 0 {
						chunk = chunk[:i] + chunk[i+len(marker):]
					}
					_, _ = f.WriteString(chunk)
				}
			} else {
				_, _ = f.WriteString("\nworkend: final log fetch failed: " + err.Error() + "\n")
			}
			return &res, nil
		case <-tick.C:
			if err := read(); err != nil {
				if wdagger.IsSessionError(err) {
					dc.MarkStale()
					newClient, reconErr := dc.Get(ctx)
					if reconErr == nil {
						client = newClient
						poller = client.Container().
							From("alpine:3.20").
							WithMountedCache("/shared", cache, dag.ContainerWithMountedCacheOpts{
								Sharing: dag.CacheSharingModeShared,
							})
						_, _ = f.WriteString("workend: session reconnected\n")
					} else {
						_, _ = f.WriteString("workend: session reconnect failed: " + reconErr.Error() + "\n")
					}
				} else {
					// Don't fail the whole run; just log via the file (the
					// producer's exit-code read is the source of truth).
					_, _ = f.WriteString("workend: stream read warning: " + err.Error() + "\n")
				}
			}
			if seenDone {
				return nil, nil
			}
		}
	}
}

// parseMetrics ingests the key=value file the wrapper script writes.
// Unrecognized lines are silently skipped so the format can grow over time
// without breaking older API binaries.
func parseMetrics(raw string) Resources {
	r := Resources{}
	for _, line := range strings.Split(raw, "\n") {
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := line[:eq]
		val := strings.TrimSpace(line[eq+1:])
		var n int64
		_, _ = fmt.Sscanf(val, "%d", &n)
		switch key {
		case "cpu_ms":
			r.CPUMs = n
		case "mem_peak_kb":
			r.MemPeakBytes = n * 1024
		case "net_rx":
			r.NetRxBytes = n
		case "net_tx":
			r.NetTxBytes = n
		}
	}
	return r
}

// baseImage chooses the OCI image used to execute tasks of a given source.
// Pinned tags only — never `latest`, so runs are reproducible.
func baseImage(source, repoPath string) string {
	switch source {
	case "npm", "pnpm", "yarn":
		return "node:22-alpine"
	case "bun":
		return "oven/bun:1-alpine"
	case "just":
		return "alpine:3.20"
	case "dagger":
		return "registry.dagger.io/cli:v0.20.5"
	case "go":
		return goImage(repoPath)
	case "python":
		return "python:3.12-alpine"
	case "rust":
		return "rust:1.80-alpine"
	case "java", "maven", "gradle":
		return "eclipse-temurin:21-jdk-alpine"
	default:
		return "alpine:3.20"
	}
}

func goImage(repoPath string) string {
	const fallback = "golang:1.25-alpine"
	if repoPath == "" {
		return fallback
	}
	data, err := os.ReadFile(filepath.Join(repoPath, "go.mod"))
	if err != nil {
		return fallback
	}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && f[0] == "go" {
			parts := strings.SplitN(f[1], ".", 3)
			if len(parts) >= 2 {
				return "golang:" + parts[0] + "." + parts[1] + "-alpine"
			}
		}
	}
	return fallback
}

// setupSteps returns the WithExec arg lists to run after `From` and before the
// task command. Used to install per-source tooling (e.g., the `just` binary or
// alternative Node package managers via corepack). jq is bundled with the
// node-based images so downstream wrapper steps can read package.json.
func setupSteps(source string) [][]string {
	switch source {
	case "npm":
		return [][]string{
			{"apk", "add", "--no-cache", "jq"},
		}
	case "pnpm":
		return [][]string{
			{"apk", "add", "--no-cache", "jq"},
			{"corepack", "enable", "pnpm"},
		}
	case "yarn":
		return [][]string{
			{"apk", "add", "--no-cache", "jq"},
			{"corepack", "enable", "yarn"},
		}
	case "bun":
		return [][]string{
			{"apk", "add", "--no-cache", "jq"},
		}
	case "just":
		return [][]string{
			{"apk", "add", "--no-cache", "just", "bash", "git"},
		}
	default:
		return nil
	}
}

func mountLanguageCaches(client *dag.Client, container *dag.Container, spec Spec) *dag.Container {
	h := sha256.New()
	h.Write([]byte(spec.GitURL))
	projHash := hex.EncodeToString(h.Sum(nil))[:12]

	switch spec.Source {
	case "npm", "pnpm", "yarn":
		return container.
			WithMountedCache("/work/node_modules", client.CacheVolume("deps-node-"+projHash)).
			WithMountedCache("/root/.npm", client.CacheVolume("deps-npm-global")).
			WithMountedCache("/root/.local/share/pnpm/store/v3", client.CacheVolume("deps-pnpm-global")).
			WithMountedCache("/usr/local/share/.cache/yarn", client.CacheVolume("deps-yarn-global"))
	case "bun":
		return container.
			WithMountedCache("/work/node_modules", client.CacheVolume("deps-node-"+projHash)).
			WithMountedCache("/root/.bun/install/cache", client.CacheVolume("deps-bun-global"))
	case "go":
		return container.
			WithMountedCache("/go/pkg/mod", client.CacheVolume("deps-go-mod-"+projHash)).
			WithMountedCache("/root/.cache/go-build", client.CacheVolume("deps-go-build-"+projHash))
	case "python":
		return container.
			WithMountedCache("/root/.cache/pip", client.CacheVolume("deps-pip-global")).
			WithMountedCache("/work/.venv", client.CacheVolume("deps-venv-"+projHash))
	case "rust":
		return container.
			WithMountedCache("/usr/local/cargo/registry", client.CacheVolume("deps-cargo-registry")).
			WithMountedCache("/work/target", client.CacheVolume("deps-rust-target-"+projHash))
	case "java", "gradle", "maven":
		return container.
			WithMountedCache("/root/.m2", client.CacheVolume("deps-maven-global")).
			WithMountedCache("/root/.gradle", client.CacheVolume("deps-gradle-global"))
	}
	return container
}

// shellQuoteArgs joins extra args into a string safe for `sh -c`. Single
// quotes are escaped using the standard `'\''` trick. Whitespace and shell
// metacharacters in arguments survive intact.
func shellQuoteArgs(args []string) string {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		parts = append(parts, "'"+strings.ReplaceAll(a, "'", `'\''`)+"'")
	}
	return strings.Join(parts, " ")
}
