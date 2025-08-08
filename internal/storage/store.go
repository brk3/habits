package storage

import "github.com/brk3/habits/pkg/habit"

type Store interface {
	Put(e habit.Habit) error
	ListHabitNames() ([]string, error)
	ListEntriesByHabit(name string) ([]habit.Habit, error)
	Close() error
}
