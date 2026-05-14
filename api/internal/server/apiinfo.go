package server

import (
	"encoding/json"
	"net/http"
	"runtime"
)

type APIInfo struct {
	Version   string   `json:"version"`
	GoVersion string   `json:"go_version"`
	Features  []string `json:"features"`
	DaggerSDK string   `json:"dagger_sdk"`
}

func apiInfoHandler() http.HandlerFunc {
	info := APIInfo{
		Version:   "0.7.0",
		GoVersion: runtime.Version(),
		Features: []string{
			"auth", "workspaces", "projects", "tasks", "runs",
			"schedules", "pipelines", "images", "compose",
			"artifacts", "notifications", "webhooks", "push",
			"ssh-keys", "pat-credentials", "board", "remote-ci",
			"log-streaming", "log-search", "secret-masking",
			"run-presets", "run-archival", "registry-push",
			"repo-search", "flaky-quarantine", "secret-rotation",
			"conditional-pipelines", "artifact-comparison",
			"audit-export", "ci-triggers", "tls",
		},
		DaggerSDK: "v0.20.5",
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(info)
	}
}
