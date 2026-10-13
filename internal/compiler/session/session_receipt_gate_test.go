package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
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

func TestReceiptGate(t *testing.T) {
	testReceiptGatePlan03(t)
	testPortedPhase32ReceiptGate(t)
}

func TestReceiptGateMalformedControls(t *testing.T) {
	root := testsupport.ProjectPath()
	base := string(mustRead(t, testsupport.ProjectPath("testdata", "receipts", "M006.manifest")))
	for _, row := range []struct{ name, controls, reason string }{
		{"empty", "", "controls_line_missing"},
		{"delimiter-only", ",", "controls_line_malformed"},
		{"leading-empty", ",guide", "controls_line_malformed"},
		{"trailing-empty", "guide,", "controls_line_malformed"},
		{"interior-empty", "guide,,codedigest", "controls_line_malformed"},
		{"whitespace-only", " ", "controls_line_malformed"},
	} {
		t.Run(row.name, func(t *testing.T) {
			manifest := filepath.Join(t.TempDir(), "controls.manifest")
			data := strings.Replace(base, "controls=guide,route-audit,codedigest,receipt-rules", "controls="+row.controls, 1)
			if err := os.WriteFile(manifest, []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}
			logs := make([]string, 2)
			for i, host := range []string{"macOS", "Linux"} {
				log := dropReceiptRows(t, phase32ReceiptLog(t, host, phase32GateDigest), "phase32 control ")
				count := strings.Count(row.controls, ",") + 1
				log = strings.Replace(log, "controls=4", fmt.Sprintf("controls=%d", count), 1)
				// Delimiter-only declarations used to accept these arbitrary rows
				// because the count matched but no declared name was checked.
				unknown := strings.Repeat("phase32 control arbitrary-row\n", count)
				logs[i] = strings.Replace(log, "phase32 host result ", unknown+"phase32 host result ", 1)
			}
			got := runGenericReceiptPair(t, root, manifest, logs[0], logs[1], phase32GateDigest)
			want := "phase32 receipt gate: status=fail reason=" + row.reason + "\n"
			if got.exit != 1 || got.stdout != "" || got.stderr != want {
				t.Fatalf("malformed controls verdict: got=%#v want exit=1 stderr=%q", got, want)
			}
		})
	}
}

