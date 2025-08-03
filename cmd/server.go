package cmd

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"

	"brk3.github.io/habits/internal/server"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		s := &server.Server{
			BucketName: "habits",
			Version:    Version,
			BuildDate:  BuildDate,
		}
		if err := s.InitDB(); err != nil {
			log.Fatalf("failed to open db: %v", err)
		}
		defer s.CloseDB()

		cmd.Println("listening on localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", s.Router()))
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
