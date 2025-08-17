package server

import (
	"fmt"
	"slices"
	"time"
)

func (s *Server) computeStreaks(habit string) (current, longest int, err error) {
	entries, err := s.Store.GetHabit(habit)
	if err != nil {
		return 0, 0, err
	}

	// collect unique days from entries
	uniq := make(map[int64]struct{}, len(entries))
	for i := range entries {
		day := toDay(entries[i].TimeStamp)
		uniq[day] = struct{}{}
	}

	if len(uniq) == 0 {
		return 0, 0, nil
	}

	// convert to slice, sort and reverse
	days := make([]int64, 0, len(uniq))
	for d := range uniq {
		days = append(days, d)
	}
	slices.Sort(days)
	slices.Reverse(days)

	today := toDay(time.Now().Unix())

	// single unique day
	if len(days) == 1 {
		longest = 1
		if days[0] == today || days[0] == today-1 {
			current = 1
		} else {
			current = 0
		}
		return current, longest, nil
	}

	streakOngoing := days[0] == today || days[0] == today-1
	longest = 1
	run := 1
	current = 0
	if streakOngoing {
		current = 1
	}

	for i := 0; i < len(days)-1; i++ {
		if days[i]-days[i+1] == 1 {
			run++
			longest = max(longest, run)
			if streakOngoing {
				current++
			}
		} else {
			run = 1
			streakOngoing = false
		}
	}

	return current, longest, nil
}

func (s *Server) getFirstLogged(habit string) (int64, error) {
	entries, err := s.Store.GetHabit(habit)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, fmt.Errorf("habit %s not found", habit)
	}

	days := make([]int64, len(entries))
	for i, entry := range entries {
		days[i] = entry.TimeStamp
	}
	slices.Sort(days)

	return days[0], nil
}

func (s *Server) computeTotalDaysDone(habit string) (int, error) {
	entries, err := s.Store.GetHabit(habit)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, fmt.Errorf("habit %s not found", habit)
	}

	days := make(map[int64]struct{}, len(entries))
	for _, e := range entries {
		days[toDay(e.TimeStamp)] = struct{}{}
	}

	return len(days), nil
}

func (s *Server) computeDaysThisMonth(habit string) (int, error) {
	entries, err := s.Store.GetHabit(habit)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, fmt.Errorf("habit %s not found", habit)
	}

	thisMonth := time.Now().UTC().Truncate(24 * time.Hour).Month()
	daysThisMonth := make(map[int64]struct{})

	for _, e := range entries {
		if time.Unix(e.TimeStamp, 0).UTC().Month() == thisMonth {
			daysThisMonth[toDay(e.TimeStamp)] = struct{}{}
		}
	}

	return len(daysThisMonth), nil
}

// toDay converts a Unix timestamp (seconds since 1970) into a "day index".
// A day index is the number of days since 1970-01-01 UTC.
// making it easy to compare days and detect consecutive streaks.
func toDay(ts int64) int64 {
	const daySec = 24 * 60 * 60
	return time.Unix(ts, 0).UTC().Truncate(24*time.Hour).Unix() / daySec
}