func testReceiptGatePlan03(t *testing.T) {
	root := testsupport.ProjectPath()
	dir := t.TempDir()
	mac := phase32ReceiptLog(t, "macOS", phase32GateDigest)
	linux := phase32ReceiptLog(t, "Linux", phase32GateDigest)
	logs := []string{filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")}
	for i, content := range []string{mac, linux} {
		if err := os.WriteFile(logs[i], []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	env := append(os.Environ(), "EXPECTED_CODEDIGEST="+phase32GateDigest)
	result, err := testsupport.RunCLIErr(context.Background(), "sh", env,
		"-c", `cd "$0" && sh scripts/check-receipts.sh testdata/receipts/M006.manifest "$1" "$2"`, root, logs[0], logs[1])
	if err != nil && result.Exit == 0 {
		t.Fatalf("spawn: %v", err)
	}
	if result.Exit != 0 {
		t.Fatalf("known-good receipt pair rejected: exit=%d stdout=%s stderr=%s", result.Exit, result.Stdout, result.Stderr)
	}
	t.Run("digest-source-computed", func(t *testing.T) {
		digest := phase32CodeDigest(t, root)
		a, b := filepath.Join(dir, "computed-a.log"), filepath.Join(dir, "computed-b.log")
		for path, host := range map[string]string{a: "macOS", b: "Linux"} {
			if err := os.WriteFile(path, []byte(phase32ReceiptLog(t, host, digest)), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		got := phase33RunAtEnv(t, root, "scripts/check-receipts.sh", envWithout("EXPECTED_CODEDIGEST"), testsupport.ProjectPath("testdata", "receipts", "M006.manifest"), a, b)
		if got.exit != 0 || !strings.Contains(got.stdout, "digest_source=computed") {
			t.Fatalf("computed digest source rejected: %#v", got)
		}
	})

	t.Run("second-label", func(t *testing.T) {
		m006, err := os.ReadFile(testsupport.ProjectPath("testdata", "receipts", "M006.manifest"))
		if err != nil {
			t.Fatal(err)
		}
		manifest := strings.Replace(string(m006), "label=phase32", "label=m999", 1)
		manifest = strings.Replace(manifest, "controls=guide,route-audit,codedigest,receipt-rules", "controls=guide,codedigest", 1)
		manifestPath := filepath.Join(dir, "m999.manifest")
		if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
			t.Fatal(err)
		}
		logsM999 := []string{filepath.Join(dir, "m999-a.log"), filepath.Join(dir, "m999-b.log")}
		for i, host := range []string{"macOS", "Linux"} {
			log := strings.ReplaceAll(phase32ReceiptLog(t, host, phase32GateDigest), "phase32 ", "m999 ")
			for _, name := range []string{"route-audit", "receipt-rules"} {
				log = dropReceiptRows(t, log, "m999 control name="+name+" ")
			}
			log = strings.Replace(log, "controls=4", "controls=2", 1)
			if err := os.WriteFile(logsM999[i], []byte(log), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		got := phase33RunAtEnv(t, root, "scripts/check-receipts.sh", envWith("EXPECTED_CODEDIGEST", phase32GateDigest), manifestPath, logsM999[0], logsM999[1])
		if got.exit != 0 || !strings.Contains(got.stdout, "m999 receipt gate: status=pass") {
			t.Fatalf("second label rejected: %#v", got)
		}
		m006Result := phase33RunAtEnv(t, root, "scripts/check-receipts.sh", envWith("EXPECTED_CODEDIGEST", phase32GateDigest), testsupport.ProjectPath("testdata", "receipts", "M006.manifest"), logsM999[0], logsM999[1])
		if m006Result.exit == 0 {
			t.Fatalf("M006 unexpectedly accepted m999 logs: %#v", m006Result)
		}
	})
	t.Run("identity-not-first", func(t *testing.T) {
		log := "phase32 toolchain go=\"go version\" clang=\"clang version\"\n" + phase32ReceiptLog(t, "macOS", phase32GateDigest)
		got := runGenericReceiptPair(t, root, "testdata/receipts/M006.manifest", log, phase32ReceiptLog(t, "Linux", phase32GateDigest), phase32GateDigest)
		if got.exit == 0 || !strings.Contains(got.stderr, "reason=identity_not_first_") {
			t.Fatalf("wrong verdict: %#v", got)
		}
	})
	t.Run("result-not-last", func(t *testing.T) {
		log := phase32ReceiptLog(t, "macOS", phase32GateDigest) + "phase32 toolchain go=\"go version\" clang=\"clang version\"\n"
		got := runGenericReceiptPair(t, root, "testdata/receipts/M006.manifest", log, phase32ReceiptLog(t, "Linux", phase32GateDigest), phase32GateDigest)
		if got.exit == 0 || !strings.Contains(got.stderr, "reason=result_not_last_") {
			t.Fatalf("wrong verdict: %#v", got)
		}
	})
	t.Run("relative-paths-from-subdirectory", func(t *testing.T) {
		temp := t.TempDir()
		files := map[string][]byte{
			"cmd/a.go":                        []byte("package a\n"),
			"scripts/check-receipts.sh":       mustRead(t, testsupport.ProjectPath("scripts", "check-receipts.sh")),
			"scripts/codedigest.sh":           mustRead(t, testsupport.ProjectPath("scripts", "codedigest.sh")),
			"testdata/receipts/M006.manifest": mustRead(t, testsupport.ProjectPath("testdata", "receipts", "M006.manifest")),
			"logs/a.log":                      []byte(mac), "logs/b.log": []byte(phase32ReceiptLog(t, "Linux", phase32GateDigest)),
		}
		for name, content := range files {
			path := filepath.Join(temp, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, content, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		phase32Git(t, temp, "init", "-q")
		phase32Git(t, temp, "add", "-A")
		phase32Git(t, temp, "commit", "-qm", "fixture")
		command := []string{"-c", `cd "$0/cmd" && exec sh ../scripts/check-receipts.sh ../testdata/receipts/M006.manifest ../logs/a.log ../logs/b.log`, temp}
		result, err := testsupport.RunCLIErr(context.Background(), "sh", envWith("EXPECTED_CODEDIGEST", phase32GateDigest), command...)
		if err != nil || result.Exit != 0 {
			t.Fatalf("relative paths rejected: exit=%d err=%v stderr=%s", result.Exit, err, result.Stderr)
		}
		command = []string{"-c", `cd "$0/cmd" && exec sh ../scripts/check-receipts.sh ../testdata/receipts/M006.manifest ../logs/missing.log ../logs/b.log`, temp}
		result, err = testsupport.RunCLIErr(context.Background(), "sh", envWith("EXPECTED_CODEDIGEST", phase32GateDigest), command...)
		if err != nil || result.Exit == 0 || !strings.Contains(string(result.Stderr), "reason=host_missing_../logs/missing.log") {
			t.Fatalf("typed path reason lost: exit=%d err=%v stderr=%s", result.Exit, err, result.Stderr)
		}
	})
	t.Run("manifest-missing-and-usage", func(t *testing.T) {
		zero := phase33RunAt(t, root, "scripts/check-receipts.sh")
		if zero.exit != 1 || zero.stdout != "" || zero.stderr != "receipt gate: status=fail reason=usage_two_logs_required\n" {
			t.Fatalf("usage bytes: %#v", zero)
		}
		one := phase33RunAt(t, root, "scripts/check-receipts.sh", testsupport.ProjectPath("testdata", "receipts", "M006.manifest"))
		if one.exit != 1 || one.stderr != "phase32 receipt gate: status=fail reason=usage_two_logs_required\n" {
			t.Fatalf("manifest-prefixed usage bytes: %#v", one)
		}
		missing := phase33RunAt(t, root, "scripts/check-receipts.sh", "/nonexistent.manifest", "a", "b")
		if missing.exit != 1 || missing.stderr != "receipt gate: status=fail reason=manifest_missing\n" {
			t.Fatalf("manifest bytes: %#v", missing)
		}
		labelLess := filepath.Join(dir, "label-less.manifest")
		if err := os.WriteFile(labelLess, []byte("programs=\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		got := phase33RunAt(t, root, "scripts/check-receipts.sh", labelLess, logs[0], logs[1])
		if got.exit != 1 || got.stderr != "receipt gate: status=fail reason=manifest_missing\n" {
			t.Fatalf("label-less manifest bytes: %#v", got)
		}
	})
	t.Run("no-repo-root", func(t *testing.T) {
		dir := t.TempDir()
		manifest := filepath.Join(dir, "manifest")
		if err := os.WriteFile(manifest, []byte("label=phase32\nprograms=x\ncontrols=x\n# case x y 0 - -\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		got := phase33RunAtEnv(t, dir, testsupport.ProjectPath("scripts", "check-receipts.sh"), envWith("EXPECTED_CODEDIGEST", phase32GateDigest), manifest, logs[0], logs[1])
		if got.exit != 1 || got.stderr != "phase32 receipt gate: status=fail reason=no_repo_root\n" {
			t.Fatalf("no-repo-root bytes: %#v", got)
		}
	})
}

func envWith(key, value string) []string {
	var env []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, key+"=") {
			env = append(env, entry)
		}
	}
	return append(env, key+"="+value)
}

func envWithout(key string) []string {
	var env []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, key+"=") {
			env = append(env, entry)
		}
	}
	return env
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func dropReceiptRows(t *testing.T, log, prefix string) string {
	t.Helper()
	var kept []string
	for _, line := range strings.Split(strings.TrimSuffix(log, "\n"), "\n") {
		if !strings.HasPrefix(line, prefix) {
			kept = append(kept, line)
		}
	}
	if len(kept) == 0 {
		t.Fatalf("drop removed all rows with prefix %q", prefix)
	}
	return strings.Join(kept, "\n") + "\n"
}

func runGenericReceiptPair(t *testing.T, root, manifest, first, second, digest string) phase33CommandResult {
	t.Helper()
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")
	if err := os.WriteFile(a, []byte(first), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte(second), 0o600); err != nil {
		t.Fatal(err)
	}
	return phase33RunAtEnv(t, root, "scripts/check-receipts.sh", envWith("EXPECTED_CODEDIGEST", digest), manifest, a, b)
}

func TestReceiptGateArchivedM006(t *testing.T) {
	matches, err := phaseArtifactGlob("32-*", "32-RECEIPT.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one archived receipt, found %d: %v", len(matches), matches)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	var receipt phase32Receipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Dir(matches[0])
	macPath := filepath.Join(archive, "receipts", "receipt-macos-latest.log")
	linuxPath := filepath.Join(archive, "receipts", "receipt-ubuntu-latest.log")
	mac, err := os.ReadFile(macPath)
	if err != nil {
		t.Fatal(err)
	}
	linux, err := os.ReadFile(linuxPath)
	if err != nil {
		t.Fatal(err)
	}
	root := testsupport.ProjectPath()
	manifest := testsupport.ProjectPath("testdata", "receipts", "M006.manifest")
	t.Run("archived-pair-passes", func(t *testing.T) {
		pass := phase33RunAtEnv(t, root, "scripts/check-receipts.sh", envWith("EXPECTED_CODEDIGEST", receipt.Codedigest), manifest, macPath, linuxPath)
		want := "phase32 receipt gate: status=pass hosts=macOS,Linux codedigest=" + receipt.Codedigest + " programs=" + strings.Join(phase32ReceiptPrograms, ",") + " cases=39 digest_source=env\n"
		if pass.exit != 0 || pass.stdout != want || pass.stderr != "" {
			t.Fatalf("archived pair: got %#v want %q", pass, want)
		}
	})
	cases := []struct {
		name, reason string
		mutate       func(string, string) (string, string)
	}{
		{"one-byte-mutated", "case_missing_or_altered_", func(a, b string) (string, string) {
			return strings.Replace(a, "phase32 case program=checksum input=empty", "phase32 case program=checksum input=emptY", 1), b
		}},
		{"digest-mismatch", "digest_mismatch_", func(a, b string) (string, string) { return a, b }},
		{"host-dropped", "host_missing_", func(a, b string) (string, string) { return a, "\x00missing" }},
		{"program-missing", "program_missing_fizzbuzz_", func(a, b string) (string, string) { return a, dropReceiptRows(t, b, "phase32 case program=fizzbuzz ") }},
		{"malformed-row", "unknown_phase32_row_", func(a, b string) (string, string) {
			return strings.Replace(a, "phase32 toolchain ", "phase32 bogus row\nphase32 toolchain ", 1), b
		}},
		{"dirty-tree", "tree_not_clean_", func(a, b string) (string, string) { return strings.Replace(a, "tree=clean", "tree=dirty", 1), b }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			a, b := string(mac), string(linux)
			a, b = tc.mutate(a, b)
			ap, bp := filepath.Join(dir, "mac.log"), filepath.Join(dir, "linux.log")
			if err := os.WriteFile(ap, []byte(a), 0o600); err != nil {
				t.Fatal(err)
			}
			if b != "\x00missing" {
				if err := os.WriteFile(bp, []byte(b), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			digest := receipt.Codedigest
			if tc.name == "digest-mismatch" {
				digest = strings.Repeat("0", 64)
			}
			got := phase33RunAtEnv(t, root, "scripts/check-receipts.sh", envWith("EXPECTED_CODEDIGEST", digest), manifest, ap, bp)
			if got.exit == 0 || !strings.Contains(got.stderr, "reason="+tc.reason) {
				t.Fatalf("unexpected verdict %#v", got)
			}
		})
	}
	t.Run("manifest-equals-producer-table", func(t *testing.T) {
		producer, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase32.sh"))
		if err != nil {
			t.Fatal(err)
		}
		producerText := string(producer)
		if !strings.Contains(producerText, "testdata/receipts/M006.manifest") || phase32ProducerHasInlineTable(producerText) {
			t.Fatal("producer does not read manifest or retains a second case/program table")
		}
	})
	if got, err := os.ReadFile(matches[0]); err != nil || !bytes.Equal(got, data) {
		t.Fatalf("archived receipt was modified: err=%v", err)
	}
	if got, err := os.ReadFile(macPath); err != nil || !bytes.Equal(got, mac) {
		t.Fatalf("archived macOS log was modified: err=%v", err)
	}
	if got, err := os.ReadFile(linuxPath); err != nil || !bytes.Equal(got, linux) {
		t.Fatalf("archived Linux log was modified: err=%v", err)
	}
}

func stripCaseRows(content string) string {
	var kept []string
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, "# case ") {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

// phase32GateDigest is the codedigest the literal logs are bound to; the gate
// runner passes the generalized EXPECTED_CODEDIGEST variable.
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
	environment := append(os.Environ(), "EXPECTED_CODEDIGEST="+expectedDigest)
	result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, testsupport.ProjectPath("scripts", "check-receipts.sh"), testsupport.ProjectPath("testdata", "receipts", "M006.manifest"), paths[0], paths[1])
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
		"cmd/a.go":                        []byte("package a\n"),
		".planning/x.md":                  []byte("plan\n"),
		"scripts/codedigest.sh":           script,
		"scripts/check-receipts.sh":       mustRead(t, testsupport.ProjectPath("scripts", "check-receipts.sh")),
		"testdata/receipts/M006.manifest": mustRead(t, testsupport.ProjectPath("testdata", "receipts", "M006.manifest")),
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

func testPortedPhase32ReceiptGate(t *testing.T) {
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
			if !strings.HasPrefix(entry, "EXPECTED_CODEDIGEST=") {
				environment = append(environment, entry)
			}
		}
		result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, testsupport.ProjectPath("scripts", "check-receipts.sh"), testsupport.ProjectPath("testdata", "receipts", "M006.manifest"), paths[0], paths[1])
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

var phase32ProgramsLine = regexp.MustCompile(`(?m)^programs=(.*)$`)

// phase32ParseCaseTable reads every case row from the milestone manifest.
func phase32ParseCaseTable(t *testing.T, manifest string) []phase32TableEntry {
	t.Helper()
	var entries []phase32TableEntry
	for _, line := range strings.Split(manifest, "\n") {
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

func phase32ProducerHasInlineTable(producer string) bool {
	for _, line := range strings.Split(producer, "\n") {
		if strings.HasPrefix(line, "# case ") || strings.HasPrefix(line, "phase32_programs=") {
			return true
		}
	}
	return false
}

func TestPhase32ProducerCaseTable(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "receipts", "M006.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := string(data)
	producerData, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase32.sh"))
	if err != nil {
		t.Fatal(err)
	}
	producer := string(producerData)
	table := phase32ParseCaseTable(t, manifest)
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
		match := phase32ProgramsLine.FindStringSubmatch(manifest)
		if match == nil {
			t.Fatal("programs= line not found")
		}
		if got := strings.Split(match[1], ","); !slices.Equal(got, phase32ReceiptPrograms) {
			t.Fatalf("programs = %v, want phase32ReceiptPrograms %v", got, phase32ReceiptPrograms)
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

	t.Run("manifest-sha256-pinned", func(t *testing.T) {
		const want = "9b6906304a74b67df57e930b297c91c6c7115ecc1e4d98dac563bdaaf446e9c7"
		got := fmt.Sprintf("%x", sha256.Sum256(data))
		if got != want {
			t.Fatalf("manifest sha256=%s want %s", got, want)
		}
	})
	t.Run("producer-reads-manifest", func(t *testing.T) {
		if !strings.Contains(producer, "testdata/receipts/M006.manifest") || phase32ProducerHasInlineTable(producer) {
			t.Fatal("producer does not read the manifest or retains a second table")
		}
	})
	t.Run("rows-in-archived-logs", func(t *testing.T) {
		logs, err := phaseArtifactGlob("32-*", "32-RECEIPT.json")
		if err != nil || len(logs) != 1 {
			t.Fatalf("archived receipt lookup: %v %v", logs, err)
		}
		base := filepath.Join(filepath.Dir(logs[0]), "receipts")
		for _, name := range []string{"receipt-macos-latest.log", "receipt-ubuntu-latest.log"} {
			content, err := os.ReadFile(filepath.Join(base, name))
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range table {
				want := fmt.Sprintf("phase32 case program=%s input=%s expected_exit=%d actual_exit=%d stdout=exact stderr=exact status=pass", c.Program, c.Key, c.Exit, c.Exit)
				if strings.Count(string(content), want) != 1 {
					t.Errorf("%s: row %q occurs %d times", name, want, strings.Count(string(content), want))
				}
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
		mutated := strings.ReplaceAll(manifest, "# case byte_index_oob ", "# dropped byte_index_oob ")
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

func phase32GateHead(t *testing.T, dir string) string {
	t.Helper()
	result, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", dir, "rev-parse", "HEAD")
	if err != nil || result.Exit != 0 {
		t.Fatalf("git rev-parse in %s: exit=%d err=%v", dir, result.Exit, err)
	}
	return strings.TrimSpace(string(result.Stdout))
}
