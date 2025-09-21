package cmd

import (
	"fmt"
	"net/http"

	"github.com/brk3/habits/internal/logger"
	"github.com/brk3/habits/internal/server"
	"github.com/brk3/habits/internal/storage/bolt"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Opening DB...")
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
		logger.Info("Listening on", "addr", addr)
		if cfg.Server.TLS.Enabled {
			return http.ListenAndServeTLS(addr, cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile, s.Router())
		} else {
			return http.ListenAndServe(addr, s.Router())
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
