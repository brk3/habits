package server

import (
	"sync"

	"github.com/brk3/habits/internal/storage"
	"github.com/brk3/habits/pkg/habit"
	"golang.org/x/oauth2"
)

type memStore struct {
	mu            sync.RWMutex
	habits        map[string][]habit.Habit
	apiKeys       map[string]string
	refreshTokens map[string]*oauth2.Token
}

func newMemStore() *memStore {
	return &memStore{
		habits:        map[string][]habit.Habit{},
		apiKeys:       map[string]string{},
		refreshTokens: map[string]*oauth2.Token{},
	}
}

func (m *memStore) PutHabit(userID string, h habit.Habit) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.habits[h.Name] = append(m.habits[h.Name], h)

	return nil
}

func (m *memStore) ListHabitNames(userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := []string{}
	for habitKey := range m.habits {
		out = append(out, habitKey)
	}

	return out, nil
}

func (m *memStore) GetHabit(userID, name string) ([]habit.Habit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]habit.Habit(nil), m.habits[name]...), nil
}

func (m *memStore) GetHabitSummary(name string) (habit.HabitSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := habit.HabitSummary{
		Name: name,
	}
	return summary, nil
}

func (m *memStore) DeleteHabit(userID, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.habits, name)
	return nil
}

func (m *memStore) PutAPIKey(keyHash, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.apiKeys[keyHash] = userID
	return nil
}

func (m *memStore) GetAPIKey(keyHash string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userID, found := m.apiKeys[keyHash]
	return userID, found, nil
}

func (m *memStore) ListAPIKeyHashes(userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var hashes []string
	for hash, storedUserID := range m.apiKeys {
		if storedUserID == userID {
			hashes = append(hashes, hash)
		}
	}
	return hashes, nil
}

func (m *memStore) DeleteAPIKey(keyHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.apiKeys, keyHash)
	return nil
}

func (m *memStore) PutRefreshToken(userID string, token *oauth2.Token) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.refreshTokens[userID] = token
	return nil
}

func (m *memStore) GetRefreshToken(userID string) (*oauth2.Token, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	token, found := m.refreshTokens[userID]
	return token, found, nil
}

func (m *memStore) DeleteRefreshToken(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.refreshTokens, userID)
	return nil
}

func (m *memStore) Close() error {
	return nil
}

var _ storage.Store = (*memStore)(nil)
