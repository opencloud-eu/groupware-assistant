package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var calendarCreateCmd = &cobra.Command{
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
		vis, err := cmd.Flags().GetBool("visible")
		if err != nil {
			return err
		}
		sortOrder, err := cmd.Flags().GetInt("sort-order")
		if err != nil {
			return err
		}
		tz, err := cmd.Flags().GetString("timezone")
		if err != nil {
			return err
		}
		color, err := cmd.Flags().GetString("color")
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

		create := jmap.NewCalendar{
			Name:         name,
			Description:  description,
			IsSubscribed: subbed,
			IsVisible:    vis,
			SortOrder:    sortOrder,
			TimeZone:     tz,
			Color:        color,
		}

		if id, err := jmap.CreateCalendar(j, AccountId, create); err != nil {
			return err
		} else {
			fmt.Printf("Created calendar %s\n", id)
			return nil
		}
	},
}

func init() {
	calendarCmd.AddCommand(calendarCreateCmd)

	calendarCreateCmd.Flags().StringP("name", "n", "", "Name of the Calendar")
	calendarCreateCmd.Flags().StringP("description", "d", "", "Description of the Calendar")
	calendarCreateCmd.Flags().BoolP("subscribed", "s", true, "Subscribed")
	calendarCreateCmd.Flags().BoolP("visible", "V", true, "Visible")
	calendarCreateCmd.Flags().Int("sort-order", 0, "Sort order")
	calendarCreateCmd.Flags().StringP("timezone", "T", "", "Timezone")
	calendarCreateCmd.Flags().StringP("color", "C", "", "Color (color name or RGB)")
}
