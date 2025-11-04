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
		alertOffset := tools.PickRandom("PT-5M", "PT-10M", "PT-15M")
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
					"href":        "https://picsum.photos/id/" + strconv.Itoa(1+rand.IntN(200)) + "/200/300",
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

type Link struct {
	Href string
}

type Location struct {
	Type         string          `json:"@type"`
	Id           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	LocationType map[string]bool `json:"locationType"`
	RelativeTo   string          `json:"relativeTo"`
	TimeZone     string          `json:"timeZone"`
	Coordinates  string          `json:"coordinates"`
	Links        map[string]Link `json:"links"`
}

var Rooms = []Location{
	{
		Type:         "Location",
		Name:         "office-upstairs",
		Description:  "Office meeting room upstairs",
		LocationType: tools.ToBoolMapS("office"),
		RelativeTo:   "start",
		TimeZone:     "CET",
		Coordinates:  "geo:52.5335389,13.4103296",
		Links: map[string]Link{
			id(): {Href: "https://www.heinlein-support.de/"},
		},
	},
	{
		Type:         "Location",
		Name:         "office-nue",
		Description:  "",
		LocationType: tools.ToBoolMapS("office"),
		RelativeTo:   "start",
		TimeZone:     "CET",
		Coordinates:  "geo:49.4723337,11.1042282",
		Links: map[string]Link{
			id(): {Href: "https://www.workandpepper.de/"},
		},
	},
	{
		Type:         "Location",
		Name:         "Meetingraum Prenzlauer Berg",
		Description:  "This is a Hero Space with great reviews, fast response-time and good quality service",
		LocationType: tools.ToBoolMapS("office"),
		RelativeTo:   "start",
		TimeZone:     "CET",
		Coordinates:  "geo:52.554222,13.4142387",
		Links: map[string]Link{
			id(): {Href: "https://www.spacebase.com/en/venue/meeting-room-prenzlauer-be-11499/"},
		},
	},
	{
		Type:         "Location",
		Name:         "Meetingraum LIANE 1",
		Description:  "Ecofriendly Bright Urban Jungle",
		LocationType: tools.ToBoolMapS("office"),
		RelativeTo:   "start",
		TimeZone:     "CET",
		Coordinates:  "geo:52.4854301,13.4224763",
		Links: map[string]Link{
			id(): {Href: "https://www.spacebase.com/en/venue/rent-a-jungle-8372/"},
		},
	},
	{
		Type:         "Location",
		Name:         "Dark Horse",
		Description:  "Collaboration and event spaces from the authors of the Workspace and Digital Innovation Playbooks.",
		LocationType: tools.ToBoolMapS("office"),
		RelativeTo:   "start",
		TimeZone:     "CET",
		Coordinates:  "geo:52.4942254,13.4346015",
		Links: map[string]Link{
			id(): {Href: "https://www.spacebase.com/en/event-venue/workshop-white-space-2667/"},
		},
	},
}

type VirtualLocation struct {
	Type        string          `json:"@type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Uri         string          `json:"uri"`
	Features    map[string]bool `json:"features"`
}

var VirtualRooms = []VirtualLocation{
	{
		Type:        "VirtualLocation",
		Name:        "opentalk",
		Description: "the main room in our opentalk instance",
		Uri:         "https://meet.opentalk.eu/fake/room/" + gofakeit.UUID(),
		Features:    tools.ToBoolMapS("audio", "chat", "video", "screen"),
	},
}

func createLocation() (string, Location) {
	locationId := id()
	room := Rooms[rand.IntN(len(Rooms))]
	return locationId, room
}

func createVirtualLocation() (string, VirtualLocation) {
	locationId := id()
	return locationId, VirtualRooms[rand.IntN(len(VirtualRooms))]
}

var ChairRoles = tools.ToBoolMapS("attendee", "chair", "owner")
var RegularRoles = tools.ToBoolMapS("attendee")

func createParticipants(locationId string, virtualLocationid string) (map[string]map[string]any, string) {
	n := 1 + rand.IntN(4)
	participants := map[string]map[string]any{}
	organizerId, organizerEmail, organizer := createParticipant(0, tools.PickRandom(locationId, virtualLocationid), "", "")
	participants[organizerId] = organizer
	for i := 1; i < n; i++ {
		id, _, participant := createParticipant(i, tools.PickRandom(locationId, virtualLocationid), organizerId, organizerEmail)
		participants[id] = participant
	}
	return participants, organizerEmail
}

func createParticipant(i int, locationId string, organizerEmail string, organizerId string) (string, string, map[string]any) {
	participantId := id()
	person := gofakeit.Person()
	roles := RegularRoles
	if i == 0 {
		roles = ChairRoles
	}
	status := "accepted"
	if i != 0 {
		status = tools.PickRandom("needs-action", "accepted", "declined", "tentative") //, delegated + set "delegatedTo"
	}
	statusComment := ""
	if rand.IntN(5) >= 3 {
		statusComment = gofakeit.HipsterSentence()
	}
	if i == 0 {
		organizerEmail = person.Contact.Email
		organizerId = participantId
	}
	m := map[string]any{
		"@type":       "Participant",
		"name":        person.FirstName + " " + person.LastName,
		"email":       person.Contact.Email,
		"description": person.Job.Title,
		"sendTo": map[string]string{
			"imip": "mailto:" + person.Contact.Email,
		},
		"kind":                 "individual",
		"roles":                roles,
		"locationId":           locationId,
		"language":             tools.PickLanguage(),
		"participationStatus":  status,
		"participationComment": statusComment,
		"expectReply":          true,
		"scheduleAgent":        "server",
		"scheduleSequence":     1,
		"scheduleStatus":       []string{"1.0"},
		"scheduleUpdated":      "2025-10-01T1:59:12Z",
		"sentBy":               organizerEmail,
		"invitedBy":            organizerId,
		"scheduleId":           "mailto:" + person.Contact.Email,
	}

	links := map[string]map[string]any{}
	for range rand.IntN(3) {
		links[id()] = map[string]any{
			"@type":       "Link",
			"href":        "https://picsum.photos/id/" + strconv.Itoa(1+rand.IntN(200)) + "/200/300",
			"contentType": "image/jpeg",
			"rel":         "icon",
			"display":     "badge",
			"title":       person.FirstName + "'s Cake Day pick",
		}
	}
	if len(links) > 0 {
		m["links"] = links
	}

	return participantId, person.Contact.Email, m
}

var Keywords = []string{
	"office",
	"important",
	"sales",
	"coordination",
	"decision",
}

func keywords() map[string]bool {
	return tools.ToBoolMap(tools.PickRandoms(Keywords...))
}

var Categories = []string{
	"secret",
	"internal",
}

func categories() map[string]bool {
	return tools.ToBoolMap(tools.PickRandoms(Categories...))
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
