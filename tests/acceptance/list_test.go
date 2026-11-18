package acceptance_test

import (
	"bytes"
	"strings"
	"testing"

	"example.com/go-todo/internal/cli"
)

func listItems(t *testing.T) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"list"}, &stdout, &stderr)
	if status != 0 || stderr.Len() != 0 {
		t.Fatalf("list returned %d with stderr %q", status, stderr.String())
	}
	return stdout.String()
}

func TestListingItems(t *testing.T) {
	t.Run(`Each incomplete item is shown as "[ ] " followed by its title, one per line. [AC0001.1]`, func(t *testing.T) {
		want := "[ ] Buy milk\n" +
			"[ ] Take cat to vet\n" +
			"[ ] Print forms at library\n" +
			"[ ] Wash the car\n"
		if got := listItems(t); got != want {
			t.Errorf("list = %q, want %q", got, want)
		}
	})

	t.Run("Done items are not shown. [AC0001.2]", func(t *testing.T) {
		if got := listItems(t); strings.Contains(got, "Get a haircut") {
			t.Errorf("list shows a done item: %q", got)
		}
	})
}
