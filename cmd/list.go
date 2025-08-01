package cmd

import (
	"io"
	"net/http"

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

func init() {
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command) {
	resp, err := http.Get("http://localhost:8080/habits")
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
