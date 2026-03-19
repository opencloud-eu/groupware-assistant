package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var tasklistCreateCmd = &cobra.Command{
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

		create := jmap.NewTaskList{
			Name:         name,
			Description:  description,
			IsSubscribed: subbed,
			SortOrder:    sortOrder,
			Role:         role,
		}

		if id, err := jmap.CreateTasklist(j, AccountId, create); err != nil {
			return err
		} else {
			fmt.Printf("Created tasklist %s\n", id)
			return nil
		}
	},
}

func init() {
	tasklistCmd.AddCommand(tasklistCreateCmd)

	tasklistCreateCmd.Flags().StringP("name", "n", "", "Name of the Tasklist")
	tasklistCreateCmd.Flags().StringP("description", "d", "", "Description of the Tasklist")
	tasklistCreateCmd.Flags().StringP("role", "R", "", "Role, if applicable")
	tasklistCreateCmd.Flags().BoolP("subscribed", "s", true, "Subscribed")
	tasklistCreateCmd.Flags().Int("sort-order", 0, "Sort order")
}
