package cmd

import (
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:        "calendar",
	Aliases:    []string{"calendars", "cal", "cals"},
	SuggestFor: []string{"calender"},
	Short:      "Operations on Calendars",
}

func init() {
	rootCmd.AddCommand(calendarCmd)
}
