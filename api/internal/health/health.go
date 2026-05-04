package health

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Status struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

type Handler struct {
	Pool           *pgxpool.Pool
	DaggerSockPath string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{
		"db":     h.checkDB(r.Context()),
		"dagger": h.checkDagger(),
	}

	overall := "ok"
	for _, v := range checks {
		if v != "ok" {
			overall = "degraded"
			break
		}
	}

	resp := Status{Status: overall, Checks: checks}
	w.Header().Set("Content-Type", "application/json")
	if overall != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) checkDB(ctx context.Context) string {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := h.Pool.Ping(pingCtx); err != nil {
		return "error: " + err.Error()
	}
	return "ok"
}

func (h *Handler) checkDagger() string {
	info, err := os.Stat(h.DaggerSockPath)
	if err != nil {
		return "error: " + err.Error()
	}
	if info.Mode()&os.ModeSocket == 0 {
		return "error: not a socket"
	}
	conn, err := net.DialTimeout("unix", h.DaggerSockPath, 2*time.Second)
	if err != nil {
		return "error: dial: " + err.Error()
	}
	_ = conn.Close()
	return "ok"
}
