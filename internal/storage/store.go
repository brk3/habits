package storage

import (
	"github.com/brk3/habits/pkg/habit"
	"golang.org/x/oauth2"
)

type Store interface {
	PutHabit(userID string, e habit.Habit) error
	ListHabitNames(userID string) ([]string, error)
	GetHabit(userID, name string) ([]habit.Habit, error)
	DeleteHabit(userID, name string) error

	PutAPIKey(keyHash, userID string) error
	GetAPIKey(keyHash string) (userID string, found bool, err error)
	ListAPIKeyHashes(userID string) ([]string, error)
	DeleteAPIKey(keyHash string) error

	PutRefreshToken(userID string, token *oauth2.Token) error
	GetRefreshToken(userID string) (*oauth2.Token, bool, error)
	DeleteRefreshToken(userID string) error

	Close() error
}
