package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var mailboxCreateCmd = &cobra.Command{
	Use: "create",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("name must be non-empty")
		}
		parentId, err := cmd.Flags().GetString("parent-id")
		if err != nil {
			return err
		}
		role, err := cmd.Flags().GetString("role")
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

		create := jmap.NewMailbox{
			Name:         name,
			IsSubscribed: &subbed,
			SortOrder:    sortOrder,
			ParentId:     parentId,
			Role:         role,
		}

		if id, err := jmap.CreateMailbox(j, AccountId, create); err != nil {
			return err
		} else {
			fmt.Printf("Created mailbox %s\n", id)
			return nil
		}
	},
}

func init() {
	mailboxCmd.AddCommand(mailboxCreateCmd)

	mailboxCreateCmd.Flags().StringP("name", "n", "", "Name of the Mailbox")
	mailboxCreateCmd.Flags().StringP("parent-id", "P", "", "ID of the parent Mailbox, if applicable")
	mailboxCreateCmd.Flags().StringP("role", "R", "", "Role, if applicable")
	mailboxCreateCmd.Flags().BoolP("subscribed", "s", true, "Subscribed")
	mailboxCreateCmd.Flags().Int("sort-order", 0, "Sort order")
}
