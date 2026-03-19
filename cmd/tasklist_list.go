package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

var tasklistListCmd = listCommand(
	jmap.ListTasklists,
	func(list []jmap.TaskList) []table.Column {
		mwId := max(len("ID"), tools.MappedLen(list, func(a jmap.TaskList) string { return a.Id }))
		mwName := tools.MappedLen(list, func(a jmap.TaskList) string { return a.Name })
		mwRole := max(len("Role"), tools.MappedLen(list, func(a jmap.TaskList) string { return a.Role }))
		mwDescription := min(50, tools.MappedLen(list, func(a jmap.TaskList) string { return a.Description }))
		mwSortOrder := tools.MappedLen(list, func(a jmap.TaskList) string { return strconv.Itoa(a.SortOrder) })
		mwIsSubscribed := max(len("Subbed"), len("false"))
		return []table.Column{
			{Title: "ID", Width: mwId},
			{Title: "Name", Width: mwName},
			{Title: "Role", Width: mwRole},
			{Title: "Description", Width: mwDescription},
			{Title: "Ord", Width: mwSortOrder},
			{Title: "Subbed", Width: mwIsSubscribed},
		}
	},
	func(t jmap.TaskList) table.Row {
		return table.Row{
			t.Id,
			t.Name,
			t.Role,
			t.Description,
			strconv.Itoa(t.SortOrder),
			strconv.FormatBool(t.IsSubscribed),
		}
	},
	func(t jmap.TaskList, titleStyle lipgloss.Style) string {
		text := ""
		text += titleStyle.Render("Rights:") + "\n"
		text += fmt.Sprintf("MayReadItems: %v\n", t.MyRights.MayReadItems)
		text += fmt.Sprintf("MayWriteOwn: %v\n", t.MyRights.MayWriteOwn)
		text += fmt.Sprintf("MayWriteAll: %v\n", t.MyRights.MayWriteAll)
		text += fmt.Sprintf("MayDelete: %v\n", t.MyRights.MayDelete)
		text += fmt.Sprintf("MayRSVP: %v\n", t.MyRights.MayRSVP)
		text += fmt.Sprintf("MayUpdatePrivate: %v\n", t.MyRights.MayUpdatePrivate)
		text += fmt.Sprintf("MayAdmin: %v\n", t.MyRights.MayAdmin)
		return text
	},
	func(t jmap.TaskList) string {
		return t.Id
	},
)

func init() {
	tasklistCmd.AddCommand(tasklistListCmd)
}
