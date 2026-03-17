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

var addressbookListCmd = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {
		return List(
			jmap.ListAddressbooks,
			func(list []jmap.Addressbook) []table.Column {
				mwId := max(len("ID"), tools.MappedLen(list, func(a jmap.Addressbook) string { return a.Id }))
				mwName := tools.MappedLen(list, func(a jmap.Addressbook) string { return a.Name })
				mwDescription := min(50, tools.MappedLen(list, func(a jmap.Addressbook) string { return a.Description }))
				mwSortOrder := tools.MappedLen(list, func(a jmap.Addressbook) string { return strconv.Itoa(a.SortOrder) })
				mwIsDefault := max(len("Deflt"), len("false"))
				mwIsSubscribed := max(len("Subbed"), len("false"))
				return []table.Column{
					{Title: "ID", Width: mwId},
					{Title: "Name", Width: mwName},
					{Title: "Description", Width: mwDescription},
					{Title: "Ord", Width: mwSortOrder},
					{Title: "Deflt", Width: mwIsDefault},
					{Title: "Subbed", Width: mwIsSubscribed},
				}
			},
			func(a jmap.Addressbook) table.Row {
				return table.Row{
					a.Id,
					a.Name,
					a.Description,
					strconv.Itoa(a.SortOrder),
					strconv.FormatBool(a.IsDefault),
					strconv.FormatBool(a.IsSubscribed),
				}
			},
			func(a jmap.Addressbook, titleStyle lipgloss.Style) string {
				text := ""
				text += titleStyle.Render("Rights:") + "\n"
				text += fmt.Sprintf("read: %v\n", a.MyRights.MayRead)
				text += fmt.Sprintf("write: %v\n", a.MyRights.MayWrite)
				text += fmt.Sprintf("delete: %v\n", a.MyRights.MayDelete)
				text += fmt.Sprintf("admin: %v\n", a.MyRights.MayAdmin)
				return text
			},
			func(a jmap.Addressbook) string {
				return a.Id
			},
		)
	},
}

func init() {
	addressbookCmd.AddCommand(addressbookListCmd)
}
