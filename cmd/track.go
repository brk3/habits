package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

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

	resp, err := http.Post("http://localhost:8080/habits", "application/json",
		bytes.NewReader(habitJson))
	if err != nil {
		cmd.Println("Error saving habit:", err)
	} else {
		cmd.Println(resp)
	}
}
