package nudge

import (
	"context"
	"strings"
	"time"

	"github.com/brk3/habits/internal/apiclient"
	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/logger"
)

func Nudge(cfg *config.Config, n Notifier, nudgeThreshold int) {
	apiclient := apiclient.New(cfg.APIBaseURL, cfg.AuthToken)
	expiring, err := GetHabitsExpiringIn(context.Background(), apiclient,
		time.Now().UTC(), time.Duration(nudgeThreshold)*time.Hour)
	if err != nil {
		logger.Error("error getting expiring habits", "err", err)
	}
	logger.Info("expiring habits", "habits", strings.Join(expiring, ", "))

	if len(expiring) > 0 {
		n.SendNudge(expiring, nudgeThreshold)
	}
}

func GetHabitsExpiringIn(ctx context.Context, q Querier, now time.Time, in time.Duration) ([]string, error) {
	habits, err := q.ListHabits(ctx)
	if err != nil {
		return nil, err
	}

	expiring := []string{}
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
