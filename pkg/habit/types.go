package habit

type Habit struct {
	Name      string `json:"name"`
	Note      string `json:"note"`
	TimeStamp int64  `json:"timestamp"`
}

type HabitSummary struct {
	Name          string `json:"name"`
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
	FirstLogged   int64  `json:"first_logged"`
	TotalDaysDone int    `json:"total_days_done"`
	BestMonth     int    `json:"best_month"`
	ThisMonth     int    `json:"this_month"`
	LastWrite     int64  `json:"last_write"`
}
