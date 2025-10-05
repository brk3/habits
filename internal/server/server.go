package server

import (
	"net/http"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/logger"
	"github.com/brk3/habits/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

type Server struct {
	store         storage.Store
	authProviders map[string]*AuthProvider
	cfg           *config.Config
	sessionCookie *securecookie.SecureCookie
}

type AuthProvider struct {
	name       string
	oauth2     *oauth2.Config
	oidcProv   *oidc.Provider
	idVerifier *oidc.IDTokenVerifier
	state      *StateStore
}

func New(cfg *config.Config, store storage.Store) (*Server, error) {
	logger.Info("Initializing server", "auth_enabled", cfg.AuthEnabled)
	srv := &Server{
		store: store,
		cfg:   cfg,
	}

	if cfg.AuthEnabled {
		var err error
		srv.authProviders, srv.sessionCookie, err = ConfigureOIDCProviders(cfg)
		if err != nil {
			return nil, err
		}
	}

	logger.Info("Server initialization complete")
	return srv, nil
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestID)
	r.Use(metricsMiddleware)

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/version", s.getVersionInfo)

	if s.cfg.AuthEnabled {
		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", s.simpleLogin)
			r.Get("/login/{id}", s.login)
			r.Get("/callback/{id}", s.callback)
			r.Get("/logout", s.logout)
			r.Get("/get_api_token", s.getAPIToken)

			// API key management (requires auth)
			r.Group(func(r chi.Router) {
				r.Use(s.authMiddleware)
				r.Post("/api_keys", s.generateAPIKey)
				r.Get("/api_keys", s.listAPIKeys)
				r.Delete("/api_keys/{keyHash}", s.deleteAPIKey)
			})
		})
	}

	r.Route("/habits", func(r chi.Router) {
		if s.cfg.AuthEnabled {
			r.Use(s.authMiddleware)
			r.Use(s.userAwareMetricsMiddleware)
		}
		r.Post("/", s.trackHabit)
		r.Get("/", s.listHabits)
		r.Get("/{habit_id}", s.getHabit)
		r.Get("/{habit_id}/summary", s.getHabitSummary)
		r.Delete("/{habit_id}", s.deleteHabit)
	})

	return r
}
