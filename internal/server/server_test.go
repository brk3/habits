package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brk3/habits/internal/storage"
	"github.com/brk3/habits/pkg/habit"
	"github.com/brk3/habits/pkg/versioninfo"
)

func TestGetVersionInfo(t *testing.T) {
	st := newMemStore()
	h := newTestServer(st)

	rr := mockRequest(h, http.MethodGet, "/version", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200 OK", rr.Code)
	}

	var resp versioninfo.VersionInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	log.Printf("response: %+v", resp)
	if resp.Version != versioninfo.Version {
		t.Fatalf("got version %s, want %s", resp.Version, versioninfo.Version)
	}
	if resp.BuildDate != versioninfo.BuildDate {
		t.Fatalf("got build date %s, want %s", resp.BuildDate, versioninfo.BuildDate)
	}
}

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

	rr = mockRequest(h, http.MethodGet, "/habits/guitar", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200 OK", rr.Code)
	}
	log.Printf("response body: %s", rr.Body.String())
	var resp HabitGetResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	log.Printf("response: %+v", resp)
	if resp.Entries[0].Name != "guitar" {
		t.Fatalf("got '%s' want guitar", resp.Entries[0].Name)
	}
	if resp.Entries[0].Note != "scales" {
		t.Fatalf("got '%s' want scales", resp.Entries[0].Note)
	}
	if resp.Entries[0].TimeStamp == 0 {
		t.Fatal("got 0 timestamp, want non-zero")
	}
}

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

func TestGetHabitSummary(t *testing.T) {
	h := newTestServer(newMemStore())

	rr := mockRequest(h, http.MethodPost, "/habits/",
		habit.Habit{
			Name:      "guitar",
			Note:      "scales",
			TimeStamp: time.Now().Unix(),
		})
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d want 201", rr.Code)
	}

	rr = mockRequest(h, http.MethodGet, "/habits/guitar/summary", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rr.Code)
	}
	log.Printf("response body: %s", rr.Body.String())
	var resp HabitSummaryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	log.Printf("response: %+v", resp)
	if resp.HabitID != "guitar" {
		t.Fatalf("got '%s' want 'guitar'", resp.HabitID)
	}
	//if resp.HabitSummary.CurrentStreak == 0 {
	//	t.Fatal("got 0 current streak, want non-zero")
	//}
	// TODO: add more checks for HabitSummary fields
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
