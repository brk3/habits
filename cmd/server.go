package cmd

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
)

var bucketName = "habits"
var db *bbolt.DB

type Habit struct {
	Name      string    `json:"Name"`
	Note      string    `json:"Note"`
	TimeStamp time.Time `json:"TimeStamp"`
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

func startServer(cmd *cobra.Command) {
	var err error
	db, err = bolt.Open("habits.db", 0600, nil)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	r := chi.NewRouter()

	r.Route("/habits", func(r chi.Router) {
		r.Post("/", TrackHabit)
	})

	cmd.Println("listening on localhost:8080")
	http.ListenAndServe(":8080", r)
}

func TrackHabit(w http.ResponseWriter, r *http.Request) {
	var h Habit
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	habitJson, _ := json.Marshal(h)

	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		key := h.Name + "/" + h.TimeStamp.Format(time.RFC3339)
		return b.Put([]byte(key), habitJson)
	})

	if err != nil {
		log.Printf("Error saving habit: %v", err)
		return
	}

	log.Printf("/track - %s", habitJson)
}
