package cmd

import (
	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use: "event",
}

func init() {
	rootCmd.AddCommand(eventCmd)
}
