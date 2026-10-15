package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

const (
	phase37GuideChildTimeout = 2 * time.Minute
	phase37GuideOutputLimit  = 32 * 1024
	phase37GuideIndexLimit   = 16 * 1024 * 1024
)

type phase37GuideBlock struct {
	Kind  string
	Attrs map[string]string
	Lines []string
}

type phase37GuideRow struct {
	Case string
	Path string
	CRC  testsupport.CRC32Tuple
	Low  testsupport.CRC32Tuple
}

type phase37Guide struct {
	Blocks      []phase37GuideBlock
	Rows        []phase37GuideRow
	Limit       int
	RefusedSize int
	ErrorExit   int
}

type phase37GuideTuple struct {
	Exit           int
	Stdout, Stderr string
}

type phase37GuideBuffer struct {
	bytes.Buffer
	limit int
	total int
}

func (b *phase37GuideBuffer) Write(p []byte) (int, error) {
	b.total += len(p)
	remaining := b.limit - b.Buffer.Len()
	if remaining > len(p) {
		remaining = len(p)
	}
	if remaining > 0 {
		_, _ = b.Buffer.Write(p[:remaining])
	}
	return len(p), nil
}

func (b *phase37GuideBuffer) overflowed() bool { return b.total > b.limit }

var (
	phase37GuideFenceRE = regexp.MustCompile("^```sh doc-test=([a-z0-9-]+)((?: [a-z0-9_]+=[a-z0-9_]+)*)$")
	phase37GuideLimitRE = regexp.MustCompile(`^Limit: files up to ([0-9,]+) bytes are read; a file of ([0-9,]+) bytes is refused\.$`)
	phase37GuideErrorRE = regexp.MustCompile(`^File-read errors exit ([0-9]+) with empty stdout and one fixed stderr line\.$`)
	phase37GuideTupleRE = regexp.MustCompile(`^\(([0-9]+), ("(?:[^"\\]|\\.)*"), ("(?:[^"\\]|\\.)*")\)$`)
)

var phase37GuideCaseIDs = []string{
	"empty", "zero", "check", "highbit-80ff", "all-bytes-00-ff", "cap-minus-one", "cap", "cap-plus-one",
	"missing", "directory", "opaque-json", "opaque-space", "opaque-shell",
}

func phase37GuideCells(line string) []string {
	line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(line), "|"), "|"))
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, part := range parts {
		cells[i] = strings.TrimSpace(part)
	}
	return cells
}

