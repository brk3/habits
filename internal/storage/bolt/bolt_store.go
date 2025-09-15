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

const rootBucket = "users"
const defaultUserID = "default"

type Store struct {
	db *bbolt.DB
}

func Open(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(rootBucket))
		return err
	}); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) getUserHabitsBucket(tx *bbolt.Tx, userID string) (*bbolt.Bucket, error) {
	if userID == "" {
		userID = defaultUserID
	}

	usersBucket := tx.Bucket([]byte(rootBucket))
	userBucket, err := usersBucket.CreateBucketIfNotExists([]byte(userID))
	if err != nil {
		return nil, err
	}
	return userBucket.CreateBucketIfNotExists([]byte("habits"))
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) PutHabit(userID string, h habit.Habit) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.getUserHabitsBucket(tx, userID)
		if err != nil {
			return err
		}
		val, _ := json.Marshal(h)
		key := fmt.Appendf(nil, "%s/%s", h.Name, time.Unix(h.TimeStamp, 0).Format(time.RFC3339))
		return bucket.Put(key, val)
	})
}

func (s *Store) ListHabitNames(userID string) ([]string, error) {
	uniq := map[string]struct{}{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := s.getUserHabitsBucket(tx, userID)
		if err != nil {
			return err
		}
		return bucket.ForEach(func(k, _ []byte) error {
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

func (s *Store) GetHabit(userID, name string) ([]habit.Habit, error) {
	var out []habit.Habit
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := s.getUserHabitsBucket(tx, userID)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
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

func (s *Store) DeleteHabit(userID, name string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.getUserHabitsBucket(tx, userID)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
		prefix := []byte(name + "/")
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := c.Delete(); err != nil {
				return err
			}
		}
		return nil
	})
}

var _ storage.Store = (*Store)(nil)
