package jmap

import (
	"net/mail"
	"time"

	"opencloud.eu/groupware-assistant/pkg/tools"
)

type attachment struct {
	name     string
	mime     string
	data     []byte
	filename string
}

type EmailBuilder struct {
	accountId   string
	mailboxId   string
	email       map[string]any
	html        string
	text        string
	attachments []attachment
}

func newEmailBuilder(accountId string, mailboxId string) (*EmailBuilder, error) {
	return &EmailBuilder{
		accountId: accountId,
		mailboxId: mailboxId,
		email: map[string]any{
			"mailboxIds": map[string]bool{
				mailboxId: true,
			},
		},
		text:        "",
		html:        "",
		attachments: []attachment{},
	}, nil
}

func (j *EmailBuilder) To(to mail.Address) {
	j.email["to"] = []map[string]any{
		{"name": to.Name, "email": to.Address},
	}
}

func (j *EmailBuilder) CC(cc []mail.Address) {
	list := make([]map[string]any, len(cc))
	for i, a := range cc {
		list[i] = map[string]any{"name": a.Name, "email": a.Address}
	}
	j.email["cc"] = list
}

func (j *EmailBuilder) BCC(bcc []mail.Address) {
	list := make([]map[string]any, len(bcc))
	for i, a := range bcc {
		list[i] = map[string]any{"name": a.Name, "email": a.Address}
	}
	j.email["bcc"] = list
}

func (j *EmailBuilder) From(from mail.Address) {
	j.email["from"] = []map[string]any{
		{"name": from.Name, "email": from.Address},
	}
}

func (j *EmailBuilder) Sender(sender mail.Address) {
	j.email["sender"] = []map[string]any{
		{"name": sender.Name, "email": sender.Address},
	}
}

func (j *EmailBuilder) MessageId(id string) {
	j.header("Message-ID", id)
}

func (j *EmailBuilder) InReplyTo(address string) {
	j.email["inReplyTo"] = []string{address}
}

func (j *EmailBuilder) Subject(value string) {
	j.email["subject"] = value
}

func (j *EmailBuilder) header(name string, value string) {
	j.email["header:"+name] = value
}

func (j *EmailBuilder) ReturnPath(returnPath string) {
	j.header("Return-Path", returnPath)
}

func (j *EmailBuilder) Received(t time.Time) {
	j.email["receivedAt"] = t.Format(time.RFC3339)
}

func (j *EmailBuilder) Sent(t time.Time) {
	j.email["sentAt"] = t.Format(time.RFC3339)
}

func (j *EmailBuilder) HTML(text string) {
	j.html = tools.ToHtml(text)
}

func (j *EmailBuilder) Text(text string) {
	j.text = text
}

func (j *EmailBuilder) Attach(content []byte, contentType string, filename string) {
	j.attachments = append(j.attachments, attachment{
		data:     content,
		mime:     contentType,
		filename: filename,
	})
}

func (j *EmailBuilder) AttachInline(content []byte, contentType string, filename string, contentId string) {
	j.attachments = append(j.attachments, attachment{
		name:     contentId,
		data:     content,
		mime:     contentType,
		filename: filename,
	})
}

func (j *EmailBuilder) keyword(k string) {
	keywords, ok := j.email["keywords"].(map[string]bool)
	if !ok {
		keywords = map[string]bool{}
	}
	keywords[k] = true
	j.email["keywords"] = keywords
}

func (j *EmailBuilder) Answered() {
	j.keyword("$answered")
}

func (j *EmailBuilder) Draft() {
	j.keyword("$draft")
}

func (j *EmailBuilder) Important() {
	j.keyword("$flagged")
}

func (j *EmailBuilder) Forwarded() {
	j.keyword("$forwarded")
}

func (j *EmailBuilder) Junk() {
	j.keyword("$junk")
}

func (j *EmailBuilder) NotJunk() {
	j.keyword("$notjunk")
}

func (j *EmailBuilder) Phishing() {
	j.keyword("$phishing")
}

func (j *EmailBuilder) Seen() {
	j.keyword("$seen")
}
