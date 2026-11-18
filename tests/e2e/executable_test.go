package e2e_test

import (
	"bytes"
	"errors"
	"os"
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

// freshItemsFile is a path to an items file that does not exist yet, so the
// executable never reads or writes the real one in the home folder.
func freshItemsFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "items.json")
}

func runBinary(t *testing.T, binary, itemsFile string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := exec.Command(binary, args...)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "GO_TODO_FILE="+itemsFile)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	return out.String(), errOut.String(), err
}

func TestStandaloneExecutableListsIncompleteItems(t *testing.T) {
	stdout, stderr, err := runBinary(t, buildBinary(t), freshItemsFile(t), "list")
	if err != nil {
		t.Fatalf("go-todo list: %v (stderr %q)", err, stderr)
	}
	want := "1. [ ] Buy milk\n2. [ ] Take cat to vet\n3. [ ] Print forms at library\n4. [ ] Wash the car\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Errorf("unexpected stderr: %q", stderr)
	}
}

func TestStandaloneExecutableKeepsAddedItemsBetweenRuns(t *testing.T) {
	binary, itemsFile := buildBinary(t), freshItemsFile(t)
	stdout, stderr, err := runBinary(t, binary, itemsFile, "add", "Buy", "bread")
	if err != nil {
		t.Fatalf("go-todo add: %v (stderr %q)", err, stderr)
	}
	if want := "5. [ ] Buy bread\n"; stdout != want {
		t.Errorf("add stdout = %q, want %q", stdout, want)
	}
	stdout, stderr, err = runBinary(t, binary, itemsFile, "list")
	if err != nil {
		t.Fatalf("go-todo list: %v (stderr %q)", err, stderr)
	}
	want := "1. [ ] Buy milk\n2. [ ] Take cat to vet\n3. [ ] Print forms at library\n4. [ ] Wash the car\n5. [ ] Buy bread\n"
	if stdout != want {
		t.Errorf("list stdout = %q, want %q", stdout, want)
	}
}

func TestStandaloneExecutableKeepsCompletedItemsBetweenRuns(t *testing.T) {
	binary, itemsFile := buildBinary(t), freshItemsFile(t)
	stdout, stderr, err := runBinary(t, binary, itemsFile, "done", "2")
	if err != nil {
		t.Fatalf("go-todo done: %v (stderr %q)", err, stderr)
	}
	if want := "[x] Take cat to vet\n"; stdout != want {
		t.Errorf("done stdout = %q, want %q", stdout, want)
	}
	stdout, stderr, err = runBinary(t, binary, itemsFile, "list")
	if err != nil {
		t.Fatalf("go-todo list: %v (stderr %q)", err, stderr)
	}
	want := "1. [ ] Buy milk\n2. [ ] Print forms at library\n3. [ ] Wash the car\n"
	if stdout != want {
		t.Errorf("list stdout = %q, want %q", stdout, want)
	}
}

func TestStandaloneExecutableWithoutCommandShowsUsage(t *testing.T) {
	stdout, stderr, err := runBinary(t, buildBinary(t), freshItemsFile(t))
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 {
		t.Fatalf("go-todo: err = %v, want exit status 2", err)
	}
	if want := "usage: go-todo list\n       go-todo add <title>\n       go-todo done <number>\n"; stderr != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}
	if stdout != "" {
		t.Errorf("unexpected stdout: %q", stdout)
	}
}
