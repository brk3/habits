package cmd

import (
	"io"
	"net/http"

	"github.com/brk3/habits/internal/config"
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
	cfg, err := config.Load("config.yaml")
	if err != nil {
		cmd.Println("Error loading config file", err)
		return
	}

	resp, err := http.Get(cfg.APIBaseURL + "/habits")
	if err != nil {
		cmd.Println("Error fetching habits:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cmd.Println("Error reading response:", err)
		return
	}

	cmd.Println(string(body))
}

func init() {
	rootCmd.AddCommand(listCmd)
}
