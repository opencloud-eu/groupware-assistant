package generator

import (
	"fmt"
	"math/rand/v2"
	"net/mail"
	"strconv"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/brianvoe/gofakeit/v7"
	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

func htmlFormat(body string, b *jmap.EmailBuilder) {
	b.HTML(body)
}

func textFormat(body string, b *jmap.EmailBuilder) {
	b.Text(body)
}

func bothFormat(body string, b *jmap.EmailBuilder) {
	htmlFormat(body, b)
	textFormat(body, b)
}

var formats = []func(string, *jmap.EmailBuilder){
	htmlFormat,
	textFormat,
	bothFormat,
}

type Sender struct {
	first  string
	last   string
	from   string
	sender string
}

func (s Sender) ToAddress() mail.Address {
	return mail.Address{
		Name:    s.first + " " + s.last,
		Address: s.from,
	}
}

func (s Sender) ToSender() string {
	return s.sender
}

type SenderGenerator struct {
	senders []Sender
}

func newSenderGenerator(numSenders uint) SenderGenerator {
	if numSenders < 1 {
		panic(fmt.Errorf("requesting less than 1 sender"))
	}
	senders := make([]Sender, numSenders)
	for i := range numSenders {
		person := newPerson()
		senders[i] = Sender{
			first:  person.GivenName,
			last:   person.SurName,
			from:   person.Email,
			sender: fmt.Sprintf("%s <%s>", person.FullName, person.Email),
		}
	}
	return SenderGenerator{
		senders: senders,
	}
}

func (s SenderGenerator) nextSender() (*Sender, error) {
	if len(s.senders) < 1 {
		return nil, fmt.Errorf("failed to determine a sender to use")
	} else {
		return &s.senders[rand.IntN(len(s.senders))], nil
	}
}

func fakeFilename(extension string) string {
	return strings.ReplaceAll(gofakeit.Product().Name, " ", "_") + extension
}

type icalAttendee struct {
	Name  string
	Email string
}

func toIcal(created time.Time, starts time.Time, duration time.Duration, summary string,
	location string, description string, url string, organizer string,
	attendees []icalAttendee, resource string) string {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)
	event := cal.AddEvent(gofakeit.UUID())
	event.SetCreatedTime(created)
	event.SetDtStampTime(created)
	event.SetModifiedAt(created)
	event.SetStartAt(starts)
	event.SetEndAt(starts.Add(duration))
	event.SetSummary(summary)
	if location != "" {
		event.SetLocation(location)
	}
	if description != "" {
		event.SetDescription(description)
	}
	if url != "" {
		event.SetURL(url)
	}
	if organizer != "" {
		event.SetOrganizer(organizer)
	}
	for _, attendee := range attendees {
		event.AddAttendee(attendee.Email, ics.CalendarUserTypeIndividual, ics.ParticipationStatusNeedsAction, ics.ParticipationRoleReqParticipant, ics.WithCN(attendee.Name), ics.WithRSVP(true))
	}
	if resource != "" {
		event.AddAttendee(resource, ics.CalendarUserTypeResource, ics.ParticipationStatusAccepted, ics.ParticipationRoleReqParticipant)
	}
	return cal.Serialize()
}

func createName(person Person) map[string]any {
	name := map[string]any{
		"@type": "Name",
	}
	n := rand.IntN(3)
	// 0: components only
	// 1: components and full
	// 2: full only
	switch n {
	case 0, 1:
		components := []map[string]string{}
		// 0: given, surname
		// 1: surname, given
		ordered := true
		switch rand.IntN(2) {
		case 0:
			components = append(components, map[string]string{"kind": "given", "value": person.GivenName})
			components = append(components, map[string]string{"kind": "surname", "value": person.SurName})
			if person.SurName2 != "" {
				components = append(components, map[string]string{"kind": "surname2", "value": person.SurName2})
			}
			ordered = true
		case 1:
			components = append(components, map[string]string{"kind": "surname", "value": person.SurName})
			if person.SurName2 != "" {
				components = append(components, map[string]string{"kind": "surname2", "value": person.SurName2})
			}
			components = append(components, map[string]string{"kind": "given", "value": person.GivenName})
			ordered = false
		}
		name["components"] = components
		if !ordered {
			switch rand.IntN(2) {
			case 0:
				name["isOrdered"] = false
			}
		} else {
			name["isOrdered"] = true
			name["defaultSeparator"] = " "
		}
	}
	switch n {
	case 1, 2:
		if person.SurName2 != "" {
			name["full"] = fmt.Sprintf("%s %s %s", person.GivenName, person.SurName, person.SurName2)
		} else {
			name["full"] = fmt.Sprintf("%s %s", person.GivenName, person.SurName)
		}
	}
	return name
}

