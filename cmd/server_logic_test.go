package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"brk3.github.io/habits/cmd"
)

func TestTrackHabit_EmptyName(t *testing.T) {
	h := &cmd.Habit{
		Name:      "",
		Note:      "note",
		TimeStamp: time.Now().Unix(),
	}
	data, _ := json.Marshal(h)
	req := httptest.NewRequest("POST", "/habits", bytes.NewReader(data))
	w := httptest.NewRecorder()

	cmd.TrackHabit(w, req)

	if w.Code != 400 {
		t.Errorf("Expected 400 for missing name, got %d", w.Code)
	}
}
