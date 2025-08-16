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

	if len(entries) == 1 {
		return 1, 1, nil
	}

	// collect unique days from entries
	truncatedToDay := make(map[int64]struct{})
	for i := range entries {
		t := time.Unix(entries[i].TimeStamp, 0).Truncate(24 * time.Hour).Unix()
		truncatedToDay[t] = struct{}{}
	}

	// sort, reverse, and convert to slice
	s_truncatedToDay := make([]int64, 0, len(truncatedToDay))
	for k := range truncatedToDay {
		s_truncatedToDay = append(s_truncatedToDay, k)
	}
	slices.Sort(s_truncatedToDay)
	slices.Reverse(s_truncatedToDay)

	// check if the most recent entry is yesterday
	streakOngoing := false
	mostRecentEntry := time.Unix(s_truncatedToDay[0], 0).UTC()
	yesterday := time.Now().UTC().Truncate(24 * time.Hour).Unix()
	if mostRecentEntry.Unix() >= yesterday {
		streakOngoing = true
	}

	longest = 0
	current = 0
	c := 0
	for i := 0; i < len(s_truncatedToDay)-1; i++ {
		t1 := time.Unix(s_truncatedToDay[i], 0).UTC()
		t2 := time.Unix(s_truncatedToDay[i+1], 0).UTC()
		d1 := t1.Day()
		d2 := t2.Day()

		if d1-d2 == 1 {
			c++
			longest = max(longest, c)
			if streakOngoing {
				current++
			}
		} else {
			c = 0
			streakOngoing = false
		}
	}

	return current, longest, nil
}
