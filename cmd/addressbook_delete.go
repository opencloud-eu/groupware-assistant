package cmd

import (
	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var addressbookDeleteCmd = deleteCommand("address book", jmap.DeleteAddressbook)

func init() {
	addressbookCmd.AddCommand(addressbookDeleteCmd)
}
