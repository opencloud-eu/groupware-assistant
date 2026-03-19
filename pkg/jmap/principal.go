package jmap

import (
	"slices"
	"strings"
)

type Principal struct {
	Id string `json:"id"`

	// `Principal` type.
	//
	// This MUST be one of the following values:
	// * `individual`: This represents a single person.
	// * `group`: This represents a group of people.
	// * `resource`: This represents some resource, e.g. a projector.
	// * `location`: This represents a location.
	// * `other`: This represents some other undefined principal.
	Type string `json:"type"`

	// The name of the principal, e.g. `"Jane Doe"`, or `"Room 4B"`.
	Name string `json:"name"`

	// A longer description of the principal, for example details about the
	// facilities of a resource, or null if no description available.
	Description string `json:"description,omitempty"`

	// An email address for the principal, or null if no email is available.
	Email string `json:"email,omitempty"`

	// The time zone for this principal, if known.
	//
	// If not null, the value MUST be a time zone id from the IANA Time Zone Database TZDB.
	TimeZone string `json:"timeZone,omitempty"`

	// A map of JMAP capability URIs to domain specific information about the principal in relation
	// to that capability, as defined in the document that registered the capability.
	Capabilities map[string]any `json:"capabilities,omitempty"`

	// A map of account id to `Account` object for each JMAP Account containing data for
	// this principal that the user has access to, or null if none.
	Accounts map[string]Account `json:"accounts,omitempty"`
}

func ListPrincipals(j *Jmap, accountId string) ([]Principal, error) {
	if accountId, err := j.mailAccountId(accountId); err != nil {
		return nil, err
	} else {
		if list, err := objects[Principal](j, accountId, PrincipalObjectType, JmapMail); err != nil {
			return nil, err
		} else {
			slices.SortFunc(list, func(a, b Principal) int { return strings.Compare(a.Id, b.Id) })
			return list, err
		}
	}
}
