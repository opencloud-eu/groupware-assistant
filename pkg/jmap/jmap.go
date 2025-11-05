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

	"github.com/tidwall/pretty"
)

const (
	JmapCore      = "urn:ietf:params:jmap:core"
	JmapMail      = "urn:ietf:params:jmap:mail"
	JmapContacts  = "urn:ietf:params:jmap:contacts"
	JmapCalendars = "urn:ietf:params:jmap:calendars"
	JmapTasks     = "urn:ietf:params:jmap:tasks"

	EmailDeletionChunkSize = 20
)

type Account struct {
	Name       string `json:"name,omitempty"`
	IsPersonal bool   `json:"isPersonal"`
	IsReadOnly bool   `json:"isReadOnly"`
}

type SessionPrimaryAccounts struct {
	Mail      string `json:"urn:ietf:params:jmap:mail,omitempty"`
	Contact   string `json:"urn:ietf:params:jmap:contacts,omitempty"`
	Calendars string `json:"urn:ietf:params:jmap:calendars,omitempty"`
	Tasks     string `json:"urn:ietf:params:jmap:tasks,omitempty"`
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
	color    bool
}

func NewJmap(baseurl *url.URL, username string, password string, trace bool, color bool) (*Jmap, error) {
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
		defer resp.Body.Close()
		var response []byte = nil
		if trace {
			if b, err := httputil.DumpResponse(resp, false); err == nil {
				response, err = io.ReadAll(resp.Body)
				if err != nil {
					return nil, err
				}
				p := pretty.Pretty(response)
				if color {
					p = pretty.Color(p, nil)
				}
				log.Printf("<== %s%s\n", b, p)
			}
		}
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("JMAP command HTTP response status is %s", resp.Status)
		}
		if response == nil {
			response, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(response, &session)
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
	defer res.Body.Close()
	var response []byte = nil
	if j.trace {
		if b, err := httputil.DumpResponse(res, false); err == nil {
			response, err = io.ReadAll(res.Body)
			if err != nil {
				return uploadedBlob{}, err
			}
			p := pretty.Pretty(response)
			if j.color {
				p = pretty.Color(p, nil)
			}
			log.Printf("<== %s%s\n", b, p)
		}
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return uploadedBlob{}, fmt.Errorf("status is %s", res.Status)
	}
	if response == nil {
		response, err = io.ReadAll(res.Body)
		if err != nil {
			return uploadedBlob{}, err
		}
	}

	var result uploadedBlob
	err = json.Unmarshal(response, &result)
	if err != nil {
		return uploadedBlob{}, err
	}

	return result, nil
}

func command[T any](j *Jmap, body map[string]any, closure func([]any) (T, error)) (T, error) {
	var zero T

	payload, err := json.Marshal(body)
	if err != nil {
		return zero, err
	}
	req, err := http.NewRequest(http.MethodPost, j.u.String(), bytes.NewReader(payload))
	if err != nil {
		return zero, err
	}

	if j.trace {
		if b, err := httputil.DumpRequestOut(req, false); err == nil {
			p := pretty.Pretty(payload)
			if j.color {
				p = pretty.Color(p, nil)
			}
			log.Printf("==> %s%s\n", b, p)
		}
	}

	req.SetBasicAuth(j.username, j.password)
	resp, err := j.h.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()
	var response []byte = nil
	if j.trace {
		if b, err := httputil.DumpResponse(resp, false); err == nil {
			response, err = io.ReadAll(resp.Body)
			if err != nil {
				return zero, err
			}
			p := pretty.Pretty(payload)
			if j.color {
				p = pretty.Color(p, nil)
			}
			log.Printf("<== %s%s\n", b, p)
		}
	}
	if resp.StatusCode >= 300 {
		return zero, fmt.Errorf("JMAP command HTTP response status is %s", resp.Status)
	}
	if response == nil {
		response, err = io.ReadAll(resp.Body)
		if err != nil {
			return zero, err
		}
	}

	r := map[string]any{}
	err = json.Unmarshal(response, &r)
	if err != nil {
		return zero, err
	}

	methodResponses := r["methodResponses"].([]any)
	return closure(methodResponses)
}

