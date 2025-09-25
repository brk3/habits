package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/logger"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

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
	sessionCookie.MaxAge(86400) // 24 hours

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

func (s *Server) simpleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
<title>OIDC Providers</title>
<style>
button {
    display: block;
    margin: 10px 0;
    padding: 10px 20px;
    font-size: 16px;
}
</style>
</head>
<body>
<h1>Available OIDC Providers</h1>
`)

	for id := range s.authConf {
		fmt.Fprintf(w, `<form action="/auth/login/%s" method="GET">
            <button type="submit">%s</button>
        </form>
`, id, s.authConf[id].name)
	}

	fmt.Fprint(w, `
</body>
</html>`)
}

func (s *Server) getAPIToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "not logged in", http.StatusUnauthorized)
		return
	}

	var prefixedToken string
	if err := s.sessionCookie.Decode("session", cookie.Value, &prefixedToken); err != nil {
		http.Error(w, "invalid session cookie", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(prefixedToken))
}

type userCtxKey struct{}

type User struct {
	Subject string
	Email   string
	Claims  map[string]any
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
			if r.Method == http.MethodGet && acceptsHTML(r.Header.Get("Accept")) {
				http.Redirect(w, r, "/auth/login", http.StatusFound)
			} else {
				w.Header().Set("WWW-Authenticate", `Bearer realm="habits"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			}
			return
		}

		// 4) Verify token with the correct provider
		idTok, err := s.authConf[providerID].idVerifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			// Clear session cookie on invalid token
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1})
			if r.Method == http.MethodGet && acceptsHTML(r.Header.Get("Accept")) {
				http.Redirect(w, r, "/auth/login", http.StatusFound)
			} else {
				w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			}
			return
		}

		// 5) Extract claims and create user
		var claims map[string]any
		_ = idTok.Claims(&claims)
		u := &User{
			Subject: idTok.Subject,
			Email:   strClaim(claims, "email"),
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

// OIDC auth flow
// Need /login, /logout, and /callback endpoints. They don't have to be called exactly that, but
// they need to exist for those purposes.
//
// This should probably be a library dep but I had fun putting it together.

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	st := randState()
	verifier, err := genCodeVerifier(48)
	if err != nil {
		http.Error(w, "pkce gen failed", http.StatusInternalServerError)
		return
	}
	challenge := codeChallengeS256(verifier)

	/// ???
	// I dont know what this is doing, or rather WHY its doing it.
	// capture an optional return path (sanitize to keep it relative)
	ret := r.URL.Query().Get("return")
	if ret == "" {
		ret = "/"
	} else {
		// keep it relative to avoid open redirects
		if u, err := url.Parse(ret); err != nil || u.IsAbs() || u.Host != "" {
			ret = "/"
		}
	}
	/// ???

	s.authConf[id].state.Put(st, authState{
		Verifier: verifier,
		Return:   ret,
		ExpireAt: time.Now().Add(5 * time.Minute),
	})

	authURL := s.authConf[id].oauth2.AuthCodeURL(
		st,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cookieName := "session"
	st := param(r, "state")
	if st == "" {
		http.Error(w, "missing state", http.StatusBadRequest)
		return
	}
	code := param(r, "code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	saved, ok := s.authConf[id].state.GetAndDelete(st)
	if !ok || saved.Verifier == "" {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	tok, err := s.authConf[id].oauth2.Exchange(
		r.Context(),
		code,
		oauth2.SetAuthURLParam("code_verifier", saved.Verifier),
	)
	if err != nil {
		http.Error(w, "code exchange failed", http.StatusBadGateway)
		return
	}
	rawIDToken, _ := tok.Extra("id_token").(string)
	if rawIDToken == "" {
		http.Error(w, "no id_token", http.StatusBadGateway)
		return
	}
	if _, err := s.authConf[id].idVerifier.Verify(r.Context(), rawIDToken); err != nil {
		http.Error(w, "id_token invalid", http.StatusUnauthorized)
		return
	}

	// Create provider-prefixed token for API use
	prefixedToken := id + ":" + rawIDToken

	// Encode prefixed token directly
	val, _ := s.sessionCookie.Encode(cookieName, prefixedToken)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((8 * time.Hour).Seconds()),
	})

	http.Redirect(w, r, saved.Return, http.StatusFound)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	// Clear single session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	logger.Info("User logout completed")
	w.WriteHeader(http.StatusNoContent)
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

func genCodeVerifier(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func codeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func randState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func param(r *http.Request, name string) string {
	if v := r.URL.Query().Get(name); v != "" {
		return v
	}
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		if v := r.PostFormValue(name); v != "" {
			return v
		}
	}
	return ""
}
