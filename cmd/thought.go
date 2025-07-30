package cmd

import (
	"encoding/json"
	"time"

	"github.com/spf13/cobra"
)

type Thought struct {
	Content   string    `json:"Content"`
	TimeStamp time.Time `json:"TimeStamp"`
}

var thoughtCmd = &cobra.Command{
	Use:   "thought",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		thought(args[0], cmd)
	},
}

func init() {
	rootCmd.AddCommand(thoughtCmd)
}

func thought(content string, cmd *cobra.Command) {
	h := &Thought{
		Content:   content,
		TimeStamp: time.Now(),
	}
	habitJson, _ := json.Marshal(h)
	cmd.Println(string(habitJson))
}
