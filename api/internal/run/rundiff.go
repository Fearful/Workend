package run

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
	"workend/api/internal/diff"
)

// DiffLine represents one line in a side-by-side diff view.
type DiffLine struct {
	Type      string `json:"type"` // "equal", "added", "removed"
	LeftLine  int    `json:"left_line,omitempty"`
	RightLine int    `json:"right_line,omitempty"`
	Content   string `json:"content"`
}

// DiffResult is the response for the run diff endpoint.
type DiffResult struct {
	LeftRunID    uuid.UUID  `json:"left_run_id"`
	RightRunID   uuid.UUID  `json:"right_run_id"`
	TaskID       uuid.UUID  `json:"task_id"`
	TaskName     string     `json:"task_name"`
	LeftStatus   string     `json:"left_status"`
	RightStatus  string     `json:"right_status"`
	LeftStarted  *time.Time `json:"left_started_at"`
	RightStarted *time.Time `json:"right_started_at"`
	TotalLines   int        `json:"total_lines"`
	ChangedLines int        `json:"changed_lines"`
	Lines        []DiffLine `json:"lines"`
	Truncated    bool       `json:"truncated"`
}

const maxDiffLines = 5000

// DiffRuns returns a rich side-by-side diff of stdout for two runs of the
// same task. Both runs must be owned by the requesting user and belong to
// the same task.
//
// GET /api/runs/{id}/diff/{other_id}
func (h *Handlers) DiffRuns(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	leftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}
	otherID, err := uuid.Parse(chi.URLParam(r, "other_id"))
	if err != nil {
		http.Error(w, "invalid other_id", http.StatusBadRequest)
		return
	}

	left, err := h.fetchOwned(r.Context(), uid, leftID)
	if err != nil {
		http.Error(w, "left run not found", http.StatusNotFound)
		return
	}
	right, err := h.fetchOwned(r.Context(), uid, otherID)
	if err != nil {
		http.Error(w, "right run not found", http.StatusNotFound)
		return
	}
	if left.TaskID != right.TaskID {
		http.Error(w, "runs belong to different tasks", http.StatusBadRequest)
		return
	}

	leftRaw, _ := os.ReadFile(left.LogPath)
	rightRaw, _ := os.ReadFile(right.LogPath)

	leftLog := MaskSecrets(string(leftRaw))
	rightLog := MaskSecrets(string(rightRaw))

	hunks := diff.Lines(leftLog, rightLog)

	lines, changed, truncated := buildDiffLines(hunks)

	result := DiffResult{
		LeftRunID:    left.ID,
		RightRunID:   right.ID,
		TaskID:       left.TaskID,
		TaskName:     left.TaskName,
		LeftStatus:   left.Status,
		RightStatus:  right.Status,
		LeftStarted:  left.StartedAt,
		RightStarted: right.StartedAt,
		TotalLines:   len(lines),
		ChangedLines: changed,
		Lines:        lines,
		Truncated:    truncated,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// buildDiffLines converts diff hunks to DiffLine entries with line numbers.
// Returns the lines, a count of changed lines, and whether the output was
// truncated.
func buildDiffLines(hunks []diff.Hunk) ([]DiffLine, int, bool) {
	lines := make([]DiffLine, 0, len(hunks))
	leftNum := 0
	rightNum := 0
	changed := 0
	truncated := false

	for _, h := range hunks {
		if len(lines) >= maxDiffLines {
			truncated = true
			break
		}

		content := strings.TrimSuffix(h.Text, "\n")

		switch h.Op {
		case diff.Equal:
			leftNum++
			rightNum++
			lines = append(lines, DiffLine{
				Type:      "equal",
				LeftLine:  leftNum,
				RightLine: rightNum,
				Content:   content,
			})
		case diff.Insert:
			rightNum++
			changed++
			lines = append(lines, DiffLine{
				Type:      "added",
				RightLine: rightNum,
				Content:   content,
			})
		case diff.Delete:
			leftNum++
			changed++
			lines = append(lines, DiffLine{
				Type:      "removed",
				LeftLine:  leftNum,
				Content:   content,
			})
		}
	}

	return lines, changed, truncated
}
