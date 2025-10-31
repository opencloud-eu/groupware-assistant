package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"opencloud.eu/groupware-assistant/pkg/generator"
)

var contactGenerateCmd = &cobra.Command{
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
		addressbookId, err := cmd.Flags().GetString("addressbook-id")
		if err != nil {
			return err
		}

		return generator.GenerateContacts(
			JmapUrl,
			Trace,
			Username,
			Password,
			AccountId,
			empty,
			addressbookId,
			count,
			func(text string) { fmt.Println(text) },
		)
	},
}

func init() {
	contactCmd.AddCommand(contactGenerateCmd)

	contactGenerateCmd.Flags().UintP("count", "c", 20, "How many emails to add to the folder")
	contactGenerateCmd.Flags().BoolP("empty", "E", false, "Whether to empty the folder before adding emails to it")
	contactGenerateCmd.Flags().String("addressbook-id", "", "ID of the JMAP AddressBook to use")
}
