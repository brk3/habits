package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.etcd.io/bbolt"

	"brk3.github.io/habits/pkg/habit"
)

type Server struct {
	DB         *bbolt.DB
	BucketName string
	Version    string
	BuildDate  string
}

type HabitListResponse struct {
	Habits []string `json:"Habits"`
}

func (s *Server) InitDB() error {
	var err error
	s.DB, err = bbolt.Open("habits.db", 0600, nil)
	if err != nil {
		return err
	}
	return s.DB.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(s.BucketName))
		return err
	})
}

func (s *Server) CloseDB() {
	if s.DB != nil {
		s.DB.Close()
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Route("/habits", func(r chi.Router) {
		r.Post("/", s.TrackHabit)
		r.Get("/", s.ListHabits)
	})
	r.Route("/version", func(r chi.Router) {
		r.Get("/", s.GetVersionInfo)
	})
	return r
}

func (s *Server) GetVersionInfo(w http.ResponseWriter, r *http.Request) {
	info := habit.VersionInfo{
		Version:   s.Version,
		BuildDate: s.BuildDate,
	}
	infoJSON, err := json.Marshal(info)
	if err != nil {
		http.Error(w, `{"error":"failed to serialize version info"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(infoJSON)
}

func (s *Server) ListHabits(w http.ResponseWriter, r *http.Request) {
	h := HabitListResponse{}
	uniqueHabits := make(map[string]struct{})

	s.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(s.BucketName))
		b.ForEach(func(k, v []byte) error {
			sanitised := strings.Split(string(k), "/")[0]
			uniqueHabits[sanitised] = struct{}{}
			return nil
		})
		return nil
	})

	for habit := range uniqueHabits {
		h.Habits = append(h.Habits, habit)
	}

	habitJSON, err := json.Marshal(h)
	if err != nil {
		http.Error(w, `{"error":"failed to serialize habit list"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(habitJSON)
}

func (s *Server) TrackHabit(w http.ResponseWriter, r *http.Request) {
	h := &habit.Habit{}
	if err := json.NewDecoder(r.Body).Decode(h); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if h.Name == "" {
		http.Error(w, `{"error":"habit name is required"}`, http.StatusBadRequest)
		return
	}

	habitJSON, err := json.Marshal(h)
	if err != nil {
		http.Error(w, `{"error":"failed to serialize habit"}`, http.StatusInternalServerError)
		return
	}

	err = s.DB.Update(func(tx *bbolt.Tx) error {
		key := h.Name + "/" + time.Now().Format(time.RFC3339)
		b := tx.Bucket([]byte(s.BucketName))
		return b.Put([]byte(key), habitJSON)
	})
	if err != nil {
		log.Printf("Error saving habit: %v", err)
		http.Error(w, `{"error":"database write failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(habitJSON)
}
