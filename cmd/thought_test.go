package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestThoughtCommand_Output(t *testing.T) {
	aThought := "reflecting on progress"

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	rootCmd.SetArgs([]string{"thought", aThought})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var h Thought
	if err := json.Unmarshal(out.Bytes(), &h); err != nil {
		t.Fatalf("Expected valid JSON output, got error: %v", err)
	}

	if h.Content != aThought {
		t.Errorf("Expected content to be %q, got %q", aThought, h.Content)
	}
	if h.TimeStamp.IsZero() {
		t.Errorf("Expected a timestamp value, got empty value")
	}
}
