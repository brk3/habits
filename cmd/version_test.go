package cmd_test

import (
	"brk3.github.io/habits/cmd"
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
