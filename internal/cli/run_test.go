package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const wantUsage = "usage: go-todo list\n       go-todo add <title>\n"

// itemsFile returns a path in a fresh temporary folder, holding contents when
// contents is not empty.
func itemsFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "items.json")
	if contents != "" {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRunListPrintsBuiltInItemsWithoutAFile(t *testing.T) {
	t.Setenv("GO_TODO_FILE", itemsFile(t, ""))
	var stdout, stderr bytes.Buffer
	if status := Run([]string{"list"}, &stdout, &stderr); status != 0 {
		t.Fatalf("exit status = %d, want 0 (stderr %q)", status, stderr.String())
	}
	want := "[ ] Buy milk\n[ ] Take cat to vet\n[ ] Print forms at library\n[ ] Wash the car\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunListReadsTheItemsFile(t *testing.T) {
	path := itemsFile(t, `[{"title":"Open","done":false},{"title":"Closed","done":true}]`)
	var stdout bytes.Buffer
	if status := run([]string{"list"}, path, &stdout, io.Discard); status != 0 {
		t.Fatalf("exit status = %d, want 0", status)
	}
	if got, want := stdout.String(), "[ ] Open\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunReportsWriteFailure(t *testing.T) {
	cases := map[string][]string{
		"list": {"list"},
		"add":  {"add", "New"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			var stderr bytes.Buffer
			reader, writer := io.Pipe()
			reader.Close()
			defer writer.Close()
			status := run(args, itemsFile(t, ""), writer, &stderr)
			if status != 1 {
				t.Errorf("exit status = %d, want 1", status)
			}
			if got, want := stderr.String(), "go-todo: write output: "+io.ErrClosedPipe.Error()+"\n"; got != want {
				t.Errorf("stderr = %q, want %q", got, want)
			}
		})
	}
}

func TestRunReportsUnreadableItems(t *testing.T) {
	cases := map[string][]string{
		"list": {"list"},
		"add":  {"add", "New"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			path := itemsFile(t, `{not json`)
			var stdout, stderr bytes.Buffer
			status := run(args, path, &stdout, &stderr)
			if status != 1 {
				t.Errorf("exit status = %d, want 1", status)
			}
			if got, prefix := stderr.String(), "go-todo: read items: "; len(got) <= len(prefix) || !strings.HasPrefix(got, prefix) {
				t.Errorf("stderr = %q, want it to start with %q", got, prefix)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if got := readFile(t, path); got != `{not json` {
				t.Errorf("file = %q, want it unchanged", got)
			}
		})
	}
}

func TestRunAddAppendsAnIncompleteItem(t *testing.T) {
	path := itemsFile(t, `[{"title":"Old","done":true}]`)
	var stdout, stderr bytes.Buffer
	if status := run([]string{"add", "New"}, path, &stdout, &stderr); status != 0 {
		t.Fatalf("exit status = %d, want 0 (stderr %q)", status, stderr.String())
	}
	if got, want := stdout.String(), "[ ] New\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
	want := "[\n  {\n    \"title\": \"Old\",\n    \"done\": true\n  },\n  {\n    \"title\": \"New\",\n    \"done\": false\n  }\n]\n"
	if got := readFile(t, path); got != want {
		t.Errorf("file = %q, want %q", got, want)
	}
}

func TestRunAddWithoutAFileKeepsTheBuiltInItems(t *testing.T) {
	path := itemsFile(t, "")
	if status := run([]string{"add", "New"}, path, io.Discard, io.Discard); status != 0 {
		t.Fatalf("exit status = %d, want 0", status)
	}
	var stdout bytes.Buffer
	if status := run([]string{"list"}, path, &stdout, io.Discard); status != 0 {
		t.Fatalf("list exit status = %d, want 0", status)
	}
	want := "[ ] Buy milk\n[ ] Take cat to vet\n[ ] Print forms at library\n[ ] Wash the car\n[ ] New\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunAddJoinsWordsWithSingleSpaces(t *testing.T) {
	cases := map[string][]string{
		"separate words":    {"add", "Buy", "bread"},
		"one quoted string": {"add", "Buy bread"},
		"extra spaces":      {"add", "  Buy ", "  bread  "},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			var stdout bytes.Buffer
			if status := run(args, itemsFile(t, "[]"), &stdout, io.Discard); status != 0 {
				t.Fatalf("exit status = %d, want 0", status)
			}
			if got, want := stdout.String(), "[ ] Buy bread\n"; got != want {
				t.Errorf("stdout = %q, want %q", got, want)
			}
		})
	}
}

func TestRunReportsSaveFailure(t *testing.T) {
	// The file can be read (it is missing) but not written.
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o700) })
	var stdout, stderr bytes.Buffer
	status := run([]string{"add", "New"}, filepath.Join(dir, "items.json"), &stdout, &stderr)
	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	if got, prefix := stderr.String(), "go-todo: save items: "; len(got) <= len(prefix) || !strings.HasPrefix(got, prefix) {
		t.Errorf("stderr = %q, want it to start with %q", got, prefix)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunReportsMissingHome(t *testing.T) {
	t.Setenv("GO_TODO_FILE", "")
	t.Setenv("HOME", "")
	var stdout, stderr bytes.Buffer
	status := Run([]string{"list"}, &stdout, &stderr)
	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	if got, prefix := stderr.String(), "go-todo: find items file: "; len(got) <= len(prefix) || !strings.HasPrefix(got, prefix) {
		t.Errorf("stderr = %q, want it to start with %q", got, prefix)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsBadUsage(t *testing.T) {
	cases := map[string][]string{
		"no command":           nil,
		"unknown command":      {"remove"},
		"extra argument":       {"list", "extra"},
		"add without title":    {"add"},
		"add with blank title": {"add", "  ", ""},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			path := itemsFile(t, `[{"title":"Keep","done":false}]`)
			var stdout, stderr bytes.Buffer
			if status := run(args, path, &stdout, &stderr); status != 2 {
				t.Errorf("exit status = %d, want 2", status)
			}
			if got := stderr.String(); got != wantUsage {
				t.Errorf("stderr = %q, want %q", got, wantUsage)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if got, want := readFile(t, path), `[{"title":"Keep","done":false}]`; got != want {
				t.Errorf("file = %q, want it unchanged", got)
			}
		})
	}
}
