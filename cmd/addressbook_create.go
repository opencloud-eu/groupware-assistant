package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var addressbookCreateCmd = &cobra.Command{
	Use: "create",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("name must be non-empty")
		}
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		subbed, err := cmd.Flags().GetBool("subscribed")
		if err != nil {
			return err
		}
		sortOrder, err := cmd.Flags().GetInt("sort-order")
		if err != nil {
			return err
		}

		u, err := url.Parse(JmapUrl)
		if err != nil {
			return err
		}

		j, err := jmap.NewJmap(u, Username, Password, Trace, Color)
		if err != nil {
			return err
		}
		defer j.Close()

		create := jmap.NewAddressbook{
			Name:         name,
			Description:  description,
			IsSubscribed: subbed,
			SortOrder:    sortOrder,
		}

		if id, err := jmap.CreateAddressbook(j, AccountId, create); err != nil {
			return err
		} else {
			fmt.Printf("Created address book %s\n", id)
			return nil
		}
	},
}

func init() {
	addressbookCmd.AddCommand(addressbookCreateCmd)

	addressbookCreateCmd.Flags().StringP("name", "n", "", "Name of the Address Book")
	addressbookCreateCmd.Flags().StringP("description", "d", "", "Description of the Address Book")
	addressbookCreateCmd.Flags().BoolP("subscribed", "s", true, "Subscribed")
	addressbookCreateCmd.Flags().Int("sort-order", 0, "Sort order")
}
