package nudge

import (
	"context"

	"github.com/brk3/habits/pkg/habit"
)

type mockClient struct {
	habits  []string
	summary map[string]*habit.HabitSummary
	err     error
}

func (f *mockClient) ListHabits(ctx context.Context) ([]string, error) {
	return f.habits, f.err
}

func (f *mockClient) GetHabitSummary(ctx context.Context, name string) (*habit.HabitSummary, error) {
	return f.summary[name], f.err
}
