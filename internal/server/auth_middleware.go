package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/logger"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

type userCtxKey struct{}

type User struct {
	Subject string
	Email   string
	UserID  string
	Claims  map[string]any
}

type StateStore struct {
	ttl time.Duration
	mu  sync.Mutex
	m   map[string]authState
}

type authState struct {
	Verifier string
	Return   string
	ExpireAt time.Time
}

func NewStateStore(ttl time.Duration) *StateStore {
	s := &StateStore{ttl: ttl, m: make(map[string]authState)}
	go func() { // janitor
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			s.mu.Lock()
			for k, v := range s.m {
				if now.After(v.ExpireAt) {
					delete(s.m, k)
				}
			}
			s.mu.Unlock()
		}
	}()
	return s
}

func ConfigureOIDCProviders(cfg *config.Config) (map[string]*AuthConfig, *securecookie.SecureCookie, error) {
	logger.Info("Configuring OIDC providers", "count", len(cfg.OIDCProviders))
	authConf := make(map[string]*AuthConfig)

	sessionHashKey := securecookie.GenerateRandomKey(32)
	sessionBlockKey := securecookie.GenerateRandomKey(32)
	sessionCookie := securecookie.New(sessionHashKey, sessionBlockKey)
	sessionCookie.MaxAge(259200) // 3 days

	for i := range cfg.OIDCProviders {
		cfgprov := cfg.OIDCProviders[i]
		id := cfgprov.Id
		name := cfgprov.Name
		clientID := cfgprov.ClientID
		clientSecret := cfgprov.ClientSecret
		issuerURL := cfgprov.IssuerURL
		redirectURL := cfgprov.RedirectURL
		scopes := cfgprov.Scopes

		logger.Debug("Setting up OIDC provider", "id", id, "name", name, "issuer", issuerURL)
		prov, err := oidc.NewProvider(context.Background(), issuerURL)
		if err != nil {
			logger.Error("Failed to create OIDC provider", "id", id, "error", err)
			return nil, nil, fmt.Errorf("failed to create OIDC provider: %w", err)
		}

		verifier := prov.Verifier(&oidc.Config{ClientID: clientID})
		oauth2Cfg := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     prov.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       scopes,
		}

		authConf[id] = &AuthConfig{
			name:       name,
			oauth2:     oauth2Cfg,
			oidcProv:   prov,
			idVerifier: verifier,
			state:      NewStateStore(5 * time.Minute),
		}
		logger.Info("OIDC provider configured successfully", "id", id, "name", name)
	}

	return authConf, sessionCookie, nil
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rawIDToken string
		var providerID string

		// 1) Try session cookie first
		if c, err := r.Cookie("session"); err == nil {
			var prefixedToken string
			if err := s.sessionCookie.Decode("session", c.Value, &prefixedToken); err == nil {
				// Parse provider-prefixed token
				if pID, token, err := parseProviderToken(prefixedToken); err == nil {
					providerID, rawIDToken = pID, token
				} else {
					logger.Debug("Failed to parse session token", "error", err)
				}
			}
		}

		// 2) Try Bearer token if no valid session cookie
		if rawIDToken == "" {
			if ah := r.Header.Get("Authorization"); strings.HasPrefix(ah, "Bearer ") {
				token := strings.TrimPrefix(ah, "Bearer ")
				// Parse provider-prefixed token: "provider:jwt"
				if parsedProviderID, parsedToken, err := parseProviderToken(token); err == nil {
					if _, exists := s.authConf[parsedProviderID]; exists {
						providerID = parsedProviderID
						rawIDToken = parsedToken
					} else {
						logger.Debug("Unknown provider in Bearer token", "provider", parsedProviderID)
					}
				} else {
					logger.Debug("Failed to parse Bearer token", "error", err)
				}
			}
		}

		// 3) No valid token → redirect for HTML, 401 for API
		if rawIDToken == "" || providerID == "" {
			s.handleAuthFailure(w, r, false)
			return
		}

		// 4) Verify token with the correct provider
		idTok, err := s.authConf[providerID].idVerifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			// Try to refresh the token before giving up
			if newIDToken, refreshed := s.tryRefreshToken(r.Context(), providerID, rawIDToken); refreshed {
				if newIdTok, verifyErr := s.authConf[providerID].idVerifier.Verify(r.Context(), newIDToken); verifyErr == nil {
					prefixedToken := providerID + ":" + newIDToken
					val, _ := s.sessionCookie.Encode("session", prefixedToken)
					http.SetCookie(w, &http.Cookie{
						Name:     "session",
						Value:    val,
						Path:     "/",
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteLaxMode,
						MaxAge:   int((3 * 24 * time.Hour).Seconds()),
					})
					idTok = newIdTok
				} else {
					logger.Debug("New ID token verification failed", "error", verifyErr)
					s.handleAuthFailure(w, r, true)
					return
				}
			} else {
				logger.Debug("Token verification failed and refresh unsuccessful", "error", err)
				s.handleAuthFailure(w, r, true)
				return
			}
		}

		// 5) Extract claims and create user
		var claims map[string]any
		_ = idTok.Claims(&claims)
		u := &User{
			Subject: idTok.Subject,
			Email:   strClaim(claims, "email"),
			UserID:  userIDFromClaims(claims),
			Claims:  claims,
		}

		// Inject user into context
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userCtxKey{}, u)))
	})
}