func phase37ParseGuide(text string) (phase37Guide, error) {
	var guide phase37Guide
	lines := strings.Split(text, "\n")
	table := false
	blockKeys := make(map[string]struct{})
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, "doc-test=") {
			match := phase37GuideFenceRE.FindStringSubmatch(line)
			if match == nil {
				return guide, fmt.Errorf("malformed doc-test fence at line %d", i+1)
			}
			block := phase37GuideBlock{Kind: match[1], Attrs: make(map[string]string)}
			for _, attr := range strings.Fields(match[2]) {
				key, value, ok := strings.Cut(attr, "=")
				if !ok || key == "" || value == "" {
					return guide, fmt.Errorf("malformed fence attribute %q", attr)
				}
				if _, exists := block.Attrs[key]; exists {
					return guide, fmt.Errorf("duplicate fence attribute %q", key)
				}
				block.Attrs[key] = value
			}
			key := fmt.Sprintf("%s|%s|%s", block.Kind, block.Attrs["app"], block.Attrs["route"])
			if _, exists := blockKeys[key]; exists {
				return guide, fmt.Errorf("duplicate doc-test fence %q", key)
			}
			blockKeys[key] = struct{}{}
			closed := false
			pending := ""
			for i++; i < len(lines); i++ {
				if lines[i] == "```" {
					closed = true
					break
				}
				if strings.Contains(lines[i], "doc-test=") {
					return guide, fmt.Errorf("nested or unclosed doc-test fence at line %d", i+1)
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
			if !closed || pending != "" {
				return guide, errors.New("unterminated doc-test fence")
			}
			guide.Blocks = append(guide.Blocks, block)
			continue
		}
		switch {
		case strings.TrimSpace(line) == "| Case | Fixture path | crc32 tuple | crc_low tuple |":
			table = true
		case strings.HasPrefix(strings.TrimSpace(line), "|") && table:
			cells := phase37GuideCells(line)
			if len(cells) == 0 || strings.HasPrefix(cells[0], "---") {
				continue
			}
			if len(cells) != 4 {
				return guide, fmt.Errorf("result row at line %d has %d columns, want 4", i+1, len(cells))
			}
			crc, err := phase37ParseGuideTuple(strings.Trim(cells[2], "`"))
			if err != nil {
				return guide, fmt.Errorf("%s crc32 tuple: %w", cells[0], err)
			}
			low, err := phase37ParseGuideTuple(strings.Trim(cells[3], "`"))
			if err != nil {
				return guide, fmt.Errorf("%s crc_low tuple: %w", cells[0], err)
			}
			guide.Rows = append(guide.Rows, phase37GuideRow{Case: strings.Trim(cells[0], "`"), Path: strings.Trim(cells[1], "`"), CRC: crc, Low: low})
		case line == "":
			table = false
		case table:
			table = false
		}
		if match := phase37GuideLimitRE.FindStringSubmatch(line); match != nil {
			guide.Limit, _ = strconv.Atoi(strings.ReplaceAll(match[1], ",", ""))
			guide.RefusedSize, _ = strconv.Atoi(strings.ReplaceAll(match[2], ",", ""))
		}
		if match := phase37GuideErrorRE.FindStringSubmatch(line); match != nil {
			guide.ErrorExit, _ = strconv.Atoi(match[1])
		}
	}
	if len(guide.Rows) == 0 {
		return guide, errors.New("guide has no case result rows")
	}
	setup := false
	for _, block := range guide.Blocks {
		setup = setup || block.Kind == "setup"
	}
	if !setup {
		return guide, errors.New("guide has no doc-test=setup fence")
	}
	if guide.Limit == 0 || guide.RefusedSize == 0 || guide.ErrorExit == 0 {
		return guide, errors.New("guide is missing its fixed input limit or error facts")
	}
	return guide, nil
}

func phase37ParseGuideTuple(cell string) (testsupport.CRC32Tuple, error) {
	match := phase37GuideTupleRE.FindStringSubmatch(cell)
	if match == nil {
		return testsupport.CRC32Tuple{}, fmt.Errorf("invalid tuple %q", cell)
	}
	exit, err := strconv.Atoi(match[1])
	if err != nil {
		return testsupport.CRC32Tuple{}, err
	}
	stdout, err := strconv.Unquote(match[2])
	if err != nil {
		return testsupport.CRC32Tuple{}, err
	}
	stderr, err := strconv.Unquote(match[3])
	if err != nil {
		return testsupport.CRC32Tuple{}, err
	}
	return testsupport.CRC32Tuple{Exit: exit, Stdout: stdout, Stderr: stderr}, nil
}

func phase37GuideRead(t testing.TB) (string, phase37Guide) {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("examples", "crc32", "README.md"))
	if err != nil {
		t.Fatalf("read CRC-32 guide: %v", err)
	}
	guide, err := phase37ParseGuide(string(data))
	if err != nil {
		t.Fatalf("parse CRC-32 guide: %v", err)
	}
	return string(data), guide
}

