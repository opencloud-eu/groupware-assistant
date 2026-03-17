package jmap

import (
	"encoding/json"
	"fmt"
)

const CalendarObjectType = "Calendar"
const EventObjectType = "CalendarEvent"

type CalendarRights struct {
	MayAdmin         bool `json:"mayAdmin"`
	MayDelete        bool `json:"mayDelete"`
	MayReadFreeBusy  bool `json:"mayReadFreeBusy"`
	MayReadItems     bool `json:"mayReadItems"`
	MayWriteAll      bool `json:"mayWriteAll"`
	MayWriteOwn      bool `json:"mayWriteOwn"`
	MayUpdatePrivate bool `json:"mayUpdatePrivate"`
	MayRSVP          bool `json:"mayRSVP"`
}

type Calendar struct {
	Id           string `json:"id,omitempty"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	IsDefault    bool   `json:"isDefault,omitzero"`
	IsSubscribed bool   `json:"isSubscribed"`
	IsVisible    bool   `json:"isVisible"`
	// IncludeInAvailability string         `json:"includeInAvailablity"`
	TimeZone  string         `json:"timeZone"`
	Color     string         `json:"color"`
	SortOrder int            `json:"sortOrder"`
	MyRights  CalendarRights `json:"myRights"`
}

type NewCalendar struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	IsSubscribed bool   `json:"isSubscribed"`
	IsVisible    bool   `json:"isVisible"`
	TimeZone     string `json:"timeZone,omitempty"`
	Color        string `json:"color,omitempty"`
	SortOrder    int    `json:"sortOrder,omitzero"`
}

func ListCalendars(j *Jmap, accountId string) ([]Calendar, error) {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Calendars
		if accountId == "" {
			return nil, fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return nil, fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}
	return objects[Calendar](j, accountId, CalendarObjectType, JmapCalendars)
}

func CreateCalendar(j *Jmap, accountId string, cal NewCalendar) (string, error) {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Calendars
		if accountId == "" {
			return "", fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return "", fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}
	b, err := json.Marshal(cal)
	if err != nil {
		return "", err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		return "", err
	}

	body := map[string]any{
		"using": []string{JmapCore, JmapCalendars},
		"methodCalls": []any{
			[]any{
				CalendarObjectType + "/set",
				map[string]any{
					"accountId": accountId,
					"create": map[string]any{
						"c": m,
					},
				},
				"0",
			},
		},
	}

	return create(j, "c", CalendarObjectType, body)
}

func DeleteCalendar(j *Jmap, accountId string, id string) error {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Calendars
		if accountId == "" {
			return fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}
	return destroy(j, accountId, CalendarObjectType, JmapCalendars, []string{id})
}

type EventSender struct {
	j          *Jmap
	accountId  string
	calendarId string
}

func (s *EventSender) CalendarId() string {
	return s.calendarId
}

func NewEventSender(j *Jmap, accountId string, calendarId string) (*EventSender, error) {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Calendars
		if accountId == "" {
			return nil, fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return nil, fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}

	calendarsById, err := objectsById(j, accountId, CalendarObjectType, JmapCalendars)
	if err != nil {
		return nil, err
	}
	if calendarId != "" {
		if _, ok := calendarsById[calendarId]; !ok {
			return nil, fmt.Errorf("calendar with id '%s' does not exist", calendarId)
		}
	} else {
		for id, calendar := range calendarsById {
			if isDefault, ok := calendar["isDefault"]; ok {
				if isDefault.(bool) {
					calendarId = id
					break
				}
			}
		}
	}
	if calendarId == "" {
		return nil, fmt.Errorf("failed to find a default Calendar")
	}

	return &EventSender{
		j:          j,
		accountId:  accountId,
		calendarId: calendarId,
	}, nil
}

func (j *EventSender) Close() error {
	return nil
}

func (j *EventSender) EmptyEvents() (uint, error) {
	return empty(j.j, j.accountId, EventObjectType, JmapMail, map[string]any{
		"inCalendar": j.calendarId,
	}, j.destroy)
}

func (j *EventSender) destroy(ids []string) error {
	return destroy(j.j, j.accountId, EventObjectType, JmapCalendars, ids)
}

func (j *EventSender) CreateEvent(c map[string]any) (string, error) {
	body := map[string]any{
		"using": []string{JmapCore, JmapContacts},
		"methodCalls": []any{
			[]any{
				EventObjectType + "/set",
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

	return create(j.j, "c", EventObjectType, body)
}
