package storage

import "github.com/brk3/habits/pkg/habit"

type Store interface {
	PutHabit(e habit.Habit) error
	ListHabitNames() ([]string, error)
	GetHabit(name string) ([]habit.Habit, error)
	Close() error
}
