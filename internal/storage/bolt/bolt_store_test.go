package bolt

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/brk3/habits/pkg/habit"
)

func newTestStore(t *testing.T) (*Store, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}

	cleanup := func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}

	return store, cleanup
}

func TestOpen(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestListHabitNames_Empty(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	names, err := store.ListHabitNames("testuser")
	if err != nil {
		t.Fatalf("ListHabitNames failed: %v", err)
	}

	if len(names) != 0 {
		t.Fatalf("expected empty list, got %d items", len(names))
	}
}

func TestListHabitNames_WithData(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	habits := []habit.Habit{
		{Name: "guitar", Note: "scales", TimeStamp: time.Now().Unix()},
		{Name: "guitar", Note: "chords", TimeStamp: time.Now().Unix() - 86400}, // yesterday
		{Name: "exercise", Note: "pushups", TimeStamp: time.Now().Unix()},
	}

	for _, h := range habits {
		if err := store.PutHabit("testuser", h); err != nil {
			t.Fatalf("PutHabit failed: %v", err)
		}
	}

	names, err := store.ListHabitNames("testuser")
	if err != nil {
		t.Fatalf("ListHabitNames failed: %v", err)
	}

	expectedNames := []string{"guitar", "exercise"}
	if len(names) != len(expectedNames) {
		t.Fatalf("expected %d names, got %d", len(expectedNames), len(names))
	}

	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, expected := range expectedNames {
		if !nameMap[expected] {
			t.Errorf("expected habit name '%s' not found in results", expected)
		}
	}
}

func TestUserIsolation(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Add habit for alice
	aliceHabit := habit.Habit{Name: "guitar", Note: "scales", TimeStamp: time.Now().Unix()}
	if err := store.PutHabit("alice", aliceHabit); err != nil {
		t.Fatalf("PutHabit failed: %v", err)
	}

	// Alice should see her habit
	aliceNames, err := store.ListHabitNames("alice")
	if err != nil {
		t.Fatalf("ListHabitNames failed: %v", err)
	}
	if len(aliceNames) != 1 || aliceNames[0] != "guitar" {
		t.Fatalf("alice should see 'guitar', got %v", aliceNames)
	}

	// Bob should see nothing
	bobNames, err := store.ListHabitNames("bob")
	if err != nil {
		t.Fatalf("ListHabitNames failed: %v", err)
	}
	if len(bobNames) != 0 {
		t.Fatalf("bob should see no habits, got %v", bobNames)
	}
}

func TestGetAPIKey_NonExistent(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	_, found, err := store.GetAPIKey("nonexistent-key")
	if err != nil {
		t.Fatalf("GetAPIKey failed: %v", err)
	}
	if found {
		t.Fatal("expected key not found, but found=true")
	}
}

func TestGetAPIKey_Valid(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	err := store.PutAPIKey("test-hash", "user-123")
	if err != nil {
		t.Fatalf("PutAPIKey failed: %v", err)
	}

	userID, found, err := store.GetAPIKey("test-hash")
	if err != nil {
		t.Fatalf("GetAPIKey failed: %v", err)
	}
	if !found {
		t.Fatal("expected key to be found")
	}
	if userID != "user-123" {
		t.Fatalf("expected userID 'user-123', got '%s'", userID)
	}
}

func TestListAPIKeyHashes(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	err := store.PutAPIKey("key1", "user1")
	if err != nil {
		t.Fatalf("PutAPIKey failed: %v", err)
	}

	err = store.PutAPIKey("key2", "user1")
	if err != nil {
		t.Fatalf("PutAPIKey failed: %v", err)
	}

	err = store.PutAPIKey("key3", "user2")
	if err != nil {
		t.Fatalf("PutAPIKey failed: %v", err)
	}

	hashes, err := store.ListAPIKeyHashes("user1")
	if err != nil {
		t.Fatalf("ListAPIKeyHashes failed: %v", err)
	}
	if len(hashes) != 2 {
		t.Fatalf("expected 2 hashes for user1, got %d", len(hashes))
	}

	hashes, err = store.ListAPIKeyHashes("user2")
	if err != nil {
		t.Fatalf("ListAPIKeyHashes failed: %v", err)
	}
	if len(hashes) != 1 {
		t.Fatalf("expected 1 hash for user2, got %d", len(hashes))
	}
}

func TestDeleteAPIKey(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	err := store.PutAPIKey("delete-me", "user-123")
	if err != nil {
		t.Fatalf("PutAPIKey failed: %v", err)
	}

	_, found, err := store.GetAPIKey("delete-me")
	if err != nil {
		t.Fatalf("GetAPIKey failed before delete: %v", err)
	}
	if !found {
		t.Fatal("expected key to be found before delete")
	}

	err = store.DeleteAPIKey("delete-me")
	if err != nil {
		t.Fatalf("DeleteAPIKey failed: %v", err)
	}

	_, found, err = store.GetAPIKey("delete-me")
	if err != nil {
		t.Fatalf("GetAPIKey failed after delete: %v", err)
	}
	if found {
		t.Fatal("expected key not to be found after delete")
	}
}
