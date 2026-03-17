package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

var calendarListCmd = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {
		return List(
			jmap.ListCalendars,
			func(list []jmap.Calendar) []table.Column {
				mwId := max(len("ID"), tools.MappedLen(list, func(a jmap.Calendar) string { return a.Id }))
				mwName := tools.MappedLen(list, func(a jmap.Calendar) string { return a.Name })
				mwDescription := min(50, tools.MappedLen(list, func(a jmap.Calendar) string { return a.Description }))
				mwColor := max(len("Color"), tools.MappedLen(list, func(a jmap.Calendar) string { return a.Color }))
				mwSortOrder := tools.MappedLen(list, func(a jmap.Calendar) string { return strconv.Itoa(a.SortOrder) })
				mwIsDefault := max(len("Deflt"), len("false"))
				mwIsSubscribed := max(len("Subbed"), len("false"))
				mwIsVisible := max(len("Vis"), len("false"))
				mwTimeZone := max(len("TZ"), tools.MappedLen(list, func(a jmap.Calendar) string { return a.TimeZone }))
				//mwIncludeInAvailability := max(len("IIA"), tools.MappedLen(list, func(a jmap.Calendar) string { return a.IncludeInAvailability }))
				return []table.Column{
					{Title: "ID", Width: mwId},
					{Title: "Name", Width: mwName},
					{Title: "Description", Width: mwDescription},
					{Title: "Color", Width: mwColor},
					{Title: "Ord", Width: mwSortOrder},
					{Title: "Deflt", Width: mwIsDefault},
					{Title: "Subbed", Width: mwIsSubscribed},
					{Title: "Vis", Width: mwIsVisible},
					//{Title: "IIA", Width: mwIncludeInAvailability},
					{Title: "TZ", Width: mwTimeZone},
				}
			},
			func(a jmap.Calendar) table.Row {
				return table.Row{
					a.Id,
					a.Name,
					a.Description,
					a.Color,
					strconv.Itoa(a.SortOrder),
					strconv.FormatBool(a.IsDefault),
					strconv.FormatBool(a.IsSubscribed),
					strconv.FormatBool(a.IsVisible),
					//a.IncludeInAvailability,
					a.TimeZone,
				}
			},
			func(a jmap.Calendar, titleStyle lipgloss.Style) string {
				text := ""
				text += titleStyle.Render("Rights:") + "\n"
				text += fmt.Sprintf("MayReadItems: %v\n", a.MyRights.MayReadItems)
				text += fmt.Sprintf("MayReadFreeBusy: %v\n", a.MyRights.MayReadFreeBusy)
				text += fmt.Sprintf("MayWriteOwn: %v\n", a.MyRights.MayWriteOwn)
				text += fmt.Sprintf("MayWriteAll: %v\n", a.MyRights.MayWriteAll)
				text += fmt.Sprintf("MayRSVP: %v\n", a.MyRights.MayRSVP)
				text += fmt.Sprintf("MayUpdatePrivate: %v\n", a.MyRights.MayUpdatePrivate)
				text += fmt.Sprintf("MayDelete: %v\n", a.MyRights.MayDelete)
				text += fmt.Sprintf("MayAdmin: %v\n", a.MyRights.MayAdmin)
				return text
			},
			func(a jmap.Calendar) string {
				return a.Id
			},
		)
	},
}

func init() {
	calendarCmd.AddCommand(calendarListCmd)
}
