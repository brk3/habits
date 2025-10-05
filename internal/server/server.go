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
	authProviders map[string]*AuthProvider
	cfg           *config.Config
	sessionCookie *securecookie.SecureCookie
	tokenStore    *TokenStore
}

// TokenStore wraps both in-memory cache and persistent storage
type TokenStore struct {
	ttl   time.Duration
	mu    sync.RWMutex
	m     map[string]*StoredToken
	store storage.Store
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
		store:      store,
		cfg:        cfg,
		tokenStore: NewTokenStore(24*time.Hour, store), // 24 hour cleanup - aligns with typical OIDC refresh token lifetimes
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

type StoredToken struct {
	Token    *oauth2.Token
	ExpireAt time.Time
}

func NewTokenStore(ttl time.Duration, store storage.Store) *TokenStore {
	s := &TokenStore{
		ttl:   ttl,
		m:     make(map[string]*StoredToken),
		store: store,
	}

	// Start janitor for in-memory cleanup
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			s.mu.Lock()
			for k, v := range s.m {
				if now.After(v.ExpireAt) {
					delete(s.m, k)
					logger.Debug("Cleaned up expired token from memory", "userID", k)
				}
			}
			s.mu.Unlock()
		}
	}()

	logger.Info("TokenStore initialized with persistent storage")
	return s
}

func (ts *TokenStore) Put(userID string, token *oauth2.Token) {
	stored := &StoredToken{
		Token:    token,
		ExpireAt: time.Now().Add(ts.ttl),
	}

	// Store in memory
	ts.mu.Lock()
	ts.m[userID] = stored
	ts.mu.Unlock()

	// Persist to storage
	if ts.store != nil {
		if err := ts.store.PutRefreshToken(userID, token); err != nil {
			logger.Error("Failed to persist refresh token", "userID", userID, "error", err)
		} else {
			logger.Debug("Refresh token persisted to storage", "userID", userID)
		}
	}
}

func (ts *TokenStore) Get(userID string) (*oauth2.Token, bool) {
	// Try memory first
	ts.mu.RLock()
	stored, ok := ts.m[userID]
	ts.mu.RUnlock()

	if ok && time.Now().Before(stored.ExpireAt) {
		logger.Debug("Token retrieved from memory cache", "userID", userID)
		return stored.Token, true
	}

	// Fall back to persistent storage
	if ts.store != nil {
		token, found, err := ts.store.GetRefreshToken(userID)
		if err != nil {
			logger.Error("Failed to retrieve token from storage", "userID", userID, "error", err)
			return nil, false
		}
		if found {
			// Populate memory cache
			ts.mu.Lock()
			ts.m[userID] = &StoredToken{
				Token:    token,
				ExpireAt: time.Now().Add(ts.ttl),
			}
			ts.mu.Unlock()
			logger.Debug("Token retrieved from persistent storage", "userID", userID)
			return token, true
		}
	}

	logger.Debug("No token found in memory or storage", "userID", userID)
	return nil, false
}

func (ts *TokenStore) Delete(userID string) {
	// Delete from memory
	ts.mu.Lock()
	delete(ts.m, userID)
	ts.mu.Unlock()

	// Delete from persistent storage
	if ts.store != nil {
		if err := ts.store.DeleteRefreshToken(userID); err != nil {
			logger.Error("Failed to delete refresh token from storage", "userID", userID, "error", err)
		} else {
			logger.Debug("Refresh token deleted from storage", "userID", userID)
		}
	}
}
