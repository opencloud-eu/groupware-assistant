package generator

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/brianvoe/gofakeit/v7"
	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

func GenerateContacts(
	jmapUrl string,
	trace bool,
	color bool,
	username string,
	password string,
	accountId string,
	empty bool,
	addressbookId string,
	count uint,
	printer func(string),
) error {
	var s *jmap.ContactSender = nil
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

		s, err = jmap.NewContactSender(j, accountId, addressbookId)
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
			"language":       tools.PickLanguage(),
			"kind":           "invidual",
			"name":           createName(person),
		}

		if rand.Intn(3) < 1 {
			contact["nicknames"] = map[string]map[string]any{id(): createNickName(person)}
		}

		{
			emails := map[string]map[string]any{}
			emailId := id()
			emails[emailId] = createEmail(person, 10)
			for i := range rand.Intn(3) {
				emails[id()] = createSecondaryEmail(gofakeit.Email(), i*100)
			}
			if len(emails) > 0 {
				contact["emails"] = emails
			}
		}
		if err := propmap(contact, "phones", 0, 2, func(i int, id string) (map[string]any, error) {
			num := person.Contact.Phone
			if i > 0 {
				num = gofakeit.Phone()
			}
			var features map[string]bool = nil
			if rand.Intn(3) < 2 {
				features = tools.ToBoolMapS("mobile", "voice", "video", "text")
			} else {
				features = tools.ToBoolMapS("voice", "main-number")
			}
			contexts := map[string]bool{}
			contexts["work"] = true
			if rand.Intn(2) < 1 {
				contexts["private"] = true
			}
			return map[string]any{
				"@type":    "Phone",
				"number":   "tel:" + "+1" + num,
				"features": features,
				"contexts": contexts,
			}, nil
		}); err != nil {
			return err
		}
		if err := propmap(contact, "addresses", 1, 2, func(i int, id string) (map[string]any, error) {
			var source *gofakeit.AddressInfo
			if i == 0 {
				source = person.Address
			} else {
				source = gofakeit.Address()
			}
			components := []map[string]string{}
			m := streetNumberRegex.FindAllStringSubmatch(source.Street, -1)
			if m != nil {
				components = append(components, map[string]string{"kind": "name", "value": m[0][2]})
				components = append(components, map[string]string{"kind": "number", "value": m[0][1]})
			} else {
				components = append(components, map[string]string{"kind": "name", "value": source.Street})
			}
			components = append(components,
				map[string]string{"kind": "locality", "value": source.City},
				map[string]string{"kind": "country", "value": source.Country},
				map[string]string{"kind": "state", "value": source.State},
				map[string]string{"kind": "postcode", "value": source.Zip},
			)
			return map[string]any{
				"@type":            "Address",
				"components":       components,
				"defaultSeparator": ", ",
				"isOrdered":        true,
				"timeZone": tools.PickRandom("America/Adak", "America/Anchorage", "America/Chicago", "America/Denver",
					"America/Detroit", "America/Indiana/Knox", "America/Kentucky/Louisville", "America/Los_Angeles", "America/New_York"),
			}, nil
		}); err != nil {
			return err
		}
		if err := propmap(contact, "onlineServices", 0, 2, func(i int, id string) (map[string]any, error) {
			switch rand.Intn(3) {
			case 0:
				return map[string]any{
					"@type":   "OnlineService",
					"service": "Mastodon",
					"user":    "@" + person.Contact.Email,
					"uri":     "https://mastodon.example.com/@" + strings.ToLower(person.FirstName),
				}, nil
			case 1:
				return map[string]any{
					"@type": "OnlineService",
					"uri":   "xmpp:" + person.Contact.Email,
				}, nil
			default:
				return map[string]any{
					"@type":   "OnlineService",
					"service": "Discord",
					"user":    person.Contact.Email,
					"uri":     "https://discord.example.com/user/" + person.Contact.Email,
				}, nil
			}
		}); err != nil {
			return err
		}

		if err := propmap(contact, "preferredLanguages", 0, 2, func(i int, id string) (map[string]any, error) {
			return map[string]any{
				"@type":    "LanguagePref",
				"language": tools.PickRandom("en", "fr", "de", "es", "it"),
				"contexts": tools.ToBoolMap(tools.PickRandoms1("work", "private")),
				"pref":     i + 1,
			}, nil
		}); err != nil {
			return err
		}

		{
			organizations := map[string]map[string]any{}
			titles := map[string]map[string]any{}
			for range rand.Intn(2) {
				orgId := id()
				org := map[string]any{
					"@type":    "Organization",
					"name":     person.Job.Company,
					"contexts": tools.ToBoolMapS("work"),
				}
				title := map[string]any{
					"@type":          "Title",
					"kind":           "title",
					"name":           person.Job.Title,
					"organizationId": orgId,
				}
				organizations[orgId] = org
				titles[id()] = title
			}
			if len(organizations) > 0 {
				contact["organizations"] = organizations
				contact["titles"] = titles
			}
		}

		if err := propmap(contact, "cryptoKeys", 0, 1, func(i int, id string) (map[string]any, error) {
			key, err := helper.GenerateKey(person.FirstName+" "+person.LastName, person.Contact.Email, []byte("secret"), "x25519", 0)
			if err != nil {
				return nil, err
			}
			keyring, err := crypto.NewKeyFromArmoredReader(strings.NewReader(key))
			if err != nil {
				return nil, err
			}
			pubkey, err := keyring.GetPublicKey()
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"@type": "CryptoKey",
				"uri":   "data:application/pgp-keys;base64," + base64.RawStdEncoding.EncodeToString(pubkey),
			}, nil
		}); err != nil {
			return err
		}
		if err := propmap(contact, "media", 0, 1, func(i int, id string) (map[string]any, error) {
			if rand.Intn(2) < 1 {
				return map[string]any{
					"@type": "Media",
					"kind":  "photo",
					"uri":   "data:image/jpeg;base64," + base64.RawStdEncoding.EncodeToString(gofakeit.ImageJpeg(64, 64)),
				}, nil
			} else {
				return map[string]any{
					"@type": "Media",
					"kind":  "photo",
					"uri":   picsum(128, 128),
				}, nil
			}
		}); err != nil {
			return err
		}
		if err := propmap(contact, "links", 0, 1, func(i int, id string) (map[string]any, error) {
			return map[string]any{
				"@type": "Link",
				"kind":  "contact",
				"uri":   "mailto" + person.Contact.Email,
				"pref":  (i + 1) * 10,
			}, nil
		}); err != nil {
			return err
		}

		uid, err := s.CreateContact(contact)
		if err != nil {
			return err
		}
		printer(fmt.Sprintf("🧑🏻 created %*s/%v uid=%v", int(math.Log10(float64(count))+1), strconv.Itoa(int(i+1)), count, uid))
	}
	return nil
}

var streetNumberRegex = regexp.MustCompile(`^(\d+)\s+(.+)$`)
