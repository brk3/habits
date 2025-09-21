package cmd

import (
	"context"

	"github.com/brk3/habits/internal/apiclient"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List habits",
	Long:  `The "list" command lets you list your tracked habits.`,
	Run: func(cmd *cobra.Command, args []string) {
		list(cmd)
	},
}

func list(cmd *cobra.Command) {
	apiclient := apiclient.New(cfg.APIBaseURL, cfg.AuthToken)

	habits, err := apiclient.ListHabits(context.Background())
	if err != nil {
		cmd.Printf("Error fetching habits: %v\n", err)
		return
	}
	for _, h := range habits {
		cmd.Println(h)
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
