package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brk3/habits/internal/storage"
	"github.com/brk3/habits/pkg/habit"
	"github.com/brk3/habits/pkg/versioninfo"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Store storage.Store
}

func New(store storage.Store) *Server {
	return &Server{Store: store}
}

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/version", s.getVersionInfo)
	r.Route("/habits", func(r chi.Router) {
		r.Post("/", s.trackHabit)
		r.Get("/", s.listHabits)
		r.Get("/{habit_id}", s.getHabit)
		r.Get("/{habit_id}/summary", s.getHabitSummary)
		//r.Get("/{habit_id}/heatmap", s.getHabitHeatmap)

	})
	return r
}

func (s *Server) getHabitSummary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "habit_id")
	if id == "" {
		http.Error(w, `{"error":"habit id is required"}`, http.StatusBadRequest)
		return
	}

	currentStreak, longestSreak, err := s.computeStreaks(id)
	if err != nil {
		http.Error(w, `{"error":"error computing streaks"}`, http.StatusInternalServerError)
		return
	}

	summary := habit.HabitSummary{
		Name: id,
		// TODO: fill in fields
		CurrentStreak: currentStreak,
		LongestStreak: longestSreak,
		FirstLogged:   0,
		TotalDaysDone: 0,
		BestMonth:     0,
		ThisMonth:     0,
		LastWrite:     time.Now().Unix(),
	}

	summaryResponse := HabitSummaryResponse{
		HabitID:      id,
		HabitSummary: summary,
	}
	if err := writeJSON(w, http.StatusOK, summaryResponse); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) getVersionInfo(w http.ResponseWriter, _ *http.Request) {
	info := versioninfo.VersionInfo{
		Version:   versioninfo.Version,
		BuildDate: versioninfo.BuildDate,
	}
	if err := writeJSON(w, http.StatusOK, info); err != nil {
		http.Error(w, `{"error":"failed to serialize version info"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) listHabits(w http.ResponseWriter, _ *http.Request) {
	names, err := s.Store.ListHabitNames()
	if err != nil {
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}
	if err := writeJSON(w, http.StatusOK, HabitListResponse{Habits: names}); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) trackHabit(w http.ResponseWriter, r *http.Request) {
	var h habit.Habit
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if h.Name == "" {
		http.Error(w, `{"error":"habit name is required"}`, http.StatusBadRequest)
		return
	}
	if h.TimeStamp == 0 {
		h.TimeStamp = time.Now().Unix()
	}
	if err := s.Store.PutHabit(h); err != nil {
		http.Error(w, `{"error":"database write failed"}`, http.StatusInternalServerError)
		return
	}
	if err := writeJSON(w, http.StatusCreated, h); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHabit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "habit_id")
	if id == "" {
		http.Error(w, `{"error":"habit id is required"}`, http.StatusBadRequest)
		return
	}

	entries, err := s.Store.GetHabit(id)
	if err != nil {
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}
	if len(entries) == 0 {
		http.Error(w, `{"error":"habit not found"}`, http.StatusNotFound)
		return
	}

	h := HabitGetResponse{
		HabitID: id,
		Entries: entries,
	}
	if err := writeJSON(w, http.StatusOK, h); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}
