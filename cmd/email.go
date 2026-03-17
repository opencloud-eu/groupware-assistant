package cmd

import (
	"github.com/spf13/cobra"
)

var emailCmd = &cobra.Command{
	Use:        "email",
	Aliases:    []string{"em", "emails"},
	SuggestFor: []string{"mail", "mails"},
	Short:      "Operations on Emails",
}

func init() {
	rootCmd.AddCommand(emailCmd)
}
