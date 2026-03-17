package cmd

import (
	"github.com/spf13/cobra"
)

var contactCmd = &cobra.Command{
	Use:     "contact",
	Aliases: []string{"contacts"},
	Short:   "Operations on Contacts within Address Books",
}

func init() {
	rootCmd.AddCommand(contactCmd)
}
