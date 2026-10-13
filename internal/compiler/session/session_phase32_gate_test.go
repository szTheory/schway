package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase32GateDigest is the codedigest the literal logs are bound to; the gate
// runner passes it as PHASE32_EXPECTED_CODEDIGEST.
var phase32GateDigest = strings.Repeat("ab", 32)

const phase32GateHeadSha = "0123456789abcdef0123456789abcdef01234567"

// phase32GateCase is one entry of the shared case table in the notation of
// scripts/verify-phase32.sh: Stdout and Stderr are "-" for empty.
type phase32GateCase struct {
	Program, Key string
	Exit         int
	Stdout       string
	Stderr       string
}

// phase32GateCases is the literal case list the known-good logs are built
// from. It is written out here, independent of the producer's table;
// TestPhase32ProducerCaseTable proves the two agree.
var phase32GateCases = []phase32GateCase{
	{"checksum", "empty", 0, "1", "-"},
	{"checksum", "zero", 0, "65537", "-"},
	{"checksum", "wikipedia", 0, "300286872", "-"},
	{"checksum", "ff65536", 0, "2006388466", "-"},
	{"checksum", "ff65537", 65, "-", "FileError.FileTooLarge"},
	{"checksum", "missing", 65, "-", "FileError.FileUnavailable"},
	{"checksum", "directory", 65, "-", "FileError.FileUnavailable"},
	{"adler_low", "empty", 0, "1", "-"},
	{"adler_low", "zero", 0, "1", "-"},
	{"adler_low", "wikipedia", 0, "920", "-"},
	{"adler_low", "ff65536", 0, "3826", "-"},
	{"adler_low", "ff65537", 65, "-", "FileError.FileTooLarge"},
	{"adler_low", "missing", 65, "-", "FileError.FileUnavailable"},
	{"adler_low", "directory", 65, "-", "FileError.FileUnavailable"},
	{"adler32_cli", "3", 0, "300286872", "-"},
	{"adler32_cli", "10", 65, "-", "-"},
	{"adler32_fixture", "0", 0, "1", "-"},
	{"adler32_fixture", "1", 0, "65537", "-"},
	{"adler32_fixture", "2", 0, "6422626", "-"},
	{"adler32_fixture", "3", 0, "300286872", "-"},
	{"adler32_fixture", "4", 0, "268435457", "-"},
	{"adler32_fixture", "5", 0, "4052720524", "-"},
	{"adler32_fixture", "6", 0, "33620352", "-"},
	{"adler32_fixture", "7", 0, "2170679522", "-"},
	{"adler32_fixture", "8", 0, "2967061426", "-"},
	{"adler32_fixture", "9", 0, "2006388466", "-"},
	{"adler32_fixture", "10", 65, "-", "-"},
	{"byte_index_oob", "0", 65, "-", "schway: byte index out of bounds"},
	{"byte_index_oob", "2", 65, "-", "schway: byte index out of bounds"},
	{"sum_to_n", "0", 0, "0", "-"},
	{"sum_to_n", "1", 0, "1", "-"},
	{"sum_to_n", "15", 0, "120", "-"},
	{"sum_to_n", "1000", 0, "500500", "-"},
	{"sum_to_n", "1001", 65, "-", "schway app: input exceeds 1000"},
	{"fizzbuzz", "0", 0, "-", "-"},
	{"fizzbuzz", "1", 0, "1", "-"},
	{"fizzbuzz", "15", 0, "@examples/phase27/fizzbuzz-15.out", "-"},
	{"fizzbuzz", "1000", 0, "@examples/phase27/fizzbuzz-1000.out", "-"},
	{"fizzbuzz", "1001", 65, "-", "schway app: input exceeds 1000"},
}

func phase32ReceiptLog(t *testing.T, host, digest string) string {
	t.Helper()
	return phase32ReceiptLogAt(t, host, digest, phase32GateHeadSha)
}

