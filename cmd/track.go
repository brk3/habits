package cmd

import (
	"encoding/json"
	"time"

	"github.com/spf13/cobra"
)

type Habit struct {
	Content   string    `json:"Content"`
	TimeStamp time.Time `json:"TimeStamp"`
}

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Record a habit entry with a timestamp",
	Long: `The "track" command lets you log a habit entry from the command line.

For example:
  habits track "guitar"

This will store the habit along with the current timestamp.`,

	Run: func(cmd *cobra.Command, args []string) {
		track(args[0], cmd)
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
}

func track(content string, cmd *cobra.Command) {
	h := &Habit{
		Content:   content,
		TimeStamp: time.Now(),
	}
	habitJson, _ := json.Marshal(h)
	cmd.Println(string(habitJson))
}
