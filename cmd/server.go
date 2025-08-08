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
		cfg := config.Load()

		log.Println("Opening DB...")
		store, err := bolt.Open(cfg.DBPath)
		if err != nil {
			return err
		}
		defer store.Close()

		s := server.New(store)
		log.Println("Listening on :8080")
		return http.ListenAndServe(":8080", s.Router())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
