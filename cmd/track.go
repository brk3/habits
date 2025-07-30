package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

type Habit struct {
	Content   string
	TimeStamp string
}

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		track(args[0], cmd)
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// trackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// trackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func track(content string, cmd *cobra.Command) {
	h := &Habit{
		Content:   content,
		TimeStamp: "00:00:00",
	}
	habitJson, _ := json.Marshal(h)
	cmd.Println(string(habitJson))
}
