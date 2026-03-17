package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var calendarDeleteCmd = &cobra.Command{
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

		if err := jmap.DeleteCalendar(j, AccountId, id); err != nil {
			return err
		} else {
			fmt.Printf("Deleted calendar %s\n", id)
			return nil
		}
	},
}

func init() {
	calendarCmd.AddCommand(calendarDeleteCmd)
}
