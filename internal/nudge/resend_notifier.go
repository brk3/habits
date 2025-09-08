package nudge

import (
	"github.com/resend/resend-go/v2"
)

type ResendNotifier struct {
	ApiKey string
	Email  string
}

func (r *ResendNotifier) SendNudge(habits []string, hoursTillExpiry int) error {
	client := resend.NewClient(r.ApiKey)
	params := &resend.SendEmailRequest{
		// TODO(pbourke): figure out good values for these
		From:    "onboarding@resend.dev",
		To:      []string{r.Email},
		Subject: "Hello World",
		Html:    "<p>Congrats on sending your <strong>first email</strong>!</p>",
	}
	client.Emails.Send(params)
	return nil
}
