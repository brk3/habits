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

	// TODO(pbourke): review
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
		MaxAge:   int((3 * 24 * time.Hour).Seconds()),
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

// helpers
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

func codeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func genCodeVerifier(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func randState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
