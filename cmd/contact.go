package cmd

import (
	"github.com/spf13/cobra"
)

var contactCmd = &cobra.Command{
	Use: "contact",
}

func init() {
	rootCmd.AddCommand(contactCmd)
}
