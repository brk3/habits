package cmd

import (
	"fmt"
	"os"

	"github.com/brk3/habits/internal/nudge"
	"github.com/spf13/cobra"
)

var (
	emailFlag string
	hoursFlag int
)

var nudgeCmd = &cobra.Command{
	Use:   "nudge",
	Short: "Send a reminder for habit streaks expiring within a certain window",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("RESEND_API_KEY") == "" {
			return fmt.Errorf("RESEND_API_KEY environment variable is not set")
		}
		if emailFlag == "" {
			return fmt.Errorf("--email is required")
		}
		if hoursFlag <= 0 {
			return fmt.Errorf("--hours must be greater than 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		nudge.Nudge(emailFlag, hoursFlag)
	},
}

func init() {
	rootCmd.AddCommand(nudgeCmd)

	nudgeCmd.Flags().StringVar(&emailFlag, "email", "", "Email address (required)")
	nudgeCmd.Flags().IntVar(&hoursFlag, "hours", 0, "Number of hours spent (required)")

	nudgeCmd.MarkFlagRequired("email")
	nudgeCmd.MarkFlagRequired("hours")
}
