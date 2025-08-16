package server

import (
	"github.com/brk3/habits/pkg/habit"
)

type HabitListResponse struct {
	Habits []string `json:"habits"`
}

type HabitGetResponse struct {
	HabitID string        `json:"habit_id"`
	Entries []habit.Habit `json:"entries"`
}

type HabitSummaryResponse struct {
	HabitID      string             `json:"habit_id"`
	HabitSummary habit.HabitSummary `json:"habit_summary"`
}
