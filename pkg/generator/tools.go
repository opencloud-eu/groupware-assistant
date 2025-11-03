package generator

import (
	"fmt"
	"math/rand"
	"net/mail"
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
	senders := make([]Sender, numSenders)
	for i := range numSenders {
		person := gofakeit.Person()
		senders[i] = Sender{
			first:  person.FirstName,
			last:   person.LastName,
			from:   person.Contact.Email,
			sender: person.FirstName + " " + person.LastName + "<" + person.Contact.Email + ">",
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
		return &s.senders[rand.Intn(len(s.senders))], nil
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

func createName(person *gofakeit.PersonInfo) map[string]any {
	name := map[string]any{
		"@type": "Name",
	}
	components := make([]map[string]string, 2)
	components[0] = map[string]string{
		"kind":  "given",
		"value": person.FirstName,
	}
	components[1] = map[string]string{
		"kind":  "surname",
		"value": person.LastName,
	}
	name["components"] = components
	name["isOrdered"] = true
	name["defaultSeparator"] = " "
	name["full"] = fmt.Sprintf("%s %s", person.FirstName, person.LastName)
	return name
}

func createNickName(_ *gofakeit.PersonInfo) map[string]any {
	return map[string]any{
		"@type":   "Name",
		"name":    gofakeit.PetName(),
		"context": tools.ToBoolMap(tools.PickRandom([]string{"work"}, []string{"work", "private"})),
	}
}

func createEmail(person *gofakeit.PersonInfo, pref int) map[string]any {
	email := person.Contact.Email
	return map[string]any{
		"@type":   "EmailAddress",
		"address": email,
		"context": tools.ToBoolMap(tools.PickRandom([]string{"work"}, []string{"work", "private"})),
		"label":   strings.ToLower(person.FirstName),
		"pref":    pref,
	}
}

func createSecondaryEmail(email string, pref int) map[string]any {
	return map[string]any{
		"@type":   "EmailAddress",
		"address": email,
		"context": tools.ToBoolMap(tools.PickRandom([]string{"work"}, []string{"work", "private"})),
		"pref":    pref,
	}
}

var idFirstLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var idOtherLetters = append(idFirstLetters, []rune("0123456789")...)

func id() string {
	n := 4 + rand.Intn(12-4+1)
	b := make([]rune, n)
	b[0] = idFirstLetters[rand.Intn(len(idFirstLetters))]
	for i := 1; i < n; i++ {
		b[i] = idOtherLetters[rand.Intn(len(idOtherLetters))]
	}
	return string(b)
}
