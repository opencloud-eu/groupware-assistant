package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var addressbookDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		u, err := url.Parse(JmapUrl)
		if err != nil {
			return err
		}
		j, err := jmap.NewJmap(u, Username, Password, Trace, Color)
		if err != nil {
			return err
		}
		defer j.Close()

		if err := jmap.DeleteAddressbook(j, AccountId, id); err != nil {
			return err
		} else {
			fmt.Printf("Deleted address book %s\n", id)
			return nil
		}
	},
}

func init() {
	addressbookCmd.AddCommand(addressbookDeleteCmd)
}
