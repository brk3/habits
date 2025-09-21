package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/brk3/habits/internal/apiclient"
	"github.com/brk3/habits/pkg/habit"
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

func track(name string, note string, cmd *cobra.Command) {
	h := &habit.Habit{
		Name:      name,
		Note:      note,
		TimeStamp: time.Now().Unix(),
	}
	apiclient := apiclient.New(cfg.APIBaseURL, cfg.AuthToken)
	err := apiclient.PutHabit(cmd.Context(), h)
	if err != nil {
		cmd.Printf("Error recording habit: %v\n", err)
		return
	}
	cmd.Printf("Recorded habit: %s - %s\n", name, note)
}

func init() {
	rootCmd.AddCommand(trackCmd)
}
