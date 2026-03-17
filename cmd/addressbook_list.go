package cmd

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

type listModel struct {
	table        table.Model
	viewport     viewport.Model
	titleStyle   lipgloss.Style
	bulletStyle  lipgloss.Style
	addressbooks map[string]jmap.Addressbook
}

func (m listModel) Init() tea.Cmd { return nil }

func (m listModel) flag(title string, b bool) string {
	return m.titleStyle.Render(title) + ": " + strconv.FormatBool(b)
}

func (m listModel) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		baseStyle.Render(m.table.View()),
		baseStyle.Render(m.viewport.View()),
	) + "\n"
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	cmds := []tea.Cmd{}
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	text := ""
	row := m.table.SelectedRow()
	if row != nil {
		id := strings.TrimSpace(row[0])
		if a, ok := m.addressbooks[id]; ok {
			text += m.titleStyle.Render("Rights:") + "\n"
			text += m.flag("read", a.MyRights.MayRead) + "\n"
			text += m.flag("write", a.MyRights.MayWrite) + "\n"
			text += m.flag("delete", a.MyRights.MayDelete) + "\n"
			text += m.flag("admin", a.MyRights.MayAdmin) + "\n"
		}
	}
	m.viewport.SetContent(text)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

var addressbookListCmd = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {
		u, err := url.Parse(JmapUrl)
		if err != nil {
			return err
		}

		j, err := jmap.NewJmap(u, Username, Password, Trace, Color)
		if err != nil {
			return err
		}
		defer j.Close()

		l, err := jmap.NewAddressbookLister(j, AccountId)
		if err != nil {
			return err
		}
		defer l.Close()

		list, err := l.ListAddressbooks()
		if err != nil {
			return err
		}

		mwId := max(len("ID"), tools.MappedLen(list, func(a jmap.Addressbook) string { return a.Id }))
		mwName := tools.MappedLen(list, func(a jmap.Addressbook) string { return a.Name })
		mwDescription := min(50, tools.MappedLen(list, func(a jmap.Addressbook) string { return a.Description }))
		mwSortOrder := tools.MappedLen(list, func(a jmap.Addressbook) string { return strconv.Itoa(a.SortOrder) })
		mwIsDefault := max(len("Deflt"), len("false"))
		mwIsSubscribed := max(len("Subbed"), len("false"))
		columns := []table.Column{
			{Title: "ID", Width: mwId},
			{Title: "Name", Width: mwName},
			{Title: "Description", Width: mwDescription},
			{Title: "Ord", Width: mwSortOrder},
			{Title: "Deflt", Width: mwIsDefault},
			{Title: "Subbed", Width: mwIsSubscribed},
		}

		rows := make([]table.Row, len(list))
		slices.SortFunc(list, func(a, b jmap.Addressbook) int { return strings.Compare(a.Id, b.Id) })
		for i, a := range list {
			rows[i] = table.Row{
				a.Id,
				a.Name,
				a.Description,
				strconv.Itoa(a.SortOrder),
				strconv.FormatBool(a.IsDefault),
				strconv.FormatBool(a.IsSubscribed),
			}
		}

		height := max(10, min(40, len(rows)))

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(height),
		)

		v := viewport.New(40, height+1)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		m := listModel{
			table:        t,
			viewport:     v,
			titleStyle:   lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("240")),
			bulletStyle:  lipgloss.NewStyle().MarginLeft(1).PaddingRight(1).Foreground(lipgloss.Color("202")),
			addressbooks: tools.ToMap(list, func(a jmap.Addressbook) string { return a.Id }),
		}
		_, err = tea.NewProgram(m).Run()
		return err
	},
}

func init() {
	addressbookCmd.AddCommand(addressbookListCmd)
}
