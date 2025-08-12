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

func (m *memStore) Put(h habit.Habit) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[h.Name] = append(m.data[h.Name], h)

	return nil
}

func (m *memStore) ListHabitNames() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := []string{}
	for habitKey := range m.data {
		out = append(out, habitKey)
	}

	return out, nil
}

func (m *memStore) ListEntriesByHabit(name string) ([]habit.Habit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]habit.Habit(nil), m.data[name]...), nil
}

func (m *memStore) Close() error {
	return nil
}

var _ storage.Store = (*memStore)(nil)
