package generator

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

var domains = []string{
	".com",
	".org",
	".net",
	".eu",
	".name",
	"mailbox.org",
	"example.com",
	"acme.com",
}

func addressize(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

type Person struct {
	GivenName string
	SurName   string
	SurName2  string
	FullName  string
	Email     string
	NickName  string
	JobTitle  string
	Company   string
	Phone     string
	Address   *gofakeit.AddressInfo
}

func newPerson() Person {
	p := gofakeit.Person()
	gn := p.FirstName
	sn := p.LastName
	sn2 := ""
	switch rand.IntN(5) {
	case 3:
		p2 := gofakeit.Person()
		sn = sn + "-" + p2.LastName
	case 4:
		p2 := gofakeit.Person()
		sn2 = p2.LastName
	default:
		if rand.IntN(5) == 4 {
			prefix := tools.PickRandom("de", "van", "ter", "zur")
			sn = prefix + " " + sn
		}
	}
	fn := gn + " " + sn
	if sn2 != "" {
		fn = fn + " " + sn2
	}
	domain := tools.PickRandom(domains...)
	csn := sn
	if sn2 != "" {
		csn += "-" + sn2
	}
	email := ""
	if strings.HasPrefix(domain, ".") {
		email = fmt.Sprintf("%s@%s%s", addressize(gn), addressize(csn), domain)
	} else {
		email = fmt.Sprintf("%s.%s@%s", addressize(gn), addressize(csn), domain)
	}
	nickname := ""
	if rand.IntN(3) == 2 {
		nickname = gofakeit.PetName()
	}

	return Person{
		GivenName: gn,
		SurName:   sn,
		SurName2:  sn2,
		FullName:  fn,
		Email:     email,
		NickName:  nickname,
		JobTitle:  p.Job.Title,
		Company:   p.Job.Company,
		Phone:     p.Contact.Phone,
		Address:   p.Address,
	}
}
