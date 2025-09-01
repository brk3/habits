package bolt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/brk3/habits/internal/storage"
	"github.com/brk3/habits/pkg/habit"
	"go.etcd.io/bbolt"
)

type Store struct {
	db     *bbolt.DB
	bucket []byte
}

func Open(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db, bucket: []byte("habits")}
	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(s.bucket)
		return err
	}); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) PutHabit(h habit.Habit) error {
	val, _ := json.Marshal(h)
	key := fmt.Appendf(nil, "%s/%s", h.Name, time.Unix(h.TimeStamp, 0).Format(time.RFC3339))
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(s.bucket).Put(key, val)
	})
}

func (s *Store) ListHabitNames() ([]string, error) {
	uniq := map[string]struct{}{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(s.bucket).ForEach(func(k, _ []byte) error {
			name := strings.SplitN(string(k), "/", 2)[0]
			uniq[name] = struct{}{}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(uniq))
	for n := range uniq {
		out = append(out, n)
	}
	return out, nil
}

func (s *Store) GetHabit(name string) ([]habit.Habit, error) {
	var out []habit.Habit
	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(s.bucket).Cursor()
		prefix := []byte(name + "/")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var e habit.Habit
			if err := json.Unmarshal(v, &e); err != nil {
				return err
			}
			out = append(out, e)
		}
		return nil
	})
	return out, err
}

func (s *Store) DeleteHabit(name string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		c := tx.Bucket(s.bucket).Cursor()
		prefix := []byte(name + "/")
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := c.Delete(); err != nil {
				return err
			}
		}
		return nil
	})
}

// compile-time check
var _ storage.Store = (*Store)(nil)
