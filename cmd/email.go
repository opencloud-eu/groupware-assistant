package cmd

import (
	"github.com/spf13/cobra"
)

var emailCmd = &cobra.Command{
	Use:     "email",
	Aliases: []string{"em", "emails"},
}

func init() {
	rootCmd.AddCommand(emailCmd)
}
