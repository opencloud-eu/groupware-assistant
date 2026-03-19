package cmd

import (
	"github.com/spf13/cobra"
)

var mailboxCmd = &cobra.Command{
	Use:     "mailbox",
	Aliases: []string{"mailboxes", "mbox", "mb"},
	Short:   "Operations on Mailboxes",
}

func init() {
	rootCmd.AddCommand(mailboxCmd)
}
