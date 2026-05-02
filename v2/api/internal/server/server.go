package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/config"
	"workend/api/internal/health"
	"workend/api/internal/workspace"
)

type Server struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func New(cfg *config.Config, pool *pgxpool.Pool, logger *slog.Logger) *Server {
	return &Server{cfg: cfg, pool: pool, logger: logger}
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
		})
	})

	return r
}
