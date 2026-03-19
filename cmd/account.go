package cmd

import (
	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:        "account",
	Aliases:    []string{"accounts", "acc", "ac"},
	SuggestFor: []string{"acount", "acounts"},
	Short:      "Operations on Accounts",
}

func init() {
	rootCmd.AddCommand(accountCmd)
}
