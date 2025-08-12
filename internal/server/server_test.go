package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brk3/habits/internal/storage"
)

/*
func TestTrackHabit_Valid(t *testing.T) {
	st := newMemStore()
	h := newTestServer(st)

	rr := mockRequest(h, http.MethodPost, "/habits/",
		habit.Habit{
			Name:      "guitar",
			Note:      "scales",
			TimeStamp: time.Now().Unix(),
		})
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d want 201", rr.Code)
	}

	// TODO(pbourke): fix here... we're unmarshalling into a habit but its a list of entries
	// looks like we need a HabitListResponse type - can also use this to extend cli cmd
	rr = mockRequest(h, http.MethodGet, "/habits/guitar", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200 OK", rr.Code)
	}
	log.Printf("response body: %s", rr.Body.String())
	var resp habit.Habit
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	log.Printf("response: %+v", resp)
	if resp.Name != "guitar" {
		t.Fatalf("got '%s' want guitar", resp.Name)
	}
	if resp.Note != "scales" {
		t.Fatalf("got '%s' want scales", resp.Note)
	}
	if resp.TimeStamp == 0 {
		t.Fatal("got 0 timestamp, want non-zero")
	}
}
*/

func TestListHabits_Empty(t *testing.T) {
	h := newTestServer(newMemStore())
	rr := mockRequest(h, http.MethodGet, "/habits/", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rr.Code)
	}
	var resp struct{ Habits []string }
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp.Habits) != 0 {
		t.Fatalf("len=%d want 0", len(resp.Habits))
	}
}

func newTestServer(st storage.Store) http.Handler {
	s := New(st)
	return s.Router()
}

func mockRequest(h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	return rr
}

var _ storage.Store = (*memStore)(nil)
