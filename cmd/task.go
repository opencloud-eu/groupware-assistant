package cmd

import (
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Operations on Tasks within Tasklists",
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