func phase37GuideStructureError(guide phase37Guide) error {
	seen := make(map[string]bool)
	for _, block := range guide.Blocks {
		key := fmt.Sprintf("%s|%s|%s", block.Kind, block.Attrs["app"], block.Attrs["route"])
		if seen[key] {
			return fmt.Errorf("duplicate block %s", key)
		}
		seen[key] = true
		switch block.Kind {
		case "setup":
			if len(block.Attrs) != 0 || len(block.Lines) != 3 || block.Lines[0] != "go build -o ./schway ./cmd/schway" || block.Lines[1] != "mkdir -p .schway/crc32/inputs .schway/crc32" || !strings.HasPrefix(block.Lines[2], "python3 -c ") {
				return errors.New("setup fence must build the CLI and create the documented binary fixtures")
			}
		case "build":
			app := block.Attrs["app"]
			if (app != "crc32" && app != "crc_low") || len(block.Attrs) != 1 || len(block.Lines) != 1 {
				return fmt.Errorf("invalid native build block for %q", app)
			}
			if !strings.Contains(block.Lines[0], "./schway build examples/crc32/"+app+".schway --manifest examples/crc32/file_bytes.bindings.json --output .schway/crc32/"+app) {
				return fmt.Errorf("%s retained build command drifted", app)
			}
		case "run":
			app, route := block.Attrs["app"], block.Attrs["route"]
			if (app != "crc32" && app != "crc_low") || len(block.Attrs) != 2 || len(block.Lines) != 1 {
				return fmt.Errorf("invalid run block for %s/%s", app, route)
			}
			want := ""
			switch route {
			case "interpreter":
				want = "./schway run --engine=interpreter examples/crc32/" + app + ".schway -- PATH"
			case "one_shot":
				want = "./schway run --engine=native examples/crc32/" + app + ".schway --manifest examples/crc32/file_bytes.bindings.json -- PATH"
			case "retained":
				want = "./schway app run .schway/crc32/" + app + " -- PATH"
			default:
				return fmt.Errorf("unknown run route %q", route)
			}
			if block.Lines[0] != want {
				return fmt.Errorf("%s/%s command drifted: %q", app, route, block.Lines[0])
			}
		default:
			return fmt.Errorf("unknown doc-test block kind %q", block.Kind)
		}
	}
	for _, key := range []string{"setup||", "build|crc32|", "build|crc_low|"} {
		if !seen[key] {
			return fmt.Errorf("required doc-test block %q is missing", key)
		}
	}
	for _, app := range []string{"crc32", "crc_low"} {
		for _, route := range []string{"interpreter", "one_shot", "retained"} {
			key := "run|" + app + "|" + route
			if !seen[key] {
				return fmt.Errorf("required doc-test block %q is missing", key)
			}
		}
	}
	return nil
}

func phase37GuideFixtureError(guide phase37Guide) error {
	if guide.Limit != testsupport.CRC32Limit || guide.RefusedSize != testsupport.CRC32Limit+1 {
		return fmt.Errorf("guide limit is %d/%d, want %d/%d", guide.Limit, guide.RefusedSize, testsupport.CRC32Limit, testsupport.CRC32Limit+1)
	}
	if guide.ErrorExit != 65 {
		return fmt.Errorf("file-error exit is %d, want 65", guide.ErrorExit)
	}
	cases, err := testsupport.CRC32CaseMatrix()
	if err != nil {
		return err
	}
	caseByID := make(map[string]testsupport.CRC32Case, len(cases))
	for _, row := range cases {
		caseByID[row.ID] = row
	}
	if len(guide.Rows) != len(phase37GuideCaseIDs) {
		return fmt.Errorf("guide has %d rows, want %d", len(guide.Rows), len(phase37GuideCaseIDs))
	}
	for index, row := range guide.Rows {
		wantID := phase37GuideCaseIDs[index]
		if row.Case != wantID {
			return fmt.Errorf("guide row %d is %q, want %q", index+1, row.Case, wantID)
		}
		fixture, ok := caseByID[row.Case]
		if !ok {
			return fmt.Errorf("guide row %q is absent from crc32.expected.json", row.Case)
		}
		wantPath := filepath.ToSlash(filepath.Join(".schway", "crc32", "inputs", fixture.Path))
		if row.Path != wantPath {
			return fmt.Errorf("guide path for %s is %q, want %q", row.Case, row.Path, wantPath)
		}
		if !phase37GuideTupleEqual(row.CRC, fixture.Expected["crc32"]) {
			return fmt.Errorf("guide crc32 tuple for %s is %+v, want %+v", row.Case, row.CRC, fixture.Expected["crc32"])
		}
		if !phase37GuideTupleEqual(row.Low, fixture.Expected["crc_low"]) {
			return fmt.Errorf("guide crc_low tuple for %s is %+v, want %+v", row.Case, row.Low, fixture.Expected["crc_low"])
		}
	}
	return nil
}

func phase37GuideTupleEqual(a, b testsupport.CRC32Tuple) bool {
	return a.Exit == b.Exit && a.Stdout == b.Stdout && a.Stderr == b.Stderr
}

