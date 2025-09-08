package nudge

import (
	"context"
	"fmt"
	"time"

	"github.com/brk3/habits/internal/apiclient"
	"github.com/brk3/habits/internal/config"
)

func Nudge(n Notifier, nudgeThreshold int) {
	cfg := config.Load()
	apiclient := apiclient.New(cfg.APIBaseURL)
	expiring, err := GetHabitsExpiringIn(context.Background(), apiclient,
		time.Now().UTC(), time.Duration(nudgeThreshold)*time.Hour)
	if err != nil {
		fmt.Println("error getting expiring habits:", err)
	}
	n.SendNudge(expiring, nudgeThreshold)
}

func GetHabitsExpiringIn(ctx context.Context, q Querier, now time.Time, in time.Duration) ([]string, error) {
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
		if h.CurrentStreak > 0 && now.Before(cutoff) && cutoff.Sub(now) <= in {
			expiring = append(expiring, h.Name)
		}
	}

	return expiring, nil
}
