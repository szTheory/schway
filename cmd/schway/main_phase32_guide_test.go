package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase32Block is one marked fence of the checksum guide. Lines are the logical
// lines after backslash continuations are joined.
type phase32Block struct {
	Kind  string
	Attrs map[string]string
	Lines []string
}

type phase32Row struct {
	Case  string
	Cells []string
}

type phase32Guide struct {
	Blocks       []phase32Block
	Results      []phase32Row // Case | checksum | adler_low | python
	Fixture      []phase32Row // Case id | adler32_cli
	Limit        int
	RefusedSize  int
	ErrorExit    int
	hasErrorExit bool
}

var (
	phase32FenceRE = regexp.MustCompile("^```sh doc-test=([a-z0-9-]+)((?: [a-z0-9_]+=[a-z0-9_]+)*)$")
	phase32LimitRE = regexp.MustCompile(`Limit: files up to ([0-9,]+) bytes are read; a file of ([0-9,]+) bytes is refused\.`)
	phase32ExitRE  = regexp.MustCompile(`Every error exits ([0-9]+)`)
)

func phase32Atoi(s string) (int, error) { return strconv.Atoi(strings.ReplaceAll(s, ",", "")) }

func phase32Cells(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	cells := make([]string, len(parts))
	for i, part := range parts {
		cells[i] = strings.TrimSpace(part)
	}
	return cells
}

// phase32ParseGuide extracts the marked fences, the case-keyed tables and the
// facts lines from the guide text. It errors when a required piece is absent.
func phase32ParseGuide(text string) (phase32Guide, error) {
	var guide phase32Guide
	lines := strings.Split(text, "\n")
	table := ""
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if m := phase32FenceRE.FindStringSubmatch(line); m != nil {
			block := phase32Block{Kind: m[1], Attrs: map[string]string{}}
			for _, attr := range strings.Fields(m[2]) {
				k, v, _ := strings.Cut(attr, "=")
				block.Attrs[k] = v
			}
			closed := false
			pending := ""
			for i++; i < len(lines); i++ {
				if lines[i] == "```" {
					closed = true
					break
				}
				body := lines[i]
				if strings.HasSuffix(body, "\\") {
					pending += strings.TrimSpace(strings.TrimSuffix(body, "\\")) + " "
					continue
				}
				logical := strings.TrimSpace(pending + strings.TrimSpace(body))
				pending = ""
				if logical != "" {
					block.Lines = append(block.Lines, logical)
				}
			}
			if !closed {
				return guide, errors.New("unterminated doc-test fence")
			}
			guide.Blocks = append(guide.Blocks, block)
			continue
		}
		switch {
		case strings.HasPrefix(line, "| Case | checksum | adler_low |"):
			table = "results"
		case strings.HasPrefix(line, "| Case id | adler32_cli |"):
			table = "fixture"
		case strings.HasPrefix(line, "|"):
			cells := phase32Cells(line)
			if table == "" || len(cells) < 2 || strings.HasPrefix(cells[0], "---") {
				continue
			}
			key := strings.Trim(cells[0], "`")
			row := phase32Row{Case: key, Cells: cells[1:]}
			if table == "results" {
				if len(cells) != 4 {
					return guide, fmt.Errorf("results row %q has %d cells, want 4", key, len(cells))
				}
				guide.Results = append(guide.Results, row)
			} else {
				guide.Fixture = append(guide.Fixture, row)
			}
		default:
			table = ""
		}
	}
	if m := phase32LimitRE.FindStringSubmatch(text); m != nil {
		guide.Limit, _ = phase32Atoi(m[1])
		guide.RefusedSize, _ = phase32Atoi(m[2])
	} else {
		return guide, errors.New("guide has no Limit: facts line")
	}
	if m := phase32ExitRE.FindStringSubmatch(text); m != nil {
		guide.ErrorExit, _ = strconv.Atoi(m[1])
		guide.hasErrorExit = true
	} else {
		return guide, errors.New("guide has no 'Every error exits' facts line")
	}
	hasSetup := false
	for _, b := range guide.Blocks {
		if b.Kind == "clean-checkout" {
			hasSetup = true
		}
	}
	if !hasSetup {
		return guide, errors.New("guide has no doc-test=clean-checkout fence")
	}
	if len(guide.Results) == 0 {
		return guide, errors.New("guide has no results rows")
	}
	return guide, nil
}

