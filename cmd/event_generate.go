package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"opencloud.eu/groupware-assistant/pkg/generator"
)

var eventGenerateCmd = &cobra.Command{
	Use: "generate",
	RunE: func(cmd *cobra.Command, args []string) error {
		count, err := cmd.Flags().GetUint("count")
		if err != nil {
			return err
		}
		empty, err := cmd.Flags().GetBool("empty")
		if err != nil {
			return err
		}
		calendarId, err := cmd.Flags().GetString("calendar-id")
		if err != nil {
			return err
		}

		return generator.GenerateEvents(
			JmapUrl,
			Trace,
			Color,
			Username,
			Password,
			AccountId,
			empty,
			calendarId,
			count,
			func(text string) { fmt.Println(text) },
		)
	},
}

func init() {
	eventCmd.AddCommand(eventGenerateCmd)

	eventGenerateCmd.Flags().UintP("count", "c", 20, "How many emails to add to the folder")
	eventGenerateCmd.Flags().BoolP("empty", "E", false, "Whether to empty the folder before adding emails to it")
	eventGenerateCmd.Flags().String("calendar-id", "", "ID of the JMAP Calendar to use")
}
