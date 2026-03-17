package jmap

import (
	"encoding/json"
	"fmt"
)

var AddressBookObjectType = "AddressBook"
var ContactCardObjectType = "ContactCard"

type AddressbookRights struct {
	MayAdmin  bool `json:"mayAdmin"`
	MayDelete bool `json:"mayDelete"`
	MayRead   bool `json:"mayRead"`
	MayWrite  bool `json:"mayWrite"`
}

type Addressbook struct {
	Id           string            `json:"id,omitempty"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	IsDefault    bool              `json:"isDefault,omitzero"`
	IsSubscribed bool              `json:"isSubscribed"`
	SortOrder    int               `json:"sortOrder,omitzero"`
	MyRights     AddressbookRights `json:"myRights"`
}

type NewAddressbook struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	IsSubscribed bool   `json:"isSubscribed"`
	SortOrder    int    `json:"sortOrder,omitzero"`
}

func ListAddressbooks(j *Jmap, accountId string) ([]Addressbook, error) {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Contacts
		if accountId == "" {
			return nil, fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return nil, fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}
	return objects[Addressbook](j, accountId, AddressBookObjectType, JmapContacts)
}

func CreateAddressbook(j *Jmap, accountId string, abook NewAddressbook) (string, error) {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Contacts
		if accountId == "" {
			return "", fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return "", fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}
	b, err := json.Marshal(abook)
	if err != nil {
		return "", err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		return "", err
	}

	body := map[string]any{
		"using": []string{JmapCore, JmapContacts},
		"methodCalls": []any{
			[]any{
				AddressBookObjectType + "/set",
				map[string]any{
					"accountId": accountId,
					"create": map[string]any{
						"c": m,
					},
				},
				"0",
			},
		},
	}

	return create(j, "c", AddressBookObjectType, body)
}

func DeleteAddressbook(j *Jmap, accountId string, id string) error {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Contacts
		if accountId == "" {
			return fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}
	return destroy(j, accountId, AddressBookObjectType, JmapContacts, []string{id})
}

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
		accountId = j.session.PrimaryAccounts.Contacts
		if accountId == "" {
			return nil, fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return nil, fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}

	addressbooksById, err := objectsById(j, accountId, AddressBookObjectType, JmapContacts)
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
	return empty(s.j, s.accountId, ContactCardObjectType, JmapContacts, map[string]any{
		"inAddressBook": s.addressbookId,
	}, s.destroy)
}

func (s *ContactSender) destroy(ids []string) error {
	return destroy(s.j, s.accountId, ContactCardObjectType, JmapContacts, ids)
}

func (s *ContactSender) CreateContact(c map[string]any) (string, error) {
	body := map[string]any{
		"using": []string{JmapCore, JmapContacts},
		"methodCalls": []any{
			[]any{
				ContactCardObjectType + "/set",
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
	return create(s.j, "c", ContactCardObjectType, body)
}
