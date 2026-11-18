package todo

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPathUsesTheEnvironmentVariable(t *testing.T) {
	t.Setenv("GO_TODO_FILE", "/somewhere/items.json")
	got, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if got != "/somewhere/items.json" {
		t.Errorf("Path = %q, want %q", got, "/somewhere/items.json")
	}
}

func TestPathDefaultsToTheHomeFolder(t *testing.T) {
	t.Setenv("GO_TODO_FILE", "")
	t.Setenv("HOME", "/Users/someone")
	got, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if want := "/Users/someone/.go-todo/items.json"; got != want {
		t.Errorf("Path = %q, want %q", got, want)
	}
}

func TestPathFailsWithoutAHomeFolder(t *testing.T) {
	t.Setenv("GO_TODO_FILE", "")
	t.Setenv("HOME", "")
	if got, err := Path(); err == nil {
		t.Errorf("Path = %q, want an error", got)
	}
}

func TestLoadOfAMissingFileGivesTheBuiltInItems(t *testing.T) {
	items, err := Load(filepath.Join(t.TempDir(), "items.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want, _ := Parse(Data)
	if !reflect.DeepEqual(items, want) {
		t.Errorf("items = %+v, want %+v", items, want)
	}
}

func TestLoadReadsTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	if err := os.WriteFile(path, []byte(`[{"title":"A","done":true}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if want := []Item{{Title: "A", Done: true}}; !reflect.DeepEqual(items, want) {
		t.Errorf("items = %+v, want %+v", items, want)
	}
}

func TestLoadRejectsCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0o644); err != nil {
		t.Fatal(err)
	}
	if items, err := Load(path); err == nil {
		t.Errorf("Load = %+v, want an error", items)
	}
}

func TestLoadReportsUnreadableFile(t *testing.T) {
	// A folder where the file should be cannot be read as a file.
	if items, err := Load(t.TempDir()); err == nil {
		t.Errorf("Load = %+v, want an error", items)
	}
}

func TestSaveThenLoadGivesTheSameItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "folder", "items.json")
	want := []Item{{Title: "A"}, {Title: "B", Done: true}}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("items = %+v, want %+v", got, want)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("folder holds %d entries, want only items.json", len(entries))
	}
}

func TestSaveOfNoItemsWritesAnEmptyList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	if err := Save(path, nil); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "[]\n" {
		t.Errorf("file = %q, want %q", got, "[]\n")
	}
}

func TestSaveReplacesTheWholeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	if err := Save(path, []Item{{Title: "A long first title"}, {Title: "B"}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := Save(path, []Item{{Title: "C"}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if want := []Item{{Title: "C"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("items = %+v, want %+v", got, want)
	}
}

func TestSaveFailsWhenTheFolderCannotBeMade(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(filepath.Join(blocker, "items.json"), []Item{{Title: "A"}}); err == nil {
		t.Error("Save succeeded under a regular file, want an error")
	}
}

func TestSaveFailsWhenTheFolderIsReadOnly(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o700) })
	if err := Save(filepath.Join(dir, "items.json"), []Item{{Title: "A"}}); err == nil {
		t.Error("Save succeeded in a read-only folder, want an error")
	}
}

func TestSaveFailsWhenTheTemporaryFileCannotBeWritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	if err := os.Mkdir(path+".tmp", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, []Item{{Title: "A"}}); err == nil {
		t.Error("Save wrote over a folder, want an error")
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("Save created the items file after a failed write")
	}
}

func TestSaveFailsWhenThePathIsAFolder(t *testing.T) {
	path := t.TempDir()
	if err := Save(path, []Item{{Title: "A"}}); err == nil {
		t.Error("Save replaced a folder, want an error")
	}
}

func TestParseReadsItems(t *testing.T) {
	items, err := Parse([]byte(`[{"title":"A","done":false},{"title":"B","done":true}]`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []Item{{Title: "A"}, {Title: "B", Done: true}}
	if !reflect.DeepEqual(items, want) {
		t.Errorf("items = %+v, want %+v", items, want)
	}
}

func TestParseRejectsInvalidJSON(t *testing.T) {
	items, err := Parse([]byte(`{not json`))
	if err == nil {
		t.Fatal("Parse accepted invalid JSON")
	}
	if items != nil {
		t.Errorf("items = %+v, want nil", items)
	}
}

func TestParseReadsBuiltInData(t *testing.T) {
	items, err := Parse(Data)
	if err != nil {
		t.Fatalf("Parse(Data): %v", err)
	}
	if len(items) != 5 {
		t.Errorf("built-in data has %d items, want 5", len(items))
	}
}

func TestIncompleteKeepsOpenItemsInOrder(t *testing.T) {
	items := []Item{
		{Title: "A"},
		{Title: "B", Done: true},
		{Title: "C"},
		{Title: "D", Done: true},
	}
	want := []Item{{Title: "A"}, {Title: "C"}}
	if got := Incomplete(items); !reflect.DeepEqual(got, want) {
		t.Errorf("Incomplete = %+v, want %+v", got, want)
	}
}

func TestIncompleteOfAllDoneIsEmpty(t *testing.T) {
	if got := Incomplete([]Item{{Title: "A", Done: true}}); len(got) != 0 {
		t.Errorf("Incomplete = %+v, want none", got)
	}
}
