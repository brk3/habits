package nudge

import (
	"context"
	"testing"
	"time"

	"github.com/brk3/habits/pkg/habit"
)

func TestGetHabitsExpiringIn(t *testing.T) {
	now := time.Now().UTC().Unix()
	f := &mockClient{
		habits: []string{"guitar", "coding"},
		summary: map[string]*habit.HabitSummary{
			"guitar": {Name: "guitar", CurrentStreak: 3, LastWrite: now - int64(23*time.Hour/time.Second)},
			"coding": {Name: "coding", CurrentStreak: 0, LastWrite: now - int64(2*time.Hour/time.Second)},
		},
	}
	got, err := GetHabitsExpiringIn(context.Background(), f, 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "guitar" {
		t.Fatalf("got %v, want [guitar]", got)
	}
}
