package jmap

import (
	"fmt"
)

type EmailSender struct {
	j         *Jmap
	accountId string
	mailboxId string
}

func NewEmailSender(j *Jmap, accountId string, mailboxId string, mailboxRole string) (*EmailSender, error) {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Mail
		if accountId == "" {
			return nil, fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return nil, fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}

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

func (s *EmailSender) Close() error {
	return nil
}

func (s *EmailSender) NewEmail() (*EmailBuilder, error) {
	return newEmailBuilder(s.accountId, s.mailboxId)
}

func (s *EmailSender) EmptyEmails() (uint, error) {
	return empty(s.j, s.accountId, "Email", JmapMail, map[string]any{
		"inMailbox": s.mailboxId,
	}, s.destroy)
}

func (s *EmailSender) destroy(ids []string) error {
	return destroy(s.j, s.accountId, "Email", JmapMail, ids)
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

	body := map[string]any{
		"using": []string{JmapCore, JmapMail},
		"methodCalls": []any{
			[]any{
				"Email/set",
				map[string]any{
					"accountId": s.accountId,
					"create": map[string]any{
						"c": e.email,
					},
				},
				"0",
			},
		},
	}

	return create(s.j, "c", "Email", body)
}
