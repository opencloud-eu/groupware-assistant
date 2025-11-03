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

func (b *EmailBuilder) To(to mail.Address) {
	b.email["to"] = []map[string]any{
		{"name": to.Name, "email": to.Address},
	}
}

func (b *EmailBuilder) CC(cc []mail.Address) {
	list := make([]map[string]any, len(cc))
	for i, a := range cc {
		list[i] = map[string]any{"name": a.Name, "email": a.Address}
	}
	b.email["cc"] = list
}

func (b *EmailBuilder) BCC(bcc []mail.Address) {
	list := make([]map[string]any, len(bcc))
	for i, a := range bcc {
		list[i] = map[string]any{"name": a.Name, "email": a.Address}
	}
	b.email["bcc"] = list
}

func (b *EmailBuilder) From(from mail.Address) {
	b.email["from"] = []map[string]any{
		{"name": from.Name, "email": from.Address},
	}
}

func (b *EmailBuilder) Sender(sender mail.Address) {
	b.email["sender"] = []map[string]any{
		{"name": sender.Name, "email": sender.Address},
	}
}

func (b *EmailBuilder) MessageId(id string) {
	b.header("Message-ID", id)
}

func (b *EmailBuilder) InReplyTo(address string) {
	b.email["inReplyTo"] = []string{address}
}

func (b *EmailBuilder) Subject(value string) {
	b.email["subject"] = value
}

func (b *EmailBuilder) header(name string, value string) {
	b.email["header:"+name] = value
}

func (b *EmailBuilder) ReturnPath(returnPath string) {
	b.header("Return-Path", returnPath)
}

func (b *EmailBuilder) Received(t time.Time) {
	b.email["receivedAt"] = t.Format(time.RFC3339)
}

func (b *EmailBuilder) Sent(t time.Time) {
	b.email["sentAt"] = t.Format(time.RFC3339)
}

func (b *EmailBuilder) HTML(text string) {
	b.html = tools.ToHtml(text)
}

func (b *EmailBuilder) Text(text string) {
	b.text = text
}

func (b *EmailBuilder) Attach(content []byte, contentType string, filename string) {
	b.attachments = append(b.attachments, attachment{
		data:     content,
		mime:     contentType,
		filename: filename,
	})
}

func (b *EmailBuilder) AttachInline(content []byte, contentType string, filename string, contentId string) {
	b.attachments = append(b.attachments, attachment{
		name:     contentId,
		data:     content,
		mime:     contentType,
		filename: filename,
	})
}

func (b *EmailBuilder) keyword(k string) {
	keywords, ok := b.email["keywords"].(map[string]bool)
	if !ok {
		keywords = map[string]bool{}
	}
	keywords[k] = true
	b.email["keywords"] = keywords
}

func (b *EmailBuilder) Answered() {
	b.keyword("$answered")
}

func (b *EmailBuilder) Draft() {
	b.keyword("$draft")
}

func (b *EmailBuilder) Important() {
	b.keyword("$flagged")
}

func (b *EmailBuilder) Forwarded() {
	b.keyword("$forwarded")
}

func (b *EmailBuilder) Junk() {
	b.keyword("$junk")
}

func (b *EmailBuilder) NotJunk() {
	b.keyword("$notjunk")
}

func (b *EmailBuilder) Phishing() {
	b.keyword("$phishing")
}

func (b *EmailBuilder) Seen() {
	b.keyword("$seen")
}
