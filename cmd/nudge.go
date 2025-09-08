package cmd

import (
	"fmt"
	"os"

	"github.com/brk3/habits/internal/nudge"
	"github.com/spf13/cobra"
)

var (
	notifyEmail    string
	resendApiKey   string
	nudgeThreshold int
)

var nudgeCmd = &cobra.Command{
	Use:   "nudge",
	Short: "Send a reminder for habit streaks expiring within a certain window",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if resendApiKey = os.Getenv("HABITS_RESEND_API_KEY"); resendApiKey == "" {
			return fmt.Errorf("RESEND_API_KEY environment variable is not set")
		}
		if notifyEmail = os.Getenv("HABITS_NOTIFY_EMAIL"); notifyEmail == "" {
			return fmt.Errorf("NOTIFY_EMAIL environment variable is not set")
		}
		if nudgeThreshold := os.Getenv("HABITS_NUDGE_THRESHOLD"); nudgeThreshold == "" {
			return fmt.Errorf("HABITS_NUDGE_THRESHOLD environment variable is not set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		nudge.Nudge(notifyEmail, nudgeThreshold, resendApiKey)
	},
}

func init() {
	rootCmd.AddCommand(nudgeCmd)
}
