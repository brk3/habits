package cmd

import (
	"log/slog"
	"os"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/logger"
	"github.com/spf13/cobra"
)

var cfg *config.Config

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
	logger.Init(slog.LevelInfo)

	var err error
	cfg, err = config.Load()
	if err != nil {
		logger.Error("Error loading config file", "err", err)
	} else {
		logger.Init(cfg.SLogLevel)
		logger.Debug("Config loaded successfully")
	}
}
