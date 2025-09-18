package cmd

import (
	"fmt"
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
		cfg, err := config.Load()
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
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Println("Listening on", addr)
		return http.ListenAndServe(addr, s.Router())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
