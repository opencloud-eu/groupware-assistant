package cmd

import (
	"github.com/spf13/cobra"
)

var emailCmd = &cobra.Command{
	Use: "email",
}

func init() {
	rootCmd.AddCommand(emailCmd)
}
