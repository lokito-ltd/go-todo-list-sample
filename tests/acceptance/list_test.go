package acceptance_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"example.com/go-todo/internal/cli"
)

// useFreshItemsFile points go-todo at an items file that does not exist yet,
// so a test never reads or writes the real one in the home folder.
func useFreshItemsFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "items.json")
	t.Setenv("GO_TODO_FILE", path)
	return path
}

func runCommand(t *testing.T, args ...string) (stdout, stderr string, status int) {
	t.Helper()
	var out, errOut bytes.Buffer
	status = cli.Run(args, &out, &errOut)
	return out.String(), errOut.String(), status
}

func listItems(t *testing.T) string {
	t.Helper()
	stdout, stderr, status := runCommand(t, "list")
	if status != 0 || stderr != "" {
		t.Fatalf("list returned %d with stderr %q", status, stderr)
	}
	return stdout
}

func TestListingItems(t *testing.T) {
	t.Run(`Each incomplete item is shown as "[ ] " followed by its title, one per line. [AC0001.1]`, func(t *testing.T) {
		useFreshItemsFile(t)
		want := "[ ] Buy milk\n" +
			"[ ] Take cat to vet\n" +
			"[ ] Print forms at library\n" +
			"[ ] Wash the car\n"
		if got := listItems(t); got != want {
			t.Errorf("list = %q, want %q", got, want)
		}
	})

	t.Run("Done items are not shown. [AC0001.2]", func(t *testing.T) {
		useFreshItemsFile(t)
		if got := listItems(t); strings.Contains(got, "Get a haircut") {
			t.Errorf("list shows a done item: %q", got)
		}
	})
}
