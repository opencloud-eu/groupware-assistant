package cmd

import (
	"github.com/spf13/cobra"
)

var principalCmd = &cobra.Command{
	Use:     "principal",
	Aliases: []string{"principals", "pr", "p"},
	Short:   "Operations on Principals",
}

func init() {
	rootCmd.AddCommand(principalCmd)
}
