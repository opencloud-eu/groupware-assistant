package jmap

import (
	"fmt"
	"slices"
	"strings"

	"opencloud.eu/groupware-assistant/pkg/tools"
)

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
	if accountId, err := j.contactAccountId(accountId); err != nil {
		return nil, err
	} else {
		if list, err := objects[Addressbook](j, accountId, AddressBookObjectType, JmapContacts); err != nil {
			return nil, err
		} else {
			slices.SortFunc(list, func(a, b Addressbook) int { return strings.Compare(a.Id, b.Id) })
			return list, err
		}
	}
}

func CreateAddressbook(j *Jmap, accountId string, abook NewAddressbook) (string, error) {
	if accountId, err := j.contactAccountId(accountId); err != nil {
		return "", err
	} else {
		if m, err := tools.Remap(abook); err != nil {
			return "", err
		} else {
			return create(j, AddressBookObjectType, createBody(accountId, AddressBookObjectType, JmapContacts, m))
		}
	}
}

func DeleteAddressbook(j *Jmap, accountId string, id string) error {
	if accountId, err := j.contactAccountId(accountId); err != nil {
		return err
	} else {
		return destroy(j, accountId, AddressBookObjectType, JmapContacts, []string{id})
	}
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
	if accountId, err := j.contactAccountId(accountId); err != nil {
		return nil, err
	} else {
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
	return create(s.j, ContactCardObjectType, createBody(s.accountId, ContactCardObjectType, JmapContacts, c))
}
