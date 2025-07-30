package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTrackCommand_Output(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	rootCmd.SetArgs([]string{"track", "guitar"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var h Habit
	if err := json.Unmarshal(out.Bytes(), &h); err != nil {
		t.Fatalf("Expected valid JSON output, got error: %v", err)
	}

	if h.Content != "guitar" {
		t.Errorf("Expected content to be 'guitar', got '%s'", h.Content)
	}
	if h.TimeStamp.IsZero() {
		t.Errorf("Expected a timestamp value, got empty string")
	}
}