func phase37GuideCommand(t testing.TB, dir, line string) phase37GuideTuple {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), phase37GuideChildTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", line)
	cmd.Dir = dir
	cmd.Env = phase37SetEnv(os.Environ(), "TMPDIR", t.TempDir())
	stdout := &phase37GuideBuffer{limit: phase37GuideOutputLimit}
	stderr := &phase37GuideBuffer{limit: phase37GuideOutputLimit}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("guide command exceeded %s: %q: %v", phase37GuideChildTimeout, line, ctx.Err())
	}
	if stdout.overflowed() || stderr.overflowed() {
		t.Fatalf("guide command output exceeded %d bytes: %q", phase37GuideOutputLimit, line)
	}
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run guide command %q: %v", line, err)
		}
		code = exitErr.ExitCode()
	}
	return phase37GuideTuple{Exit: code, Stdout: stdout.String(), Stderr: stderr.String()}
}

func phase37GuideRunCommand(t testing.TB, dir, tempDir, line string) phase37GuideTuple {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), phase37GuideChildTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", line)
	cmd.Dir = dir
	cmd.Env = phase37SetEnv(os.Environ(), "TMPDIR", tempDir)
	stdout := &phase37GuideBuffer{limit: phase37GuideOutputLimit}
	stderr := &phase37GuideBuffer{limit: phase37GuideOutputLimit}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("guide command exceeded %s: %q: %v", phase37GuideChildTimeout, line, ctx.Err())
	}
	if stdout.overflowed() || stderr.overflowed() {
		t.Fatalf("guide command output exceeded %d bytes: %q", phase37GuideOutputLimit, line)
	}
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run guide command %q: %v", line, err)
		}
		code = exitErr.ExitCode()
	}
	return phase37GuideTuple{Exit: code, Stdout: stdout.String(), Stderr: stderr.String()}
}

func phase37GuideSubstitute(line, token, value string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 || fields[len(fields)-1] != token || strings.Count(line, token) != 1 {
		return "", fmt.Errorf("template %q must end in one %s token", line, token)
	}
	return strings.TrimSuffix(line, token) + phase37GuideQuote(value), nil
}

func phase37GuideQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func phase37GuideCopyTracked(root, dest string) (int, string, map[string]struct{}, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return 0, "", nil, err
	}
	dest, err = filepath.Abs(dest)
	if err != nil {
		return 0, "", nil, err
	}
	if err := os.MkdirAll(dest, 0o700); err != nil {
		return 0, "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", root, "ls-files", "-z")
	listing := &phase37GuideBuffer{limit: phase37GuideIndexLimit}
	stderr := &phase37GuideBuffer{limit: phase37GuideOutputLimit}
	cmd.Stdout, cmd.Stderr = listing, stderr
	err = cmd.Run()
	if ctx.Err() != nil {
		return 0, "", nil, fmt.Errorf("git ls-files exceeded 30s: %w", ctx.Err())
	}
	if err != nil || listing.overflowed() || stderr.overflowed() {
		return 0, "", nil, fmt.Errorf("git ls-files failed: err=%v stderr=%q overflow=%t", err, stderr.String(), listing.overflowed() || stderr.overflowed())
	}
	raw := listing.String()
	if raw != "" && !strings.HasSuffix(raw, "\x00") {
		return 0, "", nil, errors.New("git ls-files output is not NUL terminated")
	}
	paths := strings.Split(raw, "\x00")
	tracked := make(map[string]struct{}, len(paths))
	manifest := &bytes.Buffer{}
	count := 0
	for _, rel := range paths {
		if rel == "" {
			continue
		}
		if path.IsAbs(rel) || path.Clean(rel) != rel || strings.HasPrefix(rel, "../") {
			return 0, "", nil, fmt.Errorf("unsafe tracked path %q", rel)
		}
		src := filepath.Join(root, filepath.FromSlash(rel))
		if !phase37GuideInside(root, src) {
			return 0, "", nil, fmt.Errorf("tracked source %q escapes checkout", rel)
		}
		dst := filepath.Join(dest, filepath.FromSlash(rel))
		if !phase37GuideInside(dest, dst) {
			return 0, "", nil, fmt.Errorf("tracked destination %q escapes copy", rel)
		}
		info, err := os.Lstat(src)
		if err != nil {
			return 0, "", nil, fmt.Errorf("lstat tracked source %q: %w", rel, err)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return 0, "", nil, err
		}
		var kind string
		var size int64
		var sum []byte
		switch {
		case info.Mode().IsRegular():
			in, err := os.Open(src)
			if err != nil {
				return 0, "", nil, fmt.Errorf("open tracked source %q: %w", rel, err)
			}
			out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
			if err != nil {
				_ = in.Close()
				return 0, "", nil, fmt.Errorf("create tracked copy %q: %w", rel, err)
			}
			h := sha256.New()
			size, err = io.Copy(io.MultiWriter(out, h), in)
			inErr := in.Close()
			outErr := out.Close()
			if err == nil {
				err = inErr
			}
			if err == nil {
				err = outErr
			}
			if err != nil {
				return 0, "", nil, fmt.Errorf("copy tracked source %q: %w", rel, err)
			}
			if err := os.Chmod(dst, info.Mode().Perm()); err != nil {
				return 0, "", nil, err
			}
			kind, sum = "file", h.Sum(nil)
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(src)
			if err != nil {
				return 0, "", nil, err
			}
			if filepath.IsAbs(target) {
				return 0, "", nil, fmt.Errorf("absolute tracked symlink %q is outside the reproducible copy contract", rel)
			}
			resolved, err := filepath.EvalSymlinks(filepath.Join(filepath.Dir(src), target))
			if err != nil || !phase37GuideInside(root, resolved) {
				return 0, "", nil, fmt.Errorf("tracked symlink %q does not resolve inside checkout", rel)
			}
			if err := os.Symlink(target, dst); err != nil {
				return 0, "", nil, err
			}
			digest := sha256.Sum256([]byte(target))
			kind, size, sum = "symlink", int64(len(target)), digest[:]
		default:
			return 0, "", nil, fmt.Errorf("tracked path %q is not a file or symlink", rel)
		}
		tracked[rel] = struct{}{}
		count++
		fmt.Fprintf(manifest, "%s\t%s\t%d\t%s\n", base64.RawStdEncoding.EncodeToString([]byte(rel)), kind, size, hex.EncodeToString(sum))
	}
	manifestHash := sha256.Sum256(manifest.Bytes())
	return count, hex.EncodeToString(manifestHash[:]), tracked, nil
}

