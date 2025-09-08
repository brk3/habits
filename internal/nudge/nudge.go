package nudge

import (
	"context"
	"fmt"
	"time"

	"github.com/brk3/habits/internal/apiclient"
	"github.com/brk3/habits/internal/config"
)

// TODO(pbourke): make config params
func Nudge(email string, hours int, resendApiKey string) {
	cfg := config.Load()
	apiclient := apiclient.New(cfg.APIBaseURL)
	expiring, err := GetHabitsExpiringIn(context.Background(), apiclient, time.Now().UTC(), time.Duration(hours)*time.Hour)
	if err != nil {
		fmt.Println("error getting expiring habits:", err)
	}
	/*
		client := resend.NewClient(resendApiKey)
		params := &resend.SendEmailRequest{
			From:    "onboarding@resend.dev",
			To:      []string{"pauldbourke@protonmail.com"},
			Subject: "Hello World",
			Html:    "<p>Congrats on sending your <strong>first email</strong>!</p>",
		}
		client.Emails.Send(params)
	*/
	fmt.Printf("habits expiring soon: %v\n", expiring)
}

func GetHabitsExpiringIn(ctx context.Context, q Querier, now time.Time, within time.Duration) ([]string, error) {
	habits, err := q.ListHabits(ctx)
	if err != nil {
		return nil, err
	}

	var expiring []string
	for _, habitKey := range habits {
		h, err := q.GetHabitSummary(ctx, habitKey)
		if err != nil {
			return nil, err
		}
		cutoff := time.Unix(h.LastWrite, 0).Add(24 * time.Hour)
		if h.CurrentStreak > 0 && now.Before(cutoff) && cutoff.Sub(now) <= within {
			expiring = append(expiring, h.Name)
		}
	}

	return expiring, nil
}
