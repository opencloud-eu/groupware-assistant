package generator

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"github.com/brianvoe/gofakeit/v7"
	"opencloud.eu/groupware-assistant/pkg/jmap"
)

func GenerateEmails(
	jmapUrl string,
	trace bool,
	emojis bool,
	username string,
	password string,
	accountId string,
	empty bool,
	mailboxId string,
	mailboxRole string,
	domain string,
	count uint,
	senders uint,
	minThreadSize uint,
	maxThreadSize uint,
	ccEvery uint,
	bccEvery uint,
	seenEvery uint,
	attachmentEvery uint,
	minAttachments uint,
	maxAttachments uint,
	attachmentOptionsSpec string,
	forwardedEvery uint,
	importantEvery uint,
	junkEvery uint,
	notJunkEvery uint,
	phishingEvery uint,
	draftEvery uint,
	icalEvery uint,
	printer func(string),
) error {
	var attachmentOptions []uint = nil
	if attachmentOptionsSpec != "" {
		attachmentOptionStrings := strings.Split(attachmentOptionsSpec, ",")
		attachmentOptions = make([]uint, len(attachmentOptionStrings))
		for i, o := range attachmentOptionStrings {
			value, err := strconv.Atoi(o)
			if err != nil {
				return err
			} else {
				attachmentOptions[i] = uint(value)
			}
		}
	}

	var s *jmap.JmapEmailSender = nil
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

		s, err = jmap.NewJmapEmailSender(j, accountId, mailboxId, mailboxRole)
		if err != nil {
			return err
		}
	}
	defer s.Close()

	if empty {
		deleted, err := s.EmptyEmails()
		if err != nil {
			return err
		}
		if deleted > 0 {
			printer(fmt.Sprintf("🗑️ deleted %d messages", deleted))
		} else {
			printer("ℹ️ did not delete any messages, folder is empty")
		}
	}

	toName := username
	toAddress := fmt.Sprintf("%s@%s", username, domain)
	ccName1 := "Team Lead"
	ccAddress1 := fmt.Sprintf("lead@%s", domain)
	ccName2 := "Coworker"
	ccAddress2 := fmt.Sprintf("coworker@%s", domain)
	bccName := "HR"
	bccAddress := fmt.Sprintf("corporate@%s", domain)

	sg := newSenderGenerator(senders)

	for i := uint(0); i < count; {
		threadMessageId := fmt.Sprintf("%d.%d@%s", time.Now().Unix(), 1000000+rand.Intn(8999999), domain)
		threadSubject := strings.Trim(gofakeit.Sentence(), ".") // remove the . at the end, looks weird
		threadSize := minThreadSize + uint(rand.Intn(int(maxThreadSize-minThreadSize)))
		lastMessageId := ""
		lastSubject := ""
		threadStart := time.Now().Add(time.Duration(-(24*60)-rand.Intn(7*24*60)) * time.Minute)
		received := threadStart

		for t := uint(0); i < count && t < threadSize; t++ {
			sender, err := sg.nextSender()
			if err != nil {
				return err
			}
			received = received.Add(time.Duration(rand.Intn(5)) * time.Minute)

			b, err := s.NewEmail()
			if err != nil {
				return err
			}
			b.To(mail.Address{Name: toName, Address: toAddress})

			forwarded := forwardedEvery > 0 && i%forwardedEvery == 0
			important := importantEvery > 0 && i%importantEvery == 0
			junk := junkEvery > 0 && i%junkEvery == 0
			notJunk := notJunkEvery > 0 && i%notJunkEvery == 0
			if junk {
				notJunk = false
			}
			phishing := phishingEvery > 0 && i%phishingEvery == 0
			seen := seenEvery > 0 && i%seenEvery == 0
			draft := draftEvery > 0 && i%draftEvery == 0
			ical := icalEvery > 0 && i%icalEvery == 0

			subject := ""
			answered := t < threadSize-1
			if lastMessageId == "" {
				// start a new thread
				b.MessageId(threadMessageId)
				subject = threadSubject
				lastMessageId = threadMessageId
				lastSubject = threadSubject
			} else {
				// we're continuing a thread
				messageId := fmt.Sprintf("%d.%d@%s", time.Now().Unix(), 1000000+rand.Intn(8999999), domain)
				inReplyTo := ""
				switch rand.Intn(2) {
				case 0:
					// reply to first post in thread
					if forwarded {
						subject = "Re: Fwd: " + threadSubject
					} else {
						subject = "Re: " + threadSubject
					}
					inReplyTo = threadMessageId
				default:
					// reply to last addition to thread
					if forwarded {
						subject = "Re: Fwd: " + lastSubject
					} else {
						subject = "Re: " + lastSubject
					}
					inReplyTo = lastMessageId
				}
				b.MessageId(messageId)
				b.InReplyTo(inReplyTo)
				lastMessageId = messageId
				lastSubject = subject
			}

			if answered {
				b.Answered()
			}
			if forwarded {
				b.Forwarded()
			}
			if important {
				b.Important()
			}
			if junk {
				b.Junk()
			}
			if notJunk {
				b.NotJunk()
			}
			if phishing {
				b.Phishing()
			}
			if seen {
				b.Seen()
			}
			if draft {
				b.Draft()
			}

			if i%ccEvery == 0 {
				b.CC([]mail.Address{{Name: ccName1, Address: ccAddress1}, {Name: ccName2, Address: ccAddress2}})
			}
			if i%bccEvery == 0 {
				b.BCC([]mail.Address{{Name: bccName, Address: bccAddress}})
			}

			b.ReturnPath(sender.from)
			b.Received(received.Add(time.Duration(-2) * time.Minute))
			b.Sent(received)

			numAttachments := uint(0)
			if attachmentOptions != nil {
				numAttachments = attachmentOptions[rand.Intn(len(attachmentOptions))]
			} else if maxAttachments > 0 && i%attachmentEvery == 0 {
				numAttachments = minAttachments + uint(rand.Intn(int(maxAttachments)+1))
			}

			for a := range numAttachments {
				switch rand.Intn(3) {
				case 0:
					filename := fakeFilename(".txt")
					attachment := gofakeit.Paragraph(2+rand.Intn(4), 1+rand.Intn(4), 1+rand.Intn(32), "\n")
					b.Attach([]byte(attachment), "text/plain", filename)
				case 1:
					filename := fakeFilename(".pdf")
					pdf := fpdf.New("P", "mm", "A4", "")
					pdf.AddPage()
					pdf.SetFont("Arial", "", 12)
					for range 4 + rand.Intn(5) {
						pdf.Write(2, gofakeit.Sentence())
						pdf.Write(2, "\n")
					}
					var buf bytes.Buffer
					pdf.Output(&buf)
					b.Attach(buf.Bytes(), "application/pdf", filename)
				default:
					filename := ""
					mimetype := ""
					var image []byte = nil
					switch rand.Intn(2) {
					case 0:
						filename = fakeFilename(".png")
						mimetype = "image/png"
						image = gofakeit.ImagePng(512, 512)
					default:
						filename = fakeFilename(".jpg")
						mimetype = "image/jpeg"
						image = gofakeit.ImageJpeg(400, 200)
					}
					switch rand.Intn(2) {
					case 0:
						b.Attach(image, mimetype, filename)
					default:
						b.AttachInline(image, mimetype, filename, "c"+strconv.Itoa(int(a)))
					}
				}
			}

			if ical {
				starts := time.Date(received.Year(), received.Month(), received.Day(), 9+rand.Intn(8), rand.Intn(2)*30, 0, 0, received.Location())
				duration := time.Duration(1+rand.Intn(8)) * 15 * time.Minute
				numAttendees := 1 + rand.Intn(8)
				attendees := make([]icalAttendee, numAttendees+1)
				attendees[0] = icalAttendee{Name: toName, Email: toAddress}
				for i := 1; i < numAttendees+1; i++ {
					attendees[i] = icalAttendee{Name: gofakeit.Name(), Email: gofakeit.Email()}
				}
				resource := "https://meet.opentalk.eu/room/" + gofakeit.UUID()
				text := toIcal(received, starts, duration, gofakeit.BookTitle(), gofakeit.URL(), gofakeit.Product().Description, "", attendees[rand.Intn(len(attendees))].Name, attendees, resource)
				b.Attach([]byte(text), "text/calendar", "appointment.ics")
			}

			format := formats[int(i)%len(formats)]
			text := gofakeit.Paragraph(2+rand.Intn(9), 1+rand.Intn(4), 1+rand.Intn(32), "\n")
			format(text, b)

			b.Subject(subject)
			b.Sender(sender.ToAddress())

			from := sender.ToAddress()
			if emojis {
				markers := []string{}
				if important {
					markers = append(markers, "❗")
				}
				if junk {
					markers = append(markers, "🗑️")
				}
				if phishing {
					markers = append(markers, "🐟")
				}
				if notJunk {
					markers = append(markers, "🧼")
				}
				if numAttachments > 0 {
					markers = append(markers, "📎")
				}
				if forwarded {
					markers = append(markers, "➡️")
				}
				if answered {
					markers = append(markers, "💬")
				}
				if draft {
					markers = append(markers, "✏️")
				}
				if ical {
					markers = append(markers, "📅")
				}
				if len(markers) > 0 {
					from.Name = from.Name + " " + strings.Join(markers, "")
				}
			}
			b.From(from)

			uid, err := s.SendEmail(b)
			if err != nil {
				return err
			}

			{
				attachmentStr := ""
				if numAttachments > 0 {
					attachmentStr = " " + strings.Repeat("📎", int(numAttachments)) + " "
				}
				printer(fmt.Sprintf("📩appended %*s/%v uid=%v%s'%s'", int(math.Log10(float64(count))+1), strconv.Itoa(int(i+1)), count, uid, attachmentStr, subject))
			}

			i++
		}
	}
	return nil
}
