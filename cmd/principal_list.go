package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

var principalListCmd = listCommand(
	jmap.ListPrincipals,
	func(list []jmap.Principal) []table.Column {
		mwId := max(len("ID"), tools.MappedLen(list, func(a jmap.Principal) string { return a.Id }))
		mwName := max(len("Name"), tools.MappedLen(list, func(a jmap.Principal) string { return a.Name }))
		mwType := max(len("Type"), tools.MappedLen(list, func(a jmap.Principal) string { return a.Type }))
		mwDescription := min(50, tools.MappedLen(list, func(a jmap.Principal) string { return a.Description }))
		mwEmail := max(len("Email"), tools.MappedLen(list, func(a jmap.Principal) string { return a.Email }))
		return []table.Column{
			{Title: "ID", Width: mwId},
			{Title: "Name", Width: mwName},
			{Title: "Type", Width: mwType},
			{Title: "Description", Width: mwDescription},
			{Title: "Email", Width: mwEmail},
		}
	},
	func(p jmap.Principal) table.Row {
		return table.Row{
			p.Id,
			p.Name,
			p.Type,
			p.Description,
			p.Email,
		}
	},
	func(a jmap.Principal, titleStyle lipgloss.Style) string {
		var text strings.Builder
		text.WriteString(titleStyle.Render("Accounts:"))
		text.WriteString("\n")
		for id, acc := range a.Accounts {
			text.WriteString(titleStyle.Render(id))
			text.WriteString("\n")
			fmt.Fprintf(&text, "Name: %v\n", acc.Name)
			fmt.Fprintf(&text, "IsPersonal: %v\n", acc.IsPersonal)
			fmt.Fprintf(&text, "IsReadOnly: %v\n", acc.IsReadOnly)
		}
		text.WriteString("\n")
		text.WriteString(titleStyle.Render("Capabilities:"))
		text.WriteString("\n")
		fmt.Fprintf(&text, "%v\n", a.Capabilities)
		text.WriteString("\n")
		text.WriteString(titleStyle.Render("Other:"))
		text.WriteString("\n")
		fmt.Fprintf(&text, "TZ: %v\n", a.TimeZone)
		return text.String()
	},
	func(p jmap.Principal) string {
		return p.Id
	},
)

func init() {
	principalCmd.AddCommand(principalListCmd)
}
