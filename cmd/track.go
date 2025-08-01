package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
)

type Habit struct {
	Name      string    `json:"Name"`
	Note      string    `json:"Note"`
	TimeStamp time.Time `json:"TimeStamp"`
}

var db *bbolt.DB
var bucketName = "habits"

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Record a habit entry with a timestamp",
	Long: `The "track" command lets you log a habit entry from the command line.

For example:
  habits track guitar "10 mins of major scales"

This will store the habit along with the current timestamp.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.TrimSpace(args[0])
		note := strings.TrimSpace(args[1])

		if name == "" {
			cmd.Println("Error: habit name cannot be empty")
			os.Exit(1)
		}

		if len(name) > 24 {
			cmd.Println("Error: habit name too long (max 24 characters)")
			os.Exit(1)
		}

		track(name, note, cmd)
	},
}

func InitDB(d *bbolt.DB) {
	db = d
}

func init() {
	rootCmd.AddCommand(trackCmd)
}

func track(name string, note string, cmd *cobra.Command) {
	h := &Habit{
		Name:      name,
		Note:      note,
		TimeStamp: time.Now(),
	}
	habitJson, _ := json.Marshal(h)

	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		key := name + "/" + h.TimeStamp.Format(time.RFC3339)
		return b.Put([]byte(key), habitJson)
	})
	if err != nil {
		cmd.Println("Error saving habit:", err)
	}
}
