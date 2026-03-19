package cmd

import (
	"strconv"

	"github.com/charmbracelet/bubbles/table"

	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

type Account struct {
	Id string
	jmap.Account
}

var accountListCmd = listCommand(
	func(j *jmap.Jmap, accountId string) ([]Account, error) {
		s := j.Session()
		result := []Account{}
		for id, acc := range s.Accounts {
			result = append(result, Account{Id: id, Account: acc})
		}
		return result, nil
	},
	func(list []Account) []table.Column {
		mwId := max(len("ID"), tools.MappedLen(list, func(a Account) string { return a.Id }))
		mwName := tools.MappedLen(list, func(a Account) string { return a.Name })
		return []table.Column{
			{Title: "ID", Width: mwId},
			{Title: "Name", Width: mwName},
			{Title: "Personal", Width: len("Personal")},
			{Title: "ReadOnly", Width: len("ReadOnly")},
		}
	},
	func(a Account) table.Row {
		return table.Row{
			a.Id,
			a.Name,
			strconv.FormatBool(a.IsPersonal),
			strconv.FormatBool(a.IsReadOnly),
		}
	},
	nil,
	func(a Account) string {
		return a.Id
	},
)

func init() {
	accountCmd.AddCommand(accountListCmd)
}
