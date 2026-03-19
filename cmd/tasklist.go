package cmd

import (
	"github.com/spf13/cobra"
)

var tasklistCmd = &cobra.Command{
	Use:     "tasklist",
	Aliases: []string{"tasklists", "ts"},
	Short:   "Operations on Tasklists",
}

func init() {
	rootCmd.AddCommand(tasklistCmd)
}
