package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTrackCommand_Output(t *testing.T) {
	// Capture output
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	// Provide test args
	rootCmd.SetArgs([]string{"track", "drink water"})

	// Run the root command (which includes 'track')
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check output is valid JSON
	var h Habit
	if err := json.Unmarshal(out.Bytes(), &h); err != nil {
		t.Fatalf("Expected valid JSON output, got error: %v", err)
	}

	// Basic content check
	if h.Content != "drink water" {
		t.Errorf("Expected content to be 'drink water', got '%s'", h.Content)
	}
	if h.TimeStamp.IsZero() {
		t.Errorf("Expected a timestamp value, got empty string")
	}
}
