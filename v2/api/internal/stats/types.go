package stats

import (
	"time"

	"github.com/google/uuid"
)

type Record struct {
	ID         uuid.UUID       `json:"id"`
	ProjectID  uuid.UUID       `json:"project_id"`
	ComputedAt time.Time       `json:"computed_at"`
	TotalFiles int             `json:"total_files"`
	TotalLines int             `json:"total_lines"`
	TotalCode  int             `json:"total_code"`
	Languages  map[string]Lang `json:"languages"`
}
