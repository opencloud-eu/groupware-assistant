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
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/brianvoe/gofakeit/v7"
	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

const (
	jpeg = "image/jpeg"
	png  = "image/png"
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
	includeDataUri bool,
	mediaBlobs bool,
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
			printer(fmt.Sprintf("🗑️  deleted %d contacts", deleted))
		} else {
			printer("ℹ️ did not delete any contacts, addressbook is empty")
		}
	}

	mediaOptions := []string{"picsum"}
	if includeDataUri {
		mediaOptions = append(mediaOptions, "data:jpeg", "data:png")
	}
	if mediaBlobs {
		mediaOptions = append(mediaOptions, "blob:jpeg", "blob:png")
	}

	for i := range count {
		person := newPerson()
		created := timestamp(time.Now(), -1)
		contact := map[string]any{
			"@type":          "Card",
			"version":        "1.0",
			"addressBookIds": tools.ToBoolMap([]string{s.AddressBook()}),
			"prodId":         tools.ProductName,
			"language":       tools.PickLanguage(),
			"kind":           "invidual",
			"name":           createName(person),
			"created":        created,
		}
		if rand.Intn(2) < 1 {
			contact["updated"] = timestamp(created, 1)
		}

		if nn := createNickName(person); nn != nil {
			contact["nicknames"] = map[string]map[string]any{id(): *nn}
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
			num := person.Phone
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
					"user":    "@" + person.Email,
					"uri":     "https://mastodon.example.com/@" + strings.ToLower(person.GivenName),
				}, nil
			case 1:
				return map[string]any{
					"@type": "OnlineService",
					"uri":   "xmpp:" + person.Email,
				}, nil
			default:
				return map[string]any{
					"@type":   "OnlineService",
					"service": "Discord",
					"user":    person.Email,
					"uri":     "https://discord.example.com/user/" + person.Email,
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
					"name":     person.Company,
					"contexts": tools.ToBoolMapS("work"),
				}
				title := map[string]any{
					"@type":          "Title",
					"kind":           "title",
					"name":           person.JobTitle,
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
			key, err := helper.GenerateKey(person.FullName, person.Email, []byte("secret"), "x25519", 0)
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
			switch tools.PickRandom(mediaOptions...) {
			case "picsum": // reference HTTP URI to picsum.photos
				dim := tools.PickRandom(64, 80, 100, 128, 200, 256, 384, 512)
				m := map[string]any{
					"@type": "Media",
					"kind":  "photo",
					"uri":   picsum(dim, dim),
				}
				if rand.Intn(2) == 0 {
					m["mediaType"] = jpeg
				}
				return m, nil
			case "blob:jpeg": // will generate a JPEG blob
				return map[string]any{
					"@type":     "Media",
					"kind":      "photo",
					"mediaType": jpeg,
				}, nil
			case "blob:png": // will generate a PNG blob
				return map[string]any{
					"@type":     "Media",
					"kind":      "photo",
					"mediaType": png,
				}, nil
			case "data:jpeg": // a JPEG as data: URI
				dim := tools.PickRandom(64, 80, 100, 128)
				m := map[string]any{
					"@type": "Media",
					"kind":  "photo",
					"uri":   "data:" + jpeg + ";base64," + base64.RawStdEncoding.EncodeToString(gofakeit.ImageJpeg(dim, dim)),
				}
				if rand.Intn(2) == 0 {
					m["mediaType"] = jpeg
				}
				return m, nil
			case "data:png": // a PNG as data: URI
				dim := tools.PickRandom(64, 80, 100, 128)
				m := map[string]any{
					"@type": "Media",
					"kind":  "photo",
					"uri":   "data:" + png + ";base64," + base64.RawStdEncoding.EncodeToString(gofakeit.ImagePng(dim, dim)),
				}
				if rand.Intn(2) == 0 {
					m["mediaType"] = png
				}
				return m, nil
			default:
				panic(fmt.Errorf("unsupported switch option"))
			}
		}); err != nil {
			return err
		}
		if err := propmap(contact, "links", 0, 1, func(i int, id string) (map[string]any, error) {
			return map[string]any{
				"@type": "Link",
				"kind":  "contact",
				"uri":   "mailto" + person.Email,
				"pref":  (i + 1) * 10,
			}, nil
		}); err != nil {
			return err
		}

		id, mediaBlobIds, err := s.CreateContact(contact)
		if err != nil {
			return err
		}

		mediaBlobIdsText := ""
		if len(mediaBlobIds) > 0 {
			mediaBlobIdsText = " " + strings.Join(tools.Map(mediaBlobIds, func(blobId string) string { return fmt.Sprintf("🖼️%s", blobId) }), " ")
		}
		printer(fmt.Sprintf("🪪 created %*s/%v id=%v name='%s' <%s>%s", int(math.Log10(float64(count))+1), strconv.Itoa(int(i+1)), count, id, person.FullName, person.Email, mediaBlobIdsText))
	}
	return nil
}

var streetNumberRegex = regexp.MustCompile(`^(\d+)\s+(.+)$`)
