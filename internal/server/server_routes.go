package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brk3/habits/internal/logger"
	"github.com/brk3/habits/pkg/habit"
	"github.com/brk3/habits/pkg/versioninfo"
	"github.com/go-chi/chi/v5"
)

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func (s *Server) getHabitSummary(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "habit_id")
	userID := userIDFromContext(s.cfg.AuthEnabled, r)
	logger.Debug("Getting habit summary", "habit_id", habitID, "user_id", userID)
	if userID == "" || habitID == "" {
		logger.Warn("Missing required parameters", "user_id", userID, "habit_id", habitID)
		http.Error(w, `{"error":"user id and habit id are required"}`, http.StatusBadRequest)
		return
	}

	currentStreak, longestSreak, err := s.computeStreaks(userID, habitID)
	if err != nil {
		logger.Error("Failed to compute streaks", "user_id", userID, "habit_id", habitID, "error", err)
		http.Error(w, `{"error":"error computing streaks"}`, http.StatusInternalServerError)
		return
	}

	firstLogged, err := s.getFirstLogged(userID, habitID)
	if err != nil {
		logger.Error("Failed to get first logged date", "user_id", userID, "habit_id", habitID, "error", err)
		http.Error(w, `{"error":"error retrieving first logged date"}`, http.StatusInternalServerError)
		return
	}

	totalDaysDone, err := s.computeTotalDaysDone(userID, habitID)
	if err != nil {
		logger.Error("Failed to compute total days done", "user_id", userID, "habit_id", habitID, "error", err)
		http.Error(w, `{"error":"error computing total days done"}`, http.StatusInternalServerError)
		return
	}

	daysThisMonth, err := s.computeDaysThisMonth(userID, habitID)
	if err != nil {
		logger.Error("Failed to compute days this month", "user_id", userID, "habit_id", habitID, "error", err)
		http.Error(w, `{"error":"error computing days this month"}`, http.StatusInternalServerError)
		return
	}

	bestMonth, err := s.computeBestMonth(userID, habitID)
	if err != nil {
		logger.Error("Failed to compute best month", "user_id", userID, "habit_id", habitID, "error", err)
		http.Error(w, `{"error":"error computing best month"}`, http.StatusInternalServerError)
		return
	}

	summary := habit.HabitSummary{
		Name:          habitID,
		CurrentStreak: currentStreak,
		LongestStreak: longestSreak,
		FirstLogged:   firstLogged,
		TotalDaysDone: totalDaysDone,
		BestMonth:     bestMonth,
		ThisMonth:     daysThisMonth,
		LastWrite:     time.Now().Unix(),
	}

	summaryResponse := HabitSummaryResponse{
		HabitID:      habitID,
		HabitSummary: summary,
	}
	if err := writeJSON(w, http.StatusOK, summaryResponse); err != nil {
		logger.Error("Failed to serialize habit summary response", "user_id", userID, "habit_id", habitID, "error", err)
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
		logger.Error("Failed to serialize version info response", "error", err)
		http.Error(w, `{"error":"failed to serialize version info"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) listHabits(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(s.cfg.AuthEnabled, r)
	logger.Debug("Listing habits", "user_id", userID)
	if userID == "" {
		logger.Warn("Missing user ID for list habits")
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}
	names, err := s.store.ListHabitNames(userID)
	if err != nil {
		logger.Error("Failed to list habits", "user_id", userID, "error", err)
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}
	logger.Debug("Listed habits successfully", "user_id", userID, "count", len(names))
	if err := writeJSON(w, http.StatusOK, HabitListResponse{Habits: names}); err != nil {
		logger.Error("Failed to serialize habit list response", "user_id", userID, "error", err)
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) trackHabit(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(s.cfg.AuthEnabled, r)
	logger.Debug("Tracking habit", "user_id", userID)
	if userID == "" {
		logger.Warn("Missing user ID for track habit")
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}
	var h habit.Habit
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		logger.Warn("Invalid JSON in track habit request", "error", err)
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if err := validateHabit(h); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	logger.Info("Storing habit", "user_id", userID, "habit_name", h.Name, "timestamp", h.TimeStamp)
	if err := s.store.PutHabit(userID, h); err != nil {
		logger.Error("Failed to store habit", "user_id", userID, "habit_name", h.Name, "error", err)
		http.Error(w, `{"error":"database write failed"}`, http.StatusInternalServerError)
		return
	}
	logger.Info("Habit tracked successfully", "user_id", userID, "habit_name", h.Name)

	habits, err := s.store.ListHabitNames(userID)
	if err != nil {
		logger.Warn("Failed to update active habits metric after tracking", "user_id", userID, "error", err)
	} else {
		activeHabits.Set(float64(len(habits)))
	}

	if err := writeJSON(w, http.StatusCreated, h); err != nil {
		logger.Error("Failed to serialize track habit response", "user_id", userID, "habit_name", h.Name, "error", err)
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "habit_id")
	userID := userIDFromContext(s.cfg.AuthEnabled, r)
	if userID == "" || habitID == "" {
		http.Error(w, `{"error":"user id and habit id are required"}`, http.StatusBadRequest)
		return
	}

	entries, err := s.store.GetHabit(userID, habitID)
	if err != nil {
		logger.Error("Failed to get habit entries", "user_id", userID, "habit_id", habitID, "error", err)
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}
	if len(entries) == 0 {
		http.Error(w, `{"error":"habit not found"}`, http.StatusNotFound)
		return
	}

	h := HabitGetResponse{
		HabitID: habitID,
		Entries: entries,
	}
	if err := writeJSON(w, http.StatusOK, h); err != nil {
		logger.Error("Failed to serialize get habit response", "user_id", userID, "habit_id", habitID, "error", err)
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "habit_id")
	userID := userIDFromContext(s.cfg.AuthEnabled, r)
	logger.Info("Deleting habit", "user_id", userID, "habit_id", habitID)
	if userID == "" || habitID == "" {
		logger.Warn("Missing required parameters for delete", "user_id", userID, "habit_id", habitID)
		http.Error(w, `{"error":"user id and habit id are required"}`, http.StatusBadRequest)
		return
	}

	err := s.store.DeleteHabit(userID, habitID)
	if err != nil {
		logger.Error("Failed to delete habit", "user_id", userID, "habit_id", habitID, "error", err)
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}
	logger.Info("Habit deleted successfully", "user_id", userID, "habit_id", habitID)

	habits, err := s.store.ListHabitNames(userID)
	if err != nil {
		logger.Warn("Failed to update active habits metric after deletion", "user_id", userID, "error", err)
	} else {
		activeHabits.Set(float64(len(habits)))
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateHabit(h habit.Habit) error {
	const maxNameLength = 20
	const maxNoteLength = 1024
	const minTS = 946684800
	const maxTS = 4102444800

	if len(h.Name) == 0 || len(h.Name) > maxNameLength {
		return fmt.Errorf("bad habit name: must be 1-%d characters", maxNameLength)
	}
	if len(h.Note) > maxNoteLength {
		return fmt.Errorf("bad habit note: must be 0-%d characters", maxNoteLength)
	}
	if h.TimeStamp < minTS || h.TimeStamp > maxTS {
		return fmt.Errorf("invalid timestamp")
	}

	return nil
}
