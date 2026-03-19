package cmd

import (
	"opencloud.eu/groupware-assistant/pkg/jmap"
)

var calendarDeleteCmd = deleteCommand("calendar", jmap.DeleteCalendar)

func init() {
	calendarCmd.AddCommand(calendarDeleteCmd)
}
