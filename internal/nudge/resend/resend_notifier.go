package resend

import (
	"bytes"
	"html/template"

	"github.com/resend/resend-go/v2"
)

type ResendNotifier struct {
	ApiKey string
	Email  string
}

const htmlTemplate = `
<p>The following habit streaks are expiring within the next {{.Hours}} hours:</p>
<ul>
{{range .Habits}}
  <li>{{.}}</li>
{{end}}
</ul>
`

func (r *ResendNotifier) SendNudge(habits []string, hoursTillExpiry int) error {
	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	data := struct {
		Habits []string
		Hours  int
	}{
		Habits: habits,
		Hours:  hoursTillExpiry,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	client := resend.NewClient(r.ApiKey)
	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{r.Email},
		Subject: "Streaks are expiring soon",
		Html:    buf.String(),
	}

	_, err = client.Emails.Send(params)
	return err
}
