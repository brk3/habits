package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "habits",
	Short: "Log and track personal habits or thoughts",
	Long: `
	Habits is a CLI tool to track activities over time, as well as recording ad-hoc thoughts
	or notes. It supports a growing set of commands for both structured and unstructured
	journaling.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
