package cmd

import (
	"github.com/spf13/cobra"
)

var addressbookCmd = &cobra.Command{
	Use:     "addressbook",
	Aliases: []string{"addressbooks", "abook", "ab"},
}

func init() {
	rootCmd.AddCommand(addressbookCmd)
}
