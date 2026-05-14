package run

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
)

// SearchHit is a single matching line from a log search.
type SearchHit struct {
	Line int    `json:"line"`
	Text string `json:"text"`
}

// SearchLog searches the run's log file line-by-line for a query term and
// returns matching lines with their line numbers. Max 500 results.
// GET /api/runs/{id}/log/search?q=<term>&case_sensitive=false
func (h *Handlers) SearchLog(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q parameter is required", http.StatusBadRequest)
		return
	}

	caseSensitive := r.URL.Query().Get("case_sensitive") == "true"

	run, err := h.fetchOwned(r.Context(), uid, runID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	f, err := os.Open(run.LogPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]SearchHit{})
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	const maxResults = 500
	hits := []SearchHit{}
	scanner := bufio.NewScanner(f)
	lineNum := 0

	searchQ := q
	if !caseSensitive {
		searchQ = strings.ToLower(q)
	}

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		match := text
		if !caseSensitive {
			match = strings.ToLower(text)
		}
		if strings.Contains(match, searchQ) {
			hits = append(hits, SearchHit{Line: lineNum, Text: text})
			if len(hits) >= maxResults {
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(hits)
}
