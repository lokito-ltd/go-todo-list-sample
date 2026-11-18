package e2e_test

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "go-todo")
	build := exec.Command("go", "build", "-o", binary, "../../cmd/go-todo")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, output)
	}
	return binary
}

func runBinary(t *testing.T, binary string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := exec.Command(binary, args...)
	cmd.Dir = t.TempDir()
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	return out.String(), errOut.String(), err
}

func TestStandaloneExecutableListsIncompleteItems(t *testing.T) {
	stdout, stderr, err := runBinary(t, buildBinary(t), "list")
	if err != nil {
		t.Fatalf("go-todo list: %v (stderr %q)", err, stderr)
	}
	want := "[ ] Buy milk\n[ ] Take cat to vet\n[ ] Print forms at library\n[ ] Wash the car\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Errorf("unexpected stderr: %q", stderr)
	}
}

func TestStandaloneExecutableWithoutCommandShowsUsage(t *testing.T) {
	stdout, stderr, err := runBinary(t, buildBinary(t))
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 {
		t.Fatalf("go-todo: err = %v, want exit status 2", err)
	}
	if want := "usage: go-todo list\n"; stderr != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}
	if stdout != "" {
		t.Errorf("unexpected stdout: %q", stdout)
	}
}