func phase37GuideInside(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	return err == nil && rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func TestPhase37GuideCleanCheckout(t *testing.T) {
	t.Log("RUN root=TestPhase37GuideCleanCheckout")
	_, guide := phase37GuideRead(t)
	if err := phase37GuideStructureError(guide); err != nil {
		t.Fatal(err)
	}
	if err := phase37GuideFixtureError(guide); err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("clang"); err != nil {
		t.Fatalf("Clang is required by the marked native guide commands: %v", err)
	}
	root := testsupport.ProjectPath()
	copyDir := t.TempDir()
	count, manifestHash, tracked, err := phase37GuideCopyTracked(root, copyDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("SOURCE_MANIFEST files=%d sha256=%s copied=git-ls-files-only", count, manifestHash)
	for _, required := range []string{"examples/crc32/README.md", "cmd/schway/main_phase37_guide_test.go"} {
		if _, exists := tracked[required]; !exists {
			t.Fatalf("git ls-files omitted required source %s", required)
		}
	}
	for _, absent := range []string{".planning/milestone.lock", ".planning/research/.cache"} {
		if _, exists := tracked[absent]; exists {
			t.Fatalf("untracked path %s appeared in git ls-files", absent)
		}
		if _, err := os.Lstat(filepath.Join(copyDir, filepath.FromSlash(absent))); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("untracked path %s was copied: stat error=%v", absent, err)
		}
	}
	for rel := range tracked {
		if rel == ".schway" || strings.HasPrefix(rel, ".schway/") {
			t.Fatalf("ignored .schway output %q appeared in git ls-files", rel)
		}
	}
	for _, block := range guide.Blocks {
		if block.Kind != "setup" {
			continue
		}
		for _, line := range block.Lines {
			t.Logf("RUN guide setup: %s", line)
			got := phase37GuideCommand(t, copyDir, line)
			if got.Exit != 0 {
				t.Fatalf("setup tuple for %q = (%d,%q,%q), want exit 0", line, got.Exit, got.Stdout, got.Stderr)
			}
			t.Logf("PASS guide setup: %s", line)
		}
	}
	cliInfo, err := os.Stat(filepath.Join(copyDir, "schway"))
	if err != nil || !cliInfo.Mode().IsRegular() || cliInfo.Mode().Perm()&0o111 == 0 {
		t.Fatalf("setup did not produce executable schway CLI: info=%v err=%v", cliInfo, err)
	}
	if err := phase37GuideCheckInputs(t, copyDir, guide); err != nil {
		t.Fatal(err)
	}
	for _, block := range guide.Blocks {
		if block.Kind != "build" {
			continue
		}
		t.Logf("RUN guide retained build app=%s", block.Attrs["app"])
		got := phase37GuideCommand(t, copyDir, block.Lines[0])
		if got.Exit != 0 {
			t.Fatalf("retained build %s tuple = (%d,%q,%q)", block.Attrs["app"], got.Exit, got.Stdout, got.Stderr)
		}
		artifact := filepath.Join(copyDir, ".schway", "crc32", block.Attrs["app"])
		info, err := os.Stat(artifact)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
			t.Fatalf("retained build %s did not produce executable artifact: info=%v err=%v", block.Attrs["app"], info, err)
		}
		t.Logf("PASS guide retained build app=%s artifact=%s", block.Attrs["app"], filepath.ToSlash(strings.TrimPrefix(artifact, copyDir+string(filepath.Separator))))
	}
	cases, err := testsupport.CRC32CaseMatrix()
	if err != nil {
		t.Fatal(err)
	}
	caseByID := make(map[string]testsupport.CRC32Case, len(cases))
	for _, row := range cases {
		caseByID[row.ID] = row
	}
	for _, block := range guide.Blocks {
		if block.Kind != "run" {
			continue
		}
		app, route := block.Attrs["app"], block.Attrs["route"]
		for _, advertised := range guide.Rows {
			fixture := caseByID[advertised.Case]
			line, err := phase37GuideSubstitute(block.Lines[0], "PATH", advertised.Path)
			if err != nil {
				t.Fatal(err)
			}
			label := fmt.Sprintf("%s/%s/%s", advertised.Case, app, route)
			t.Logf("RUN guide case=%s command=%s", label, line)
			tmp := t.TempDir()
			got := phase37GuideRunCommand(t, copyDir, tmp, line)
			want := fixture.Expected[app]
			if got.Exit != want.Exit || got.Stdout != want.Stdout || got.Stderr != want.Stderr {
				t.Fatalf("%s tuple = (%d,%q,%q), want (%d,%q,%q)", label, got.Exit, got.Stdout, got.Stderr, want.Exit, want.Stdout, want.Stderr)
			}
			if route == "one_shot" {
				phase37RequireTempClean(t, tmp)
			}
			t.Logf("PASS guide case=%s tuple=(%d,%q,%q)", label, got.Exit, got.Stdout, got.Stderr)
		}
	}
	t.Log("PASS root=TestPhase37GuideCleanCheckout")
}