func createNickName(person Person) *map[string]any {
	if person.NickName != "" {
		return &map[string]any{
			"@type":   "Name",
			"name":    person.NickName,
			"context": tools.ToBoolMap(tools.PickRandoms("work", "private")),
		}
	} else {
		return nil
	}
}

func createEmail(person Person, pref int) map[string]any {
	email := person.Email
	label := ""
	switch rand.IntN(3) {
	case 2:
		label = strings.ToLower(person.GivenName)
	}
	r := map[string]any{
		"@type":   "EmailAddress",
		"address": email,
		"context": tools.ToBoolMap(tools.PickRandoms("work", "private")),
		"pref":    pref,
	}
	if label != "" {
		r["label"] = label
	}
	return r
}

func createSecondaryEmail(email string, pref int) map[string]any {
	return map[string]any{
		"@type":   "EmailAddress",
		"address": email,
		"context": tools.ToBoolMap(tools.PickRandoms("work", "private")),
		"pref":    pref,
	}
}

var idFirstLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var idOtherLetters = append(idFirstLetters, []rune("0123456789")...)

func id() string {
	n := 4 + rand.IntN(12-4+1)
	b := make([]rune, n)
	b[0] = idFirstLetters[rand.IntN(len(idFirstLetters))]
	for i := 1; i < n; i++ {
		b[i] = idOtherLetters[rand.IntN(len(idOtherLetters))]
	}
	return string(b)
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
	Coordinates  string          `json:"coordinates"`
	Links        map[string]Link `json:"links"`
}

var Rooms = []Location{
	{
		Type:         "Location",
		Name:         "office-upstairs",
		Description:  "Office meeting room upstairs",
		LocationType: tools.ToBoolMapS("office"),
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
		Coordinates:  "geo:49.4723337,11.1042282",
		Links: map[string]Link{
			id(): {Href: "https://www.workandpepper.de/"},
		},
	},
	{
		Type:         "Location",
		Name:         "Meetingraum Prenzlauer Berg",
		Description:  "This is a Hero Space with great reviews, fast response-time and good quality service",
		LocationType: tools.ToBoolMapS("office", "public"),
		Coordinates:  "geo:52.554222,13.4142387",
		Links: map[string]Link{
			id(): {Href: "https://www.spacebase.com/en/venue/meeting-room-prenzlauer-be-11499/"},
		},
	},
	{
		Type:         "Location",
		Name:         "Meetingraum LIANE 1",
		Description:  "Ecofriendly Bright Urban Jungle",
		LocationType: tools.ToBoolMapS("office", "library"),
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
	person := newPerson()
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
		organizerEmail = person.Email
		organizerId = participantId
	}
	m := map[string]any{
		"@type":       "Participant",
		"name":        person.FullName,
		"email":       person.Email,
		"description": person.JobTitle,
		"sendTo": map[string]string{
			"imip": "mailto:" + person.Email,
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
		"scheduleId":           "mailto:" + person.Email,
	}

	links := map[string]map[string]any{}
	for range rand.IntN(3) {
		links[id()] = map[string]any{
			"@type":       "Link",
			"href":        "https://picsum.photos/id/" + strconv.Itoa(1+rand.IntN(200)) + "/200/300",
			"contentType": "image/jpeg",
			"rel":         "icon",
			"display":     "badge",
			"title":       person.GivenName + "'s Cake Day pick",
		}
	}
	if len(links) > 0 {
		m["links"] = links
	}

	return participantId, person.Email, m
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

func propmap(container map[string]any, name string, min int, max int, generator func(int, string) (map[string]any, error)) error {
	n := min + rand.IntN(max-min+1)
	if n < 1 {
		return nil
	}

	m := make(map[string]map[string]any, n)
	for i := range n {
		id := id()
		item, err := generator(i, id)
		if err != nil {
			return err
		}
		if item != nil {
			m[id] = item
		}
	}
	if len(m) > 0 {
		container[name] = m
	}
	return nil
}

func picsum(w, h int) string {
	return fmt.Sprintf("https://picsum.photos/id/%d/%d/%d", 1+rand.IntN(200), h, w)
}
