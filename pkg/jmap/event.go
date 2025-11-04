package jmap

import (
	"fmt"
)

const CalendarObjectType = "Calendar"
const EventObjectType = "CalendarEvent"

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
