package server

import (
	"sync"

	"github.com/brk3/habits/internal/storage"
	"github.com/brk3/habits/pkg/habit"
)

type memStore struct {
	mu   sync.RWMutex
	data map[string][]habit.Habit
}

func newMemStore() *memStore {
	return &memStore{data: map[string][]habit.Habit{}}
}

func (m *memStore) PutHabit(userID string, h habit.Habit) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[h.Name] = append(m.data[h.Name], h)

	return nil
}

func (m *memStore) ListHabitNames(userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := []string{}
	for habitKey := range m.data {
		out = append(out, habitKey)
	}

	return out, nil
}

func (m *memStore) GetHabit(userID, name string) ([]habit.Habit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]habit.Habit(nil), m.data[name]...), nil
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

	delete(m.data, name)
	return nil
}

func (m *memStore) PutAPIKey(key, userID string) error {
	return nil // TODO
}

func (m *memStore) GetAPIKey(key string) (string, error) {
	return "", nil // TODO
}

func (m *memStore) Close() error {
	return nil
}

var _ storage.Store = (*memStore)(nil)
