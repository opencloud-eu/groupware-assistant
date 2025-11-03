package generator

import (
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"

	"github.com/brianvoe/gofakeit/v7"
	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

func GenerateContacts(
	jmapUrl string,
	trace bool,
	username string,
	password string,
	accountId string,
	empty bool,
	addressbookId string,
	count uint,
	printer func(string),
) error {
	var s *jmap.JmapContactSender = nil
	{
		u, err := url.Parse(jmapUrl)
		if err != nil {
			return err
		}

		j, err := jmap.NewJmap(u, username, password, trace)
		if err != nil {
			return err
		}
		defer j.Close()

		s, err = jmap.NewJmapContactSender(j, accountId, addressbookId)
		if err != nil {
			return err
		}
	}
	defer s.Close()

	if empty {
		deleted, err := s.EmptyContacts()
		if err != nil {
			return err
		}
		if deleted > 0 {
			printer(fmt.Sprintf("🗑️ deleted %d contacts", deleted))
		} else {
			printer("ℹ️ did not delete any contacts, addressbook is empty")
		}
	}

	for i := range count {
		person := gofakeit.Person()
		contact := map[string]any{
			"@type":          "Card",
			"version":        "1.0",
			"addressBookIds": tools.ToBoolMap([]string{s.AddressBook()}),
			"prodId":         tools.ProductName,
			"language":       tools.PickRandom("en-US", "en-GB", "en-AU"),
		}

		contact["@type"] = "individual"
		contact["name"] = createName(person)
		nickNameId := id()
		contact["nicknames"] = map[string]any{
			nickNameId: createNickName(person),
		}
		emails := map[string]any{}
		emailId := id()
		emails[emailId] = createEmail(person, 10)
		for i := range rand.Intn(3) {
			emails[id()] = createSecondaryEmail(gofakeit.Email(), i*100)
		}

		uid, err := s.CreateContact(contact)
		if err != nil {
			return err
		}
		printer(fmt.Sprintf("🧑🏻 created %*s/%v uid=%v", int(math.Log10(float64(count))+1), strconv.Itoa(int(i+1)), count, uid))
	}
	return nil
}
