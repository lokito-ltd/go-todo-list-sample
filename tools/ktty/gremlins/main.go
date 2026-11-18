// Command gremlins runs Gremlins mutation testing for Katatay.
//
// Usage, from the module root:
//
//	go run ./tools/ktty/gremlins -report reports/mutation/mutation.json < question.json
//
// The question on stdin is
//
//	{"targets": [{"file": "a/b.go", "lines": [3, 9]} | {"file": "a/b.go"}], "budget": 300 | null, "verbose": false}
//
// For each package folder holding a target, it runs `go tool gremlins unleash`
// in that folder with every other Go file excluded, keeps the mutants on the
// target lines, prints them, and writes the standard mutation testing report
// (schema version 2). Gremlins runs only the tests of the mutated package.
//
// Exit codes: 0 when no mutant survived, 1 when some survived or no test
// reached them, 2 when the run proves nothing (no mutant was tested, or the
// budget ran out first) or failed.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
)

type question struct {
	Targets []target `json:"targets"`
	Budget  *float64 `json:"budget"`
	Verbose bool     `json:"verbose"`
}

type target struct {
	File  string `json:"file"`
	Lines []int  `json:"lines"`
}

// gremlinsOutput is the part of Gremlins' --output file this wrapper reads.
type gremlinsOutput struct {
	Files []struct {
		Filename  string `json:"file_name"`
		Mutations []struct {
			Type   string `json:"type"`
			Status string `json:"status"`
			Line   int    `json:"line"`
			Column int    `json:"column"`
		} `json:"mutations"`
	} `json:"files"`
}

var statuses = map[string]string{
	"KILLED":      "Killed",
	"LIVED":       "Survived",
	"NOT COVERED": "NoCoverage",
	"TIMED OUT":   "Timeout",
	"NOT VIABLE":  "CompileError",
	"SKIPPED":     "Ignored",
	"RUNNABLE":    "Pending",
}

type report struct {
	SchemaVersion string                `json:"schemaVersion"`
	Thresholds    thresholds            `json:"thresholds"`
	Files         map[string]reportFile `json:"files"`
}

type thresholds struct {
	High int `json:"high"`
	Low  int `json:"low"`
}

type reportFile struct {
	Language string   `json:"language"`
	Source   string   `json:"source"`
	Mutants  []mutant `json:"mutants"`
}

type mutant struct {
	ID          string   `json:"id"`
	MutatorName string   `json:"mutatorName"`
	Location    location `json:"location"`
	Status      string   `json:"status"`
}

type location struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func main() {
	reportPath := flag.String("report", "", "mutation testing report to write")
	flag.Parse()
	if *reportPath == "" {
		fmt.Println("gremlins wrapper: -report is required")
		os.Exit(2)
	}
	code, err := run(*reportPath, os.Stdin)
	if err != nil {
		fmt.Println("gremlins wrapper:", err)
		os.Exit(2)
	}
	os.Exit(code)
}

