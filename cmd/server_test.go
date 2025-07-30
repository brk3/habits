package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestThoughtEndpoint(t *testing.T) {
	router := chi.NewRouter()

	router.Post("/thought", func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Content string `json:"content"`
		}

		var body reqBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		h := &Thought{
			Content:   body.Content,
			TimeStamp: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h)
	})

	// Prepare request
	body := []byte(`{"content":"test from server"}`)
	req := httptest.NewRequest("POST", "/thought", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result Thought
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Content != "test from server" {
		t.Errorf("expected content 'test from server', got '%s'", result.Content)
	}
}
