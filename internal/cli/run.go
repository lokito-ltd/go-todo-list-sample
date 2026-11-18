package cli

import (
	"fmt"
	"io"
	"strings"

	"example.com/go-todo/internal/todo"
)

const usage = "usage: go-todo list\n       go-todo add <title>"

// Run executes the command named in args and returns the process exit status.
func Run(args []string, stdout, stderr io.Writer) int {
	path, err := todo.Path()
	if err != nil {
		fmt.Fprintf(stderr, "go-todo: find items file: %v\n", err)
		return 1
	}
	return run(args, path, stdout, stderr)
}

func run(args []string, path string, stdout, stderr io.Writer) int {
	// Plain if statements rather than a switch: Go's coverage records no
	// block for case expressions, so Gremlins would skip their mutants
	// (go-gremlins/gremlins#298).
	if len(args) == 1 && args[0] == "list" {
		return list(path, stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "add" {
		if title := strings.Join(strings.Fields(strings.Join(args[1:], " ")), " "); title != "" {
			return add(path, title, stdout, stderr)
		}
	}
	fmt.Fprintln(stderr, usage)
	return 2
}

func list(path string, stdout, stderr io.Writer) int {
	items, err := todo.Load(path)
	if err != nil {
		fmt.Fprintf(stderr, "go-todo: read items: %v\n", err)
		return 1
	}
	for _, item := range todo.Incomplete(items) {
		if err := printItem(stdout, item); err != nil {
			fmt.Fprintf(stderr, "go-todo: write output: %v\n", err)
			return 1
		}
	}
	return 0
}

func add(path, title string, stdout, stderr io.Writer) int {
	items, err := todo.Load(path)
	if err != nil {
		fmt.Fprintf(stderr, "go-todo: read items: %v\n", err)
		return 1
	}
	item := todo.Item{Title: title}
	if err := todo.Save(path, append(items, item)); err != nil {
		fmt.Fprintf(stderr, "go-todo: save items: %v\n", err)
		return 1
	}
	if err := printItem(stdout, item); err != nil {
		fmt.Fprintf(stderr, "go-todo: write output: %v\n", err)
		return 1
	}
	return 0
}

func printItem(w io.Writer, item todo.Item) error {
	_, err := fmt.Fprintf(w, "[ ] %s\n", item.Title)
	return err
}
