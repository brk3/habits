package nudge

import (
	"context"
	"github.com/brk3/habits/pkg/habit"
)

type Querier interface {
	ListHabits(ctx context.Context) ([]string, error)
	GetHabitSummary(ctx context.Context, name string) (*habit.HabitSummary, error)
}
