package jmap

import (
	"fmt"
	"slices"
	"strings"

	"opencloud.eu/groupware-assistant/pkg/tools"
)

type MailboxRights struct {
	MayReadItems   bool `json:"mayReadItems"`
	MayAddItems    bool `json:"mayAddItems"`
	MayRemoveItems bool `json:"mayRemoveItems"`
	MaySetSeen     bool `json:"maySetSeen"`
	MaySetKeywords bool `json:"maySetKeywords"`
	MayCreateChild bool `json:"mayCreateChild"`
	MayRename      bool `json:"mayRename"`
	MayDelete      bool `json:"mayDelete"`
	MaySubmit      bool `json:"maySubmit"`
}

type Mailbox struct {
	Id            string         `json:"id,omitempty"`
	Name          string         `json:"name,omitempty"`
	ParentId      string         `json:"parentId,omitempty"`
	Role          string         `json:"role,omitempty"`
	SortOrder     int            `json:"sortOrder,omitempty"`
	TotalEmails   int            `json:"totalEmails"`
	UnreadEmails  int            `json:"unreadEmails"`
	TotalThreads  int            `json:"totalThreads"`
	UnreadThreads int            `json:"unreadThreads"`
	MyRights      *MailboxRights `json:"myRights,omitempty"`
	IsSubscribed  *bool          `json:"isSubscribed,omitempty"`
}

type NewMailbox struct {
	Name         string `json:"name,omitempty"`
	ParentId     string `json:"parentId,omitempty"`
	Role         string `json:"role,omitempty"`
	SortOrder    int    `json:"sortOrder,omitempty"`
	IsSubscribed *bool  `json:"isSubscribed,omitempty"`
}

func ListMailboxes(j *Jmap, accountId string) ([]Mailbox, error) {
	if accountId, err := j.contactAccountId(accountId); err != nil {
		return nil, err
	} else {
		if list, err := objects[Mailbox](j, accountId, MailboxObjectType, JmapMail); err != nil {
			return nil, err
		} else {
			slices.SortFunc(list, func(a, b Mailbox) int { return strings.Compare(a.Id, b.Id) })
			return list, nil
		}
	}
}

func CreateMailbox(j *Jmap, accountId string, mbox NewMailbox) (string, error) {
	if accountId, err := j.mailAccountId(accountId); err != nil {
		return "", err
	} else {
		if m, err := tools.Remap(mbox); err != nil {
			return "", err
		} else {
			return create(j, MailboxObjectType, createBody(accountId, MailboxObjectType, JmapMail, m))
		}
	}
}

func DeleteMailbox(j *Jmap, accountId string, id string) error {
	if accountId, err := j.mailAccountId(accountId); err != nil {
		return err
	} else {
		return destroy(j, accountId, MailboxObjectType, JmapMail, []string{id})
	}
}

type EmailSender struct {
	j         *Jmap
	accountId string
	mailboxId string
}

func NewEmailSender(j *Jmap, accountId string, mailboxId string, mailboxRole string) (*EmailSender, error) {
	if accountId, err := j.mailAccountId(accountId); err != nil {
		return nil, err
	} else {
		mailboxesById, err := objectsById(j, accountId, "Mailbox", JmapMail)
		if err != nil {
			return nil, err
		}
		if mailboxId != "" {
			if _, ok := mailboxesById[mailboxId]; !ok {
				return nil, fmt.Errorf("mailbox with id '%s' does not exist", mailboxId)
			}
		}
		if mailboxRole != "" {
			if mailboxId == "" {
				for id, mailbox := range mailboxesById {
					role := ""
					if r := mailbox["role"]; r != nil {
						role = r.(string)
					}
					if role == mailboxRole {
						mailboxId = id
						break
					}
				}
				if mailboxId == "" {
					return nil, fmt.Errorf("there is no mailbox with role '%s'", mailboxRole)
				}
			} else {
				mailbox := mailboxesById[mailboxId]
				if mailboxRole != mailbox["role"].(string) {
					return nil, fmt.Errorf("mailbox with id '%s' does not have role '%s' but '%v'", mailboxId, mailboxRole, mailbox["role"])
				}
			}
		}

		return &EmailSender{
			j:         j,
			accountId: accountId,
			mailboxId: mailboxId,
		}, nil
	}
}

func (s *EmailSender) Close() error {
	return nil
}

func (s *EmailSender) NewEmail() (*EmailBuilder, error) {
	return newEmailBuilder(s.accountId, s.mailboxId)
}

func (s *EmailSender) EmptyEmails() (uint, error) {
	return empty(s.j, s.accountId, EmailObjectType, JmapMail, map[string]any{
		"inMailbox": s.mailboxId,
	}, s.destroy)
}

func (s *EmailSender) destroy(ids []string) error {
	return destroy(s.j, s.accountId, EmailObjectType, JmapMail, ids)
}

func (s *EmailSender) SendEmail(e *EmailBuilder) (string, error) {
	bodyValues := map[string]map[string]any{}
	if e.text != "" {
		bodyValues["t"] = map[string]any{"value": e.text}
		e.email["textBody"] = []map[string]any{{
			"partId": "t",
			"type":   "text/plain",
		}}
	}
	if e.html != "" {
		bodyValues["h"] = map[string]any{"value": e.html}
		e.email["htmlBody"] = []map[string]any{{
			"partId": "h",
			"type":   "text/html",
		}}
	}

	attachments := []map[string]any{}
	for _, a := range e.attachments {
		upload, err := s.j.uploadBlob(s.accountId, a.data, a.mime)
		if err != nil {
			return "", err
		}
		ao := map[string]any{
			"blobId":      upload.BlobId,
			"name":        a.filename,
			"type":        a.mime,
			"disposition": "attachment",
		}
		attachments = append(attachments, ao)
	}
	if len(attachments) > 0 {
		e.email["attachments"] = attachments
	}

	if len(bodyValues) > 0 {
		e.email["bodyValues"] = bodyValues
	}

	return create(s.j, EmailObjectType, createBody(s.accountId, EmailObjectType, JmapMail, e.email))
}
