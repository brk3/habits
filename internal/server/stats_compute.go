package server

import (
	"slices"
	"time"
)

func (s *Server) computeStreaks(habit string) (current, longest int, err error) {
	entries, err := s.Store.GetHabit(habit)
	if err != nil {
		return 0, 0, err
	}

	const daySec int64 = 24 * 60 * 60

	// collect unique days from entries
	uniq := make(map[int64]struct{}, len(entries))
	for i := range entries {
		day := time.Unix(entries[i].TimeStamp, 0).UTC().Truncate(24*time.Hour).Unix() / daySec
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

	today := time.Now().UTC().Truncate(24*time.Hour).Unix() / daySec

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
