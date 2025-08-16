package habit

type Habit struct {
	Name      string `json:"name"`
	Note      string `json:"note"`
	TimeStamp int64  `json:"timestamp"`
}

type HabitSummary struct {
	Name              string `json:"name"`
	CurrentStreakDays int    `json:"current_streak_days"`
	LongestStreakDays int    `json:"longest_streak_days"`
	DaysThisMonth     int    `json:"days_this_month"`
	TotalDaysCount    int    `json:"total_days_count"`
	BestMonth         int    `json:"best_month"`
	LastUpdated       int64  `json:"last_updated"`
}
