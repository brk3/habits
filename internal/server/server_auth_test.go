package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/storage"
)

func TestLogin_RedirectsToIDP(t *testing.T) {
	// Setup test server with mock OIDC provider. This also tests the provider validation logic.
	h := newTestServerWithAuth(t, newMemStore())

	rr := mockRequest(h, http.MethodGet, "/auth/login/test", nil)
	if rr.Code != http.StatusFound {
		t.Fatalf("got %d want 304", rr.Code)
	}
	loc, err := rr.Result().Location()
	if err != nil {
		t.Fatalf("error getting location: %v", err)
	}
	if loc.Path != "/auth" {
		t.Fatalf("got redirect to %s, want /auth on test host", loc.String())
	}
}

func TestAuthEnabled_NotLoggedIn_Forbidden(t *testing.T) {
	h := newTestServerWithAuth(t, newMemStore())

	req := httptest.NewRequest(http.MethodGet, "/habits/", nil)
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", rr.Code)
	}
}

func TestAuthEnabled_NotLoggedIn_Redirect(t *testing.T) {
	h := newTestServerWithAuth(t, newMemStore())

	req := httptest.NewRequest(http.MethodGet, "/habits/", nil)
	req.Header.Set("Accept", "text/html")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("got %d want 304", rr.Code)
	}
}

func TestGetUserID_WithValidUser(t *testing.T) {
	// Test with auth enabled and valid user in context
	claims := map[string]any{
		"iss": "https://test-issuer.com",
		"sub": "test-subject",
	}
	user := &User{
		Subject: "test-subject",
		Email:   "test@example.com",
		UserID:  userIDFromClaims(claims),
		Claims:  claims,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), userCtxKey{}, user)
	req = req.WithContext(ctx)

	userID := userIDFromContext(true, req)
	if userID == "" {
		t.Fatal("userIDFromContext returned empty string for valid user")
	}
	if !strings.HasPrefix(userID, "user-") {
		t.Fatalf("userIDFromContext returned %q, expected to start with 'user-'", userID)
	}
}

func TestGetUserID_AuthDisabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	userID := userIDFromContext(false, req)
	if userID != "anonymous" {
		t.Fatalf("userIDFromContext returned %q, expected 'anonymous'", userID)
	}
}

func TestGetUserID_NoUserInContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	userID := userIDFromContext(true, req)
	if userID != "" {
		t.Fatalf("userIDFromContext returned %q, expected empty string when no user in context", userID)
	}
}

func newTestServerWithAuth(t *testing.T, st storage.Store) http.Handler {
	mockOIDC := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			baseURL := "http://" + r.Host
			w.Write([]byte(`{
				"issuer": "` + baseURL + `",
				"authorization_endpoint": "` + baseURL + `/auth",
				"token_endpoint": "` + baseURL + `/token",
				"jwks_uri": "` + baseURL + `/keys"
			}`))
		}
	}))
	t.Cleanup(mockOIDC.Close)

	cfg := config.Config{
		AuthEnabled: true,
		OIDCProviders: []config.OIDCProviderConfig{{
			Id:        "test",
			IssuerURL: mockOIDC.URL,
			ClientID:  "test",
		}},
	}
	s, err := New(&cfg, st)
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}
	return s.Router()
}
