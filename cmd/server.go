package cmd

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
)

var (
	bucketName = "habits"
	db         *bbolt.DB
)

var (
	Version   = "dev"
	BuildDate = "unknown"
)

type Habit struct {
	Name      string    `json:"Name"`
	Note      string    `json:"Note"`
	TimeStamp time.Time `json:"TimeStamp"`
}

type HabitListResponse struct {
	Habits []string `json:"Habits"`
}

type VersionInfo struct {
	Version   string `json:"Version"`
	BuildDate string `json:"BuildDate"`
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func initDB() {
	log.Println("Opening DB...")
	var err error
	db, err = bbolt.Open("habits.db", 0600, nil)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	log.Println("Creating bucket if needed...")
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		log.Fatalf("failed to create db bucket: %v", err)
	}
}

func InjectDB(d *bbolt.DB) {
	db = d
}

func startServer() {
	log.Println("Starting server...")
	initDB()
	defer db.Close()

	r := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.Logger)

	r.Route("/habits", func(r chi.Router) {
		r.Post("/", TrackHabit)
		r.Get("/", ListHabits)
		r.Get("/{habit_id}", GetHabit) // Add this line
	})

	r.Route("/version", func(r chi.Router) {
		r.Get("/", GetVersionInfo)
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Printf("server error: %v", err)
	}
}

func GetVersionInfo(w http.ResponseWriter, r *http.Request) {
	info := VersionInfo{
		Version:   Version,
		BuildDate: BuildDate,
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

func ListHabits(w http.ResponseWriter, r *http.Request) {
	h := HabitListResponse{}
	uniqueHabits := make(map[string]struct{})

	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

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

func TrackHabit(w http.ResponseWriter, r *http.Request) {
	h := &Habit{}
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

	err = db.Update(func(tx *bbolt.Tx) error {
		key := h.Name + "/" + time.Now().Format(time.RFC3339)
		b := tx.Bucket([]byte(bucketName))
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

func GetHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "habit_id")
	if habitID == "" {
		http.Error(w, `{"error":"habit id is required"}`, http.StatusBadRequest)
		return
	}

	var entries []Habit
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()

		prefix := []byte(habitID + "/")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var habit Habit
			if err := json.Unmarshal(v, &habit); err != nil {
				return err
			}
			entries = append(entries, habit)
		}
		return nil
	})

	if err != nil {
		http.Error(w, `{"error":"failed to read habits"}`, http.StatusInternalServerError)
		return
	}

	if len(entries) == 0 {
		http.Error(w, `{"error":"habit not found"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"habit_id": habitID,
		"entries":  entries,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}
