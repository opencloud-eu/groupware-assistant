package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"opencloud.eu/groupware-assistant/pkg/generator"
)

var emailGenerateCmd = &cobra.Command{
	Use: "generate",
	RunE: func(cmd *cobra.Command, args []string) error {
		count, err := cmd.Flags().GetUint("count")
		if err != nil {
			return err
		}
		senders, err := cmd.Flags().GetUint("senders")
		if err != nil {
			return err
		}
		emojis, err := cmd.Flags().GetBool("emojis")
		if err != nil {
			return err
		}
		empty, err := cmd.Flags().GetBool("empty")
		if err != nil {
			return err
		}
		domain, err := cmd.Flags().GetString("domain")
		if err != nil {
			return err
		}
		mailboxId, err := cmd.Flags().GetString("mailbox-id")
		if err != nil {
			return err
		}
		mailboxRole, err := cmd.Flags().GetString("mailbox-role")
		if err != nil {
			return err
		}
		minThreadSize, err := cmd.Flags().GetUint("min-thread-size")
		if err != nil {
			return err
		}
		maxThreadSize, err := cmd.Flags().GetUint("max-thread-size")
		if err != nil {
			return err
		}
		ccEvery, err := cmd.Flags().GetUint("cc-every")
		if err != nil {
			return err
		}
		bccEvery, err := cmd.Flags().GetUint("bcc-every")
		if err != nil {
			return err
		}
		seenEvery, err := cmd.Flags().GetUint("seen-every")
		if err != nil {
			return err
		}
		attachEvery, err := cmd.Flags().GetUint("attach-every")
		if err != nil {
			return err
		}
		minAttachments, err := cmd.Flags().GetUint("min-attachments")
		if err != nil {
			return err
		}
		maxAttachments, err := cmd.Flags().GetUint("max-attachments")
		if err != nil {
			return err
		}
		attachmentOptionsSpec, err := cmd.Flags().GetString("attachment-options")
		if err != nil {
			return err
		}
		forwardedEvery, err := cmd.Flags().GetUint("forwarded-every")
		if err != nil {
			return err
		}
		importantEvery, err := cmd.Flags().GetUint("important-every")
		if err != nil {
			return err
		}
		junkEvery, err := cmd.Flags().GetUint("junk-every")
		if err != nil {
			return err
		}
		notJunkEvery, err := cmd.Flags().GetUint("not-junk-every")
		if err != nil {
			return err
		}
		phishingEvery, err := cmd.Flags().GetUint("phishing-every")
		if err != nil {
			return err
		}
		draftEvery, err := cmd.Flags().GetUint("draft-every")
		if err != nil {
			return err
		}
		icalEvery, err := cmd.Flags().GetUint("ical-every")
		if err != nil {
			return err
		}

		if senders == 0 {
			senders = min(1, count/4)
		}

		return generator.GenerateEmails(
			JmapUrl,
			Trace,
			emojis,
			Username,
			Password,
			AccountId,
			empty,
			mailboxId,
			mailboxRole,
			domain,
			count,
			senders,
			minThreadSize,
			maxThreadSize,
			ccEvery,
			bccEvery,
			seenEvery,
			attachEvery,
			minAttachments,
			maxAttachments,
			attachmentOptionsSpec,
			forwardedEvery,
			importantEvery,
			junkEvery,
			notJunkEvery,
			phishingEvery,
			draftEvery,
			icalEvery,
			func(text string) { fmt.Println(text) },
		)
	},
}

func init() {
	emailCmd.AddCommand(emailGenerateCmd)

	emailGenerateCmd.Flags().UintP("count", "c", 20, "How many emails to add to the folder")
	emailGenerateCmd.Flags().UintP("senders", "s", 0, "How many senders to use, spread randomly across the emails; 0 is the default and is then computed to be <count>/4")
	emailGenerateCmd.Flags().BoolP("empty", "E", false, "Whether to empty the folder before adding emails to it")
	emailGenerateCmd.Flags().StringP("domain", "d", "example.com", "The domain to use for all email addresses (From, CC, ...)")
	emailGenerateCmd.Flags().String("mailbox-id", "", "ID of the JMAP Mailbox to use")
	emailGenerateCmd.Flags().String("mailbox-role", "inbox", "Role of the JMAP Mailbox to use when no ID is specified")
	emailGenerateCmd.Flags().Uint("min-thread-size", 1, "Minimum number of emails in one thread")
	emailGenerateCmd.Flags().Uint("max-thread-size", 6, "Maximum number of emails in one thread")
	emailGenerateCmd.Flags().Uint("cc-every", 3, "Add CC: headers every n emails")
	emailGenerateCmd.Flags().Uint("bcc-every", 2, "Add BCC: headers every n emails")
	emailGenerateCmd.Flags().Uint("seen-every", 3, "Mark emails as seen (read) every n emails")
	emailGenerateCmd.Flags().Uint("attach-every", 2, "Add a random number of attachments every n emails")
	emailGenerateCmd.Flags().Uint("min-attachments", 1, "Minimum number of attachments per email")
	emailGenerateCmd.Flags().Uint("max-attachments", 4, "Maximum number of attachments per email")
	emailGenerateCmd.Flags().String("attachment-options", "", "Specifies a comma-separated list of numbers of attachments of which a random value is picked for every email; when set, overrides --min-attachments, --max-attachments and --attachment-every")
	emailGenerateCmd.Flags().Uint("forwarded-every", 4, "Mark emails as forwarded every n emails")
	emailGenerateCmd.Flags().Uint("important-every", 4, "Mark emails as important every n emails")
	emailGenerateCmd.Flags().Uint("junk-every", 10, "Mark emails as junk every n emails")
	emailGenerateCmd.Flags().Uint("not-junk-every", 3, "Mark emails as not-junk every n emails")
	emailGenerateCmd.Flags().Uint("phishing-every", 7, "Mark emails as phishing every n emails")
	emailGenerateCmd.Flags().Uint("draft-every", 10, "Mark emails as draft every n emails")
	emailGenerateCmd.Flags().Uint("ical-every", 4, "Add ical attachment every n emails")
	emailGenerateCmd.Flags().Bool("emojis", true, "Whether to include emojis in the From name to easily find emails that match certain criteria")
}
