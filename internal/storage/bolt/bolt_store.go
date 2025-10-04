package bolt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/brk3/habits/internal/logger"
	"github.com/brk3/habits/internal/storage"
	"github.com/brk3/habits/pkg/habit"
	"go.etcd.io/bbolt"
)

const rootBucket = "users"

type Store struct {
	db *bbolt.DB
}

func Open(path string) (*Store, error) {
	logger.Debug("Opening BoltDB", "path", path)
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		logger.Error("Failed to open BoltDB", "path", path, "error", err)
		return nil, err
	}

	s := &Store{db: db}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(rootBucket))
		return err
	}); err != nil {
		logger.Error("Failed to create root bucket", "bucket", rootBucket, "error", err)
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("Failed to close database after bucket creation error", "error", closeErr)
		}
		return nil, err
	}

	logger.Info("BoltDB opened successfully", "path", path)
	return s, nil
}

func (s *Store) ensureUserHabitsBucketExists(userID string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		usersBucket := tx.Bucket([]byte(rootBucket))
		if usersBucket == nil {
			return fmt.Errorf("root bucket does not exist")
		}

		userBucket, err := usersBucket.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}

		_, err = userBucket.CreateBucketIfNotExists([]byte("habits"))
		return err
	})
}

func (s *Store) ensureAPIKeyBucketExists() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		apiKeysBucket := tx.Bucket([]byte(rootBucket))
		if apiKeysBucket == nil {
			return fmt.Errorf("root bucket does not exist")
		}

		apiKeysBucket, err := apiKeysBucket.CreateBucketIfNotExists([]byte("api_keys"))
		return err
	})
}

func (s *Store) getUserHabitsBucket(tx *bbolt.Tx, userID string) (*bbolt.Bucket, error) {
	usersBucket := tx.Bucket([]byte(rootBucket))
	if usersBucket == nil {
		return nil, fmt.Errorf("root bucket does not exist")
	}

	userBucket := usersBucket.Bucket([]byte(userID))
	if userBucket == nil {
		return nil, fmt.Errorf("user bucket for %s does not exist", userID)
	}

	habitsBucket := userBucket.Bucket([]byte("habits"))
	if habitsBucket == nil {
		return nil, fmt.Errorf("habits bucket for %s does not exist", userID)
	}
	return habitsBucket, nil
}

func (s *Store) Close() error {
	logger.Debug("Closing BoltDB")
	err := s.db.Close()
	if err != nil {
		logger.Error("Failed to close BoltDB", "error", err)
		return fmt.Errorf("failed to close database: %w", err)
	}
	logger.Debug("BoltDB closed successfully")
	return nil
}

func (s *Store) PutHabit(userID string, h habit.Habit) error {
	logger.Debug("Storing habit", "user_id", userID, "habit_name", h.Name)
	if err := s.ensureUserHabitsBucketExists(userID); err != nil {
		logger.Error("Failed to ensure bucket exists", "user_id", userID, "error", err)
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.getUserHabitsBucket(tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user habits bucket: %w", err)
		}
		val, err := json.Marshal(h)
		if err != nil {
			return fmt.Errorf("failed to marshal habit %s: %w", h.Name, err)
		}
		key := fmt.Appendf(nil, "%s/%s", h.Name, time.Unix(h.TimeStamp, 0).Format(time.RFC3339))
		err = bucket.Put(key, val)
		if err != nil {
			return fmt.Errorf("failed to store habit %s: %w", h.Name, err)
		}
		logger.Debug("Habit stored successfully", "key", string(key))
		return nil
	})
}

