package acceptance_test

import (
	"os"
	"testing"
)

func TestAddingItems(t *testing.T) {
	t.Run("An added item is shown as incomplete at the end of the list. [AC0002.1]", func(t *testing.T) {
		useFreshItemsFile(t)
		before := listItems(t)
		stdout, stderr, status := runCommand(t, "add", "Water the plants")
		if status != 0 || stderr != "" {
			t.Fatalf("add returned %d with stderr %q", status, stderr)
		}
		if want := "[ ] Water the plants\n"; stdout != want {
			t.Errorf("add printed %q, want %q", stdout, want)
		}
		if got, want := listItems(t), before+"[ ] Water the plants\n"; got != want {
			t.Errorf("list = %q, want %q", got, want)
		}
	})

	t.Run("The words given after the command form the title, separated by single spaces. [AC0002.2]", func(t *testing.T) {
		useFreshItemsFile(t)
		if _, stderr, status := runCommand(t, "add", "Call", "  the ", "plumber"); status != 0 {
			t.Fatalf("add returned %d with stderr %q", status, stderr)
		}
		if _, stderr, status := runCommand(t, "add", "Pay the  rent"); status != 0 {
			t.Fatalf("add returned %d with stderr %q", status, stderr)
		}
		want := "[ ] Buy milk\n" +
			"[ ] Take cat to vet\n" +
			"[ ] Print forms at library\n" +
			"[ ] Wash the car\n" +
			"[ ] Call the plumber\n" +
			"[ ] Pay the rent\n"
		if got := listItems(t); got != want {
			t.Errorf("list = %q, want %q", got, want)
		}
	})

	t.Run("Adding with no title is refused and leaves the list unchanged. [AC0002.3]", func(t *testing.T) {
		path := useFreshItemsFile(t)
		if _, _, status := runCommand(t, "add", "Existing"); status != 0 {
			t.Fatalf("setup add returned %d", status)
		}
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"add"}, {"add", " ", ""}} {
			stdout, stderr, status := runCommand(t, args...)
			if status != 2 {
				t.Errorf("%q returned %d, want 2", args, status)
			}
			if want := "usage: go-todo list\n       go-todo add <title>\n"; stderr != want {
				t.Errorf("%q stderr = %q, want %q", args, stderr, want)
			}
			if stdout != "" {
				t.Errorf("%q stdout = %q, want empty", args, stdout)
			}
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(after) != string(before) {
			t.Errorf("items file changed from %q to %q", before, after)
		}
	})
}
