package server

import (
	"net/http"
	"sync"
	"time"

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
	authConf      map[string]*AuthConfig
	cfg           *config.Config
	sessionCookie *securecookie.SecureCookie
	tokenStore    *TokenStore
}

type AuthConfig struct {
	name       string
	oauth2     *oauth2.Config
	oidcProv   *oidc.Provider
	idVerifier *oidc.IDTokenVerifier
	state      *StateStore
}

func New(cfg *config.Config, store storage.Store) (*Server, error) {
	logger.Info("Initializing server", "auth_enabled", cfg.AuthEnabled)
	srv := &Server{
		store:      store,
		cfg:        cfg,
		tokenStore: NewTokenStore(24 * time.Hour), // 24 hour cleanup interval
	}

	if cfg.AuthEnabled {
		var err error
		srv.authConf, srv.sessionCookie, err = ConfigureOIDCProviders(cfg)
		if err != nil {
			return nil, err
		}
		// TokenStore handles its own cleanup
	}

	logger.Info("Server initialization complete")
	return srv, nil
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
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
		})
	}

	r.Route("/habits", func(r chi.Router) {
		if s.cfg.AuthEnabled {
			r.Use(s.authMiddleware)
		}
		r.Post("/", s.trackHabit)
		r.Get("/", s.listHabits)
		r.Get("/{habit_id}", s.getHabit)
		r.Get("/{habit_id}/summary", s.getHabitSummary)
		r.Delete("/{habit_id}", s.deleteHabit)
	})

	return r
}

type TokenStore struct {
	ttl time.Duration
	mu  sync.RWMutex
	m   map[string]*StoredToken
}

type StoredToken struct {
	Token    *oauth2.Token
	ExpireAt time.Time
}

func NewTokenStore(ttl time.Duration) *TokenStore {
	s := &TokenStore{ttl: ttl, m: make(map[string]*StoredToken)}
	go func() { // janitor - similar to StateStore
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			s.mu.Lock()
			for k, v := range s.m {
				if now.After(v.ExpireAt) {
					delete(s.m, k)
					logger.Debug("Cleaned up expired token", "userID", k)
				}
			}
			s.mu.Unlock()
		}
	}()
	return s
}

func (s *TokenStore) Put(userID string, token *oauth2.Token) {
	s.mu.Lock()
	s.m[userID] = &StoredToken{
		Token:    token,
		ExpireAt: time.Now().Add(s.ttl),
	}
	s.mu.Unlock()
}

func (s *TokenStore) Get(userID string) (*oauth2.Token, bool) {
	s.mu.RLock()
	stored, ok := s.m[userID]
	s.mu.RUnlock()
	if !ok {
		logger.Debug("No token found in store", "userID", userID)
		return nil, false
	}
	if time.Now().After(stored.ExpireAt) {
		logger.Debug("Token in store has expired", "userID", userID, "expireAt", stored.ExpireAt, "now", time.Now())
		return nil, false
	}
	logger.Debug("Token retrieved from store", "userID", userID, "expireAt", stored.ExpireAt)
	return stored.Token, true
}

func (s *TokenStore) Delete(userID string) {
	s.mu.Lock()
	delete(s.m, userID)
	s.mu.Unlock()
}
