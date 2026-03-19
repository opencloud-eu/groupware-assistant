package cmd

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.yaml.in/yaml/v4"

	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

type ListModel[T any] struct {
	table       table.Model
	hasViewport bool
	viewport    viewport.Model
	titleStyle  lipgloss.Style
	bulletStyle lipgloss.Style
	itemsById   map[string]T
	detailer    func(T, lipgloss.Style) string
}

func (m ListModel[T]) Init() tea.Cmd { return nil }

func (m ListModel[T]) View() string {
	if m.hasViewport {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyle.Render(m.table.View()),
			baseStyle.Render(m.viewport.View()),
		) + "\n"
	} else {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyle.Render(m.table.View()),
		) + "\n"
	}
}

func (m ListModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	if m.hasViewport {
		text := ""
		row := m.table.SelectedRow()
		if row != nil {
			id := strings.TrimSpace(row[0])
			if a, ok := m.itemsById[id]; ok {
				text = m.detailer(a, m.titleStyle)
			}
		}
		m.viewport.SetContent(text)
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func ListJson[T any](
	lister func(*jmap.Jmap, string) ([]T, error),
) error {
	u, err := url.Parse(JmapUrl)
	if err != nil {
		return err
	}

	j, err := jmap.NewJmap(u, Username, Password, Trace, Color)
	if err != nil {
		return err
	}
	defer j.Close()

	list, err := lister(j, AccountId)
	if err != nil {
		return err
	}

	if b, err := json.MarshalIndent(list, "", "  "); err != nil {
		return err
	} else {
		_, err := os.Stdout.Write(b)
		return err
	}
}

func ListYaml[T any](
	lister func(*jmap.Jmap, string) ([]T, error),
) error {
	u, err := url.Parse(JmapUrl)
	if err != nil {
		return err
	}

	j, err := jmap.NewJmap(u, Username, Password, Trace, Color)
	if err != nil {
		return err
	}
	defer j.Close()

	list, err := lister(j, AccountId)
	if err != nil {
		return err
	}

	if b, err := yaml.Marshal(list); err != nil {
		return err
	} else {
		_, err := os.Stdout.Write(b)
		return err
	}
}

func List[T any](
	lister func(*jmap.Jmap, string) ([]T, error),
	columner func([]T) []table.Column,
	rowMapper func(T) table.Row,
	detailer func(T, lipgloss.Style) string,
	idMapper func(T) string,
) error {
	u, err := url.Parse(JmapUrl)
	if err != nil {
		return err
	}

	j, err := jmap.NewJmap(u, Username, Password, Trace, Color)
	if err != nil {
		return err
	}
	defer j.Close()

	list, err := lister(j, AccountId)
	if err != nil {
		return err
	}

	columns := columner(list)
	rows := make([]table.Row, len(list))
	for i, a := range list {
		rows[i] = rowMapper(a)
	}

	height := max(10, min(40, len(rows)))

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	var v viewport.Model
	hasViewport := false
	if detailer != nil {
		v = viewport.New(40, height+1)
		hasViewport = true
	}

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

	m := ListModel[T]{
		table:       t,
		viewport:    v,
		hasViewport: hasViewport,
		titleStyle:  lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("240")),
		bulletStyle: lipgloss.NewStyle().MarginLeft(1).PaddingRight(1).Foreground(lipgloss.Color("202")),
		itemsById:   tools.ToMap(list, idMapper),
		detailer:    detailer,
	}
	_, err = tea.NewProgram(m).Run()
	return err
}
