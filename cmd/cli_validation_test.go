package cmd_test

import (
	"os/exec"
	"testing"
)

func TestTrackCommand_NameTooLong(t *testing.T) {
	cmd := exec.Command("../habits", "track", "averyveryverylonghabitnamethatexceedslimit", "note")
	err := cmd.Run()
	if err == nil {
		t.Error("Expected error for long habit name")
	}
}