func (s *Store) ListHabitNames(userID string) ([]string, error) {
	if err := s.ensureUserHabitsBucketExists(userID); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists for user %s: %w", userID, err)
	}
	uniq := map[string]struct{}{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := s.getUserHabitsBucket(tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user habits bucket for listing: %w", err)
		}
		return bucket.ForEach(func(k, _ []byte) error {
			name := strings.SplitN(string(k), "/", 2)[0]
			uniq[name] = struct{}{}
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list habit names for user %s: %w", userID, err)
	}
	out := make([]string, 0, len(uniq))
	for n := range uniq {
		out = append(out, n)
	}
	return out, nil
}

func (s *Store) GetHabit(userID, name string) ([]habit.Habit, error) {
	if err := s.ensureUserHabitsBucketExists(userID); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists for user %s: %w", userID, err)
	}
	var out []habit.Habit
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := s.getUserHabitsBucket(tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user habits bucket for retrieval: %w", err)
		}
		c := bucket.Cursor()
		prefix := []byte(name + "/")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var e habit.Habit
			if err := json.Unmarshal(v, &e); err != nil {
				return fmt.Errorf("failed to unmarshal habit entry for %s: %w", name, err)
			}
			out = append(out, e)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get habit %s for user %s: %w", name, userID, err)
	}
	return out, nil
}

func (s *Store) DeleteHabit(userID, name string) error {
	if err := s.ensureUserHabitsBucketExists(userID); err != nil {
		return fmt.Errorf("failed to ensure bucket exists for user %s: %w", userID, err)
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.getUserHabitsBucket(tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user habits bucket for deletion: %w", err)
		}
		c := bucket.Cursor()
		prefix := []byte(name + "/")
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := c.Delete(); err != nil {
				return fmt.Errorf("failed to delete habit entry %s: %w", string(k), err)
			}
		}
		return nil
	})
}

func (s *Store) PutAPIKey(keyHash, userID string) error {
	if err := s.ensureAPIKeyBucketExists(); err != nil {
		return fmt.Errorf("failed to ensure API key bucket exists: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(rootBucket)).Bucket([]byte("api_keys"))
		if bucket == nil {
			return fmt.Errorf("api_keys bucket not found")
		}

		err := bucket.Put([]byte(keyHash), []byte(userID))
		if err != nil {
			return fmt.Errorf("failed to store API key: %w", err)
		}

		hashPreview := keyHash
		if len(hashPreview) > 8 {
			hashPreview = hashPreview[:8] + "..."
		}
		logger.Debug("API key stored", "keyHash", hashPreview, "userID", userID)
		return nil
	})
}

func (s *Store) GetAPIKey(keyHash string) (string, bool, error) {
	if err := s.ensureAPIKeyBucketExists(); err != nil {
		return "", false, fmt.Errorf("failed to ensure API key bucket exists: %w", err)
	}

	var userID string
	var found bool
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(rootBucket)).Bucket([]byte("api_keys"))
		if bucket == nil {
			return fmt.Errorf("api_keys bucket not found")
		}

		userIDBytes := bucket.Get([]byte(keyHash))
		if userIDBytes != nil {
			userID = string(userIDBytes)
			found = true
		}
		return nil
	})

	return userID, found, err
}

func (s *Store) ListAPIKeyHashes(userID string) ([]string, error) {
	if err := s.ensureAPIKeyBucketExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure API key bucket exists: %w", err)
	}

	var hashes []string
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(rootBucket)).Bucket([]byte("api_keys"))
		if bucket == nil {
			return fmt.Errorf("api_keys bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			storedUserID := string(v)
			if storedUserID == userID {
				hashes = append(hashes, string(k))
			}
			return nil
		})
	})

	return hashes, err
}

func (s *Store) DeleteAPIKey(keyHash string) error {
	if err := s.ensureAPIKeyBucketExists(); err != nil {
		return fmt.Errorf("failed to ensure API key bucket exists: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(rootBucket)).Bucket([]byte("api_keys"))
		if bucket == nil {
			return fmt.Errorf("api_keys bucket not found")
		}
		return bucket.Delete([]byte(keyHash))
	})
}

var _ storage.Store = (*Store)(nil)
