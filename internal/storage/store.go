package storage

import "github.com/brk3/habits/pkg/habit"

type Store interface {
	PutHabit(e habit.Habit) error
	ListHabitNames() ([]string, error)
	GetHabit(name string) ([]habit.Habit, error)
	DeleteHabit(name string) error
	Close() error
}
