package audit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminHandlers owns admin-only audit endpoints.
type AdminHandlers struct {
	Pool *pgxpool.Pool
}

type exportEntry struct {
	ID          int64           `json:"id"`
	OccurredAt  time.Time       `json:"occurred_at"`
	ActorID     *uuid.UUID      `json:"actor_id"`
	Action      string          `json:"action"`
	TargetKind  *string         `json:"target_kind"`
	TargetID    *string         `json:"target_id"`
	IP          *string         `json:"ip"`
	Metadata    json.RawMessage `json:"metadata"`
	RetainedUntil *time.Time    `json:"retained_until"`
}

// ExportAuditLog streams audit entries between two dates as a JSON array.
// GET /api/admin/audit-log/export?from=<date>&to=<date>&format=json
//
// Max 90-day window per request. Admin-only (caller must be behind
// admin.RequireAdmin middleware).
func (h *AdminHandlers) ExportAuditLog(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		http.Error(w, "from and to query parameters are required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		http.Error(w, "invalid from date (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		http.Error(w, "invalid to date (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	// End of the 'to' day.
	to = to.Add(24*time.Hour - time.Nanosecond)

	if to.Sub(from) > 90*24*time.Hour {
		http.Error(w, "max 90-day window per export request", http.StatusBadRequest)
		return
	}
	if from.After(to) {
		http.Error(w, "from must be before to", http.StatusBadRequest)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, occurred_at, actor_id, action, target_kind, target_id, ip, metadata, retained_until
		FROM audit_log
		WHERE occurred_at >= $1 AND occurred_at <= $2
		ORDER BY occurred_at ASC
	`, from, to)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []exportEntry{}
	for rows.Next() {
		var e exportEntry
		var meta []byte
		if err := rows.Scan(&e.ID, &e.OccurredAt, &e.ActorID, &e.Action,
			&e.TargetKind, &e.TargetID, &e.IP, &meta, &e.RetainedUntil); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if len(meta) > 0 {
			e.Metadata = meta
		} else {
			e.Metadata = json.RawMessage("null")
		}
		out = append(out, e)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="audit-log_%s_%s.json"`, fromStr, toStr))
	_ = json.NewEncoder(w).Encode(out)
}

// SetRetention updates retained_until for audit rows that have no retention
// set yet.
// POST /api/admin/audit-log/retention
//
//	body: {"retain_days": 365}
func (h *AdminHandlers) SetRetention(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RetainDays int `json:"retain_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.RetainDays < 1 {
		http.Error(w, "retain_days must be a positive integer", http.StatusBadRequest)
		return
	}

	interval := strconv.Itoa(body.RetainDays) + " days"
	tag, err := h.Pool.Exec(r.Context(), `
		UPDATE audit_log
		SET retained_until = now() + $1::INTERVAL
		WHERE retained_until IS NULL
	`, interval)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"updated_rows": tag.RowsAffected(),
		"retain_days":  body.RetainDays,
	})
}
