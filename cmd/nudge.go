package cmd

import (
	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/nudge"
	"github.com/brk3/habits/internal/nudge/resend"

	"github.com/spf13/cobra"
)

var nudgeCmd = &cobra.Command{
	Use:   "nudge",
	Short: "Send a reminder for habit streaks expiring within a certain window",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			cmd.Printf("error loading config file: %v\n", err)
			return
		}
		if cfg.Nudge.ResendAPIKey == "" {
			cmd.Println("nudge.resend_api_key is required")
			return
		}
		if cfg.Nudge.NotifyEmail == "" {
			cmd.Println("nudge.notify_email is required")
			return
		}
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
