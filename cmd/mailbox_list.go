package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

var mailboxListCmd = listCommand(
	jmap.ListMailboxes,
	func(list []jmap.Mailbox) []table.Column {
		mwId := max(len("ID"), tools.MappedLen(list, func(a jmap.Mailbox) string { return a.Id }))
		mwParentId := max(len("PID"), tools.MappedLen(list, func(a jmap.Mailbox) string { return a.ParentId }))
		mwName := max(len("Name"), tools.MappedLen(list, func(a jmap.Mailbox) string { return a.Name }))
		mwRole := max(len("Role"), tools.MappedLen(list, func(a jmap.Mailbox) string { return a.Role }))
		mwSortOrder := tools.MappedLen(list, func(a jmap.Mailbox) string { return strconv.Itoa(a.SortOrder) })
		mwIsSubscribed := max(len("Subbed"), len("false"))
		return []table.Column{
			{Title: "ID", Width: mwId},
			{Title: "PID", Width: mwParentId},
			{Title: "Name", Width: mwName},
			{Title: "Role", Width: mwRole},
			{Title: "Ord", Width: mwSortOrder},
			{Title: "Subbed", Width: mwIsSubscribed},
			{Title: "Unread", Width: 6},
			{Title: "Total", Width: 6},
		}
	},
	func(a jmap.Mailbox) table.Row {
		sub := ""
		if a.IsSubscribed != nil {
			sub = strconv.FormatBool(*a.IsSubscribed)
		}
		return table.Row{
			a.Id,
			a.ParentId,
			a.Name,
			a.Role,
			strconv.Itoa(a.SortOrder),
			sub,
			fmt.Sprintf("% 6d", a.UnreadEmails),
			fmt.Sprintf("% 6d", a.TotalEmails),
		}
	},
	func(a jmap.Mailbox, titleStyle lipgloss.Style) string {
		text := ""
		text += titleStyle.Render("Rights:") + "\n"
		text += fmt.Sprintf("MayReadItems: %v\n", a.MyRights.MayReadItems)
		text += fmt.Sprintf("MayRename: %v\n", a.MyRights.MayRename)
		text += fmt.Sprintf("MayDelete: %v\n", a.MyRights.MayDelete)
		text += fmt.Sprintf("MayAddItems: %v\n", a.MyRights.MayAddItems)
		text += fmt.Sprintf("MayCreateChild: %v\n", a.MyRights.MayCreateChild)
		text += fmt.Sprintf("MayRemoveItems: %v\n", a.MyRights.MayRemoveItems)
		text += fmt.Sprintf("MaySetKeywords: %v\n", a.MyRights.MaySetKeywords)
		text += fmt.Sprintf("MaySetSeen: %v\n", a.MyRights.MaySetSeen)
		text += fmt.Sprintf("MaySubmit: %v\n", a.MyRights.MaySubmit)
		return text
	},
	func(a jmap.Mailbox) string {
		return a.Id
	},
)

func init() {
	mailboxCmd.AddCommand(mailboxListCmd)
}
