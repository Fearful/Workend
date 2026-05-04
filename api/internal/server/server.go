package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/admin"
	"workend/api/internal/artifact"
	"workend/api/internal/audit"
	"workend/api/internal/auth"
	"workend/api/internal/board"
	"workend/api/internal/cidetect"
	"workend/api/internal/comment"
	"workend/api/internal/compose"
	"workend/api/internal/config"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/dashboard"
	"workend/api/internal/deps"
	"workend/api/internal/events"
	"workend/api/internal/health"
	"workend/api/internal/image"
	"workend/api/internal/notify"
	"workend/api/internal/oauth"
	"workend/api/internal/oidc"
	"workend/api/internal/patcred"
	"workend/api/internal/pin"
	"workend/api/internal/pipeline"
	"workend/api/internal/project"
	"workend/api/internal/push"
	"workend/api/internal/remoteci"
	"workend/api/internal/quota"
	"workend/api/internal/run"
	"workend/api/internal/schedule"
	"workend/api/internal/secret"
	"workend/api/internal/secscan"
	"workend/api/internal/sshkey"
	"workend/api/internal/stats"
	"workend/api/internal/task"
	"workend/api/internal/workspace"
)

type Server struct {
	cfg       *config.Config
	pool      *pgxpool.Pool
	dagger    *wdagger.Client
	oauth     *oauth.Registry
	box       *secret.Box
	logger    *slog.Logger
	auditLog  *audit.Logger
	notif     *notify.Dispatcher
	eventsD   *events.Dispatcher
	push      *push.Sender
	runH      *run.Handlers
	schedH    *schedule.Handlers
	oidcDoc   *oidc.DiscoveryDoc
}

func New(cfg *config.Config, pool *pgxpool.Pool, dc *wdagger.Client, oauthReg *oauth.Registry, box *secret.Box, logger *slog.Logger, oidcDoc *oidc.DiscoveryDoc) *Server {
	auditLog := &audit.Logger{Pool: pool, Log: logger}

	var smtpCfg *notify.SMTP
	if cfg.SMTPHost != "" {
		smtpCfg = &notify.SMTP{
			Host: cfg.SMTPHost, From: cfg.SMTPFrom,
			Username: cfg.SMTPUsername, Password: cfg.SMTPPassword,
		}
	}
	notif := notify.New(pool, logger, smtpCfg, cfg.WebPublicURL)
	eventsD := events.New(pool, logger)
	pushS := push.New(pool, push.Config{
		PublicKey:  cfg.VAPIDPublicKey,
		PrivateKey: cfg.VAPIDPrivateKey,
		Subject:    cfg.VAPIDSubject,
	}, logger)

	runH := &run.Handlers{
		Pool: pool, Dagger: dc, LogsRoot: cfg.LogsRoot, ArtifactsRoot: cfg.ArtifactsRoot,
		Logger: logger, Audit: auditLog,
		Notify: notif, Events: eventsD, Push: pushS, WebURL: cfg.WebPublicURL,
		DefaultTimeout: time.Duration(cfg.DefaultTaskTimeoutSec) * time.Second,
		OAuth:          oauthReg,
	}
	runH.Init()
	schedH := schedule.NewHandlers(pool, dc, logger, auditLog)
	return &Server{
		cfg: cfg, pool: pool, dagger: dc, oauth: oauthReg, box: box, logger: logger,
		auditLog: auditLog, notif: notif, eventsD: eventsD, push: pushS,
		runH: runH, schedH: schedH, oidcDoc: oidcDoc,
	}
}

// Events exposes the events dispatcher so other packages can emit.
func (s *Server) Events() *events.Dispatcher { return s.eventsD }

// Schedules returns the schedule handlers (so main can start the ticker).
func (s *Server) Schedules() *schedule.Handlers { return s.schedH }

