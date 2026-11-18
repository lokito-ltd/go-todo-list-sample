package acceptance_test

import (
	"os"
	"testing"
)

const usage = "usage: go-todo list\n       go-todo add <title>\n       go-todo done <number>\n"

func completeItem(t *testing.T, number string) {
	t.Helper()
	stdout, stderr, status := runCommand(t, "done", number)
	if status != 0 || stderr != "" {
		t.Fatalf("done %s returned %d with stderr %q", number, status, stderr)
	}
	if stdout == "" {
		t.Errorf("done %s printed nothing", number)
	}
}

func readItemsFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestCompletingItems(t *testing.T) {
	t.Run("A completed item is no longer shown on the list. [AC0003.1]", func(t *testing.T) {
		useFreshItemsFile(t)
		completeItem(t, "2")
		want := "1. [ ] Buy milk\n" +
			"2. [ ] Print forms at library\n" +
			"3. [ ] Wash the car\n"
		if got := listItems(t); got != want {
			t.Errorf("list = %q, want %q", got, want)
		}
	})

	t.Run("The other items keep their order after one is completed. [AC0003.2]", func(t *testing.T) {
		useFreshItemsFile(t)
		if _, _, status := runCommand(t, "add", "Water the plants"); status != 0 {
			t.Fatalf("setup add returned %d", status)
		}
		completeItem(t, "1")
		completeItem(t, "3")
		want := "1. [ ] Take cat to vet\n" +
			"2. [ ] Print forms at library\n" +
			"3. [ ] Water the plants\n"
		if got := listItems(t); got != want {
			t.Errorf("list = %q, want %q", got, want)
		}
	})

	t.Run("Completing a number that is not on the list is refused and leaves the list unchanged. [AC0003.3]", func(t *testing.T) {
		path := useFreshItemsFile(t)
		completeItem(t, "4")
		before := readItemsFile(t, path)
		// Four built-in items are open, one of them was completed: 3 is the last.
		stdout, stderr, status := runCommand(t, "done", "4")
		if status != 1 {
			t.Errorf("done 4 returned %d, want 1", status)
		}
		if want := "go-todo: no item 4 on the list\n"; stderr != want {
			t.Errorf("stderr = %q, want %q", stderr, want)
		}
		if stdout != "" {
			t.Errorf("stdout = %q, want empty", stdout)
		}
		if after := readItemsFile(t, path); after != before {
			t.Errorf("items file changed from %q to %q", before, after)
		}
	})

	t.Run("Completing without a single whole number is refused and leaves the list unchanged. [AC0003.4]", func(t *testing.T) {
		path := useFreshItemsFile(t)
		if _, _, status := runCommand(t, "add", "Existing"); status != 0 {
			t.Fatalf("setup add returned %d", status)
		}
		before := readItemsFile(t, path)
		for _, args := range [][]string{{"done"}, {"done", "two"}, {"done", "0"}, {"done", "-1"}, {"done", "1", "2"}, {"done", "1.5"}} {
			stdout, stderr, status := runCommand(t, args...)
			if status != 2 {
				t.Errorf("%q returned %d, want 2", args, status)
			}
			if stderr != usage {
				t.Errorf("%q stderr = %q, want %q", args, stderr, usage)
			}
			if stdout != "" {
				t.Errorf("%q stdout = %q, want empty", args, stdout)
			}
		}
		if after := readItemsFile(t, path); after != before {
			t.Errorf("items file changed from %q to %q", before, after)
		}
	})
}
