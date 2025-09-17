package cmd

import (
	"log"
	"net/http"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/server"
	"github.com/brk3/habits/internal/storage/bolt"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load("config.yaml")
		if err != nil {
			return err
		}

		log.Println("Opening DB...")
		store, err := bolt.Open(cfg.DBPath)
		if err != nil {
			return err
		}
		defer store.Close()

		s, err := server.New(cfg, store)
		if err != nil {
			return err
		}
		log.Println("Listening on", cfg.Server.Addr())
		return http.ListenAndServe(cfg.Server.Addr(), s.Router())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
