package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"brk3.github.io/habits/cmd"
	"go.etcd.io/bbolt"
)

func setupTestDB(t *testing.T) *bbolt.DB {
	db, err := bbolt.Open("test.db", 0600, nil)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("habits"))
		return err
	})
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}

	cmd.InjectDB(db) // assumes you add a helper to inject test DB
	return db
}

func teardownTestDB(db *bbolt.DB) {
	db.Close()
	_ = os.Remove("test.db")
}

func TestGetVersionInfo(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	cmd.GetVersionInfo(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var data cmd.VersionInfo
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if data.Version == "" {
		t.Error("version should not be empty")
	}
}

func TestTrackHabit(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	habit := cmd.Habit{
		Name:      "guitar",
		Note:      "practice",
		TimeStamp: time.Now(),
	}
	body, _ := json.Marshal(habit)

	req := httptest.NewRequest(http.MethodPost, "/habits", bytes.NewReader(body))
	w := httptest.NewRecorder()

	cmd.TrackHabit(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
}

func TestTrackHabitInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/habits", bytes.NewReader([]byte("{invalid")))
	w := httptest.NewRecorder()

	cmd.TrackHabit(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestListHabits(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	// Insert a test habit directly
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("habits"))
		return b.Put([]byte("guitar/"+time.Now().Format(time.RFC3339)), []byte(`{"Name":"guitar"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/habits", nil)
	w := httptest.NewRecorder()

	cmd.ListHabits(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var response cmd.HabitListResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Habits) == 0 {
		t.Error("expected at least one habit in response")
	}
}
