package cli

import (
	"bytes"
	"io"
	"testing"
)

func TestRunListPrintsIncompleteItems(t *testing.T) {
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

func TestRunListSkipsDoneItems(t *testing.T) {
	var stdout bytes.Buffer
	data := []byte(`[{"title":"Open","done":false},{"title":"Closed","done":true}]`)
	if status := run([]string{"list"}, data, &stdout, io.Discard); status != 0 {
		t.Fatalf("exit status = %d, want 0", status)
	}
	if got, want := stdout.String(), "[ ] Open\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunReportsWriteFailure(t *testing.T) {
	var stderr bytes.Buffer
	reader, writer := io.Pipe()
	reader.Close()
	defer writer.Close()
	status := Run([]string{"list"}, writer, &stderr)
	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	if got, want := stderr.String(), "go-todo: write output: "+io.ErrClosedPipe.Error()+"\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestRunReportsUnreadableItems(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := run([]string{"list"}, []byte(`{not json`), &stdout, &stderr)
	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	if got, prefix := stderr.String(), "go-todo: read items: "; len(got) <= len(prefix) || got[:len(prefix)] != prefix {
		t.Errorf("stderr = %q, want it to start with %q", got, prefix)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsBadUsage(t *testing.T) {
	cases := map[string][]string{
		"no command":      nil,
		"unknown command": {"add"},
		"extra argument":  {"list", "extra"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if status := Run(args, &stdout, &stderr); status != 2 {
				t.Errorf("exit status = %d, want 2", status)
			}
			if got, want := stderr.String(), "usage: go-todo list\n"; got != want {
				t.Errorf("stderr = %q, want %q", got, want)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}
