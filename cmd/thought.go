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

// thoughtCmd represents the thought command
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// thoughtCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// thoughtCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func thought(content string, cmd *cobra.Command) {
	h := &Thought{
		Content:   content,
		TimeStamp: time.Now(),
	}
	habitJson, _ := json.Marshal(h)
	cmd.Println(string(habitJson))
}
