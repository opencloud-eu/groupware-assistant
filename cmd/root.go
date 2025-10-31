package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "groupware-assistant",
	Short: "Creates all sorts of Groupware related data for testing",
	Long: `Groupware-Assistant is a CLI application that is capable of
creating several types of realistic-ish Groupware data, to populate an
IMAP and JMAP server in order to develop applications or run tests.
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	JmapUrl   string
	Username  string
	Password  string
	AccountId string
	Trace     bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&JmapUrl, "url", "b", "https://stalwart.opencloud.test", "JMAP base URL")
	rootCmd.PersistentFlags().StringVarP(&Username, "username", "u", "alan", "JMAP basic authentication username")
	rootCmd.PersistentFlags().StringVarP(&Password, "password", "p", "demo", "JMAP basic authentication password")
	rootCmd.PersistentFlags().StringVarP(&AccountId, "account-id", "A", "", "JMAP account ID to use, default behavior is to use the default account")
	rootCmd.PersistentFlags().BoolVar(&Trace, "trace", false, "Show JMAP HTTP traffic")
}
