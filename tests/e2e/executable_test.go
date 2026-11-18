package e2e_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestStandaloneExecutable(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "go-todo")
	build := exec.Command("go", "build", "-o", binary, "../../cmd/go-todo")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, output)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(binary)
	cmd.Dir = t.TempDir()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go-todo: %v (stderr %q)", err, stderr.String())
	}
	if got, want := stdout.String(), "[ ] Hello World\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("unexpected stderr: %q", stderr.String())
	}
}