// phase32GuideOutputTargets collects every output path the block writes (same
// logic as the Phase 27.3 checker; a cmd test cannot import session_test).
func phase32GuideOutputTargets(lines []string) []string {
	var targets []string
	for _, line := range lines {
		fields := strings.Fields(line)
		for i, field := range fields {
			next := ""
			if i+1 < len(fields) {
				next = fields[i+1]
			}
			switch {
			case field == "-o" && next != "":
				targets = append(targets, next)
			case field == "--output" && next != "":
				targets = append(targets, next, next+".schway-build.json")
			case (field == ">" || field == ">>") && next != "":
				targets = append(targets, next)
			case strings.HasPrefix(field, "--output=") && len(field) > len("--output="):
				value := strings.TrimPrefix(field, "--output=")
				targets = append(targets, value, value+".schway-build.json")
			case strings.HasPrefix(field, "-o=") && len(field) > len("-o="):
				targets = append(targets, strings.TrimPrefix(field, "-o="))
			case field == "run" && i > 0 && fields[i-1] == "app" && next != "":
				targets = append(targets, next)
			}
		}
	}
	return targets
}

func phase32TargetAllowed(target string) bool {
	cleaned := path.Clean(target)
	return strings.HasPrefix(cleaned, ".schway/") || cleaned == "schway"
}

// phase32NativePolicy reports whether a missing clang must fail (true) rather
// than skip: hosted CI or SCHWAY_REQUIRE_NATIVE=1.
func phase32NativePolicy(clangErr error, getenv func(string) string) bool {
	if clangErr == nil {
		return false
	}
	return getenv("CI") != "" || getenv("GITHUB_ACTIONS") != "" || getenv("SCHWAY_REQUIRE_NATIVE") == "1"
}

// phase32CopyCheckout copies the tracked and untracked-but-not-ignored files
// (minus .planning/) into a fresh temp directory and returns it.
func phase32CopyCheckout(t *testing.T, root string) string {
	t.Helper()
	dest := t.TempDir()
	var paths []string
	result, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", root, "ls-files", "-co", "--exclude-standard", "-z")
	if err == nil && result.Exit == 0 {
		for _, p := range strings.Split(string(result.Stdout), "\x00") {
			if p != "" {
				paths = append(paths, p)
			}
		}
	} else {
		walkErr := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				switch d.Name() {
				case ".git", ".schway", ".planning":
					return filepath.SkipDir
				}
				return nil
			}
			rel, relErr := filepath.Rel(root, p)
			if relErr != nil {
				return relErr
			}
			paths = append(paths, filepath.ToSlash(rel))
			return nil
		})
		if walkErr != nil {
			t.Fatalf("walk checkout: %v", walkErr)
		}
	}
	for _, rel := range paths {
		if rel == ".planning" || strings.HasPrefix(rel, ".planning/") {
			continue
		}
		src := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Lstat(src)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			t.Fatalf("stat %s: %v", rel, err)
		}
		dst := filepath.Join(dest, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			target, err := os.Readlink(src)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, dst); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		in, err := os.Open(src)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			t.Fatal(err)
		}
		out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm()|0o200)
		if err != nil {
			_ = in.Close()
			t.Fatal(err)
		}
		_, copyErr := io.Copy(out, in)
		_ = in.Close()
		if closeErr := out.Close(); copyErr == nil {
			copyErr = closeErr
		}
		if copyErr != nil {
			t.Fatalf("copy %s: %v", rel, copyErr)
		}
		if err := os.Chmod(dst, info.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
	}
	return dest
}

func phase32Quote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// phase32Substitute replaces the trailing token (PATH or CASE) of a template
// line with a single-quoted value.
func phase32Substitute(line, token, value string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 || fields[len(fields)-1] != token || strings.Count(line, token) != 1 {
		return "", fmt.Errorf("template line %q must end with exactly one %s token", line, token)
	}
	return strings.TrimSuffix(line, token) + phase32Quote(value), nil
}

func phase32RunLine(t *testing.T, dir, line string) testsupport.CLIResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	result, err := testsupport.RunCLIErr(ctx, "sh", os.Environ(), "-c", `cd "$0" && eval "$1"`, dir, line)
	if err != nil {
		t.Fatalf("run %q: %v", line, err)
	}
	return result
}

