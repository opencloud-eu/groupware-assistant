package jmap

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"
)

type Account struct {
	Name       string `json:"name,omitempty"`
	IsPersonal bool   `json:"isPersonal"`
	IsReadOnly bool   `json:"isReadOnly"`
}

type SessionPrimaryAccounts struct {
	Mail    string `json:"urn:ietf:params:jmap:mail,omitempty"`
	Contact string `json:"urn:ietf:params:jmap:contacts,omitempty"`
}

type Session struct {
	Accounts        map[string]Account     `json:"accounts,omitempty"`
	PrimaryAccounts SessionPrimaryAccounts `json:"primaryAccounts"`
	Username        string                 `json:"username,omitempty"`
	ApiUrl          string                 `json:"apiUrl,omitempty"`
	UploadUrl       string                 `json:"uploadUrl,omitempty"`
}

type Jmap struct {
	h        *http.Client
	username string
	password string
	session  Session
	u        *url.URL
	trace    bool
}

func NewJmap(baseurl *url.URL, username string, password string, trace bool) (*Jmap, error) {
	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	httpTransport.TLSClientConfig = tlsConfig
	h := http.DefaultClient
	h.Transport = httpTransport

	var session Session
	{
		wku := baseurl.JoinPath("/.well-known/jmap")

		req, err := http.NewRequest(http.MethodGet, wku.String(), nil)
		if err != nil {
			return nil, err
		}
		if trace {
			if b, err := httputil.DumpRequestOut(req, true); err == nil {
				log.Printf("==> %s\n", string(b))
			}
		}
		req.SetBasicAuth(username, password)
		resp, err := h.Do(req)
		if err != nil {
			return nil, err
		}
		if trace {
			if b, err := httputil.DumpResponse(resp, true); err == nil {
				log.Printf("<== %s\n", string(b))
			}
		}
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("status is %s", resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &session)
		if err != nil {
			return nil, err
		}
	}

	u, err := url.Parse(session.ApiUrl)
	if err != nil {
		return nil, err
	}

	return &Jmap{
		h:        h,
		trace:    trace,
		username: username,
		password: password,
		session:  session,
		u:        u,
	}, nil
}

func (j *Jmap) Close() error {
	return nil
}

type JmapEmailSender struct {
	j         *Jmap
	accountId string
	mailboxId string
}

func NewJmapEmailSender(j *Jmap, accountId string, mailboxId string, mailboxRole string) (*JmapEmailSender, error) {
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

	mailboxesById := map[string]map[string]any{}
	mailboxesByRole := map[string]string{}
	{
		body := map[string]any{
			"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail", "urn:ietf:params:jmap:contacts"},
			"methodCalls": []any{
				[]any{
					"Mailbox/get",
					map[string]any{
						"accountId": accountId,
					},
					"0",
				},
			},
		}
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, j.u.String(), bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		if j.trace {
			if b, err := httputil.DumpRequestOut(req, true); err == nil {
				log.Printf("==> %s\n", string(b))
			}
		}
		req.SetBasicAuth(j.username, j.password)
		resp, err := j.h.Do(req)
		if err != nil {
			return nil, err
		}
		if j.trace {
			if b, err := httputil.DumpResponse(resp, true); err == nil {
				log.Printf("<== %s\n", string(b))
			}
		}
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("status is %s", resp.Status)
		}
		defer resp.Body.Close()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		r := map[string]any{}
		err = json.Unmarshal(response, &r)
		if err != nil {
			return nil, err
		}
		l := r["methodResponses"].([]any)
		z := l[0].([]any)
		f := z[1].(map[string]any)
		mailboxes := f["list"].([]any)

		for _, a := range mailboxes {
			mailbox := a.(map[string]any)
			id := mailbox["id"].(string)
			role := ""
			if r := mailbox["role"]; r != nil {
				role = r.(string)
			}
			mailboxesById[id] = mailbox
			if role != "" {
				mailboxesByRole[role] = id
			}
		}
	}
	if mailboxId != "" {
		if _, ok := mailboxesById[mailboxId]; !ok {
			return nil, fmt.Errorf("mailbox with id '%s' does not exist", mailboxId)
		}
	}
	if mailboxRole != "" {
		if mailboxId == "" {
			id, ok := mailboxesByRole[mailboxRole]
			if !ok {
				return nil, fmt.Errorf("there is no mailbox with role '%s'", mailboxRole)
			}
			mailboxId = id
		} else {
			mailbox := mailboxesById[mailboxId]
			if mailboxRole != mailbox["role"].(string) {
				return nil, fmt.Errorf("mailbox with id '%s' does not have role '%s' but '%v'", mailboxId, mailboxRole, mailbox["role"])
			}
		}
	}

	return &JmapEmailSender{
		j:         j,
		accountId: accountId,
		mailboxId: mailboxId,
	}, nil
}