// phase32ReceiptLogAt builds a literal known-good log for one host. Nothing
// here is read from scripts/verify-phase32.sh, so the gate cannot pass by
// echoing its own producer.
func phase32ReceiptLogAt(t *testing.T, host, digest, headsha string) string {
	t.Helper()
	var lines []string
	lines = append(lines, "phase32 host identity host="+host+" arch=x86_64 go=linux/amd64 clang_target=x86_64-pc-linux-gnu headsha="+headsha+" codedigest="+digest+" tree=clean")
	lines = append(lines, `phase32 toolchain go="go version go1.24.0" clang="clang version"`)
	for _, c := range phase32GateCases {
		lines = append(lines, fmt.Sprintf("phase32 case program=%s input=%s expected_exit=%d actual_exit=%d stdout=exact stderr=exact status=pass", c.Program, c.Key, c.Exit, c.Exit))
	}
	lines = append(lines,
		"phase32 control name=guide package=./cmd/schway tests=TestPhase32GuideCleanCheckout status=pass",
		"phase32 control name=route-audit package=./cmd/schway tests=TestPhase32PublicRouteAudit status=pass",
		"phase32 control name=codedigest package=./internal/compiler/session tests=TestPhase32CodeDigest status=pass",
		"phase32 control name=receipt-rules package=./internal/compiler/session tests=TestPhase32ReceiptLivingClaims status=pass",
	)
	lines = append(lines, fmt.Sprintf("phase32 host result host=%s codedigest=%s tree=clean cases=%d controls=4 programs=%s status=pass",
		host, digest, len(phase32GateCases), strings.Join(phase32ReceiptPrograms, ",")))
	return strings.Join(lines, "\n") + "\n"
}

