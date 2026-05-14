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
	"workend/api/internal/graph"
	"workend/api/internal/health"
	"workend/api/internal/monorepo"
	"workend/api/internal/image"
	"workend/api/internal/repo"
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
	"workend/api/internal/rbac"
	"workend/api/internal/run"
	"workend/api/internal/sandbox"
	"workend/api/internal/schedule"
	"workend/api/internal/secret"
	"workend/api/internal/secscan"
	"workend/api/internal/signing"
	"workend/api/internal/sshkey"
	"workend/api/internal/stats"
	"workend/api/internal/task"
	"workend/api/internal/widget"
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
		Notify:    s.notif,
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
	dashH := &dashboard.Handlers{Pool: s.pool, ReposRoot: s.cfg.ReposRoot}
	imgH := &image.Handlers{Pool: s.pool, Box: s.box}
	notifH := &notify.Handlers{D: s.notif}
	quotaH := &quota.Handlers{Pool: s.pool, ReposRoot: s.cfg.ReposRoot, Logger: s.logger, Notify: s.notif}
	pinH := &pin.Handlers{Pool: s.pool}
	commentH := &comment.Handlers{Pool: s.pool}
	boardH := &board.Handlers{Pool: s.pool, OAuth: s.oauth}
	pipelineH := &pipeline.Handlers{Pool: s.pool, Runner: s.runH, Logger: s.logger}
	artifactH := &artifact.Handlers{Pool: s.pool, ArtifactsRoot: s.cfg.ArtifactsRoot, Logger: s.logger}
	remoteciH := &remoteci.Handlers{Pool: s.pool, OAuth: s.oauth, Logger: s.logger}
	cidetectH := &cidetect.Handlers{Pool: s.pool}
	composeH := &compose.Handlers{Pool: s.pool, Dagger: s.dagger, Box: s.box, Logger: s.logger}
	depsH := &deps.Handlers{Pool: s.pool}
	repoH := &repo.Handlers{Pool: s.pool}
	auditAdminH := &audit.AdminHandlers{Pool: s.pool}
	favH := &project.FavoriteHandlers{Pool: s.pool}
	wsSecretH := &workspace.SecretHandlers{Pool: s.pool, Box: s.box}
	activityH := &workspace.ActivityFeedHandlers{Pool: s.pool, Logger: s.logger}
	sandboxH := &sandbox.Handlers{Pool: s.pool, Dagger: s.dagger, ReposRoot: s.cfg.ReposRoot, Logger: s.logger, WebURL: s.cfg.WebPublicURL}
	shellH := &sandbox.ShellHandlers{Pool: s.pool, Logger: s.logger, WebURL: s.cfg.WebPublicURL}
	previewH := &sandbox.PreviewHandlers{Pool: s.pool, WebURL: s.cfg.WebPublicURL, Logger: s.logger}
	graphH := &graph.Handlers{Pool: s.pool, ReposRoot: s.cfg.ReposRoot, Logger: s.logger}
	monorepoH := &monorepo.Handlers{Pool: s.pool, ReposRoot: s.cfg.ReposRoot, Logger: s.logger}
	rbacH := &rbac.Handlers{Pool: s.pool, Logger: s.logger}
	signingH := &signing.Handlers{Pool: s.pool, Logger: s.logger}
	widgetH := &widget.Handlers{Pool: s.pool, ReposRoot: s.cfg.ReposRoot, Logger: s.logger}
	runH := s.runH
	schedH := s.schedH

	r.Route("/api", func(r chi.Router) {
		r.Get("/info", apiInfoHandler())
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

		// Public, token-authed: shared run read-only view.
		r.Get("/shared/runs/{token}", runH.GetSharedRun)

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
			r.Get("/workspaces/{id}/retention", wsH.GetRetention)
			r.Put("/workspaces/{id}/retention", wsH.SetRetention)
			r.Get("/workspaces/{id}/secrets", wsSecretH.ListSecrets)
			r.Post("/workspaces/{id}/secrets", wsSecretH.CreateSecret)
			r.Put("/workspaces/{id}/secrets/{secret_id}", wsSecretH.UpdateSecret)
			r.Delete("/workspaces/{id}/secrets/{secret_id}", wsSecretH.DeleteSecret)
			r.Get("/workspaces/{id}/task-templates", taskH.ListTemplates)
			r.Post("/workspaces/{id}/task-templates", taskH.CreateTemplate)
			r.Get("/workspaces/{id}/feed", activityH.GetFeed)
			r.Get("/workspaces/{id}/feed/stream", activityH.GetFeedSSE)
			r.Get("/workspaces/{id}/sandboxes", sandboxH.List)
			r.Get("/workspaces/{id}/dependency-map", graphH.GetDependencyMap)
			r.Post("/workspaces/{id}/detect-incidents", graphH.DetectIncidents)
			r.Get("/workspaces/{id}/incidents", graphH.ListIncidents)
			r.Get("/workspaces/{id}/incident-timeline", graphH.GetTimeline)
			r.Get("/workspaces/{id}/roles", rbacH.ListRoles)
			r.Post("/workspaces/{id}/roles", rbacH.CreateRole)
			r.Post("/workspaces/{id}/members/{user_id}/role", rbacH.AssignRole)
			r.Get("/workspaces/{id}/members/{user_id}/role", rbacH.GetMemberRole)
			r.Get("/workspaces/{id}/signing-keys", signingH.ListKeys)
			r.Post("/workspaces/{id}/signing-keys", signingH.GenerateKey)

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

			r.Get("/projects/{id}/ci-triggers", remoteciH.ListTriggerRules)
			r.Post("/projects/{id}/ci-triggers", remoteciH.CreateTriggerRule)
			r.Delete("/ci-triggers/{id}", remoteciH.DeleteTriggerRule)
			r.Post("/ci-triggers/{id}/toggle", remoteciH.ToggleTriggerRule)

			r.Get("/projects/{id}/registry", imgH.GetRegistryConfig)
			r.Put("/projects/{id}/registry", imgH.SetRegistryConfig)
			r.Delete("/projects/{id}/registry", imgH.DeleteRegistryConfig)

			r.Get("/projects/{id}/search", repoH.SearchInProject)
			r.Get("/projects/{id}/metrics", runH.GetProjectMetrics)
			r.Post("/projects/{id}/favorite", favH.AddFavorite)
			r.Delete("/projects/{id}/favorite", favH.RemoveFavorite)
			r.Post("/projects/{id}/view", favH.RecordView)
			r.Get("/projects/{id}/previews", previewH.List)
			r.Post("/projects/{id}/previews", previewH.Create)
			r.Post("/projects/{id}/scan-dependencies", graphH.ScanProjectDeps)
			r.Get("/projects/{id}/dependencies-graph", graphH.GetProjectDeps)
			r.Post("/projects/{id}/detect-inputs", graphH.AutoDetectInputs)
			r.Post("/projects/{id}/predict-impact", graphH.PredictImpact)
			r.Post("/projects/{id}/predict-impact/diff", graphH.PredictImpactFromDiff)
			r.Post("/projects/{id}/sync-commits", graphH.SyncCommitHistory)
			r.Get("/projects/{id}/blame-timeline", graphH.GetBlameTimeline)
			r.Post("/projects/{id}/monorepo/detect", monorepoH.DetectAndSync)
			r.Get("/projects/{id}/monorepo/packages", monorepoH.ListPackages)
			r.Post("/projects/{id}/monorepo/auto-map", monorepoH.AutoMapTasks)

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
			r.Get("/me/favorites", favH.ListFavorites)
			r.Get("/me/recent-projects", favH.ListRecent)
			r.Get("/me/alert-rules", runH.ListAlertRules)
			r.Delete("/alert-rules/{id}", runH.DeleteAlertRule)
			r.Post("/alert-rules/{id}/toggle", runH.ToggleAlertRule)
			r.Delete("/task-templates/{id}", taskH.DeleteTemplate)
			r.Post("/task-templates/{id}/apply", taskH.ApplyTemplate)
			r.Delete("/run-shares/{id}", runH.RevokeShare)
			r.Patch("/roles/{id}", rbacH.UpdateRole)
			r.Delete("/roles/{id}", rbacH.DeleteRole)
			r.Post("/signing-keys/{id}/revoke", signingH.RevokeKey)
			r.Get("/commits/{id}", graphH.GetCommitDetail)
			r.Get("/packages/{id}", monorepoH.GetPackage)
			r.Post("/packages/{id}/task-scopes", monorepoH.AddTaskScope)
			r.Delete("/task-scopes/{id}", monorepoH.RemoveTaskScope)
			r.Get("/me/sandboxes", sandboxH.ListForUser)
			r.Post("/sandboxes", sandboxH.Create)
			r.Get("/sandboxes/{id}", sandboxH.Get)
			r.Post("/sandboxes/{id}/destroy", sandboxH.Destroy)
			r.Post("/sandboxes/{id}/extend", sandboxH.Extend)
			r.Post("/sandboxes/{id}/shell", shellH.CreateSession)
			r.Get("/sandboxes/{id}/shell-sessions", shellH.ListSessions)
			r.Post("/shell-sessions/{id}/exec", shellH.ExecCommand)
			r.Post("/shell-sessions/{id}/close", shellH.CloseSession)
			r.Get("/previews/{id}", previewH.Get)
			r.Post("/previews/{id}/redeploy", previewH.Redeploy)
			r.Post("/previews/{id}/stop", previewH.Stop)
			r.Delete("/previews/{id}", previewH.Delete)
			r.Post("/previews/{id}/auto-deploy", previewH.ToggleAutoDeploy)
			r.Get("/incidents/{id}", graphH.GetIncident)
			r.Patch("/incidents/{id}", graphH.UpdateIncident)
			r.Delete("/file-inputs/{id}", graphH.RemoveInput)

			r.Post("/tasks/{id}/runs", runH.Create)
			r.Patch("/tasks/{id}", taskH.Update)
			r.Post("/tasks/{id}/pin", pinH.Pin)
			r.Delete("/tasks/{id}/pin", pinH.Unpin)
			r.Post("/tasks/{id}/quarantine", taskH.Quarantine)
			r.Post("/tasks/{id}/unquarantine", taskH.Unquarantine)
			r.Get("/tasks/{id}/presets", runH.ListPresets)
			r.Post("/tasks/{id}/presets", runH.CreatePreset)
			r.Delete("/presets/{id}", runH.DeletePreset)
			r.Get("/tasks/{id}/concurrency", runH.GetConcurrencyStatus)
			r.Get("/tasks/{id}/metrics", runH.GetTaskMetrics)
			r.Get("/tasks/{id}/metrics/trends", runH.GetTaskTrends)
			r.Get("/tasks/{id}/resources", runH.GetResourceTrends)
			r.Post("/tasks/{id}/alert-rules", runH.CreateAlertRule)
			r.Get("/tasks/{id}/file-inputs", graphH.ListInputs)
			r.Post("/tasks/{id}/file-inputs", graphH.AddInput)
			r.Get("/me/pinned-tasks", pinH.ListForUser)
			r.Get("/runs/{id}", runH.Get)
			r.Get("/runs/{id}/log", runH.GetLog)
			r.Get("/runs/{id}/log/search", runH.SearchLog)
			r.Get("/runs/{id}/log/stream", runH.Stream)
			r.Get("/runs/{id}/compare", runH.Compare)
			r.Get("/runs/{id}/artifacts/compare/{right_id}", artifactH.Compare)
			r.Get("/runs/{id}/diff/{other_id}", runH.DiffRuns)
			r.Post("/runs/{id}/share", runH.CreateShare)
			r.Get("/runs/{id}/shares", runH.ListShares)
			r.Post("/runs/{id}/sign", signingH.SignRun)
			r.Get("/runs/{id}/verify", signingH.VerifyRun)
			r.Get("/runs/{id}/provenance", signingH.GetProvenance)
			r.Post("/runs/{id}/archive", runH.ArchiveRun)
			r.Post("/runs/{id}/unarchive", runH.UnarchiveRun)
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
			r.Post("/me/ssh-keys/{id}/rotate", sshH.RotateKey)

			r.Get("/me/pat-credentials", patH.List)
			r.Post("/me/pat-credentials", patH.Create)
			r.Delete("/me/pat-credentials/{id}", patH.Delete)
			r.Post("/me/pat-credentials/{id}/rotate", patH.RotateCredential)

			r.Get("/me/outbound-webhooks", eventsH.List)
			r.Post("/me/outbound-webhooks", eventsH.Create)
			r.Delete("/me/outbound-webhooks/{id}", eventsH.Delete)
			r.Post("/me/outbound-webhooks/{id}/test", eventsH.Test)

			r.Get("/me/push/vapid", pushH.VAPID)
			r.Post("/me/push/subscribe", pushH.Subscribe)
			r.Delete("/me/push/subscribe", pushH.Unsubscribe)
			r.Get("/me/push/subscriptions", pushH.List)
			r.Delete("/me/push/subscriptions/{id}", pushH.DeleteByID)

			r.Get("/widgets/dashboard/run-pulse", widgetH.RunPulse)
			r.Get("/widgets/dashboard/failure-heatmap", widgetH.FailureHeatmap)
			r.Get("/widgets/dashboard/my-queue", widgetH.MyQueue)
			r.Get("/widgets/dashboard/sprint-velocity", widgetH.SprintVelocity)
			r.Get("/widgets/dashboard/sandbox-status", widgetH.SandboxStatus)
			r.Get("/widgets/dashboard/quota-meter", widgetH.QuotaMeter)
			r.Get("/widgets/workspaces/health", widgetH.HealthCards)
			r.Get("/widgets/workspaces/comparison", widgetH.Comparison)
			r.Get("/widgets/workspaces/incidents", widgetH.IncidentBanner)
			r.Get("/widgets/workspaces/storage", widgetH.StorageBreakdown)
			r.Get("/widgets/workspaces/templates", widgetH.Templates)
			r.Get("/widgets/workspaces/pending-members", widgetH.InvitePending)
			r.Get("/widgets/workspace/{id}/project-grid", widgetH.ProjectGrid)
			r.Get("/widgets/workspace/{id}/team-presence", widgetH.TeamPresence)
			r.Get("/widgets/workspace/{id}/dependency-overview", widgetH.DependencyOverview)
			r.Get("/widgets/workspace/{id}/secrets-summary", widgetH.SecretsSummary)
			r.Get("/widgets/workspace/{id}/pipeline-board", widgetH.PipelineBoard)
			r.Get("/widgets/workspace/{id}/sandbox-preview-rack", widgetH.SandboxPreviewRack)
			r.Get("/widgets/project/{id}/impact-radar", widgetH.ImpactRadar)
			r.Get("/widgets/project/{id}/dependency-tree", widgetH.DependencyTree)
			r.Get("/widgets/project/{id}/live-run", widgetH.LiveLogPreview)
			r.Get("/widgets/project/{id}/alert-rules", widgetH.AlertRulesPanel)

			r.Group(func(r chi.Router) {
				r.Use(admin.RequireAdmin(s.pool))
				r.Get("/admin/users", adminH.ListUsers)
				r.Get("/admin/audit-log", adminH.AuditLog)
				r.Get("/admin/audit-log/export", auditAdminH.ExportAuditLog)
				r.Post("/admin/audit-log/retention", auditAdminH.SetRetention)
			})
		})
	})

	return r
}
