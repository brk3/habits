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

		// TODO: These options should be configurable somehow

		// OIDC issuer url
		issuer := "https://idm.example.com/oauth2/openid/idm-provided-id"

		// OIDC provider gives you this information
		clientID := "idm-provided-id"
		clientSecret := "<secret token provided by oidc provided>"

		// redirectURL in current form it would only support a single OIDC provider
		// When making this configurable, that URL will change to `/auth/callback/{ID}`
		// to support multiple providers (i.e. GitLab && GitHub && Google)
		redirectURL := "https://habits.example.com/auth/callback"

		s, err := server.New(store, issuer, clientID, clientSecret, redirectURL)
		if err != nil {
			return err
		}
		log.Println("Listening on 127.0.0.1:9999")
		return http.ListenAndServe("127.0.0.1:9999", s.Router())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
