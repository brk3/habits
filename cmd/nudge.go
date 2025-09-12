package cmd

import (
	"fmt"
	"os"
	"strconv"

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
			return fmt.Errorf("HABITS_RESEND_API_KEY environment variable is not set")
		}

		if notifyEmail = os.Getenv("HABITS_NOTIFY_EMAIL"); notifyEmail == "" {
			return fmt.Errorf("HABITS_NOTIFY_EMAIL environment variable is not set")
		}

		nudgeThresholdStr := os.Getenv("HABITS_NUDGE_THRESHOLD")
		if nudgeThresholdStr == "" {
			return fmt.Errorf("HABITS_NUDGE_THRESHOLD environment variable is not set")
		}
		var err error
		nudgeThreshold, err = strconv.Atoi(nudgeThresholdStr)
		if err != nil {
			return fmt.Errorf("HABITS_NUDGE_THRESHOLD must be a valid integer: %v", err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		n := nudge.ResendNotifier{
			ApiKey: resendApiKey,
			Email:  notifyEmail,
		}
		nudge.Nudge(&n, nudgeThreshold)
	},
}

func init() {
	rootCmd.AddCommand(nudgeCmd)
}