func run(reportPath string, stdin io.Reader) (int, error) {
	_ = os.Remove(reportPath)

	var q question
	if err := json.NewDecoder(stdin).Decode(&q); err != nil {
		return 0, fmt.Errorf("read question: %w", err)
	}

	// ranges per target file; nil means the whole file.
	ranges := map[string][][2]int{}
	whole := map[string]bool{}
	dirs := map[string][]string{}
	for _, t := range q.Targets {
		file := path.Clean(filepath.ToSlash(t.File))
		if !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			fmt.Printf("Skipped %s: not Go code\n", file)
			continue
		}
		if _, ok := ranges[file]; !ok && !whole[file] {
			dir := path.Dir(file)
			dirs[dir] = append(dirs[dir], file)
		}
		switch len(t.Lines) {
		case 0:
			whole[file] = true
		case 2:
			ranges[file] = append(ranges[file], [2]int{t.Lines[0], t.Lines[1]})
		default:
			return 0, fmt.Errorf("target %s: lines must be [start, end]", file)
		}
	}

	ctx := context.Background()
	if q.Budget != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*q.Budget*float64(time.Second)))
		defer cancel()
	}

	rep := report{SchemaVersion: "2", Thresholds: thresholds{High: 80, Low: 60}, Files: map[string]reportFile{}}
	counts := map[string]int{}
	outOfTime := false
	id := 0

	dirNames := make([]string, 0, len(dirs))
	for d := range dirs {
		dirNames = append(dirNames, d)
	}
	sort.Strings(dirNames)

	for _, dir := range dirNames {
		files := dirs[dir]
		for _, file := range files {
			source, err := os.ReadFile(filepath.FromSlash(file))
			if err != nil {
				return 0, err
			}
			rep.Files[file] = reportFile{Language: "go", Source: string(source), Mutants: []mutant{}}
		}
		if outOfTime {
			continue
		}
		out, err := unleash(ctx, dir, files, q.Verbose)
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			outOfTime = true
			fmt.Printf("Budget reached before %s was tested\n", dir)
			continue
		}
		if err != nil {
			return 0, fmt.Errorf("%s: %w", dir, err)
		}
		for _, f := range out.Files {
			file := path.Join(dir, filepath.ToSlash(f.Filename))
			rf, ok := rep.Files[file]
			if !ok {
				continue
			}
			for _, m := range f.Mutations {
				if !whole[file] && !inRanges(m.Line, ranges[file]) {
					continue
				}
				status, ok := statuses[m.Status]
				if !ok {
					status = "RuntimeError"
				}
				id++
				rf.Mutants = append(rf.Mutants, mutant{
					ID:          fmt.Sprint(id),
					MutatorName: m.Type,
					Location: location{
						Start: position{Line: m.Line, Column: m.Column},
						End:   position{Line: m.Line, Column: m.Column + 1},
					},
					Status: status,
				})
				counts[status]++
				fmt.Printf("%s %s at %s:%d:%d\n", status, m.Type, file, m.Line, m.Column)
			}
			rep.Files[file] = rf
		}
	}

	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return 0, err
	}
	if err := os.WriteFile(reportPath, data, 0o644); err != nil {
		return 0, err
	}

	fmt.Printf("Killed: %d, Survived: %d, No coverage: %d, Timeout: %d, Not viable: %d\n",
		counts["Killed"], counts["Survived"], counts["NoCoverage"], counts["Timeout"], counts["CompileError"])

	detected := counts["Killed"] + counts["Timeout"]
	switch {
	case counts["Survived"] > 0:
		return 1, nil
	case outOfTime:
		fmt.Println("The budget ran out, so the run proves nothing.")
		return 2, nil
	case detected == 0:
		fmt.Println("No mutant on the target lines was tested, so the run proves nothing.")
		return 2, nil
	case counts["NoCoverage"] > 0:
		return 1, nil
	}
	return 0, nil
}

// unleash runs Gremlins in dir with every Go file except files excluded.
func unleash(ctx context.Context, dir string, files []string, verbose bool) (*gremlinsOutput, error) {
	keep := map[string]bool{}
	for _, f := range files {
		keep[strings.TrimPrefix(f, dir+"/")] = true
	}
	tmp, err := os.MkdirTemp("", "ktty-gremlins-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	outFile := filepath.Join(tmp, "gremlins.json")

	args := []string{"tool", "gremlins", "unleash", "--output", outFile, "--timeout-coefficient", "30"}
	err = filepath.WalkDir(filepath.FromSlash(dir), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(filepath.FromSlash(dir), p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasSuffix(rel, ".go") && !strings.HasSuffix(rel, "_test.go") && !keep[rel] {
			args = append(args, "--exclude-files", "^"+regexp.QuoteMeta(rel)+"$")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = filepath.FromSlash(dir)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	var log bytes.Buffer
	cmd.Stdout = &log
	cmd.Stderr = &log
	runErr := cmd.Run()
	if verbose || runErr != nil {
		os.Stderr.Write(log.Bytes())
	}
	if runErr != nil {
		return nil, fmt.Errorf("gremlins: %w", runErr)
	}

	data, err := os.ReadFile(outFile)
	if errors.Is(err, fs.ErrNotExist) {
		// Gremlins succeeded but found nothing to mutate, so it wrote no file.
		return &gremlinsOutput{}, nil
	}
	if err != nil {
		return nil, err
	}
	var out gremlinsOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func inRanges(line int, ranges [][2]int) bool {
	for _, r := range ranges {
		if line >= r[0] && line <= r[1] {
			return true
		}
	}
	return false
}