// phase32Expect decodes a results/fixture cell into the expected exit, stdout
// and stderr under the guide grammar.
func phase32Expect(cell string, errorExit int) (exit int, stdout, stderr string, isError bool, err error) {
	switch {
	case cell == "refused, no output":
		return errorExit, "", "", true, nil
	case strings.HasPrefix(cell, "`FileError.") && strings.HasSuffix(cell, "`"):
		return errorExit, "", strings.Trim(cell, "`") + "\n", true, nil
	}
	if _, convErr := strconv.ParseUint(cell, 10, 64); convErr != nil {
		return 0, "", "", false, fmt.Errorf("unrecognised cell %q", cell)
	}
	return 0, cell + "\n", "", false, nil
}

func phase32ReadGuide(t *testing.T) phase32Guide {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	guide, err := phase32ParseGuide(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	return guide
}

// TestPhase32GuideCleanCheckout runs the checksum guide, line by line, from a
// simulated clean checkout and compares every result against the README cells.
func TestPhase32GuideCleanCheckout(t *testing.T) {
	// D-04: the static check gates execution; a drifted guide never runs.
	if problems := phase32ReadStaticProblems(t); len(problems) != 0 {
		t.Fatalf("guide fails the static check:\n%s", strings.Join(problems, "\n"))
	}
	_, clangErr := exec.LookPath("clang")
	if clangErr != nil {
		if phase32NativePolicy(clangErr, os.Getenv) {
			t.Fatalf("clang is required on this host: %v", clangErr)
		}
		t.Skip("env:clang toolchain unavailable on this host")
	}
	guide := phase32ReadGuide(t)
	cases := map[string]testsupport.ChecksumCase{}
	for _, c := range testsupport.ChecksumCaseMatrix() {
		cases[c.Name] = c
	}

	var setup []string
	var runs []phase32Block
	for _, b := range guide.Blocks {
		switch b.Kind {
		case "clean-checkout":
			setup = append(setup, b.Lines...)
		case "run":
			runs = append(runs, b)
		}
	}

	copyDir := phase32CopyCheckout(t, testsupport.ProjectPath())
	for _, line := range setup {
		if result := phase32RunLine(t, copyDir, line); result.Exit != 0 {
			t.Fatalf("setup line %q exit=%d stdout=%q stderr=%q", line, result.Exit, result.Stdout, result.Stderr)
		}
	}

	inputs := t.TempDir()
	executed := 0
	for _, b := range runs {
		if len(b.Lines) != 1 {
			t.Fatalf("run fence for %s holds %d lines, want exactly one", b.Attrs["app"], len(b.Lines))
		}
		template := b.Lines[0]
		app := b.Attrs["app"]
		switch app {
		case "checksum", "adler_low":
			column := 0
			if app == "adler_low" {
				column = 1
			}
			for _, row := range guide.Results {
				c, ok := cases[row.Case]
				if !ok {
					t.Fatalf("results row %q is not a checksum case", row.Case)
				}
				exit, stdout, stderr, isError, err := phase32Expect(row.Cells[column], guide.ErrorExit)
				if err != nil {
					t.Fatalf("%s %s: %v", app, row.Case, err)
				}
				input := testsupport.MaterializeChecksumCase(t, filepath.Join(inputs, app), c)
				line, err := phase32Substitute(template, "PATH", input)
				if err != nil {
					t.Fatal(err)
				}
				result := phase32RunLine(t, copyDir, line)
				executed++
				if result.Exit != exit || string(result.Stdout) != stdout || string(result.Stderr) != stderr {
					t.Fatalf("%s %s: exit=%d stdout=%q stderr=%q, want exit=%d stdout=%q stderr=%q",
						app, row.Case, result.Exit, result.Stdout, result.Stderr, exit, stdout, stderr)
				}
				if isError && strings.Contains(string(result.Stderr), testsupport.ChecksumSecretSegment) {
					t.Fatalf("%s %s: stderr echoes the private path segment: %q", app, row.Case, result.Stderr)
				}
			}
		case "adler32_cli":
			for _, row := range guide.Fixture {
				if _, err := strconv.ParseUint(row.Case, 10, 64); err != nil {
					t.Fatalf("fixture case id %q: %v", row.Case, err)
				}
				exit, stdout, stderr, _, err := phase32Expect(row.Cells[0], guide.ErrorExit)
				if err != nil {
					t.Fatalf("adler32_cli %s: %v", row.Case, err)
				}
				line, err := phase32Substitute(template, "CASE", row.Case)
				if err != nil {
					t.Fatal(err)
				}
				result := phase32RunLine(t, copyDir, line)
				executed++
				if result.Exit != exit || string(result.Stdout) != stdout || string(result.Stderr) != stderr {
					t.Fatalf("adler32_cli %s: exit=%d stdout=%q stderr=%q, want exit=%d stdout=%q stderr=%q",
						row.Case, result.Exit, result.Stdout, result.Stderr, exit, stdout, stderr)
				}
			}
		default:
			t.Fatalf("unknown run app %q", app)
		}
	}
	want := len(guide.Results)*2 + len(guide.Fixture)
	if executed != want {
		t.Fatalf("executed %d rows, want %d", executed, want)
	}
}

// phase32ExpectedJSON mirrors the TestPhase31ExpectedMatrixJSON struct.
type phase32ExpectedJSON struct {
	Schema    string `json:"schema"`
	Limit     int    `json:"limit"`
	ErrorExit struct {
		Code int    `json:"code"`
		Name string `json:"name"`
	} `json:"error_exit"`
	Cases []struct {
		Name    string `json:"name"`
		Kind    string `json:"kind"`
		Exit    int    `json:"exit"`
		Stdout  string `json:"stdout"`
		Stderr  string `json:"stderr"`
		Outcome string `json:"outcome"`
	} `json:"cases"`
}

// phase32GuideJSONProblems cross-checks the parsed guide against the expected
// JSON, the independent oracle rows and the Phase 29 fixture table. Empty means
// consistent.
func phase32GuideJSONProblems(guide phase32Guide, doc phase32ExpectedJSON) []string {
	var problems []string
	if guide.Limit != doc.Limit {
		problems = append(problems, fmt.Sprintf("README limit %d != json limit %d", guide.Limit, doc.Limit))
	}
	if guide.RefusedSize != doc.Limit+1 {
		problems = append(problems, fmt.Sprintf("README refused size %d != limit+1 (%d)", guide.RefusedSize, doc.Limit+1))
	}
	if guide.ErrorExit != doc.ErrorExit.Code {
		problems = append(problems, fmt.Sprintf("README error exit %d != json error_exit %d", guide.ErrorExit, doc.ErrorExit.Code))
	}
	byName := map[string]int{}
	for i, c := range doc.Cases {
		byName[c.Name] = i
	}
	inputs := map[string][]byte{}
	for _, c := range testsupport.ChecksumCaseMatrix() {
		inputs[c.Name] = c.Content
	}
	oracleLow := map[string]uint64{}
	for _, row := range testsupport.ChecksumOracleRows() {
		if row.App == "adler_low" && row.Outcome == "returned" {
			oracleLow[string(row.Input)] = row.Want
		}
	}
	inTable := map[string]bool{}
	for _, row := range guide.Results {
		inTable[row.Case] = true
		i, ok := byName[row.Case]
		if !ok {
			problems = append(problems, fmt.Sprintf("table case %q is not in json cases", row.Case))
			continue
		}
		c := doc.Cases[i]
		exit, stdout, stderr, _, err := phase32Expect(row.Cells[0], guide.ErrorExit)
		if err != nil {
			problems = append(problems, fmt.Sprintf("case %q checksum cell: %v", row.Case, err))
			continue
		}
		if exit != c.Exit || stdout != c.Stdout || stderr != c.Stderr {
			problems = append(problems, fmt.Sprintf("case %q checksum cell gives exit=%d stdout=%q stderr=%q, json has exit=%d stdout=%q stderr=%q",
				row.Case, exit, stdout, stderr, c.Exit, c.Stdout, c.Stderr))
		}
		if c.Exit == 0 {
			sum, convErr := strconv.ParseUint(row.Cells[0], 10, 64)
			low, lowErr := strconv.ParseUint(row.Cells[1], 10, 64)
			if convErr != nil || lowErr != nil {
				problems = append(problems, fmt.Sprintf("case %q has a non-decimal success cell", row.Case))
				continue
			}
			if low != sum%65536 {
				problems = append(problems, fmt.Sprintf("case %q adler_low %d != checksum %d %% 65536", row.Case, low, sum))
			}
			if want, ok := oracleLow[string(inputs[row.Case])]; !ok || want != low {
				problems = append(problems, fmt.Sprintf("case %q adler_low %d disagrees with the oracle (%d, found=%v)", row.Case, low, want, ok))
			}
		} else if row.Cells[1] != row.Cells[0] {
			problems = append(problems, fmt.Sprintf("case %q adler_low error cell %q != checksum cell %q", row.Case, row.Cells[1], row.Cells[0]))
		}
	}
	for _, name := range []string{"empty", "zero", "wikipedia", "ff65536", "ff65537", "missing", "directory"} {
		if _, ok := byName[name]; !ok {
			problems = append(problems, fmt.Sprintf("json has no case %q", name))
		} else if !inTable[name] {
			problems = append(problems, fmt.Sprintf("json case %q is missing from the table", name))
		}
	}
	for _, row := range guide.Fixture {
		id, err := strconv.ParseUint(row.Case, 10, 64)
		if err != nil {
			problems = append(problems, fmt.Sprintf("fixture id %q is not a number", row.Case))
			continue
		}
		want, ok := testsupport.Phase29FixtureExpected(id)
		if row.Cells[0] == "refused, no output" {
			if ok {
				problems = append(problems, fmt.Sprintf("fixture id %d is refused in the README but Phase29FixtureExpected has %d", id, want))
			}
			continue
		}
		got, convErr := strconv.ParseUint(row.Cells[0], 10, 64)
		if convErr != nil || !ok || got != want {
			problems = append(problems, fmt.Sprintf("fixture id %d cell %q != Phase29FixtureExpected %d (found=%v)", id, row.Cells[0], want, ok))
		}
	}
	return problems
}

// TestPhase32GuideMatchesExpectedJSON is clang-free: every number in the guide
// agrees with checksum.expected.json and the independent oracle.
func TestPhase32GuideMatchesExpectedJSON(t *testing.T) {
	guide := phase32ReadGuide(t)
	data, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "checksum.expected.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc phase32ExpectedJSON
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if problems := phase32GuideJSONProblems(guide, doc); len(problems) != 0 {
		t.Fatalf("guide disagrees with the expected data:\n%s", strings.Join(problems, "\n"))
	}
	t.Run("armed-json", func(t *testing.T) {
		mutated := doc
		mutated.Limit = 65535
		if len(phase32GuideJSONProblems(guide, mutated)) == 0 {
			t.Fatal("a mutated json limit produced no problem")
		}
	})
}

func phase32ReadStaticProblems(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "checksum.expected.json"))
	if err != nil {
		t.Fatal(err)
	}
	return phase32GuideStaticProblems(string(raw), data)
}