func create(j *Jmap, id string, objectType string, body map[string]any) (string, error) {
	return command(j, body, func(methodResponses []any) (string, error) {
		z := methodResponses[0].([]any)
		f := z[1].(map[string]any)
		if x, ok := f["created"]; ok {
			created := x.(map[string]any)
			if c, ok := created[id].(map[string]any); ok {
				return c["id"].(string), nil
			} else {
				fmt.Println(f)
				return "", fmt.Errorf("failed to create %v", objectType)
			}
		} else {
			if ncx, ok := f["notCreated"]; ok {
				nc := ncx.(map[string]any)
				c := nc[id].(map[string]any)
				return "", fmt.Errorf("failed to create %v: %v", objectType, c["description"])
			} else {
				fmt.Println(f)
				return "", fmt.Errorf("failed to create %v", objectType)
			}
		}
	})
}

func destroy(j *Jmap, accountId string, objectType string, scope string, ids []string) error {
	body := map[string]any{
		"using": []string{JmapCore, scope},
		"methodCalls": []any{
			[]any{
				objectType + "/set",
				map[string]any{
					"accountId": accountId,
					"destroy":   ids,
				},
				"0",
			},
		},
	}

	f, err := command(j, body, func(methodResponses []any) (map[string]any, error) {
		z := methodResponses[0].([]any)
		return z[1].(map[string]any), nil
	})
	if err != nil {
		return err
	}

	if x, ok := f["destroyed"]; ok {
		destroyed := x.([]any)
		if len(destroyed) == len(ids) {
			return nil
		} else {
			return fmt.Errorf("failed to destroy %ss: %v", objectType, f)
		}
	} else {
		if ncx, ok := f["notDestroyed"]; ok {
			nc := ncx.(map[string]any)
			for id, setErrorObj := range nc {
				setError := setErrorObj.(map[string]any)
				if description, ok := setError["description"]; ok {
					return fmt.Errorf("failed to destroy %ss: %s: %v", objectType, id, description)
				}
			}
			keys := make([]string, len(nc))
			i := 0
			for k := range nc {
				keys[i] = k
				i++
			}
			return fmt.Errorf("failed to destroy %ss: [%s]", objectType, strings.Join(keys, ", "))
		} else {
			return fmt.Errorf("failed to destroy %ss: %v", objectType, f)
		}
	}
}

func empty(j *Jmap, accountId string, objectType string, scope string, filter map[string]any, destroyer func([]string) error) (uint, error) {
	body := map[string]any{
		"using": []string{JmapCore, scope},
		"methodCalls": []any{
			[]any{
				objectType + "/query",
				map[string]any{
					"accountId": accountId,
					"filter":    filter,
				},
				"0",
			},
		},
	}

	f, err := command(j, body, func(methodResponses []any) (map[string]any, error) {
		z := methodResponses[0].([]any)
		return z[1].(map[string]any), nil
	})
	if idsObj, ok := f["ids"]; ok {
		anies := idsObj.([]any)
		ids := make([]string, len(anies))
		for i, a := range anies {
			ids[i] = a.(string)
		}
		destroyed := uint(0)
		for chunk := range slices.Chunk(ids, 20) {
			err = destroyer(chunk)
			if err != nil {
				return destroyed, err
			}
			destroyed += uint(len(chunk))
		}
		return destroyed, nil
	} else {
		return uint(0), fmt.Errorf("failed to destroy %vs: %v", objectType, f)
	}
}

func objectsById(j *Jmap, accountId string, objectType string, scope string) (map[string]map[string]any, error) {
	m := map[string]map[string]any{}
	{
		body := map[string]any{
			"using": []string{JmapCore, scope},
			"methodCalls": []any{
				[]any{
					objectType + "/get",
					map[string]any{
						"accountId": accountId,
					},
					"0",
				},
			},
		}
		result, err := command(j, body, func(methodResponses []any) ([]any, error) {
			z := methodResponses[0].([]any)
			f := z[1].(map[string]any)
			if list, ok := f["list"]; ok {
				return list.([]any), nil
			} else {
				return nil, fmt.Errorf("methodResponse[1] has no 'list' attribute: %v", f)
			}
		})
		if err != nil {
			return nil, err
		}
		for _, a := range result {
			obj := a.(map[string]any)
			id := obj["id"].(string)
			m[id] = obj
		}
	}
	return m, nil
}
