package cmd

import (
	"encoding/json"
	"net/http"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/pkg/versioninfo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `The "version" command displays the current version info for both client
and server if available.`,
	Run: func(cmd *cobra.Command, args []string) {
		version(cmd)
	},
}

func version(cmd *cobra.Command) {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		cmd.Println("Error loading config file", err)
		return
	}

	resp, err := http.Get(cfg.APIBaseURL + "/version")
	if err != nil {
		cmd.Println("Error fetching server version:", err)
		return
	}
	defer resp.Body.Close()

	serverVersion := &versioninfo.VersionInfo{}
	if err := json.NewDecoder(resp.Body).Decode(serverVersion); err != nil {
		cmd.Println("Error decoding version response:", err)
		return
	}

	cmd.Printf("Server Version: %s\n", serverVersion.Version)
	cmd.Printf("Client Version: %s\n", versioninfo.Version)
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
