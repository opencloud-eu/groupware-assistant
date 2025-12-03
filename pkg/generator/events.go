package generator

import (
	"fmt"
	"math"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

func GenerateEvents(
	jmapUrl string,
	trace bool,
	color bool,
	username string,
	password string,
	accountId string,
	empty bool,
	calendarId string,
	count uint,
	printer func(string),
) error {
	var s *jmap.EventSender = nil
	{
		u, err := url.Parse(jmapUrl)
		if err != nil {
			return err
		}

		j, err := jmap.NewJmap(u, username, password, trace, color)
		if err != nil {
			return err
		}
		defer j.Close()

		s, err = jmap.NewEventSender(j, accountId, calendarId)
		if err != nil {
			return err
		}
	}
	defer s.Close()

	if empty {
		deleted, err := s.EmptyEvents()
		if err != nil {
			return err
		}
		if deleted > 0 {
			printer(fmt.Sprintf("🗑️ deleted %d events", deleted))
		} else {
			printer("ℹ️ did not delete any events, calendar is empty")
		}
	}

	for i := range count {
		linkId := id()
		locationId, location := createLocation()
		virtualLocationId, virtualLocation := createVirtualLocation()
		alertId := id()
		participants, organizerEmail := createParticipants(locationId, virtualLocationId)
		alertOffset := tools.PickRandom("-PT5M", "-PT10M", "-PT15M")
		duration := tools.PickRandom("PT30M", "PT45M", "PT1H", "PT90M")
		tz := tools.PickRandom("Europe/Paris", "Europe/Brussels", "Europe/Berlin")
		daysDiff := rand.IntN(31) - 15
		t := time.Now().Add(time.Duration(daysDiff) * time.Hour * 24)
		h := tools.PickRandom(9, 10, 11, 14, 15, 16, 18)
		m := tools.PickRandom(0, 30)
		t = time.Date(t.Year(), t.Month(), t.Day(), h, m, 0, 0, t.Location())
		start := strings.ReplaceAll(t.Format(time.DateTime), " ", "T")
		title := gofakeit.Sentence()
		description := gofakeit.Paragraph()
		descriptionFormat := tools.PickRandom("text/plain", "text/html")
		if descriptionFormat == "text/html" {
			description = tools.ToHtml(description)
		}
		status := tools.PickRandom("confirmed", "tentative", "cancelled")
		freeBusy := tools.PickRandom("busy", "busy", "busy", "busy", "free")
		privacy := tools.PickRandom("public", "private", "secret")

		event := map[string]any{
			"@type":                  "Event",
			"calendarIds":            tools.ToBoolMap([]string{s.CalendarId()}),
			"isDraft":                false,
			"start":                  start,
			"duration":               duration,
			"status":                 status,
			"uid":                    gofakeit.UUID(),
			"prodId":                 tools.ProductName,
			"title":                  title,
			"description":            description,
			"descriptionContentType": descriptionFormat,
			"links": map[string]map[string]any{
				linkId: {
					"@type":       "Link",
					"href":        picsum(300, 200),
					"rel":         "about",
					"contentType": "image/jpeg",
				},
			},
			"locale":          tools.PickLanguage(),
			"keywords":        keywords(),
			"categories":      categories(),
			"color":           gofakeit.Color(),
			"sequence":        0,
			"showWithoutTime": false,
			"locations": map[string]Location{
				locationId: location,
			},
			"virtualLocations": map[string]VirtualLocation{
				virtualLocationId: virtualLocation,
			},
			"freeBusyStatus": freeBusy,
			"privacy":        privacy,
			"replyTo": map[string]string{
				"imip": "mailto:" + organizerEmail,
			},
			"sentBy":       organizerEmail,
			"participants": participants,
			"alerts": map[string]map[string]any{
				alertId: {
					"@type": "Alert",
					"trigger": map[string]any{
						"@type":      "OffsetTrigger",
						"offset":     alertOffset,
						"relativeTo": "start",
					},
				},
			},
			"timeZone":        tz,
			"mayInviteSelf":   true,
			"mayInviteOthers": true,
			"hideAttendees":   false,
		}

		recurrenceRule := createRecurrenceRule()
		if recurrenceRule != nil {
			event["recurrenceRule"] = recurrenceRule
		}

		uid, err := s.CreateEvent(event)
		if err != nil {
			return err
		}
		printer(fmt.Sprintf("🧑🏻 created %*s/%v uid=%v", int(math.Log10(float64(count))+1), strconv.Itoa(int(i+1)), count, uid))
	}
	return nil
}

func createRecurrenceRule() map[string]any {
	if rand.IntN(10) <= 7 {
		return nil
	}
	frequency := tools.PickRandom("weekly", "daily")
	interval := tools.PickRandom(1, 2)
	count := 1
	if frequency == "weekly" {
		count = 1 + rand.IntN(8)
	} else {
		count = 1 + rand.IntN(4)
	}
	return map[string]any{
		"@type":          "RecurrenceRule",
		"frequency":      frequency,
		"interval":       interval,
		"rscale":         "iso8601",
		"skip":           "omit",
		"firstDayOfWeek": "mo",
		"count":          count,
	}
}