func phase37GuideCheckInputs(t testing.TB, copyDir string, guide phase37Guide) error {
	t.Helper()
	cases, err := testsupport.CRC32CaseMatrix()
	if err != nil {
		return err
	}
	caseByID := make(map[string]testsupport.CRC32Case, len(cases))
	for _, row := range cases {
		caseByID[row.ID] = row
	}
	oracles, err := testsupport.CRC32OracleRows()
	if err != nil {
		return err
	}
	oracleByID := make(map[string]testsupport.CRC32OracleRow, len(oracles))
	for _, row := range oracles {
		oracleByID[row.ID] = row
	}
	for _, row := range guide.Rows {
		fixture := caseByID[row.Case]
		fullPath := filepath.Join(copyDir, filepath.FromSlash(row.Path))
		switch fixture.Kind {
		case "missing":
			if _, err := os.Lstat(fullPath); !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("missing fixture %s exists or cannot be checked: %v", row.Case, err)
			}
		case "directory":
			info, err := os.Stat(fullPath)
			if err != nil || !info.IsDir() {
				return fmt.Errorf("directory fixture %s: info=%v err=%v", row.Case, info, err)
			}
		default:
			oracle, ok := oracleByID[fixture.Recipe]
			if !ok {
				return fmt.Errorf("fixture %s references absent oracle recipe %s", row.Case, fixture.Recipe)
			}
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return fmt.Errorf("read binary fixture %s: %w", row.Case, err)
			}
			if !bytes.Equal(data, oracle.Input) {
				return fmt.Errorf("binary fixture %s differs from its recipe bytes", row.Case)
			}
			digest := sha256.Sum256(data)
			if hex.EncodeToString(digest[:]) != oracle.SHA256 {
				return fmt.Errorf("binary fixture %s SHA-256 differs from the independent recipe", row.Case)
			}
		}
	}
	return nil
}

