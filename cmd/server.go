package cmd

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var bucketName = "habits"

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
	r := chi.NewRouter()

	db, err := bolt.Open("habits.db", 0600, nil)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	r.Post("/track", func(w http.ResponseWriter, r *http.Request) {
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
			cmd.Println("Error saving habit:", err)
			return
		}

		cmd.Printf("/track - %s\n", habitJson)
	})

	cmd.Println("listening on localhost:8080")
	http.ListenAndServe(":8080", r)
}
