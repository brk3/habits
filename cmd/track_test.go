package cmd_test

import (
	"os/exec"
	"testing"
)

func TestTrackCommand_InvalidArgs(t *testing.T) {
	cmd := exec.Command("../habits", "track")
	err := cmd.Run()
	if err == nil {
		t.Error("Expected error due to missing args")
	}
}
