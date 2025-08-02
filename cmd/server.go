package cmd

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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
		startServer(cmd)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func initDB() {
	var err error
	db, err = bbolt.Open("habits.db", 0600, nil)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Fatalf("failed to create db bucket: %v", err)
		}
		return nil
	})
}

func InjectDB(d *bbolt.DB) {
	db = d
}

func startServer(cmd *cobra.Command) {
	initDB()
	defer db.Close()

	r := chi.NewRouter()

	r.Route("/habits", func(r chi.Router) {
		r.Post("/", TrackHabit)
		r.Get("/", ListHabits)
	})

	r.Route("/version", func(r chi.Router) {
		r.Get("/", GetVersionInfo)
	})

	cmd.Println("listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
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
