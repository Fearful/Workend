package quota

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var thresholds = []int{80, 95}

type AlertSender interface {
	SendQuotaAlert(ctx context.Context, userID uuid.UUID, pct int, usedBytes, quotaBytes int64)
}

func CheckAndAlert(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger, reposRoot string, userID uuid.UUID, sender AlertSender) {
	var quotaBytes int64
	if err := pool.QueryRow(ctx,
		`SELECT quota_bytes FROM users WHERE id = $1`, userID,
	).Scan(&quotaBytes); err != nil || quotaBytes <= 0 {
		return
	}

	rows, err := pool.Query(ctx, `SELECT id FROM workspaces WHERE user_id = $1`, userID)
	if err != nil {
		return
	}
	defer rows.Close()

	var usedBytes int64
	for rows.Next() {
		var wsID uuid.UUID
		if err := rows.Scan(&wsID); err == nil {
			usedBytes += dirSize(filepath.Join(reposRoot, wsID.String()))
		}
	}

	pct := int(usedBytes * 100 / quotaBytes)

	for _, t := range thresholds {
		if pct < t {
			continue
		}
		alreadySent := hasRecentAlert(ctx, pool, userID, t)
		if alreadySent {
			continue
		}
		if err := recordAlert(ctx, pool, userID, t, usedBytes, quotaBytes); err != nil {
			logger.Warn("quota: failed to record alert", "err", err, "threshold", t)
			continue
		}
		logger.Info("quota: threshold reached", "user_id", userID, "pct", pct, "threshold", t)
		if sender != nil {
			sender.SendQuotaAlert(ctx, userID, t, usedBytes, quotaBytes)
		}
	}
}

func hasRecentAlert(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, threshold int) bool {
	var createdAt time.Time
	err := pool.QueryRow(ctx, `
		SELECT created_at FROM quota_alerts
		WHERE user_id = $1 AND threshold = $2
	`, userID, threshold).Scan(&createdAt)
	if err == pgx.ErrNoRows {
		return false
	}
	if err != nil {
		return true
	}
	return time.Since(createdAt) < 24*time.Hour
}

func recordAlert(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, threshold int, usedBytes, quotaBytes int64) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO quota_alerts (user_id, threshold, used_bytes, quota_bytes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, threshold)
		DO UPDATE SET used_bytes = $3, quota_bytes = $4, created_at = now()
	`, userID, threshold, usedBytes, quotaBytes)
	return err
}

func formatBytes(b int64) string {
	const gb = 1024 * 1024 * 1024
	const mb = 1024 * 1024
	if b >= gb {
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	}
	return fmt.Sprintf("%.0f MB", float64(b)/float64(mb))
}
