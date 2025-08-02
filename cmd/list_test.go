package cmd_test

import (
	"testing"
)

func TestListHabits_Empty(t *testing.T) {
	//db, _ := bbolt.Open("test.db", 0600, nil)
	//defer db.Close()
	//defer func() { _ = db.Update(func(tx *bbolt.Tx) error { return tx.DeleteBucket([]byte("habits")) }) }()

	//cmd.InjectDB(db)

	//req := httptest.NewRequest("GET", "/habits", nil)
	//w := httptest.NewRecorder()
	//cmd.ListHabits(w, req)

	//if w.Code != 200 {
	//	t.Errorf("Expected 200 OK, got %d", w.Code)
	//	}
}