func (j *JmapEmailSender) Close() error {
	return nil
}

func (j *JmapEmailSender) NewEmail() (*JmapEmailBuilder, error) {
	return newJmapEmailBuilder(j.accountId, j.mailboxId)
}

func (j *JmapEmailSender) EmptyEmails() (int, error) {
	var ids []string
	{
		body := map[string]any{
			"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"},
			"methodCalls": []any{
				[]any{
					"Email/query",
					map[string]any{
						"accountId":  j.accountId,
						"inMailbox":  map[string]bool{j.mailboxId: true},
						"properties": "id",
					},
					"0",
				},
			},
		}
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		req, err := http.NewRequest(http.MethodPost, j.j.u.String(), bytes.NewReader(payload))
		if err != nil {
			return 0, err
		}
		if j.j.trace {
			if b, err := httputil.DumpRequestOut(req, true); err == nil {
				log.Printf("==> %s\n", string(b))
			}
		}
		req.SetBasicAuth(j.j.username, j.j.password)
		resp, err := j.j.h.Do(req)
		if err != nil {
			return 0, err
		}
		if j.j.trace {
			if b, err := httputil.DumpResponse(resp, true); err == nil {
				log.Printf("<== %s\n", string(b))
			}
		}
		if resp.StatusCode >= 300 {
			return 0, fmt.Errorf("status is %s", resp.Status)
		}
		defer resp.Body.Close()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		r := map[string]any{}
		err = json.Unmarshal(response, &r)
		if err != nil {
			return 0, err
		}
		l := r["methodResponses"].([]any)
		z := l[0].([]any)
		f := z[1].(map[string]any)
		v := f["ids"].([]any)
		ids = make([]string, len(v))
		for i, a := range v {
			ids[i] = a.(string)
		}
	}
	for chunk := range slices.Chunk(ids, 20) {
		body := map[string]any{
			"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"},
			"methodCalls": []any{
				[]any{
					"Email/set",
					map[string]any{
						"accountId": j.accountId,
						"destroy":   chunk,
					},
					"0",
				},
			},
		}
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		req, err := http.NewRequest(http.MethodPost, j.j.u.String(), bytes.NewReader(payload))
		if err != nil {
			return 0, err
		}
		if j.j.trace {
			if b, err := httputil.DumpRequestOut(req, true); err == nil {
				log.Printf("==> %s\n", string(b))
			}
		}
		req.SetBasicAuth(j.j.username, j.j.password)
		resp, err := j.j.h.Do(req)
		if err != nil {
			return 0, err
		}
		if j.j.trace {
			if b, err := httputil.DumpResponse(resp, true); err == nil {
				log.Printf("<== %s\n", string(b))
			}
		}
		if resp.StatusCode >= 300 {
			return 0, fmt.Errorf("status is %s", resp.Status)
		}
		defer resp.Body.Close()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		r := map[string]any{}
		err = json.Unmarshal(response, &r)
		if err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

type uploadedBlob struct {
	BlobId string `json:"blobId"`
	Size   int    `json:"size"`
	Type   string `json:"type"`
	Sha512 string `json:"sha:512"`
}

func (j *Jmap) uploadBlob(accountId string, data []byte, mimetype string) (uploadedBlob, error) {
	uploadUrl := strings.ReplaceAll(j.session.UploadUrl, "{accountId}", accountId)
	req, err := http.NewRequest(http.MethodPost, uploadUrl, bytes.NewReader(data))
	if err != nil {
		return uploadedBlob{}, err
	}
	req.Header.Add("Content-Type", mimetype)
	req.SetBasicAuth(j.username, j.password)
	res, err := j.h.Do(req)
	if err != nil {
		return uploadedBlob{}, err
	}
	if j.trace {
		if b, err := httputil.DumpResponse(res, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return uploadedBlob{}, fmt.Errorf("status is %s", res.Status)
	}
	defer res.Body.Close()
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return uploadedBlob{}, err
	}

	var result uploadedBlob
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return uploadedBlob{}, err
	}

	return result, nil
}

