package cmd

import (
	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var mailboxDeleteCmd = deleteCommand("mailbox", jmap.DeleteMailbox)

func init() {
	mailboxCmd.AddCommand(mailboxDeleteCmd)
}
