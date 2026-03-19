package jmap

import (
	"fmt"
	"slices"
	"strings"

	"opencloud.eu/groupware-assistant/pkg/tools"
)

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
	if accountId, err := j.calendarAccountId(accountId); err != nil {
		return nil, err
	} else {
		if list, err := objects[Calendar](j, accountId, CalendarObjectType, JmapCalendars); err != nil {
			return nil, err
		} else {
			slices.SortFunc(list, func(a, b Calendar) int { return strings.Compare(a.Id, b.Id) })
			return list, err
		}
	}
}

func CreateCalendar(j *Jmap, accountId string, cal NewCalendar) (string, error) {
	if accountId, err := j.calendarAccountId(accountId); err != nil {
		return "", err
	} else {
		if m, err := tools.Remap(cal); err != nil {
			return "", err
		} else {
			return create(j, CalendarObjectType, createBody(accountId, CalendarObjectType, JmapCalendars, m))
		}
	}
}

func DeleteCalendar(j *Jmap, accountId string, id string) error {
	if accountId, err := j.calendarAccountId(accountId); err != nil {
		return err
	} else {
		return destroy(j, accountId, CalendarObjectType, JmapCalendars, []string{id})
	}
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
	if accountId, err := j.calendarAccountId(accountId); err != nil {
		return nil, err
	} else {
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
			return nil, fmt.Errorf("failed to find a default %s", CalendarObjectType)
		}

		return &EventSender{
			j:          j,
			accountId:  accountId,
			calendarId: calendarId,
		}, nil
	}
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
	return create(j.j, EventObjectType, createBody(j.accountId, EventObjectType, JmapContacts, c))
}
