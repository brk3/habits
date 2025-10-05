package server

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brk3/habits/internal/config"
)

func TestAPIKeyGeneration(t *testing.T) {
	store := newMemStore()

	cfg := &config.Config{AuthEnabled: true}
	srv, err := New(cfg, store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	claims := map[string]any{
		"iss": "https://test.com",
		"sub": "test-user",
	}
	userID := userIDFromClaims(claims)

	req := httptest.NewRequest(http.MethodPost, "/auth/api_keys", nil)
	req = withAuthenticatedUser(req, userID, "test@example.com")

	rr := httptest.NewRecorder()
	srv.generateAPIKey(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200, body: %s", rr.Code, rr.Body.String())
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	apiKey, ok := response["api_key"]
	if !ok || apiKey == "" {
		t.Fatal("response missing api_key field")
	}

	if !strings.HasPrefix(apiKey, "hab_live_") {
		t.Fatalf("API key has wrong prefix: %s", apiKey)
	}

	// Verify it was stored
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := fmt.Sprintf("%x", hash)
	storedUserID, found, err := store.GetAPIKey(keyHash)
	if err != nil {
		t.Fatalf("failed to get API key from store: %v", err)
	}
	if !found {
		t.Fatal("API key not found in store")
	}
	if storedUserID != userID {
		t.Fatalf("stored userID %s doesn't match expected %s", storedUserID, userID)
	}
}

// TestAPIKeyAuthentication tests using an API key for authentication
func TestAPIKeyAuthentication(t *testing.T) {
	store := newMemStore()
	h := newTestServerWithAuth(t, store)

	// Generate and store an API key
	apiKey := "hab_live_test123456789012345678901234"
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := fmt.Sprintf("%x", hash)

	claims := map[string]any{
		"iss": "https://test.com",
		"sub": "test-user",
	}
	userID := userIDFromClaims(claims)

	if err := store.PutAPIKey(keyHash, userID); err != nil {
		t.Fatalf("failed to store API key: %v", err)
	}

	// Make a request with the API key
	req := httptest.NewRequest(http.MethodGet, "/habits/", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// Should succeed (200 for empty list)
	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200, body: %s", rr.Code, rr.Body.String())
	}
}

// TestAPIKeyAuthentication_InvalidKey tests authentication with invalid API key
func TestAPIKeyAuthentication_InvalidKey(t *testing.T) {
	store := newMemStore()
	h := newTestServerWithAuth(t, store)

	// Make a request with an invalid API key
	req := httptest.NewRequest(http.MethodGet, "/habits/", nil)
	req.Header.Set("Authorization", "Bearer hab_live_invalid_key_not_in_db")
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", rr.Code)
	}
}

// TestAPIKeyAuthentication_WrongPrefix tests that non-hab_ tokens are handled correctly
func TestAPIKeyAuthentication_WrongPrefix(t *testing.T) {
	store := newMemStore()
	h := newTestServerWithAuth(t, store)

	// Make a request with a Bearer token that's not an API key
	req := httptest.NewRequest(http.MethodGet, "/habits/", nil)
	req.Header.Set("Authorization", "Bearer some_random_token")
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// Should fail (not a valid OIDC token either)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", rr.Code)
	}
}

// TestListAPIKeys tests listing a user's API keys
func TestListAPIKeys(t *testing.T) {
	store := newMemStore()

	cfg := &config.Config{AuthEnabled: true}
	srv, err := New(cfg, store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	claims := map[string]any{
		"iss": "https://test.com",
		"sub": "test-user",
	}
	userID := userIDFromClaims(claims)

	// Store some API keys for this user
	key1Hash := fmt.Sprintf("%x", sha256.Sum256([]byte("key1")))
	key2Hash := fmt.Sprintf("%x", sha256.Sum256([]byte("key2")))
	store.PutAPIKey(key1Hash, userID)
	store.PutAPIKey(key2Hash, userID)

	// Store a key for a different user
	otherUserID := "user-other"
	key3Hash := fmt.Sprintf("%x", sha256.Sum256([]byte("key3")))
	store.PutAPIKey(key3Hash, otherUserID)

	// Make authenticated request
	req := httptest.NewRequest(http.MethodGet, "/auth/api_keys", nil)
	req = withAuthenticatedUser(req, userID, "test@example.com")

	rr := httptest.NewRecorder()
	srv.listAPIKeys(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200, body: %s", rr.Code, rr.Body.String())
	}

	var response map[string][]map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	keys, ok := response["keys"]
	if !ok {
		t.Fatal("response missing keys field")
	}

	// Should have 2 keys (not the other user's key)
	if len(keys) != 2 {
		t.Fatalf("got %d keys, want 2", len(keys))
	}
}

// TestAuthenticateAPIKey_ValidKey tests the authenticateAPIKey function directly
func TestAuthenticateAPIKey_ValidKey(t *testing.T) {
	store := newMemStore()

	cfg := &config.Config{AuthEnabled: true}
	srv, err := New(cfg, store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Store an API key
	apiKey := "hab_live_testkey123456789"
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := fmt.Sprintf("%x", hash)
	userID := "user-test123"

	if err := store.PutAPIKey(keyHash, userID); err != nil {
		t.Fatalf("failed to store API key: %v", err)
	}

	// Authenticate
	user, authenticated := srv.authenticateAPIKey(apiKey)
	if !authenticated {
		t.Fatal("authentication should have succeeded")
	}

	if user.UserID != userID {
		t.Fatalf("got userID %s, want %s", user.UserID, userID)
	}

	if user.Email != "" {
		t.Fatal("API key auth should not have email")
	}

	if !strings.HasPrefix(user.Subject, "apikey:") {
		t.Fatalf("Subject should start with 'apikey:', got: %s", user.Subject)
	}
}

// TestAuthenticateAPIKey_InvalidKey tests authentication with non-existent key
func TestAuthenticateAPIKey_InvalidKey(t *testing.T) {
	store := newMemStore()

	cfg := &config.Config{AuthEnabled: true}
	srv, err := New(cfg, store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Try to authenticate with a key that doesn't exist
	apiKey := "hab_live_doesnotexist"
	_, authenticated := srv.authenticateAPIKey(apiKey)
	if authenticated {
		t.Fatal("authentication should have failed for non-existent key")
	}
}

// TestAPIKeyGeneration_Unauthenticated tests that unauthenticated users can't generate keys
func TestAPIKeyGeneration_Unauthenticated(t *testing.T) {
	store := newMemStore()

	cfg := &config.Config{AuthEnabled: true}
	srv, err := New(cfg, store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/api_keys", nil)
	// No user in context

	rr := httptest.NewRecorder()
	srv.generateAPIKey(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", rr.Code)
	}
}

// Helper function to add authenticated user to request context
func withAuthenticatedUser(req *http.Request, userID, email string) *http.Request {
	user := &User{
		UserID:  userID,
		Email:   email,
		Subject: "test-subject",
		Claims:  map[string]any{},
	}
	ctx := req.Context()
	ctx = context.WithValue(ctx, userCtxKey{}, user)
	return req.WithContext(ctx)
}
