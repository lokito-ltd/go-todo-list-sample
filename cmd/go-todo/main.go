package main

import (
	"os"

	"example.com/go-todo/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Stdout, os.Stderr))
}