// phase32GuideStaticProblems is the pure, clang-free guide check: parse,
// anti-vacuity (D-04), output targets (D-01a), command shapes, and the JSON and
// oracle cross-check. Empty means the guide is acceptable.
func phase32GuideStaticProblems(text string, expectedJSON []byte) []string {
	guide, err := phase32ParseGuide(text)
	if err != nil {
		return []string{"parse: " + err.Error()}
	}
	var problems []string
	if len(guide.Blocks) < 4 {
		problems = append(problems, fmt.Sprintf("guide has %d marked blocks, want at least 4", len(guide.Blocks)))
	}
	apps := map[string]bool{}
	var setup []string
	for _, b := range guide.Blocks {
		for _, target := range phase32GuideOutputTargets(b.Lines) {
			if !phase32TargetAllowed(target) {
				problems = append(problems, fmt.Sprintf("output target %q is outside the ignored .schway/ tree", target))
			}
		}
		switch b.Kind {
		case "clean-checkout":
			setup = append(setup, b.Lines...)
		case "run":
			apps[b.Attrs["app"]] = true
			if len(b.Lines) != 1 {
				problems = append(problems, fmt.Sprintf("run fence for %s holds %d lines, want exactly one", b.Attrs["app"], len(b.Lines)))
				continue
			}
			if !strings.HasPrefix(b.Lines[0], "./schway app run ") {
				problems = append(problems, fmt.Sprintf("run line %q must start with `./schway app run `", b.Lines[0]))
			}
		}
	}
	for _, app := range []string{"checksum", "adler_low", "adler32_cli"} {
		if !apps[app] {
			problems = append(problems, fmt.Sprintf("guide has no run fence for app %q", app))
		}
	}
	mkdirAt, buildAt := -1, -1
	for i, line := range setup {
		if line == "mkdir -p .schway/checksum" && mkdirAt < 0 {
			mkdirAt = i
		}
		if strings.HasPrefix(line, "./schway ") && !strings.HasPrefix(line, "./schway build ") {
			problems = append(problems, fmt.Sprintf("setup line %q must start with `./schway build `", line))
		}
		if strings.HasPrefix(line, "./schway build ") && buildAt < 0 {
			buildAt = i
		}
	}
	if mkdirAt < 0 || buildAt < 0 || mkdirAt > buildAt {
		problems = append(problems, fmt.Sprintf("guide must run `mkdir -p .schway/checksum` before the first build (mkdir=%d build=%d)", mkdirAt, buildAt))
	}
	var doc phase32ExpectedJSON
	if err := json.Unmarshal(expectedJSON, &doc); err != nil {
		return append(problems, "expected json: "+err.Error())
	}
	return append(problems, phase32GuideJSONProblems(guide, doc)...)
}