// middleware helpers
func acceptsHTML(accept string) bool {
	return strings.Contains(accept, "text/html") || accept == ""
}

// parseProviderToken parses a provider-prefixed token of the format "provider:jwt"
// Returns the provider ID and JWT token, or empty strings and error if invalid format
func parseProviderToken(token string) (providerID, jwt string, err error) {
	if token == "" {
		return "", "", fmt.Errorf("empty token")
	}

	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid token format: expected 'provider:jwt'")
	}

	providerID, jwt = parts[0], parts[1]
	if providerID == "" {
		return "", "", fmt.Errorf("empty provider ID")
	}
	if jwt == "" {
		return "", "", fmt.Errorf("empty JWT token")
	}

	return providerID, jwt, nil
}

func strClaim(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

// userIDFromClaims generates a consistent user ID from OIDC token claims
func userIDFromClaims(claims map[string]any) string {
	iss, ok := claims["iss"].(string)
	if !ok {
		return ""
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return ""
	}

	userInfo := iss + "|" + sub
	hash := sha256.Sum256([]byte(userInfo))
	return fmt.Sprintf("user-%x", hash[:8])
}

// userIDFromContext extracts user ID from authenticated request context
func userIDFromContext(authEnabled bool, r *http.Request) string {
	if !authEnabled {
		logger.Debug("Auth disabled, using anonymous userid")
		return "anonymous"
	}

	user, ok := r.Context().Value(userCtxKey{}).(*User)
	if !ok {
		logger.Error("No user in context")
		return ""
	}

	return user.UserID
}

// parseTokenClaims extracts claims from a JWT token without verification
func (s *Server) parseTokenClaims(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload
	payload := parts[1]
	if mod := len(payload) % 4; mod != 0 {
		payload += strings.Repeat("=", 4-mod)
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return claims, nil
}

func (s *StateStore) Put(key string, v authState) {
	s.mu.Lock()
	s.m[key] = v
	s.mu.Unlock()
}

func (s *StateStore) GetAndDelete(key string) (authState, bool) {
	s.mu.Lock()
	v, ok := s.m[key]
	if ok {
		delete(s.m, key)
	}
	s.mu.Unlock()
	if ok && time.Now().After(v.ExpireAt) {
		return authState{}, false
	}
	return v, ok
}

// handleAuthFailure centralizes auth failure handling
func (s *Server) handleAuthFailure(w http.ResponseWriter, r *http.Request, clearCookie bool) {
	if clearCookie {
		// Clear session cookie on invalid token
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1})
	}

	if r.Method == http.MethodGet && acceptsHTML(r.Header.Get("Accept")) {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
	} else {
		if clearCookie {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
		} else {
			w.Header().Set("WWW-Authenticate", `Bearer realm="habits"`)
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
}

// tryRefreshToken attempts to refresh an expired ID token using stored refresh token
func (s *Server) tryRefreshToken(ctx context.Context, providerID, expiredIDToken string) (string, bool) {
	// Extract claims from expired token (without verification since it's expired)
	claims, err := s.parseTokenClaims(expiredIDToken)
	if err != nil {
		logger.Debug("Failed to parse token claims", "error", err)
		return "", false
	}

	userID := userIDFromClaims(claims)
	if userID == "" {
		logger.Debug("Failed to calculate user ID from claims")
		return "", false
	}

	// Look up refresh token
	s.refreshMutex.RLock()
	refreshToken, exists := s.refreshTokens[userID]
	s.refreshMutex.RUnlock()

	if !exists {
		logger.Debug("No refresh token found for user", "userID", userID)
		return "", false
	}

	// Attempt to refresh
	authConfig := s.authConf[providerID]
	token := &oauth2.Token{RefreshToken: refreshToken}
	newToken, err := authConfig.oauth2.TokenSource(ctx, token).Token()
	if err != nil {
		logger.Debug("Token refresh failed", "error", err)
		// Remove invalid refresh token
		s.refreshMutex.Lock()
		delete(s.refreshTokens, userID)
		s.refreshMutex.Unlock()
		return "", false
	}

	// Extract new ID token
	newIDToken, ok := newToken.Extra("id_token").(string)
	if !ok || newIDToken == "" {
		logger.Debug("No id_token in refreshed token")
		return "", false
	}

	// Update stored refresh token if it changed
	if newToken.RefreshToken != "" && newToken.RefreshToken != refreshToken {
		s.refreshMutex.Lock()
		s.refreshTokens[userID] = newToken.RefreshToken
		s.refreshMutex.Unlock()
		logger.Debug("Updated refresh token for user", "userID", userID)
	}

	logger.Debug("Successfully refreshed token for user", "userID", userID)
	return newIDToken, true
}
