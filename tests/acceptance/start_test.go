package acceptance_test

import (
	"bytes"
	"testing"

	"example.com/go-todo/internal/cli"
)

func TestStartingAppShowsOneUncheckedItem(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := cli.Run(&stdout, &stderr)
	if status != 0 || stderr.Len() != 0 {
		t.Fatalf("startup returned %d with stderr %q", status, stderr.String())
	}
	if got, want := stdout.String(), "[ ] Hello World\n"; got != want {
		t.Errorf("list = %q, want %q", got, want)
	}
}
