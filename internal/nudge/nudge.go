package nudge

import (
	"github.com/resend/resend-go/v2"
)

func Nudge(email string, hours int, resendApiKey string) {
	client := resend.NewClient(resendApiKey)
	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{"pauldbourke@protonmail.com"},
		Subject: "Hello World",
		Html:    "<p>Congrats on sending your <strong>first email</strong>!</p>",
	}
	client.Emails.Send(params)
}
