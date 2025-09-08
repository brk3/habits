package cmd

import (
	"fmt"
	"os"

	"github.com/brk3/habits/internal/nudge"
	"github.com/spf13/cobra"
)

var (
	notifyEmail  string
	resendApiKey string
)

var nudgeCmd = &cobra.Command{
	Use:   "nudge",
	Short: "Send a reminder for habit streaks expiring within a certain window",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if resendApiKey = os.Getenv("RESEND_API_KEY"); resendApiKey == "" {
			return fmt.Errorf("RESEND_API_KEY environment variable is not set")
		}
		if notifyEmail = os.Getenv("NOTIFY_EMAIL"); notifyEmail == "" {
			return fmt.Errorf("NOTIFY_EMAIL environment variable is not set")
		}
		if notifyHourBeforeMidnightEnv := os.Getenv("NOTIFY_HOUR_BEFORE_MIDNIGHT"); notifyHourBeforeMidnightEnv == "" {
			return fmt.Errorf("NOTIFY_HOUR_BEFORE_MIDNIGHT environment variable is not set")
		}
		// TODO(pbourke): finish rest of cfg vars
		/*
			if emailFlag == "" {
				return fmt.Errorf("--email is required")
			}
			if hoursFlag <= 0 {
				return fmt.Errorf("--hours must be greater than 0")
			}
		*/
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		nudge.Nudge("", 2, resendApiKey)
	},
}

func init() {
	rootCmd.AddCommand(nudgeCmd)
}
