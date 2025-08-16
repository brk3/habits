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

func testGetHabitSummary_ActiveStreakSince(t *testing.T, streakLength, since, expectedCurrent,
	expectedLongest int) {
	h := newTestServer(newMemStore())

	for i := 0 + since; i < streakLength+since; i++ {
		rr := mockRequest(h, http.MethodPost, "/habits/",
			habit.Habit{
				Name:      "guitar",
				Note:      "practice",
				TimeStamp: time.Now().AddDate(0, 0, -i).Unix(),
			})
		if rr.Code != http.StatusCreated {
			t.Fatalf("got %d want 201", rr.Code)
		}
	}

	rr := mockRequest(h, http.MethodGet, "/habits/guitar/summary", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rr.Code)
	}
	log.Printf("response body: %s", rr.Body.String())
	var resp HabitSummaryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	log.Printf("response: %+v", resp)

	if resp.HabitSummary.CurrentStreak != expectedCurrent {
		t.Fatalf("got current streak %d, want %d", resp.HabitSummary.CurrentStreak, expectedCurrent)
	}
	if resp.HabitSummary.LongestStreak != expectedLongest {
		t.Fatalf("got longest streak %d, want %d", resp.HabitSummary.LongestStreak, expectedLongest)
	}
}

func TestGetHabitSummary_Streak_ActiveSinceToday(t *testing.T) {
	const today = 0
	testGetHabitSummary_ActiveStreakSince(t, 5, today, 5, 5)
}

func TestGetHabitSummary_Streak_ActiveSinceYesterday(t *testing.T) {
	const yesterday = 1
	testGetHabitSummary_ActiveStreakSince(t, 5, yesterday, 5, 5)
}

func TestGetHabitSummary_Streak_NotActive(t *testing.T) {
	const twoDaysAgo = 2
	testGetHabitSummary_ActiveStreakSince(t, 5, twoDaysAgo, 0, 5)
}

func TestGetHabitSummary_MultipleStreaks(t *testing.T) {
	h := newTestServer(newMemStore())

	// streak of 5 starting today
	for i := 0; i < 5; i++ {
		rr := mockRequest(h, http.MethodPost, "/habits/",
			habit.Habit{
				Name:      "guitar",
				Note:      "practice",
				TimeStamp: time.Now().AddDate(0, 0, -i).Unix(),
			})
		if rr.Code != http.StatusCreated {
			t.Fatalf("got %d want 201", rr.Code)
		}
	}

	// streak of 10 starting 10 days ago
	for i := 10; i < 20; i++ {
		rr := mockRequest(h, http.MethodPost, "/habits/",
			habit.Habit{
				Name:      "guitar",
				Note:      "practice",
				TimeStamp: time.Now().AddDate(0, 0, -i).Unix(),
			})
		if rr.Code != http.StatusCreated {
			t.Fatalf("got %d want 201", rr.Code)
		}
	}

	rr := mockRequest(h, http.MethodGet, "/habits/guitar/summary", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rr.Code)
	}
	log.Printf("response body: %s", rr.Body.String())
	var resp HabitSummaryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	log.Printf("response: %+v", resp)

	if resp.HabitSummary.CurrentStreak != 5 {
		t.Fatalf("got current streak %d, want 5", resp.HabitSummary.CurrentStreak)
	}
	if resp.HabitSummary.LongestStreak != 10 {
		t.Fatalf("got longest streak %d, want 10", resp.HabitSummary.LongestStreak)
	}
}

func TestTrackHabit_WithInvalidTimeStamp(t *testing.T) {
	st := newMemStore()
	h := newTestServer(st)

	rr := mockRequest(h, http.MethodPost, "/habits/",
		habit.Habit{
			Name:      "guitar",
			Note:      "scales",
			TimeStamp: -1,
		})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got %d want 400 Bad Request", rr.Code)
	}
}

func TestGetHabitSummary_FirstLogged(t *testing.T) {
	h := newTestServer(newMemStore())

	for i := 0; i <= 5; i++ {
		rr := mockRequest(h, http.MethodPost, "/habits/",
			habit.Habit{
				Name:      "guitar",
				Note:      "practice",
				TimeStamp: time.Now().AddDate(0, 0, -i).Unix(),
			})
		if rr.Code != http.StatusCreated {
			t.Fatalf("got %d want 201", rr.Code)
		}
	}

	rr := mockRequest(h, http.MethodGet, "/habits/guitar/summary", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rr.Code)
	}
	log.Printf("response body: %s", rr.Body.String())
	var resp HabitSummaryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	log.Printf("response: %+v", resp)

	expectedFirstLogged := time.Now().AddDate(0, 0, -5).Unix()
	if resp.HabitSummary.FirstLogged != expectedFirstLogged {
		t.Fatalf("got first logged %d, want %d", resp.HabitSummary.FirstLogged, expectedFirstLogged)
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