func (j *JmapEmailSender) SendEmail(e *JmapEmailBuilder) (string, error) {
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
		upload, err := j.j.uploadBlob(j.accountId, a.data, a.mime)
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
		"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"},
		"methodCalls": []any{
			[]any{
				"Email/set",
				map[string]any{
					"accountId": j.accountId,
					"create": map[string]any{
						"c": e.email,
					},
				},
				"0",
			},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, j.j.u.String(), bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	if j.j.trace {
		if b, err := httputil.DumpRequestOut(req, true); err == nil {
			log.Printf("==> %s\n", string(b))
		}
	}

	req.SetBasicAuth(j.j.username, j.j.password)
	resp, err := j.j.h.Do(req)
	if err != nil {
		return "", err
	}
	if j.j.trace {
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("status is %s", resp.Status)
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	r := map[string]any{}
	err = json.Unmarshal(response, &r)
	if err != nil {
		return "", err
	}

	if j.j.trace {
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}

	l := r["methodResponses"].([]any)
	z := l[0].([]any)
	f := z[1].(map[string]any)
	if x, ok := f["created"]; ok {
		created := x.(map[string]any)
		if c, ok := created["c"].(map[string]any); ok {
			return c["id"].(string), nil
		} else {
			fmt.Println(f)
			return "", fmt.Errorf("failed to create email")
		}
	} else {
		if ncx, ok := f["notCreated"]; ok {
			nc := ncx.(map[string]any)
			c := nc["c"].(map[string]any)
			return "", fmt.Errorf("failed to create email: %v", c["description"])
		} else {
			fmt.Println(f)
			return "", fmt.Errorf("failed to create email")
		}
	}
}

type JmapContactSender struct {
	j             *Jmap
	accountId     string
	addressbookId string
}

func (s *JmapContactSender) AddressBook() string {
	return s.addressbookId
}

func NewJmapContactSender(j *Jmap, accountId string, addressbookId string) (*JmapContactSender, error) {
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

	addressbooksById := map[string]map[string]any{}
	{
		body := map[string]any{
			"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:contacts"},
			"methodCalls": []any{
				[]any{
					"AddressBook/get",
					map[string]any{
						"accountId": accountId,
					},
					"0",
				},
			},
		}
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, j.u.String(), bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		if j.trace {
			if b, err := httputil.DumpRequestOut(req, true); err == nil {
				log.Printf("==> %s\n", string(b))
			}
		}
		req.SetBasicAuth(j.username, j.password)
		resp, err := j.h.Do(req)
		if err != nil {
			return nil, err
		}
		if j.trace {
			if b, err := httputil.DumpResponse(resp, true); err == nil {
				log.Printf("<== %s\n", string(b))
			}
		}
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("status is %s", resp.Status)
		}
		defer resp.Body.Close()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		r := map[string]any{}
		err = json.Unmarshal(response, &r)
		if err != nil {
			return nil, err
		}
		l := r["methodResponses"].([]any)
		z := l[0].([]any)
		f := z[1].(map[string]any)
		addressbooks := f["list"].([]any)

		for _, a := range addressbooks {
			addressbook := a.(map[string]any)
			id := addressbook["id"].(string)
			addressbooksById[id] = addressbook
		}
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

	return &JmapContactSender{
		j:             j,
		accountId:     accountId,
		addressbookId: addressbookId,
	}, nil
}

func (j *JmapContactSender) Close() error {
	return nil
}

