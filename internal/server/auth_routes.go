package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/brk3/habits/internal/logger"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Generate PKCE challenge
	verifier := make([]byte, 48)
	if _, err := rand.Read(verifier); err != nil {
		http.Error(w, "pkce gen failed", http.StatusInternalServerError)
		return
	}
	verifierStr := base64.RawURLEncoding.EncodeToString(verifier)
	hash := sha256.Sum256([]byte(verifierStr))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// Generate state
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		http.Error(w, "state gen failed", http.StatusInternalServerError)
		return
	}
	st := hex.EncodeToString(stateBytes)

	// Capture return path (sanitize to keep it relative)
	ret := r.URL.Query().Get("return")
	if ret == "" {
		ret = "/"
	} else if u, err := url.Parse(ret); err != nil || u.IsAbs() || u.Host != "" {
		ret = "/"
	}

	s.authConf[id].state.Put(st, authState{
		Verifier: verifierStr,
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
	st := r.URL.Query().Get("state")
	if st == "" {
		http.Error(w, "missing state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
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
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in response", http.StatusBadGateway)
		return
	}
	if rawIDToken == "" {
		http.Error(w, "no id_token", http.StatusBadGateway)
		return
	}
	idToken, err := s.authConf[id].idVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "id_token invalid", http.StatusUnauthorized)
		return
	}

	// Store complete token for future refresh
	logger.Debug("Processing token storage", "hasRefreshToken", tok.RefreshToken != "", "expiry", tok.Expiry)
	if tok.RefreshToken != "" {
		var claims map[string]any
		if err := idToken.Claims(&claims); err != nil {
			logger.Error("Failed to extract claims from ID token", "error", err)
			http.Error(w, "token claims invalid", http.StatusUnauthorized)
			return
		}

		userID := userIDFromClaims(claims)
		if userID != "" {
			s.tokenStore.Put(userID, tok)
			logger.Debug("Stored oauth2 token for user", "userID", userID, "hasRefresh", tok.RefreshToken != "", "expiry", tok.Expiry)
		} else {
			logger.Debug("Failed to calculate userID from claims")
		}
	} else {
		logger.Debug("No refresh token in oauth2 token - refresh will not be possible")
	}

	// Set session cookie
	prefixedToken := id + ":" + rawIDToken
	val, err := s.sessionCookie.Encode("session", prefixedToken)
	if err != nil {
		logger.Error("Failed to encode session cookie", "error", err)
		http.Error(w, "session encoding failed", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((3 * 24 * time.Hour).Seconds()),
	})

	http.Redirect(w, r, saved.Return, http.StatusFound)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) simpleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<h1>Login</h1><style>button{display:block;margin:10px 0;padding:10px 20px;}</style>`)
	for id := range s.authConf {
		fmt.Fprintf(w, `<form action="/auth/login/%s"><button>%s</button></form>`, id, s.authConf[id].name)
	}
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

