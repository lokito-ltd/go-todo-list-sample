package cli

import (
	"bytes"
	"io"
	"testing"
)

func TestRunWritesItemLine(t *testing.T) {
	var stdout bytes.Buffer
	if status := Run(&stdout, io.Discard); status != 0 {
		t.Fatalf("exit status = %d, want 0", status)
	}
	if got, want := stdout.String(), "[ ] Hello World\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunReportsWriteFailure(t *testing.T) {
	var stderr bytes.Buffer
	reader, writer := io.Pipe()
	reader.Close()
	defer writer.Close()
	status := Run(writer, &stderr)
	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	if got, want := stderr.String(), "go-todo: write output: "+io.ErrClosedPipe.Error()+"\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}
