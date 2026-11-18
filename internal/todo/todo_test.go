package todo

import (
	"reflect"
	"testing"
)

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
