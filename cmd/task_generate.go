package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"opencloud.eu/groupware-assistant/pkg/generator"
)

var taskGenerateCmd = &cobra.Command{
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
		tasklistId, err := cmd.Flags().GetString("tasklist-id")
		if err != nil {
			return err
		}

		return generator.GenerateTasks(
			JmapUrl,
			Trace,
			Color,
			Username,
			Password,
			AccountId,
			empty,
			tasklistId,
			count,
			func(text string) { fmt.Println(text) },
		)
	},
}

func init() {
	taskCmd.AddCommand(taskGenerateCmd)

	taskGenerateCmd.Flags().UintP("count", "c", 20, "How many tasks to add to the tasklist")
	taskGenerateCmd.Flags().BoolP("empty", "E", false, "Whether to empty the tasklist before adding tasks to it")
	taskGenerateCmd.Flags().String("tasklist-id", "", "ID of the JMAP TaskList to use")
}
