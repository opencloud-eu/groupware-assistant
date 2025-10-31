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

type JmapEmailBuilder struct {
	accountId   string
	mailboxId   string
	email       map[string]any
	html        string
	text        string
	attachments []attachment
}

func newJmapEmailBuilder(accountId string, mailboxId string) (*JmapEmailBuilder, error) {
	return &JmapEmailBuilder{
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

func (j *JmapEmailBuilder) To(to mail.Address) {
	j.email["to"] = []map[string]any{
		{"name": to.Name, "email": to.Address},
	}
}

func (j *JmapEmailBuilder) CC(cc []mail.Address) {
	list := make([]map[string]any, len(cc))
	for i, a := range cc {
		list[i] = map[string]any{"name": a.Name, "email": a.Address}
	}
	j.email["cc"] = list
}

func (j *JmapEmailBuilder) BCC(bcc []mail.Address) {
	list := make([]map[string]any, len(bcc))
	for i, a := range bcc {
		list[i] = map[string]any{"name": a.Name, "email": a.Address}
	}
	j.email["bcc"] = list
}

func (j *JmapEmailBuilder) From(from mail.Address) {
	j.email["from"] = []map[string]any{
		{"name": from.Name, "email": from.Address},
	}
}

func (j *JmapEmailBuilder) Sender(sender mail.Address) {
	j.email["sender"] = []map[string]any{
		{"name": sender.Name, "email": sender.Address},
	}
}

func (j *JmapEmailBuilder) MessageId(id string) {
	j.header("Message-ID", id)
}

func (j *JmapEmailBuilder) InReplyTo(address string) {
	j.email["inReplyTo"] = []string{address}
}

func (j *JmapEmailBuilder) Subject(value string) {
	j.email["subject"] = value
}

func (j *JmapEmailBuilder) header(name string, value string) {
	j.email["header:"+name] = value
}

func (j *JmapEmailBuilder) ReturnPath(returnPath string) {
	j.header("Return-Path", returnPath)
}

func (j *JmapEmailBuilder) Received(t time.Time) {
	j.email["receivedAt"] = t.Format(time.RFC3339)
}

func (j *JmapEmailBuilder) Sent(t time.Time) {
	j.email["sentAt"] = t.Format(time.RFC3339)
}

func (j *JmapEmailBuilder) HTML(text string) {
	j.html = tools.ToHtml(text)
}

func (j *JmapEmailBuilder) Text(text string) {
	j.text = text
}

func (j *JmapEmailBuilder) Attach(content []byte, contentType string, filename string) {
	j.attachments = append(j.attachments, attachment{
		data:     content,
		mime:     contentType,
		filename: filename,
	})
}

func (j *JmapEmailBuilder) AttachInline(content []byte, contentType string, filename string, contentId string) {
	j.attachments = append(j.attachments, attachment{
		name:     contentId,
		data:     content,
		mime:     contentType,
		filename: filename,
	})
}

func (j *JmapEmailBuilder) keyword(k string) {
	keywords, ok := j.email["keywords"].(map[string]bool)
	if !ok {
		keywords = map[string]bool{}
	}
	keywords[k] = true
	j.email["keywords"] = keywords
}

func (j *JmapEmailBuilder) Answered() {
	j.keyword("$answered")
}

func (j *JmapEmailBuilder) Draft() {
	j.keyword("$draft")
}

func (j *JmapEmailBuilder) Important() {
	j.keyword("$flagged")
}

func (j *JmapEmailBuilder) Forwarded() {
	j.keyword("$forwarded")
}

func (j *JmapEmailBuilder) Junk() {
	j.keyword("$junk")
}

func (j *JmapEmailBuilder) NotJunk() {
	j.keyword("$notjunk")
}

func (j *JmapEmailBuilder) Phishing() {
	j.keyword("$phishing")
}

func (j *JmapEmailBuilder) Seen() {
	j.keyword("$seen")
}