func TestPhase37GuideMatchesExpectedJSON(t *testing.T) {
	_, guide := phase37GuideRead(t)
	if err := phase37GuideStructureError(guide); err != nil {
		t.Fatal(err)
	}
	if err := phase37GuideFixtureError(guide); err != nil {
		t.Fatal(err)
	}
	t.Logf("PASS root=TestPhase37GuideMatchesExpectedJSON rows=%d", len(guide.Rows))
}

func TestPhase37GuideParserNegativeControls(t *testing.T) {
	text, guide := phase37GuideRead(t)
	if err := phase37GuideFixtureError(guide); err != nil {
		t.Fatal(err)
	}
	t.Run("missing-fence-close", func(t *testing.T) {
		index := strings.LastIndex(text, "\n```")
		if index < 0 {
			t.Fatal("guide has no closing fence to mutate")
		}
		if _, err := phase37ParseGuide(text[:index] + text[index+len("\n```"):]); err == nil {
			t.Fatal("unterminated marked fence was accepted")
		}
	})
	t.Run("duplicate-fence", func(t *testing.T) {
		start := strings.Index(text, "```sh doc-test=setup")
		end := strings.Index(text[start+1:], "\n```")
		if start < 0 || end < 0 {
			t.Fatal("setup fence is absent")
		}
		end += start + 1 + len("\n```")
		if _, err := phase37ParseGuide(text + "\n" + text[start:end]); err == nil {
			t.Fatal("duplicate marked fence was accepted")
		}
	})
	t.Run("malformed-fence", func(t *testing.T) {
		mutated := strings.Replace(text, "```sh doc-test=setup", "```sh doc-test=setup app=", 1)
		if _, err := phase37ParseGuide(mutated); err == nil {
			t.Fatal("malformed marked fence was accepted")
		}
	})
	t.Run("unknown-block-kind", func(t *testing.T) {
		mutated := strings.Replace(text, "```sh doc-test=run app=crc32 route=interpreter", "```sh doc-test=run-extra app=crc32 route=interpreter", 1)
		parsed, err := phase37ParseGuide(mutated)
		if err != nil {
			t.Fatal(err)
		}
		if err := phase37GuideStructureError(parsed); err == nil || !strings.Contains(err.Error(), "unknown doc-test block kind") {
			t.Fatalf("unknown marked block did not reach the closed-structure assertion: %v", err)
		}
	})
	t.Run("expected-answer-drift", func(t *testing.T) {
		mutated := strings.Replace(text, `(0, "3421780262\n", "")`, `(0, "3421780263\n", "")`, 1)
		changed, err := phase37ParseGuide(mutated)
		if err != nil {
			t.Fatal(err)
		}
		if err := phase37GuideFixtureError(changed); err == nil || !strings.Contains(err.Error(), "tuple") {
			t.Fatalf("changed expected answer did not reach tuple/fixture assertion: %v", err)
		}
	})
	t.Run("limit-drift", func(t *testing.T) {
		mutated := strings.Replace(text, "Limit: files up to 65,536 bytes are read; a file of 65,537 bytes is refused.", "Limit: files up to 65,535 bytes are read; a file of 65,536 bytes is refused.", 1)
		changed, err := phase37ParseGuide(mutated)
		if err != nil {
			t.Fatal(err)
		}
		if err := phase37GuideFixtureError(changed); err == nil || !strings.Contains(err.Error(), "limit") {
			t.Fatalf("changed limit did not reach the fixture assertion: %v", err)
		}
	})
	t.Run("fixed-error-drift", func(t *testing.T) {
		mutated := strings.Replace(text, "File-read errors exit 65 with empty stdout and one fixed stderr line.", "File-read errors exit 64 with empty stdout and one fixed stderr line.", 1)
		changed, err := phase37ParseGuide(mutated)
		if err != nil {
			t.Fatal(err)
		}
		if err := phase37GuideFixtureError(changed); err == nil || !strings.Contains(err.Error(), "error") {
			t.Fatalf("changed error exit did not reach the fixture assertion: %v", err)
		}
	})
	t.Run("wrong-runnable-command-reaches-tuple", func(t *testing.T) {
		cli := testsupport.BuildCLI(t)
		var template string
		for _, block := range guide.Blocks {
			if block.Kind == "run" && block.Attrs["app"] == "crc32" && block.Attrs["route"] == "interpreter" {
				template = block.Lines[0]
				break
			}
		}
		if template == "" {
			t.Fatal("interpreter command is absent")
		}
		work := t.TempDir()
		input := filepath.Join(work, ".schway", "crc32", "inputs", "empty.bin")
		if err := os.MkdirAll(filepath.Dir(input), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(input, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		line := strings.Replace(template, "./schway", phase37GuideQuote(cli), 1)
		line = strings.Replace(line, "examples/crc32/crc32.schway", phase37GuideQuote(testsupport.ProjectPath("examples", "crc32", "crc32.schway")), 1)
		line, err := phase37GuideSubstitute(line, "PATH", ".schway/crc32/inputs/empty.bin")
		if err != nil {
			t.Fatal(err)
		}
		got := phase37GuideCommand(t, work, line)
		if got.Exit != 0 || got.Stdout != "0\n" || got.Stderr != "" {
			t.Fatalf("mutated command failed before its semantic assertion: %+v", got)
		}
		cases, err := testsupport.CRC32CaseMatrix()
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range cases {
			if row.ID == "check" && (got.Exit == row.Expected["crc32"].Exit && got.Stdout == row.Expected["crc32"].Stdout && got.Stderr == row.Expected["crc32"].Stderr) {
				t.Fatal("wrong runnable input unexpectedly matched the check tuple")
			}
		}
		t.Logf("PASS wrong command reached value mismatch: exit=%d stdout=%q", got.Exit, got.Stdout)
	})
	t.Run("untracked-helper-is-excluded", func(t *testing.T) {
		root := t.TempDir()
		readme := strings.Join([]string{
			"# private guide",
			"",
			"Limit: files up to 65,536 bytes are read; a file of 65,537 bytes is refused.",
			"File-read errors exit 65 with empty stdout and one fixed stderr line.",
			"",
			"| Case | Fixture path | crc32 tuple | crc_low tuple |",
			"| --- | --- | --- | --- |",
			"| empty | empty.bin | (0, \"0\\n\", \"\") | (0, \"0\\n\", \"\") |",
			"",
			"```sh doc-test=setup",
			"./phase37-private-helper",
			"```",
			"",
		}, "\n")
		if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(readme), 0o644); err != nil {
			t.Fatal(err)
		}
		helper := filepath.Join(root, "phase37-private-helper")
		if err := os.WriteFile(helper, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"init", "-q", root}, {"-C", root, "add", "--", "README.md"}} {
			cmd := exec.Command("git", args...)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("prepare private repo fixture: %v: %s", err, output)
			}
		}
		copyDir := t.TempDir()
		if _, _, tracked, err := phase37GuideCopyTracked(root, copyDir); err != nil {
			t.Fatal(err)
		} else {
			if _, ok := tracked["phase37-private-helper"]; ok {
				t.Fatal("untracked helper appeared in tracked path set")
			}
		}
		if _, err := os.Lstat(filepath.Join(copyDir, "phase37-private-helper")); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("untracked helper was copied: %v", err)
		}
		privateGuide, err := os.ReadFile(filepath.Join(copyDir, "README.md"))
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := phase37ParseGuide(string(privateGuide))
		if err != nil || len(parsed.Blocks) != 1 {
			t.Fatalf("parse private guide: blocks=%d err=%v", len(parsed.Blocks), err)
		}
		got := phase37GuideCommand(t, copyDir, parsed.Blocks[0].Lines[0])
		if got.Exit != 127 || (!strings.Contains(strings.ToLower(got.Stderr), "not found") && !strings.Contains(strings.ToLower(got.Stderr), "no such file")) {
			t.Fatalf("guide command requiring untracked helper did not fail at missing helper: %+v", got)
		}
		t.Log("PASS untracked helper excluded and the guide command reached the missing-helper failure")
	})
}