// TestPhase32GuideParserNegativeControls shows the static check rejecting each
// kind of drift on an in-memory mutated copy of the real guide.
func TestPhase32GuideParserNegativeControls(t *testing.T) {
	raw, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "checksum.expected.json"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if problems := phase32GuideStaticProblems(text, data); len(problems) != 0 {
		t.Fatalf("baseline guide has problems:\n%s", strings.Join(problems, "\n"))
	}
	controls := []struct {
		name   string
		mutate func(string) (string, bool)
	}{
		{"mutated-value", func(s string) (string, bool) {
			return strings.Replace(s, "300286872 | 920", "300286873 | 920", 1), strings.Contains(s, "300286872 | 920")
		}},
		{"renamed-command", func(s string) (string, bool) {
			const anchor = "./schway app run .schway/checksum/checksum -- PATH"
			return strings.Replace(s, anchor, "./schway app runx .schway/checksum/checksum -- PATH", 1), strings.Contains(s, anchor)
		}},
		{"dropped-fence", func(s string) (string, bool) {
			const anchor = "```sh doc-test=run app=adler_low\n./schway app run .schway/checksum/adler_low -- PATH\n```\n"
			return strings.Replace(s, anchor, "", 1), strings.Contains(s, anchor)
		}},
		{"outside-target", func(s string) (string, bool) {
			const anchor = "--output .schway/checksum/checksum"
			return strings.Replace(s, anchor, "--output ./checksum", 1), strings.Contains(s, anchor)
		}},
		{"dropped-row", func(s string) (string, bool) {
			var kept []string
			found := false
			for _, line := range strings.Split(s, "\n") {
				if strings.HasPrefix(line, "| `ff65537` |") {
					found = true
					continue
				}
				kept = append(kept, line)
			}
			return strings.Join(kept, "\n"), found
		}},
	}
	for _, c := range controls {
		t.Run(c.name, func(t *testing.T) {
			mutated, anchored := c.mutate(text)
			if !anchored {
				t.Fatalf("control %s: anchor absent from the README; the control would pass vacuously", c.name)
			}
			if mutated == text {
				t.Fatalf("control %s: mutation changed nothing", c.name)
			}
			if len(phase32GuideStaticProblems(mutated, data)) == 0 {
				t.Fatalf("control %s: the mutated guide produced no problem", c.name)
			}
		})
	}
}

// TestPhase32NativePolicy pins the hosted-runner policy as a pure function.
func TestPhase32NativePolicy(t *testing.T) {
	missing := errors.New("clang not found")
	rows := []struct {
		name string
		err  error
		env  map[string]string
		fail bool
	}{
		{"missing-on-github-actions", missing, map[string]string{"GITHUB_ACTIONS": "true"}, true},
		{"missing-with-require-native", missing, map[string]string{"SCHWAY_REQUIRE_NATIVE": "1"}, true},
		{"missing-local-skips", missing, map[string]string{}, false},
		{"present-never-fails", nil, map[string]string{"GITHUB_ACTIONS": "true", "SCHWAY_REQUIRE_NATIVE": "1"}, false},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			getenv := func(k string) string { return row.env[k] }
			if got := phase32NativePolicy(row.err, getenv); got != row.fail {
				t.Fatalf("phase32NativePolicy = %v, want %v", got, row.fail)
			}
		})
	}
}
