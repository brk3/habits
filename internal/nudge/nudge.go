package nudge

import (
	"context"
	"time"

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

func GetHabitsExpiringIn(ctx context.Context, q Querier, within time.Duration) ([]string, error) {
	now := time.Now().UTC()
	habits, err := q.ListHabits(ctx)
	if err != nil {
		return nil, err
	}
	var expiring []string
	for _, h := range habits {
		summary, err := q.GetHabitSummary(ctx, h)
		if err != nil {
			return nil, err
		}
		if summary.CurrentStreak > 0 {
			lastWrite := time.Unix(summary.LastWrite, 0).UTC()
			nextExpiry := lastWrite.Add(24 * time.Hour)
			if nextExpiry.Sub(now) <= within {
				expiring = append(expiring, h)
			}
		}
	}
	return expiring, nil
}
