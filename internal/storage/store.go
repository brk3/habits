package storage

import "github.com/brk3/habits/pkg/habit"

type Store interface {
	PutHabit(userID string, e habit.Habit) error
	ListHabitNames(userID string) ([]string, error)
	GetHabit(userID, name string) ([]habit.Habit, error)
	DeleteHabit(userID, name string) error

	GetAPIKey(key string) (string, error)
	PutAPIKey(key, userID string) error

	Close() error
}
