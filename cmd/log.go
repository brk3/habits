package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type Habit struct {
	Content   string
	TimeStamp string
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log(args[0])
	},
}

func init() {
	rootCmd.AddCommand(logCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func log(content string) {
	fmt.Println("log called!")
	h := &Habit{
		Content:   content,
		TimeStamp: "00:00:00",
	}
	habitJson, _ := json.Marshal(h)
	fmt.Println(string(habitJson))
}
