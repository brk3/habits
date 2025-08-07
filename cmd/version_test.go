package cmd_test

import (
	"github.com/brk3/habits/cmd"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVersionEndpointFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()
	cmd.GetVersionInfo(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)

	if !strings.Contains(string(body), "Version") {
		t.Error("Expected version info in response")
	}
}