func (j *JmapContactSender) EmptyContacts() (uint, error) {
	body := map[string]any{
		"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:contacts"},
		"methodCalls": []any{
			[]any{
				"ContactCard/query",
				map[string]any{
					"accountId": j.accountId,
					"filter": map[string]any{
						"inAddressBook": j.addressbookId,
					},
				},
				"0",
			},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return uint(0), err
	}
	req, err := http.NewRequest(http.MethodPost, j.j.u.String(), bytes.NewReader(payload))
	if err != nil {
		return uint(0), err
	}

	if j.j.trace {
		if b, err := httputil.DumpRequestOut(req, true); err == nil {
			log.Printf("==> %s\n", string(b))
		}
	}

	req.SetBasicAuth(j.j.username, j.j.password)
	resp, err := j.j.h.Do(req)
	if err != nil {
		return uint(0), err
	}
	if j.j.trace {
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}
	if resp.StatusCode >= 300 {
		return uint(0), fmt.Errorf("status is %s", resp.Status)
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return uint(0), err
	}

	r := map[string]any{}
	err = json.Unmarshal(response, &r)
	if err != nil {
		return uint(0), err
	}

	if j.j.trace {
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}

	l := r["methodResponses"].([]any)
	z := l[0].([]any)
	f := z[1].(map[string]any)
	if idsObj, ok := f["ids"]; ok {
		anies := idsObj.([]any)
		ids := make([]string, len(anies))
		for i, a := range anies {
			ids[i] = a.(string)
		}
		destroyed := uint(0)
		for chunk := range slices.Chunk(ids, 20) {
			err = j.destroy(chunk)
			if err != nil {
				return destroyed, err
			}
			destroyed += uint(len(chunk))
		}
		return destroyed, nil
	} else {
		fmt.Println(f)
		return uint(0), fmt.Errorf("failed to destroy ContactCards")
	}
}

func (j *JmapContactSender) destroy(ids []string) error {
	body := map[string]any{
		"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:contacts"},
		"methodCalls": []any{
			[]any{
				"ContactCard/set",
				map[string]any{
					"accountId": j.accountId,
					"destroy":   ids,
				},
				"0",
			},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, j.j.u.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	if j.j.trace {
		if b, err := httputil.DumpRequestOut(req, true); err == nil {
			log.Printf("==> %s\n", string(b))
		}
	}

	req.SetBasicAuth(j.j.username, j.j.password)
	resp, err := j.j.h.Do(req)
	if err != nil {
		return err
	}
	if j.j.trace {
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status is %s", resp.Status)
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	r := map[string]any{}
	err = json.Unmarshal(response, &r)
	if err != nil {
		return err
	}

	if j.j.trace {
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}

	l := r["methodResponses"].([]any)
	z := l[0].([]any)
	f := z[1].(map[string]any)
	if x, ok := f["destroyed"]; ok {
		destroyed := x.([]any)
		if len(destroyed) == len(ids) {
			return nil
		} else {
			fmt.Println(f)
			return fmt.Errorf("failed to destroy ContactCards")
		}
	} else {
		if ncx, ok := f["notDestroyed"]; ok {
			nc := ncx.(map[string]any)
			c := nc["c"].(map[string]any)
			return fmt.Errorf("failed to destroy ContactCards: %v", c["description"])
		} else {
			fmt.Println(f)
			return fmt.Errorf("failed to destroy ContactCards")
		}
	}
}

func (j *JmapContactSender) CreateContact(c map[string]any) (string, error) {
	body := map[string]any{
		"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:contacts"},
		"methodCalls": []any{
			[]any{
				"ContactCard/set",
				map[string]any{
					"accountId": j.accountId,
					"create": map[string]any{
						"c": c,
					},
				},
				"0",
			},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, j.j.u.String(), bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	if j.j.trace {
		if b, err := httputil.DumpRequestOut(req, true); err == nil {
			log.Printf("==> %s\n", string(b))
		}
	}

	req.SetBasicAuth(j.j.username, j.j.password)
	resp, err := j.j.h.Do(req)
	if err != nil {
		return "", err
	}
	if j.j.trace {
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("status is %s", resp.Status)
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	r := map[string]any{}
	err = json.Unmarshal(response, &r)
	if err != nil {
		return "", err
	}

	if j.j.trace {
		if b, err := httputil.DumpResponse(resp, true); err == nil {
			log.Printf("<== %s\n", string(b))
		}
	}

	l := r["methodResponses"].([]any)
	z := l[0].([]any)
	f := z[1].(map[string]any)
	if x, ok := f["created"]; ok {
		created := x.(map[string]any)
		if c, ok := created["c"].(map[string]any); ok {
			return c["id"].(string), nil
		} else {
			fmt.Println(f)
			return "", fmt.Errorf("failed to create ContactCard")
		}
	} else {
		if ncx, ok := f["notCreated"]; ok {
			nc := ncx.(map[string]any)
			c := nc["c"].(map[string]any)
			return "", fmt.Errorf("failed to create ContactCard: %v", c["description"])
		} else {
			fmt.Println(f)
			return "", fmt.Errorf("failed to create ContactCard")
		}
	}
}
