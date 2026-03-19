package cmd

import (
	"fmt"
	"net/url"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"opencloud.eu/groupware-assistant/pkg/jmap"
)

func deleteCommand(objectName string, closure func(*jmap.Jmap, string, string) error) *cobra.Command {
	return &cobra.Command{
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

			if err := closure(j, AccountId, id); err != nil {
				return err
			} else {
				fmt.Printf("Deleted %s %s\n", objectName, id)
				return nil
			}
		},
	}
}

func listCommand[T any](
	lister func(*jmap.Jmap, string) ([]T, error),
	columner func([]T) []table.Column,
	rowMapper func(T) table.Row,
	detailer func(T, lipgloss.Style) string,
	idMapper func(T) string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "ll", "ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}
			switch format {
			case "json":
				return ListJson(lister)
			case "yaml":
				return ListYaml(lister)
			default:
				return List(lister, columner, rowMapper, detailer, idMapper)
			}
		},
	}
	cmd.Flags().StringP("format", "F", "tui", "Output format, one of 'tui', 'json', 'yaml'")
	return cmd
}
