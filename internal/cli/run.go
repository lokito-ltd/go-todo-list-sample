package cli

import (
	"fmt"
	"io"

	"example.com/go-todo/internal/todo"
)

const usage = "usage: go-todo list"

// Run executes the command named in args and returns the process exit status.
func Run(args []string, stdout, stderr io.Writer) int {
	return run(args, todo.Data, stdout, stderr)
}

func run(args []string, data []byte, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] != "list" {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	items, err := todo.Parse(data)
	if err != nil {
		fmt.Fprintf(stderr, "go-todo: read items: %v\n", err)
		return 1
	}
	for _, item := range todo.Incomplete(items) {
		if _, err := fmt.Fprintf(stdout, "[ ] %s\n", item.Title); err != nil {
			fmt.Fprintf(stderr, "go-todo: write output: %v\n", err)
			return 1
		}
	}
	return 0
}
