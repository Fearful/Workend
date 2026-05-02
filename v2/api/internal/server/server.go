package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/config"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/health"
	"workend/api/internal/project"
	"workend/api/internal/run"
	"workend/api/internal/task"
	"workend/api/internal/workspace"
)

type Server struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	dagger *wdagger.Client
	logger *slog.Logger
}

func New(cfg *config.Config, pool *pgxpool.Pool, dc *wdagger.Client, logger *slog.Logger) *Server {
	return &Server{cfg: cfg, pool: pool, dagger: dc, logger: logger}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Handle("/healthz", &health.Handler{
		Pool:           s.pool,
		DaggerSockPath: s.cfg.DaggerSockPath,
	})

	authH := &auth.Handlers{Pool: s.pool, Secure: s.cfg.CookieSecure}
	wsH := &workspace.Handlers{Pool: s.pool}
	projH := &project.Handlers{
		Pool:      s.pool,
		Dagger:    s.dagger,
		ReposRoot: s.cfg.ReposRoot,
		Logger:    s.logger,
	}
	taskH := &task.Handlers{Pool: s.pool}
	runH := &run.Handlers{
		Pool:     s.pool,
		Dagger:   s.dagger,
		LogsRoot: s.cfg.LogsRoot,
		Logger:   s.logger,
	}
	runH.Init()

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/signup", authH.Signup)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/logout", authH.Logout)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireUser(s.pool))
			r.Get("/me", authH.Me)

			r.Get("/workspaces", wsH.List)
			r.Post("/workspaces", wsH.Create)
			r.Get("/workspaces/{id}", wsH.Get)
			r.Delete("/workspaces/{id}", wsH.Delete)

			r.Get("/workspaces/{workspace_id}/projects", projH.List)
			r.Post("/workspaces/{workspace_id}/projects", projH.Create)
			r.Get("/projects/{id}", projH.Get)
			r.Delete("/projects/{id}", projH.Delete)
			r.Post("/projects/{id}/sync", projH.Sync)

			r.Get("/projects/{project_id}/tasks", taskH.ListByProject)
			r.Get("/projects/{project_id}/runs", runH.ListByProject)

			r.Post("/tasks/{id}/runs", runH.Create)
			r.Get("/runs/{id}", runH.Get)
			r.Get("/runs/{id}/log", runH.GetLog)
		})
	})

	return r
}
