package cli

import (
	"fmt"
	"io"
)

// Run prints the initial to-do list and returns the process exit status.
func Run(stdout, stderr io.Writer) int {
	if _, err := fmt.Fprintln(stdout, "[ ] Hello World"); err != nil {
		fmt.Fprintf(stderr, "go-todo: write output: %v\n", err)
		return 1
	}
	return 0
}
