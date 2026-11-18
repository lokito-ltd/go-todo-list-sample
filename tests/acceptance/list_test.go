package acceptance_test

import (
	"bytes"
	"os"
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
	t.Run(`Each incomplete item is shown on its own line as its number, ". [ ] " and its title. [AC0001.1]`, func(t *testing.T) {
		useFreshItemsFile(t)
		want := "1. [ ] Buy milk\n" +
			"2. [ ] Take cat to vet\n" +
			"3. [ ] Print forms at library\n" +
			"4. [ ] Wash the car\n"
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

	t.Run("Items are numbered from 1 in the order they are listed. [AC0001.3]", func(t *testing.T) {
		path := useFreshItemsFile(t)
		// A done item between two open ones gets no number and leaves no gap.
		items := `[{"title":"First","done":false},{"title":"Finished","done":true},{"title":"Second","done":false}]`
		if err := os.WriteFile(path, []byte(items), 0o644); err != nil {
			t.Fatal(err)
		}
		if got, want := listItems(t), "1. [ ] First\n2. [ ] Second\n"; got != want {
			t.Errorf("list = %q, want %q", got, want)
		}
	})
}
