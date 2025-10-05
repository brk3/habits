package server

import (
	"crypto/sha256"
	"fmt"
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
