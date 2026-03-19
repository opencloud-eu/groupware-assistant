package cmd

import (
	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var tasklistDeleteCmd = deleteCommand("tasklist", jmap.DeleteTasklist)

func init() {
	tasklistCmd.AddCommand(tasklistDeleteCmd)
}
