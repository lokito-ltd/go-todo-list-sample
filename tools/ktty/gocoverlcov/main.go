// Command gocoverlcov prints a Go coverage profile as LCOV on stdout.
//
// Usage, from the module root:
//
//	go run ./tools/ktty/gocoverlcov -profile reports/test/cover.out ./cmd/... ./internal/...
//
// Every non-test Go file in the given packages gets a record, at zero when no
// test loaded it. A line counts as hit when any coverage block on it ran.
// Go measures no branches, so the output has no branch records.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	profile := flag.String("profile", "", "Go coverage profile to read")
	flag.Parse()
	if *profile == "" {
		fmt.Fprintln(os.Stderr, "gocoverlcov: -profile is required")
		os.Exit(2)
	}
	if err := run(*profile, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "gocoverlcov:", err)
		os.Exit(1)
	}
}

func run(profile string, patterns []string) error {
	module, err := modulePath()
	if err != nil {
		return err
	}
	files := map[string]map[int]int{} // file -> line -> hits

	if len(patterns) > 0 {
		known, err := packageFiles(patterns)
		if err != nil {
			return err
		}
		for _, f := range known {
			files[f] = map[int]int{}
		}
	}

	in, err := os.Open(profile)
	if err != nil {
		return err
	}
	defer in.Close()
	scanner := bufio.NewScanner(in)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			if !strings.HasPrefix(line, "mode: ") {
				return errors.New("profile does not start with a mode line")
			}
			continue
		}
		if line == "" {
			continue
		}
		name, start, end, count, err := parseBlock(line)
		if err != nil {
			return fmt.Errorf("%q: %w", line, err)
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(name, module), "/")
		if rel == name {
			continue // outside this module
		}
		lines, ok := files[rel]
		if !ok {
			lines = map[int]int{}
			files[rel] = lines
		}
		for l := start; l <= end; l++ {
			if hits, seen := lines[l]; !seen || count > hits {
				lines[l] = count
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for _, name := range names {
		lines := files[name]
		numbers := make([]int, 0, len(lines))
		for l := range lines {
			numbers = append(numbers, l)
		}
		sort.Ints(numbers)
		hit := 0
		fmt.Fprintf(out, "TN:\nSF:%s\n", name)
		for _, l := range numbers {
			fmt.Fprintf(out, "DA:%d,%d\n", l, lines[l])
			if lines[l] > 0 {
				hit++
			}
		}
		fmt.Fprintf(out, "LF:%d\nLH:%d\nend_of_record\n", len(numbers), hit)
	}
	return nil
}

// parseBlock reads "path/file.go:12.3,14.5 2 1".
func parseBlock(line string) (name string, start, end, count int, err error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return "", 0, 0, 0, errors.New("expected three fields")
	}
	colon := strings.LastIndex(fields[0], ":")
	if colon < 0 {
		return "", 0, 0, 0, errors.New("no colon in block")
	}
	name = fields[0][:colon]
	span := strings.Split(fields[0][colon+1:], ",")
	if len(span) != 2 {
		return "", 0, 0, 0, errors.New("bad span")
	}
	if start, err = strconv.Atoi(strings.SplitN(span[0], ".", 2)[0]); err != nil {
		return
	}
	if end, err = strconv.Atoi(strings.SplitN(span[1], ".", 2)[0]); err != nil {
		return
	}
	count, err = strconv.Atoi(fields[2])
	return
}

// packageFiles lists the non-test Go files of the packages, as paths from the
// module root.
func packageFiles(patterns []string) ([]string, error) {
	args := append([]string{"list", "-e", "-f", "{{.Dir}}{{range .GoFiles}}\t{{.}}{{end}}{{range .CgoFiles}}\t{{.}}{{end}}"}, patterns...)
	data, err := exec.Command("go", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		dir, err := filepath.Rel(root, parts[0])
		if err != nil {
			return nil, err
		}
		for _, f := range parts[1:] {
			files = append(files, filepath.ToSlash(filepath.Join(dir, f)))
		}
	}
	return files, nil
}

// modulePath reads the module path from go.mod in the current folder.
func modulePath() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return strings.Trim(fields[1], `"`), nil
		}
	}
	return "", errors.New("go.mod has no module line")
}
