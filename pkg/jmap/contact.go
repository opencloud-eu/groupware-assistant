package jmap

import (
	"fmt"
)

type ContactSender struct {
	j             *Jmap
	accountId     string
	addressbookId string
}

func (s *ContactSender) AddressBook() string {
	return s.addressbookId
}

func NewContactSender(j *Jmap, accountId string, addressbookId string) (*ContactSender, error) {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Contact
		if accountId == "" {
			return nil, fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return nil, fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}

	addressbooksById, err := objectsById(j, accountId, "AddressBook", JmapContacts)
	if err != nil {
		return nil, err
	}
	if addressbookId != "" {
		if _, ok := addressbooksById[addressbookId]; !ok {
			return nil, fmt.Errorf("addressbook with id '%s' does not exist", addressbookId)
		}
	} else {
		for id, addressbook := range addressbooksById {
			if isDefault, ok := addressbook["isDefault"]; ok {
				if isDefault.(bool) {
					addressbookId = id
					break
				}
			}
		}
	}
	if addressbookId == "" {
		return nil, fmt.Errorf("failed to find a default AddressBook")
	}

	return &ContactSender{
		j:             j,
		accountId:     accountId,
		addressbookId: addressbookId,
	}, nil
}

func (s *ContactSender) Close() error {
	return nil
}

func (s *ContactSender) EmptyContacts() (uint, error) {
	return empty(s.j, s.accountId, "ContactCard", JmapContacts, map[string]any{
		"inAddressBook": s.addressbookId,
	}, s.destroy)
}

func (s *ContactSender) destroy(ids []string) error {
	return destroy(s.j, s.accountId, "ContactCard", JmapContacts, ids)
}

func (s *ContactSender) CreateContact(c map[string]any) (string, error) {
	body := map[string]any{
		"using": []string{JmapCore, JmapContacts},
		"methodCalls": []any{
			[]any{
				"ContactCard/set",
				map[string]any{
					"accountId": s.accountId,
					"create": map[string]any{
						"c": c,
					},
				},
				"0",
			},
		},
	}
	return create(s.j, "c", "ContactCard", body)
}
