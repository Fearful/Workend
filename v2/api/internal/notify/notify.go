// Package notify dispatches notifications when runs complete.
//
// Each user can configure multiple targets (webhook URL, Slack webhook URL,
// email address). The dispatcher fires fan-out per matching config when the
// run.markFinished call invokes Notify.OnRunComplete.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	KindWebhook = "webhook"
	KindSlack   = "slack"
	KindEmail   = "email"

	TriggerOnFailure      = "on_failure"
	TriggerOnStatusChange = "on_status_change"
	TriggerAlways         = "always"
)

type Config struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Kind      string    `json:"kind"`
	Target    string    `json:"target"`
	Trigger   string    `json:"trigger"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type Event struct {
	RunID       uuid.UUID
	UserID      uuid.UUID
	ProjectID   uuid.UUID
	ProjectName string
	TaskName    string
	TaskSource  string
	Status      string  // succeeded | failed | cancelled
	PrevStatus  string  // status of previous run for this task; "" if first
	ExitCode    int
	DurationSec int
	URL         string  // public link to /runs/<id> on the web frontend
}

type SMTP struct {
	Host     string // host:port
	From     string
	Username string
	Password string
}

type Dispatcher struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
	HTTP   *http.Client
	SMTP   *SMTP
	WebURL string // e.g., http://localhost:3000
}

func New(pool *pgxpool.Pool, logger *slog.Logger, smtp *SMTP, webURL string) *Dispatcher {
	return &Dispatcher{
		Pool:   pool,
		Logger: logger,
		HTTP:   &http.Client{Timeout: 10 * time.Second},
		SMTP:   smtp,
		WebURL: webURL,
	}
}

// OnRunComplete is called from run.markFinished after status persists.
// Best-effort: failures are logged, never returned.
func (d *Dispatcher) OnRunComplete(e Event) {
	if d == nil {
		return
	}
	go d.dispatch(e)
}

func (d *Dispatcher) dispatch(e Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := d.Pool.Query(ctx, `
		SELECT id, kind, target, trigger
		FROM notification_configs
		WHERE user_id = $1 AND enabled = true
	`, e.UserID)
	if err != nil {
		d.Logger.Warn("notify: load configs", "err", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var c Config
		if err := rows.Scan(&c.ID, &c.Kind, &c.Target, &c.Trigger); err != nil {
			continue
		}
		if !shouldFire(c.Trigger, e) {
			continue
		}
		if err := d.send(ctx, c, e); err != nil {
			d.Logger.Warn("notify: send failed", "kind", c.Kind, "target", c.Target, "err", err)
		}
	}
}

func shouldFire(trigger string, e Event) bool {
	switch trigger {
	case TriggerAlways:
		return true
	case TriggerOnFailure:
		return e.Status == "failed"
	case TriggerOnStatusChange:
		return e.PrevStatus != "" && e.PrevStatus != e.Status
	}
	return false
}

func (d *Dispatcher) send(ctx context.Context, c Config, e Event) error {
	switch c.Kind {
	case KindWebhook:
		return d.sendWebhook(ctx, c.Target, e)
	case KindSlack:
		return d.sendSlack(ctx, c.Target, e)
	case KindEmail:
		return d.sendEmail(c.Target, e)
	}
	return errors.New("unknown notification kind: " + c.Kind)
}

func (d *Dispatcher) sendWebhook(ctx context.Context, url string, e Event) error {
	body, _ := json.Marshal(map[string]any{
		"run_id":       e.RunID,
		"project_name": e.ProjectName,
		"task_name":    e.TaskName,
		"task_source":  e.TaskSource,
		"status":       e.Status,
		"exit_code":    e.ExitCode,
		"duration_sec": e.DurationSec,
		"url":          e.URL,
	})
	return d.postJSON(ctx, url, body)
}

func (d *Dispatcher) sendSlack(ctx context.Context, url string, e Event) error {
	emoji := ":white_check_mark:"
	color := "good"
	if e.Status == "failed" {
		emoji = ":x:"
		color = "danger"
	} else if e.Status == "cancelled" {
		emoji = ":no_entry_sign:"
		color = "warning"
	}
	text := fmt.Sprintf("%s *%s* %s in %s (exit %d, %ds)\n<%s|view run>",
		emoji, e.TaskName, e.Status, e.ProjectName, e.ExitCode, e.DurationSec, e.URL)
	body, _ := json.Marshal(map[string]any{
		"text": text,
		"attachments": []map[string]any{{
			"color": color,
			"text":  text,
		}},
	})
	return d.postJSON(ctx, url, body)
}

func (d *Dispatcher) sendEmail(to string, e Event) error {
	if d.SMTP == nil || d.SMTP.Host == "" {
		return errors.New("SMTP not configured")
	}
	subject := fmt.Sprintf("[Workend] %s %s in %s", e.TaskName, e.Status, e.ProjectName)
	body := fmt.Sprintf(
		"Task: %s (%s)\nProject: %s\nStatus: %s\nExit code: %d\nDuration: %ds\n\n%s\n",
		e.TaskName, e.TaskSource, e.ProjectName, e.Status, e.ExitCode, e.DurationSec, e.URL,
	)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		d.SMTP.From, to, subject, body))
	auth := smtp.PlainAuth("", d.SMTP.Username, d.SMTP.Password, hostOnly(d.SMTP.Host))
	return smtp.SendMail(d.SMTP.Host, auth, d.SMTP.From, []string{to}, msg)
}

func (d *Dispatcher) postJSON(ctx context.Context, url string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

func hostOnly(hostPort string) string {
	for i := 0; i < len(hostPort); i++ {
		if hostPort[i] == ':' {
			return hostPort[:i]
		}
	}
	return hostPort
}
