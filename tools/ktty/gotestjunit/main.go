// Command gotestjunit runs `go test -json` and writes a JUnit XML report in
// which every test case names its test file and the lines of its function.
//
// Usage, from the module root:
//
//	go run ./tools/ktty/gotestjunit -o reports/test/junit.xml -- [go test flags] [packages]
//
// It prints the package summaries and the output of failed tests, and exits
// non-zero when go test fails or any test or package fails.
package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type event struct {
	Action     string
	Package    string
	ImportPath string
	Test       string
	Elapsed    float64
	Output     string
}

type result struct {
	name    string
	status  string // pass, fail, skip, or "" when it never finished
	elapsed float64
	output  strings.Builder
}

type suite struct {
	name    string
	failed  bool
	elapsed float64
	order   []string
	results map[string]*result
}

type location struct {
	file      string
	line, end int
}

func main() {
	out := flag.String("o", "", "JUnit XML file to write")
	flag.Parse()
	if *out == "" {
		fmt.Fprintln(os.Stderr, "gotestjunit: -o is required")
		os.Exit(2)
	}
	code, err := run(*out, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "gotestjunit:", err)
		os.Exit(2)
	}
	os.Exit(code)
}

func run(out string, args []string) (int, error) {
	module, err := modulePath()
	if err != nil {
		return 0, err
	}
	// The report folder also holds files go test writes, such as a cover profile.
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return 0, err
	}
	cmd := exec.Command("go", append([]string{"test", "-json"}, args...)...)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	suites := map[string]*suite{}
	var order []string
	getSuite := func(name string) *suite {
		s, ok := suites[name]
		if !ok {
			s = &suite{name: name, results: map[string]*result{}}
			suites[name] = s
			order = append(order, name)
		}
		return s
	}
	anyFailed := false

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var ev event
		if err := json.Unmarshal(line, &ev); err != nil || ev.Action == "" {
			fmt.Println(string(line))
			continue
		}
		switch ev.Action {
		case "build-output":
			fmt.Print(ev.Output)
			continue
		case "build-fail":
			anyFailed = true
			continue
		}
		s := getSuite(ev.Package)
		if ev.Test == "" {
			switch ev.Action {
			case "output":
				if !strings.HasPrefix(ev.Output, "=== ") && ev.Output != "PASS\n" {
					fmt.Print(ev.Output)
				}
			case "fail":
				s.failed = true
				anyFailed = true
				s.elapsed = ev.Elapsed
			case "pass", "skip":
				s.elapsed = ev.Elapsed
			}
			continue
		}
		r, ok := s.results[ev.Test]
		if !ok {
			r = &result{name: ev.Test}
			s.results[ev.Test] = r
			s.order = append(s.order, ev.Test)
		}
		switch ev.Action {
		case "output":
			r.output.WriteString(ev.Output)
		case "pass", "fail", "skip":
			r.status = ev.Action
			r.elapsed = ev.Elapsed
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	waitErr := cmd.Wait()
	goCode := 0
	if waitErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(waitErr, &exitErr) {
			return 0, waitErr
		}
		goCode = exitErr.ExitCode()
	}

	report := testsuites{}
	for _, name := range order {
		s := suites[name]
		if len(s.order) == 0 {
			continue
		}
		funcs, err := testFunctions(module, name)
		if err != nil {
			return 0, err
		}
		ts := testsuite{Name: name, Time: seconds(s.elapsed)}
		for _, testName := range s.order {
			r := s.results[testName]
			tc := testcase{Classname: name, Name: r.name, Time: seconds(r.elapsed)}
			if loc, ok := funcs[strings.SplitN(r.name, "/", 2)[0]]; ok {
				tc.File = loc.file
				tc.Properties = &properties{Property: []property{
					{Name: "line", Value: fmt.Sprint(loc.line)},
					{Name: "endLine", Value: fmt.Sprint(loc.end)},
				}}
			}
			switch r.status {
			case "pass":
			case "skip":
				tc.Skipped = &skipped{Message: "skipped"}
				ts.Skipped++
			default:
				message := "failed"
				if r.status == "" {
					message = "did not finish"
				}
				tc.Failure = &failure{Message: message, Text: r.output.String()}
				ts.Failures++
				anyFailed = true
				fmt.Printf("--- %s: %s %s\n%s", strings.ToUpper(message), name, r.name, r.output.String())
			}
			ts.Tests++
			ts.Cases = append(ts.Cases, tc)
		}
		report.Tests += ts.Tests
		report.Failures += ts.Failures
		report.Skipped += ts.Skipped
		report.Suites = append(report.Suites, ts)
	}

	data, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return 0, err
	}
	data = append([]byte(xml.Header), append(data, '\n')...)
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return 0, err
	}

	switch {
	case goCode != 0:
		return goCode, nil
	case anyFailed:
		return 1, nil
	}
	return 0, nil
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

// testFunctions maps each top-level function in the package's test files to
// its file, as a path from the module root, and its first and last lines.
func testFunctions(module, pkg string) (map[string]location, error) {
	funcs := map[string]location{}
	rel := strings.TrimPrefix(strings.TrimPrefix(pkg, module), "/")
	if pkg != module && rel == pkg {
		return funcs, nil // not in this module
	}
	dir := filepath.FromSlash(rel)
	if dir == "" {
		dir = "."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file := path.Join(rel, entry.Name())
		parsed, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.SkipObjectResolution)
		if err != nil {
			continue // a file that does not parse has no tests to name
		}
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			funcs[fn.Name.Name] = location{
				file: file,
				line: fset.Position(fn.Pos()).Line,
				end:  fset.Position(fn.End()).Line,
			}
		}
	}
	return funcs, nil
}

func seconds(s float64) string { return fmt.Sprintf("%.3f", s) }

type testsuites struct {
	XMLName  xml.Name    `xml:"testsuites"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Skipped  int         `xml:"skipped,attr"`
	Suites   []testsuite `xml:"testsuite"`
}

type testsuite struct {
	Name     string     `xml:"name,attr"`
	Tests    int        `xml:"tests,attr"`
	Failures int        `xml:"failures,attr"`
	Skipped  int        `xml:"skipped,attr"`
	Time     string     `xml:"time,attr"`
	Cases    []testcase `xml:"testcase"`
}

type testcase struct {
	Classname  string      `xml:"classname,attr"`
	Name       string      `xml:"name,attr"`
	Time       string      `xml:"time,attr"`
	File       string      `xml:"file,attr,omitempty"`
	Properties *properties `xml:"properties,omitempty"`
	Failure    *failure    `xml:"failure,omitempty"`
	Skipped    *skipped    `xml:"skipped,omitempty"`
}

type properties struct {
	Property []property `xml:"property"`
}

type property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type failure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

type skipped struct {
	Message string `xml:"message,attr"`
}