// Runs returns the run handlers (so the scheduler can enqueue runs).
func (s *Server) Runs() *run.Handlers { return s.runH }

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Handle("/healthz", &health.Handler{
		Pool:           s.pool,
		DaggerSockPath: s.cfg.DaggerSockPath,
	})

	authH := &auth.Handlers{Pool: s.pool, Secure: s.cfg.CookieSecure, Audit: s.auditLog, OIDCDisableSignup: s.cfg.OIDCConfigured()}
	wsH := &workspace.Handlers{Pool: s.pool}
	adminH := &admin.Handlers{Pool: s.pool, Audit: s.auditLog}
	projH := &project.Handlers{
		Pool:      s.pool,
		Dagger:    s.dagger,
		OAuth:     s.oauth,
		Secret:    s.box,
		ReposRoot: s.cfg.ReposRoot,
		Logger:    s.logger,
	}
	sshH := &sshkey.Handlers{Pool: s.pool, Box: s.box}
	patH := &patcred.Handlers{Pool: s.pool, Box: s.box}
	secScanH := &secscan.Handlers{Pool: s.pool}
	eventsH := &events.Handlers{Pool: s.pool, D: s.eventsD}
	pushH := &push.Handlers{Pool: s.pool, Cfg: push.Config{
		PublicKey:  s.cfg.VAPIDPublicKey,
		PrivateKey: s.cfg.VAPIDPrivateKey,
		Subject:    s.cfg.VAPIDSubject,
	}}
	taskH := &task.Handlers{Pool: s.pool}
	statsH := &stats.Handlers{Pool: s.pool}
	dashH := &dashboard.Handlers{Pool: s.pool}
	imgH := &image.Handlers{Pool: s.pool}
	notifH := &notify.Handlers{D: s.notif}
	quotaH := &quota.Handlers{Pool: s.pool, ReposRoot: s.cfg.ReposRoot}
	pinH := &pin.Handlers{Pool: s.pool}
	commentH := &comment.Handlers{Pool: s.pool}
	boardH := &board.Handlers{Pool: s.pool, OAuth: s.oauth}
	pipelineH := &pipeline.Handlers{Pool: s.pool, Runner: s.runH, Logger: s.logger}
	artifactH := &artifact.Handlers{Pool: s.pool, ArtifactsRoot: s.cfg.ArtifactsRoot, Logger: s.logger}
	remoteciH := &remoteci.Handlers{Pool: s.pool, OAuth: s.oauth, Logger: s.logger}
	cidetectH := &cidetect.Handlers{Pool: s.pool}
	composeH := &compose.Handlers{Pool: s.pool, Dagger: s.dagger, Box: s.box, Logger: s.logger}
	depsH := &deps.Handlers{Pool: s.pool}
	runH := s.runH
	schedH := s.schedH

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/signup", authH.Signup)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/logout", authH.Logout)
		r.Get("/auth/config", oidc.AuthConfigHandler(s.cfg.OIDCConfigured(), s.cfg.OIDCProviderName))

		if s.oidcDoc != nil {
			oidcH := &oidc.Handlers{
				Pool:         s.pool,
				Doc:          s.oidcDoc,
				ClientID:     s.cfg.OIDCClientID,
				ClientSecret: s.cfg.OIDCClientSecret,
				RedirectBase: s.cfg.WebPublicURL,
				ProviderName: s.cfg.OIDCProviderName,
				Secure:       s.cfg.CookieSecure,
				Audit:        s.auditLog,
				Logger:       s.logger,
			}
			r.Get("/auth/oidc/start", oidcH.Start)
			r.Get("/auth/oidc/callback", oidcH.Callback)
		}

		// Public, token-authed: incoming webhooks from git providers.
		r.Post("/webhooks/projects/{token}", projH.Webhook)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireUser(s.pool))
			r.Get("/me", authH.Me)

			if s.oauth != nil && len(s.oauth.All()) > 0 {
				oauthH := &oauth.Handlers{
					Registry:     s.oauth,
					RedirectBase: s.cfg.WebPublicURL,
					Secure:       s.cfg.CookieSecure,
				}
				r.Post("/auth/{provider}/start", oauthH.Start)
				r.Get("/auth/{provider}/callback", oauthH.Callback)
				r.Get("/me/connections", oauthH.ListConnections)
				r.Delete("/me/connections/{id}", oauthH.DeleteConnection)
			}

			r.Get("/workspaces", wsH.List)
			r.Post("/workspaces", wsH.Create)
			r.Get("/workspaces/{id}", wsH.Get)
			r.Delete("/workspaces/{id}", wsH.Delete)
			r.Get("/workspaces/{id}/members", wsH.ListMembers)
			r.Post("/workspaces/{id}/members", wsH.AddMember)
			r.Delete("/workspaces/{id}/members/{user_id}", wsH.RemoveMember)
			r.Get("/workspaces/{id}/activity", wsH.Activity)
			r.Get("/workspaces/{id}/recent-issues", boardH.RecentIssues)

			r.Get("/workspaces/{workspace_id}/projects", projH.List)
			r.Post("/workspaces/{workspace_id}/projects", projH.Create)
			r.Get("/projects/{id}", projH.Get)
			r.Get("/projects/{id}/overview", projH.Overview)
			r.Delete("/projects/{id}", projH.Delete)
			r.Post("/projects/{id}/sync", projH.Sync)
			r.Post("/projects/{id}/webhook-secret", projH.SetWebhookSecret)
			r.Get("/projects/{id}/stats", statsH.GetLatest)
			r.Get("/projects/{id}/branches", projH.ListBranches)
			r.Post("/projects/{id}/branch", projH.SwitchBranch)
			r.Post("/projects/{id}/pull-requests", projH.CreatePullRequest)
			r.Post("/projects/{id}/ci-trigger", projH.TriggerCI)

			r.Get("/projects/{id}/secret-findings", secScanH.ListByProject)
			r.Get("/projects/{id}/dependencies", depsH.ListByProject)
			r.Get("/projects/{id}/board", boardH.Get)
			r.Post("/projects/{id}/board", boardH.Setup)
			r.Post("/projects/{id}/board/sync", boardH.Sync)
			r.Get("/issues/{id}", boardH.GetIssue)
			r.Patch("/issues/{id}/move", boardH.MoveIssue)
			r.Post("/issues/{id}/comments", boardH.CreateComment)
			r.Post("/issues/{id}/close", boardH.CloseIssue)

			r.Get("/projects/{project_id}/pipelines", pipelineH.ListByProject)
			r.Post("/projects/{project_id}/pipelines", pipelineH.Create)
			r.Delete("/pipelines/{id}", pipelineH.Delete)
			r.Post("/pipelines/{id}/runs", pipelineH.Run)
			r.Get("/pipeline-runs/{id}", pipelineH.GetRun)

			r.Get("/projects/{id}/remote-pipelines", remoteciH.ListByProject)
			r.Post("/projects/{id}/remote-pipelines/trigger", remoteciH.TriggerPipeline)

			r.Get("/projects/{id}/pipeline-configs", cidetectH.ListByProject)

			r.Get("/projects/{id}/compose", composeH.GetConfig)
			r.Post("/projects/{id}/compose/start", composeH.Start)
			r.Post("/projects/{id}/compose/stop", composeH.Stop)
			r.Get("/projects/{id}/compose/status", composeH.GetStatus)

			r.Get("/projects/{project_id}/tasks", taskH.ListByProject)
			r.Get("/projects/{project_id}/runs", runH.ListByProject)
			r.Get("/projects/{project_id}/trends", runH.Trends)
			r.Get("/projects/{project_id}/activity", runH.Activity)
			r.Get("/me/runs", runH.ListForUser)
			r.Get("/me/dashboard", dashH.Get)
			r.Get("/workspaces/{id}/dashboard", dashH.WorkspaceSummary)
			r.Get("/runs/{id}/artifacts", artifactH.ListByRun)
			r.Get("/artifacts/{id}/download", artifactH.Download)
			r.Get("/me/flaky-tasks", runH.ListFlaky)
			r.Get("/me/usage", quotaH.Get)

			r.Post("/tasks/{id}/runs", runH.Create)
			r.Patch("/tasks/{id}", taskH.Update)
			r.Post("/tasks/{id}/pin", pinH.Pin)
			r.Delete("/tasks/{id}/pin", pinH.Unpin)
			r.Get("/me/pinned-tasks", pinH.ListForUser)
			r.Get("/runs/{id}", runH.Get)
			r.Get("/runs/{id}/log", runH.GetLog)
			r.Get("/runs/{id}/log/stream", runH.Stream)
			r.Get("/runs/{id}/compare", runH.Compare)
			r.Post("/runs/{id}/cancel", runH.Cancel)
			r.Post("/runs/{id}/approve", runH.Approve)
			r.Get("/runs/{id}/comments", commentH.List)
			r.Post("/runs/{id}/comments", commentH.Create)
			r.Delete("/runs/{run_id}/comments/{id}", commentH.Delete)
			r.Get("/me/mentions", commentH.ListMentions)
			r.Get("/me/mentions/count", commentH.MentionsCount)
			r.Post("/me/mentions/read", commentH.MarkRead)

			r.Get("/projects/{project_id}/schedules", schedH.ListByProject)
			r.Post("/projects/{project_id}/schedules", schedH.Create)
			r.Delete("/schedules/{id}", schedH.Delete)
			r.Post("/schedules/{id}/toggle", schedH.Toggle)

			r.Get("/projects/{project_id}/images", imgH.ListByProject)
			r.Get("/images/{id}/sbom", imgH.SBOM)

			r.Get("/me/notifications", notifH.List)
			r.Post("/me/notifications", notifH.Create)
			r.Delete("/me/notifications/{id}", notifH.Delete)
			r.Post("/me/notifications/{id}/test", notifH.Test)
			r.Get("/me/notification-subscriptions", notifH.ListSubscriptions)
			r.Post("/me/notification-subscriptions", notifH.CreateSubscription)
			r.Delete("/me/notification-subscriptions/{id}", notifH.DeleteSubscription)

			r.Get("/me/ssh-keys", sshH.List)
			r.Post("/me/ssh-keys", sshH.Create)
			r.Delete("/me/ssh-keys/{id}", sshH.Delete)

			r.Get("/me/pat-credentials", patH.List)
			r.Post("/me/pat-credentials", patH.Create)
			r.Delete("/me/pat-credentials/{id}", patH.Delete)

			r.Get("/me/outbound-webhooks", eventsH.List)
			r.Post("/me/outbound-webhooks", eventsH.Create)
			r.Delete("/me/outbound-webhooks/{id}", eventsH.Delete)
			r.Post("/me/outbound-webhooks/{id}/test", eventsH.Test)

			r.Get("/me/push/vapid", pushH.VAPID)
			r.Post("/me/push/subscribe", pushH.Subscribe)
			r.Delete("/me/push/subscribe", pushH.Unsubscribe)
			r.Get("/me/push/subscriptions", pushH.List)
			r.Delete("/me/push/subscriptions/{id}", pushH.DeleteByID)

			r.Group(func(r chi.Router) {
				r.Use(admin.RequireAdmin(s.pool))
				r.Get("/admin/users", adminH.ListUsers)
				r.Get("/admin/audit-log", adminH.AuditLog)
			})
		})
	})

	return r
}
