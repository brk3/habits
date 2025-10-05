package server

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"
)

// hashAPIKey creates a SHA256 hash of an API key for storage
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", hash)
}

// truncateHash returns a truncated hash for display/logging
// Returns first 16 chars + "..." or the full hash if shorter
func truncateHash(hash string) string {
	if len(hash) <= 16 {
		return hash
	}
	return hash[:16] + "..."
}

// createSessionCookie creates a session cookie with standard security settings
func createSessionCookie(name, value string, maxAge time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
	}
}
