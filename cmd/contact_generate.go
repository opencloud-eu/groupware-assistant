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
		includeDataUri, err := cmd.Flags().GetBool("data-uri")
		if err != nil {
			return err
		}
		mediaBlobs, err := cmd.Flags().GetBool("media-blobs")
		if err != nil {
			return err
		}
		addressbookId, err := cmd.Flags().GetString("addressbook-id")
		if err != nil {
			return err
		}

		if mediaBlobs {
			return fmt.Errorf("media blobs are currently unsupported by Stalwart")
		}

		return generator.GenerateContacts(
			JmapUrl,
			Trace,
			Color,
			Username,
			Password,
			AccountId,
			empty,
			addressbookId,
			count,
			includeDataUri,
			mediaBlobs,
			func(text string) { fmt.Println(text) },
		)
	},
}

func init() {
	contactCmd.AddCommand(contactGenerateCmd)

	contactGenerateCmd.Flags().UintP("count", "c", 20, "How many contacts to add to the address book")
	contactGenerateCmd.Flags().BoolP("empty", "E", false, "Whether to empty the address book before adding contacts to it")
	contactGenerateCmd.Flags().String("addressbook-id", "", "ID of the JMAP AddressBook to use, autodiscovers when omitted")
	contactGenerateCmd.Flags().Bool("data-uri", true, "Include media photos that make use of data: URIs")
	contactGenerateCmd.Flags().Bool("media-blobs", false, "Include media photos that are uploaded as blobs")
}
