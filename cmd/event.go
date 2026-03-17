package cmd

import (
	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use:     "event",
	Aliases: []string{"events"},
	Short:   "Operations on Events within Calendars",
}

func init() {
	rootCmd.AddCommand(eventCmd)
}
