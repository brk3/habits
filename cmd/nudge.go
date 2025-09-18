package cmd

import (
	"fmt"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/nudge"
	"github.com/brk3/habits/internal/nudge/resend"

	"github.com/spf13/cobra"
)

var (
	cfg            *config.Config
	notifyEmail    string
	resendApiKey   string
	nudgeThreshold int
)

var nudgeCmd = &cobra.Command{
	Use:   "nudge",
	Short: "Send a reminder for habit streaks expiring within a certain window",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("error loading config file: %v", err)
		}
		if cfg.Nudge.ResendAPIKey == "" {
			return fmt.Errorf("nudge.resend_api_key is required")
		}
		if cfg.Nudge.NotifyEmail == "" {
			return fmt.Errorf("nudge.notify_email is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		n := resend.ResendNotifier{
			ApiKey: cfg.Nudge.ResendAPIKey,
			Email:  cfg.Nudge.NotifyEmail,
		}
		nudge.Nudge(cfg, &n, cfg.Nudge.ThresholdHours)
	},
}

func init() {
	rootCmd.AddCommand(nudgeCmd)
}
