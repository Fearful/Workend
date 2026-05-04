package run

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	wdagger "workend/api/internal/dagger"
)

// VulnSummary is the digested form of the Trivy report we store on each
// image. The full JSON is large; the summary is enough to drive the UI
// "5 critical, 12 high" badges and the dashboard alert.
type VulnSummary struct {
	Critical int              `json:"critical"`
	High     int              `json:"high"`
	Medium   int              `json:"medium"`
	Low      int              `json:"low"`
	Unknown  int              `json:"unknown"`
	Top      []VulnSummaryRow `json:"top"`
}

// VulnSummaryRow is one notable finding (criticals first, then high).
type VulnSummaryRow struct {
	ID       string `json:"id"`
	Package  string `json:"package"`
	Severity string `json:"severity"`
	FixedIn  string `json:"fixed_in,omitempty"`
}

// ScanImageAsync runs Trivy against the project's filesystem (the same tree
// used to build the image). This catches dependency CVEs in the lockfiles
// and any installed packages, which is the most common class of finding.
// Full image-tarball scans are deferred until the Dagger SDK exposes a
// stable per-engine `Tarball()`.
//
// Best-effort: persists scan_status='error' on failure so the UI can
// surface the issue without making the build itself fail.
func ScanImageAsync(pool *pgxpool.Pool, dc *wdagger.Client, imageID uuid.UUID, repoPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if _, err := pool.Exec(ctx,
		`UPDATE images SET scan_status = 'pending' WHERE id = $1`, imageID); err != nil {
		return
	}

	summary, err := runTrivy(ctx, dc, repoPath)
	if err != nil {
		_, _ = pool.Exec(ctx, `
			UPDATE images SET scan_status = 'error', scan_completed_at = now() WHERE id = $1
		`, imageID)
		return
	}
	encoded, _ := json.Marshal(summary)
	_, _ = pool.Exec(ctx, `
		UPDATE images
		SET scan_status = 'ok', scan_completed_at = now(), vuln_summary = $1::jsonb
		WHERE id = $2
	`, string(encoded), imageID)
}

func runTrivy(ctx context.Context, dc *wdagger.Client, repoPath string) (*VulnSummary, error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return nil, err
	}

	repoDir := client.Host().Directory(repoPath)
	out, err := client.Container().
		From("aquasec/trivy:0.55.2").
		WithMountedDirectory("/scan", repoDir).
		// `--quiet` suppresses progress; `--no-progress` for older versions;
		// `--format json` is what we parse. Filesystem scan covers OS pkgs
		// + language lockfiles.
		WithExec([]string{
			"trivy", "fs",
			"--quiet",
			"--no-progress",
			"--format", "json",
			"--severity", "CRITICAL,HIGH,MEDIUM,LOW,UNKNOWN",
			"--scanners", "vuln",
			"/scan",
		}).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("trivy: %w", err)
	}
	return parseTrivyJSON([]byte(out))
}

// parseTrivyJSON walks the Trivy report and tallies severities, picking up
// to 10 most-severe findings for the summary.
func parseTrivyJSON(raw []byte) (*VulnSummary, error) {
	var report struct {
		Results []struct {
			Vulnerabilities []struct {
				VulnerabilityID  string `json:"VulnerabilityID"`
				PkgName          string `json:"PkgName"`
				Severity         string `json:"Severity"`
				FixedVersion     string `json:"FixedVersion"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, fmt.Errorf("parse trivy json: %w", err)
	}

	sev := func(s string) int {
		switch s {
		case "CRITICAL":
			return 4
		case "HIGH":
			return 3
		case "MEDIUM":
			return 2
		case "LOW":
			return 1
		}
		return 0
	}

	all := []VulnSummaryRow{}
	out := &VulnSummary{}
	for _, r := range report.Results {
		for _, v := range r.Vulnerabilities {
			switch v.Severity {
			case "CRITICAL":
				out.Critical++
			case "HIGH":
				out.High++
			case "MEDIUM":
				out.Medium++
			case "LOW":
				out.Low++
			default:
				out.Unknown++
			}
			all = append(all, VulnSummaryRow{
				ID: v.VulnerabilityID, Package: v.PkgName,
				Severity: v.Severity, FixedIn: v.FixedVersion,
			})
		}
	}
	sort.SliceStable(all, func(i, j int) bool { return sev(all[i].Severity) > sev(all[j].Severity) })
	if len(all) > 10 {
		all = all[:10]
	}
	out.Top = all
	return out, nil
}
