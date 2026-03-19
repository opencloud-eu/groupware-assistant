package jmap

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
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
			// use the addressbook with the lowest id
			ids := slices.Collect(maps.Keys(addressbooksById))
			if len(ids) > 0 {
				slices.Sort(ids)
				addressbookId = ids[0]
			} else {
				return nil, fmt.Errorf("no addressbooks found")
			}
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

func (s *ContactSender) CreateContact(c map[string]any) (string, []string, error) {
	mediaBlobIds := []string{}
	if media, ok := c["media"]; ok && media != nil {
		if media, ok := media.(map[string]map[string]any); ok {
			for _, obj := range media {
				if kind, ok := obj["kind"]; ok && kind == "photo" {
					if uri, ok := obj["uri"]; !ok || uri == "" {
						if mediaType, ok := obj["mediaType"].(string); ok && mediaType != "" {
							dim := tools.PickRandom(100, 128, 200, 256, 384, 512)
							var image []byte = nil
							switch mediaType {
							case "image/jpeg":
								image = gofakeit.ImageJpeg(dim, dim)
							case "image/png":
								image = gofakeit.ImagePng(dim, dim)
							}
							if blob, err := s.j.uploadBlob(s.accountId, image, mediaType); err != nil {
								return "", nil, err
							} else {
								obj["blobId"] = blob.BlobId
								mediaBlobIds = append(mediaBlobIds, blob.BlobId)
							}
						} else {
							panic("contact media has no uri but neither does it have a mediaType")
						}
					}
				}
			}
		} else {
			panic("contact media is not a map")
		}
	}
	id, err := create(s.j, ContactCardObjectType, createBody(s.accountId, ContactCardObjectType, JmapContacts, c))
	return id, mediaBlobIds, err
}
