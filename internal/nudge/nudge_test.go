package nudge

import (
	"context"
	"testing"
	"time"

	"github.com/brk3/habits/pkg/habit"
)

func TestGetHabitsExpiringIn(t *testing.T) {
	within := 2 * time.Hour

	// last write was 20:00 on Jan 1, 2024 UTC
	lastWrite := time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)

	// now is 18:00 on Jan 1, 2024 UTC - 2 hours to go until threshold
	now := time.Date(2024, 1, 2, 18, 0, 0, 0, time.UTC)

	f := &mockClient{
		habits: []string{"guitar", "coding"},
		summary: map[string]*habit.HabitSummary{
			"guitar": {Name: "guitar", CurrentStreak: 3, LastWrite: lastWrite.Unix()},
			"coding": {Name: "coding", CurrentStreak: 0, LastWrite: lastWrite.Unix()},
		},
	}

	got, err := GetHabitsExpiringIn(context.Background(), f, now, within)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "guitar" {
		t.Fatalf("got %v, want [guitar]", got)
	}
}

func TestGetHabitsExpiringIn_NoneExpiring(t *testing.T) {
	within := 2 * time.Hour

	// last write was today (extending streak, no nudge needed)
	lastWrite := time.Date(2024, 1, 2, 20, 0, 0, 0, time.UTC)

	// now is 10pm on Jan 1, 2024 UTC
	now := time.Date(2024, 1, 2, 22, 0, 0, 0, time.UTC)

	f := &mockClient{
		habits: []string{"guitar", "coding"},
		summary: map[string]*habit.HabitSummary{
			"guitar": {Name: "guitar", CurrentStreak: 3, LastWrite: lastWrite.Unix()},
			"coding": {Name: "coding", CurrentStreak: 0, LastWrite: lastWrite.Unix()},
		},
	}

	got, err := GetHabitsExpiringIn(context.Background(), f, now, within)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want []", got)
	}
}
