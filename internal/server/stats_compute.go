package server

import "time"

func (s *Server) computeStreaks(habit string) (current, longest int, err error) {
	entries, err := s.Store.GetHabit(habit)
	if err != nil {
		return 0, 0, err
	}

	truncatedToDay := make([]int64, len(entries))
	for i := range entries {
		truncatedToDay[i] = time.Unix(entries[i].TimeStamp, 0).Truncate(24 * time.Hour).Unix()
	}

	return 10, 10, nil

	// TODO:
	// - current: walk back from today (UTC-truncated) until a gap
	// - longest: scan sorted days for max run length where consecutive days differ by 24h
}