func runPhase32Gate(t *testing.T, expectedDigest, first, second string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	paths := []string{filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")}
	for index, content := range []string{first, second} {
		if content == "\x00missing" {
			continue
		}
		if err := os.WriteFile(paths[index], []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	environment := append(os.Environ(), "PHASE32_EXPECTED_CODEDIGEST="+expectedDigest)
	result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, testsupport.ProjectPath("scripts", "check-phase32-receipts.sh"), paths[0], paths[1])
	if err != nil {
		return "", err
	}
	output := string(result.Stdout) + string(result.Stderr)
	if result.Exit != 0 {
		return output, errors.New("gate rejected the receipt pair")
	}
	return output, nil
}

// phase32GateTempRepo makes a repo whose code digest the sh script computes.
func phase32GateTempRepo(t *testing.T) string {
	t.Helper()
	script, err := os.ReadFile(testsupport.ProjectPath("scripts", "codedigest.sh"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for name, content := range map[string][]byte{
		"cmd/a.go":              []byte("package a\n"),
		".planning/x.md":        []byte("plan\n"),
		"scripts/codedigest.sh": script,
	} {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	phase32Git(t, dir, "init", "-q")
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "base")
	return dir
}

func phase32GateHead(t *testing.T, dir string) string {
	t.Helper()
	result, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", dir, "rev-parse", "HEAD")
	if err != nil || result.Exit != 0 {
		t.Fatalf("git rev-parse in %s: exit=%d err=%v", dir, result.Exit, err)
	}
	return strings.TrimSpace(string(result.Stdout))
}

func TestPhase32ReceiptGate(t *testing.T) {
	mac := phase32ReceiptLog(t, "macOS", phase32GateDigest)
	linux := phase32ReceiptLog(t, "Linux", phase32GateDigest)

	mutate := func(t *testing.T, log, old, replacement string) string {
		t.Helper()
		if !strings.Contains(log, old) {
			t.Fatalf("mutation anchor %q not present", old)
		}
		return strings.Replace(log, old, replacement, 1)
	}
	withoutAll := func(t *testing.T, log, prefix string) string {
		t.Helper()
		var kept []string
		dropped := 0
		for _, line := range strings.Split(strings.TrimSuffix(log, "\n"), "\n") {
			if strings.HasPrefix(line, prefix) {
				dropped++
				continue
			}
			kept = append(kept, line)
		}
		if dropped == 0 {
			t.Fatalf("no line with prefix %q", prefix)
		}
		return strings.Join(kept, "\n") + "\n"
	}
	reject := func(t *testing.T, digest, first, second, reason string) {
		t.Helper()
		output, err := runPhase32Gate(t, digest, first, second)
		if err == nil || !strings.Contains(output, "status=fail") || !strings.Contains(output, reason) {
			t.Fatalf("mutated pair was not rejected with %q: err=%v output=%s", reason, err, output)
		}
		if strings.Contains(output, "status=pass") {
			t.Fatalf("rejected pair also reported pass: %s", output)
		}
	}

	t.Run("known-good", func(t *testing.T) {
		output, err := runPhase32Gate(t, phase32GateDigest, mac, linux)
		if err != nil || !strings.Contains(output, "status=pass hosts=macOS,Linux") {
			t.Fatalf("known-good pair rejected: err=%v output=%s", err, output)
		}
		if !strings.Contains(output, "programs="+strings.Join(phase32ReceiptPrograms, ",")) {
			t.Fatalf("pass line lacks the program list: %s", output)
		}
		if !strings.HasSuffix(strings.TrimSpace(output), "digest_source=env") {
			t.Fatalf("pass line must end with digest_source=env when the env digest is supplied: %s", output)
		}
	})
	t.Run("digest-source-computed", func(t *testing.T) {
		digest := phase32CodeDigest(t, testsupport.ProjectPath())
		dir := t.TempDir()
		paths := []string{filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")}
		for index, content := range []string{phase32ReceiptLog(t, "macOS", digest), phase32ReceiptLog(t, "Linux", digest)} {
			if err := os.WriteFile(paths[index], []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		var environment []string
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "PHASE32_EXPECTED_CODEDIGEST=") {
				environment = append(environment, entry)
			}
		}
		result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, testsupport.ProjectPath("scripts", "check-phase32-receipts.sh"), paths[0], paths[1])
		output := string(result.Stdout) + string(result.Stderr)
		if err != nil || result.Exit != 0 || !strings.Contains(output, "status=pass hosts=macOS,Linux") || !strings.HasSuffix(strings.TrimSpace(output), "digest_source=computed") {
			t.Fatalf("computed-digest pass line wrong: err=%v exit=%d output=%s", err, result.Exit, output)
		}
	})
	t.Run("host-order-independent", func(t *testing.T) {
		if output, err := runPhase32Gate(t, phase32GateDigest, linux, mac); err != nil {
			t.Fatalf("host order must not matter: err=%v output=%s", err, output)
		}
	})
	t.Run("digest-mismatch", func(t *testing.T) {
		other := strings.Repeat("cd", 32)
		reject(t, phase32GateDigest, phase32ReceiptLog(t, "macOS", other), phase32ReceiptLog(t, "Linux", other), "digest_mismatch")
	})
	t.Run("code-edit", func(t *testing.T) {
		dir := phase32GateTempRepo(t)
		before := phase32ShDigest(t, dir)
		path := filepath.Join(dir, "cmd", "a.go")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(content, '!'), 0o644); err != nil {
			t.Fatal(err)
		}
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "edit one byte")
		after := phase32ShDigest(t, dir)
		if before == after {
			t.Fatal("a .go byte edit did not move the digest")
		}
		reject(t, after, phase32ReceiptLog(t, "macOS", before), phase32ReceiptLog(t, "Linux", before), "digest_mismatch")
	})
	t.Run("tree-modified", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, mutate(t, linux, "tree=clean", "tree=modified"), "tree_not_clean")
	})
	t.Run("host-missing", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, "\x00missing", "host_missing")
	})
	t.Run("duplicate-host", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, mac, "duplicate_host_macOS")
	})
	t.Run("program-rows-dropped", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, withoutAll(t, linux, "phase32 case program=adler_low "), "program_missing_adler_low")
	})
	t.Run("oob-rows-dropped", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, withoutAll(t, linux, "phase32 case program=byte_index_oob "), "program_missing_byte_index_oob")
	})
	t.Run("programs-list-swapped", func(t *testing.T) {
		swapped := slices.Clone(phase32ReceiptPrograms)
		swapped[0], swapped[len(swapped)-1] = swapped[len(swapped)-1], swapped[0]
		reject(t, phase32GateDigest, mac, mutate(t, linux, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs="+strings.Join(swapped, ",")), "result_programs_mismatch")
	})
	t.Run("unknown-row", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, mutate(t, linux, "phase32 host result", "phase32 forged row status=pass\nphase32 host result"), "unknown_phase32_row")
	})
	t.Run("case-altered", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, mutate(t, linux, "stdout=exact", "stdout=mismatch"), "case_missing_or_altered")
	})
	t.Run("case-exit-altered", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, mutate(t, linux, "program=sum_to_n input=1001 expected_exit=65 actual_exit=65", "program=sum_to_n input=1001 expected_exit=0 actual_exit=0"), "case_missing_or_altered")
	})
	t.Run("failed-case-status", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, mutate(t, linux, "stderr=exact status=pass", "stderr=exact status=fail"), "failed_or_unavailable_row")
	})
	t.Run("control-dropped", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, withoutAll(t, linux, "phase32 control name=route-audit "), "control_row_count")
	})
	t.Run("control-altered", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, mutate(t, linux, "name=guide package=./cmd/schway", "name=guide package="), "control_missing_or_altered_guide")
	})
	t.Run("result-not-last", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, linux+"phase32 toolchain trailing\n", "result_not_last")
	})
	t.Run("log-digest-differs", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, phase32ReceiptLog(t, "Linux", strings.Repeat("cd", 32)), "log_digest_differs")
	})
	t.Run("missing-identity", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, withoutAll(t, linux, "phase32 host identity "), "identity_row_count")
	})
	t.Run("empty-log", func(t *testing.T) {
		reject(t, phase32GateDigest, mac, "", "log_empty")
	})
	t.Run("planning-only-head-move", func(t *testing.T) {
		dir := phase32GateTempRepo(t)
		digest1, head1 := phase32ShDigest(t, dir), phase32GateHead(t, dir)
		path := filepath.Join(dir, ".planning", "x.md")
		if err := os.WriteFile(path, []byte("plan, edited\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "planning only")
		digest2, head2 := phase32ShDigest(t, dir), phase32GateHead(t, dir)
		if head1 == head2 || digest1 != digest2 {
			t.Fatalf("setup: head moved=%v digest moved=%v", head1 != head2, digest1 != digest2)
		}
		first := phase32ReceiptLogAt(t, "macOS", digest1, head1)
		second := phase32ReceiptLogAt(t, "Linux", digest1, head1)
		output, err := runPhase32Gate(t, digest2, first, second)
		if err != nil || !strings.Contains(output, "status=pass") {
			t.Fatalf("a .planning-only HEAD move must keep the receipt current: err=%v output=%s", err, output)
		}
	})
}

type phase32TableEntry struct {
	Program, Key string
	Exit         int
	Stdout       string
	Stderr       string
}

var phase32ProgramsLine = regexp.MustCompile(`(?m)^phase32_programs=(.*)$`)

// phase32ParseCaseTable reads the entries between the table markers of the
// producer script text.
func phase32ParseCaseTable(t *testing.T, script string) []phase32TableEntry {
	t.Helper()
	begin := strings.Index(script, "\n# phase32-cases-begin\n")
	end := strings.Index(script, "\n# phase32-cases-end\n")
	if begin < 0 || end < begin {
		t.Fatal("case table markers not found")
	}
	var entries []phase32TableEntry
	for _, line := range strings.Split(script[begin:end], "\n") {
		rest, ok := strings.CutPrefix(line, "# case ")
		if !ok {
			continue
		}
		fields := strings.SplitN(rest, " ", 5)
		if len(fields) != 5 {
			t.Fatalf("malformed case line %q", line)
		}
		code, err := strconv.Atoi(fields[2])
		if err != nil {
			t.Fatalf("case line %q: %v", line, err)
		}
		entries = append(entries, phase32TableEntry{fields[0], fields[1], code, fields[3], fields[4]})
	}
	return entries
}

func phase32NotationOf(text string) string {
	if text == "" {
		return "-"
	}
	return strings.TrimSuffix(text, "\n")
}

func TestPhase32ProducerCaseTable(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase32.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	table := phase32ParseCaseTable(t, script)
	byProgram := map[string][]phase32TableEntry{}
	for _, entry := range table {
		byProgram[entry.Program] = append(byProgram[entry.Program], entry)
	}
	find := func(program, key string) phase32TableEntry {
		t.Helper()
		for _, entry := range byProgram[program] {
			if entry.Key == key {
				return entry
			}
		}
		t.Fatalf("no table entry for %s %s", program, key)
		return phase32TableEntry{}
	}
	wantCounts := map[string]int{"checksum": 7, "adler_low": 7, "adler32_cli": 2, "adler32_fixture": 11, "byte_index_oob": 2, "sum_to_n": 5, "fizzbuzz": 5}

	t.Run("program-set", func(t *testing.T) {
		match := phase32ProgramsLine.FindStringSubmatch(script)
		if match == nil {
			t.Fatal("phase32_programs= line not found")
		}
		if got := strings.Split(match[1], ","); !slices.Equal(got, phase32ReceiptPrograms) {
			t.Fatalf("phase32_programs = %v, want phase32ReceiptPrograms %v", got, phase32ReceiptPrograms)
		}
		if len(table) != 39 {
			t.Fatalf("table has %d entries, want 39", len(table))
		}
		for program, want := range wantCounts {
			if len(byProgram[program]) != want {
				t.Errorf("%s has %d entries, want %d", program, len(byProgram[program]), want)
			}
		}
		for program := range byProgram {
			if _, ok := wantCounts[program]; !ok {
				t.Errorf("table names unknown program %q", program)
			}
		}
	})

	t.Run("go-literal-list-agrees", func(t *testing.T) {
		var got []phase32TableEntry
		for _, c := range phase32GateCases {
			got = append(got, phase32TableEntry{c.Program, c.Key, c.Exit, c.Stdout, c.Stderr})
		}
		if !slices.Equal(got, table) {
			t.Fatalf("phase32GateCases differ from the producer table:\n got %v\nwant %v", got, table)
		}
	})

	t.Run("checksum-expected-json", func(t *testing.T) {
		raw, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "checksum.expected.json"))
		if err != nil {
			t.Fatal(err)
		}
		var expected struct {
			Cases []struct {
				Name   string `json:"name"`
				Exit   int    `json:"exit"`
				Stdout string `json:"stdout"`
				Stderr string `json:"stderr"`
			} `json:"cases"`
		}
		if err := json.Unmarshal(raw, &expected); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"empty", "zero", "wikipedia", "ff65536", "ff65537", "missing", "directory"} {
			found := false
			for _, c := range expected.Cases {
				if c.Name != key {
					continue
				}
				found = true
				entry := find("checksum", key)
				if entry.Exit != c.Exit || entry.Stdout != phase32NotationOf(c.Stdout) || entry.Stderr != phase32NotationOf(c.Stderr) {
					t.Errorf("checksum %s = %+v, want exit=%d stdout=%q stderr=%q", key, entry, c.Exit, c.Stdout, c.Stderr)
				}
			}
			if !found {
				t.Errorf("checksum.expected.json lacks %s", key)
			}
		}
	})

	t.Run("adler-low-oracle", func(t *testing.T) {
		names := map[string]string{"empty": "empty", "zero": "0x00", "wikipedia": "Wikipedia", "ff65536": "65536x0xFF", "ff65537": "65537x0xFF"}
		full := map[string]testsupport.ChecksumOracleRow{}
		low := map[string]testsupport.ChecksumOracleRow{}
		for _, row := range testsupport.ChecksumOracleRows() {
			switch row.App {
			case "checksum":
				full[row.Name] = row
			case "adler_low":
				low[row.Name] = row
			}
		}
		for key, name := range names {
			entry, whole := find("adler_low", key), find("checksum", key)
			if low[name].Outcome == "returned" {
				if entry.Exit != 0 || entry.Stdout != strconv.FormatUint(low[name].Want, 10) || entry.Stderr != "-" {
					t.Errorf("adler_low %s = %+v, want oracle %d", key, entry, low[name].Want)
				}
				if low[name].Want != full[name].Want%65536 {
					t.Errorf("oracle rows disagree for %s: low %d, full %d", key, low[name].Want, full[name].Want)
				}
				continue
			}
			if entry.Exit != whole.Exit || entry.Stdout != whole.Stdout || entry.Stderr != whole.Stderr {
				t.Errorf("adler_low %s error row %+v differs from checksum %+v", key, entry, whole)
			}
		}
		for _, key := range []string{"missing", "directory"} {
			entry, whole := find("adler_low", key), find("checksum", key)
			if entry.Exit != whole.Exit || entry.Stdout != whole.Stdout || entry.Stderr != whole.Stderr {
				t.Errorf("adler_low %s error row %+v differs from checksum %+v", key, entry, whole)
			}
		}
	})

	t.Run("phase29-fixture-vectors", func(t *testing.T) {
		for _, program := range []string{"adler32_cli", "adler32_fixture"} {
			for _, entry := range byProgram[program] {
				id, err := strconv.ParseUint(entry.Key, 10, 64)
				if err != nil {
					t.Fatalf("%s key %q: %v", program, entry.Key, err)
				}
				want, ok := testsupport.Phase29FixtureExpected(id)
				if !ok {
					if entry.Exit != 65 || entry.Stdout != "-" || entry.Stderr != "-" {
						t.Errorf("%s %d is outside the table and must exit 65 silently: %+v", program, id, entry)
					}
					continue
				}
				if entry.Exit != 0 || entry.Stdout != strconv.FormatUint(want, 10) || entry.Stderr != "-" {
					t.Errorf("%s %d = %+v, want %d", program, id, entry, want)
				}
			}
		}
		for id := uint64(0); id <= 9; id++ {
			if _, ok := testsupport.Phase29FixtureExpected(id); !ok {
				t.Fatalf("fixture case %d has no pinned checksum", id)
			}
			find("adler32_fixture", strconv.FormatUint(id, 10))
		}
		find("adler32_fixture", "10")
		find("adler32_cli", "3")
		find("adler32_cli", "10")
	})

	t.Run("byte-index-oob", func(t *testing.T) {
		const want = "schway: byte index out of bounds"
		source, err := os.ReadFile(testsupport.ProjectPath("cmd", "schway", "main_phase29_test.go"))
		if err != nil || !strings.Contains(string(source), `stderr != "`+want+`\n"`) {
			t.Fatalf("TestPhase29OOBControlAppCLI no longer pins the BYTE-02 stderr %q (err=%v)", want, err)
		}
		for _, key := range []string{"0", "2"} {
			if entry := find("byte_index_oob", key); entry.Exit != 65 || entry.Stdout != "-" || entry.Stderr != want {
				t.Errorf("byte_index_oob %s = %+v", key, entry)
			}
		}
	})

	t.Run("phase27-fixtures", func(t *testing.T) {
		const over = "schway app: input exceeds 1000"
		for key, want := range map[string]string{"0": "0", "1": "1", "15": "120", "1000": "500500"} {
			if entry := find("sum_to_n", key); entry.Exit != 0 || entry.Stdout != want || entry.Stderr != "-" {
				t.Errorf("sum_to_n %s = %+v, want stdout %s", key, entry, want)
			}
		}
		if entry := find("sum_to_n", "1001"); entry.Exit != 65 || entry.Stdout != "-" || entry.Stderr != over {
			t.Errorf("sum_to_n 1001 = %+v", entry)
		}
		fizz := map[string]string{"0": "-", "1": "1", "15": "@examples/phase27/fizzbuzz-15.out", "1000": "@examples/phase27/fizzbuzz-1000.out"}
		for key, want := range fizz {
			if entry := find("fizzbuzz", key); entry.Exit != 0 || entry.Stdout != want || entry.Stderr != "-" {
				t.Errorf("fizzbuzz %s = %+v, want stdout %s", key, entry, want)
			}
			if strings.HasPrefix(want, "@") {
				if _, err := os.Stat(testsupport.ProjectPath(strings.Split(strings.TrimPrefix(want, "@"), "/")...)); err != nil {
					t.Errorf("fizzbuzz %s fixture: %v", key, err)
				}
			}
		}
		if entry := find("fizzbuzz", "1001"); entry.Exit != 65 || entry.Stdout != "-" || entry.Stderr != over {
			t.Errorf("fizzbuzz 1001 = %+v", entry)
		}
	})

	t.Run("dropping-a-program-is-detected", func(t *testing.T) {
		mutated := strings.ReplaceAll(script, "# case byte_index_oob ", "# dropped byte_index_oob ")
		count := 0
		for _, entry := range phase32ParseCaseTable(t, mutated) {
			if entry.Program == "byte_index_oob" {
				count++
			}
		}
		if count == 2 || len(phase32ParseCaseTable(t, mutated)) == len(table) {
			t.Fatal("the table parser did not notice dropped rows")
		}
	})
}
