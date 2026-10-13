package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase32DigestPaths is the receipt contract (D-08a): the paths whose HEAD
// listing is hashed. scripts/codedigest.sh names the same set.
var phase32DigestPaths = []string{"cmd", "internal", "examples", "native", "scripts", "testdata", "go.mod", ".github/workflows/ci.yml"}

// phase32CodeDigest is the Go twin of scripts/codedigest.sh.
func phase32CodeDigest(t testing.TB, root string) string {
	t.Helper()
	arguments := append([]string{"-C", root, "-c", "core.quotePath=false", "ls-tree", "-r", "HEAD", "--"}, phase32DigestPaths...)
	result, err := testsupport.RunCLIErr(context.Background(), "git", nil, arguments...)
	if err != nil || result.Exit != 0 {
		t.Fatalf("git ls-tree in %s: exit=%d err=%v stderr=%s", root, result.Exit, err, result.Stderr)
	}
	var kept []string
	for _, line := range strings.Split(strings.TrimRight(string(result.Stdout), "\n"), "\n") {
		if line == "" || strings.HasSuffix(line, ".md") {
			continue
		}
		kept = append(kept, line)
	}
	sum := sha256.Sum256([]byte(strings.Join(kept, "\n") + "\n"))
	return hex.EncodeToString(sum[:])
}

func phase32ShDigest(t testing.TB, root string) string {
	t.Helper()
	result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, "-c", `cd "$0" && sh scripts/codedigest.sh`, root)
	if err != nil || result.Exit != 0 {
		t.Fatalf("codedigest.sh in %s: exit=%d err=%v stderr=%s", root, result.Exit, err, result.Stderr)
	}
	return strings.TrimSpace(string(result.Stdout))
}

func phase32Git(t testing.TB, dir string, arguments ...string) {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=phase32", "-c", "user.email=phase32@example.invalid"}, arguments...)
	result, err := testsupport.RunCLIErr(context.Background(), "git", nil, full...)
	if err != nil || result.Exit != 0 {
		t.Fatalf("git %v: exit=%d err=%v stderr=%s", arguments, result.Exit, err, result.Stderr)
	}
}

func TestPhase32CodeDigest(t *testing.T) {
	root := testsupport.ProjectPath()
	t.Run("sh-equals-go", func(t *testing.T) {
		sh, goDigest := phase32ShDigest(t, root), phase32CodeDigest(t, root)
		if sh != goDigest {
			t.Fatalf("sh digest %s != go digest %s", sh, goDigest)
		}
	})

	script, err := os.ReadFile(testsupport.ProjectPath("scripts", "codedigest.sh"))
	if err != nil {
		t.Fatal(err)
	}
	tempRepo := func(t *testing.T) string {
		dir := t.TempDir()
		for name, content := range map[string][]byte{
			"cmd/a.go":              []byte("package a\n"),
			"cmd/README.md":         []byte("readme\n"),
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
	t.Run("subdirectory-equals-root", func(t *testing.T) {
		dir := tempRepo(t)
		root := phase32ShDigest(t, dir)
		result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, "-c", `cd "$0/cmd" && sh ../scripts/codedigest.sh`, dir)
		if err != nil || result.Exit != 0 {
			t.Fatalf("codedigest.sh from subdirectory: exit=%d err=%v stderr=%s", result.Exit, err, result.Stderr)
		}
		if sub := strings.TrimSpace(string(result.Stdout)); sub != root || sub != phase32CodeDigest(t, dir) {
			t.Fatalf("subdirectory digest %s != root %s", sub, root)
		}
	})
	t.Run("empty-listing-refused", func(t *testing.T) {
		dir := t.TempDir()
		for name, content := range map[string][]byte{"tools/codedigest.sh": script, "other.txt": []byte("x\n")} {
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
		result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, "-c", `cd "$0" && sh tools/codedigest.sh`, dir)
		if err != nil && result.Exit == 0 {
			t.Fatalf("spawn: %v", err)
		}
		out := string(result.Stdout) + string(result.Stderr)
		if result.Exit == 0 || !strings.Contains(out, "reason=empty_listing") {
			t.Fatalf("expected reason=empty_listing, exit=%d out=%s", result.Exit, out)
		}
	})
	for _, row := range []struct {
		name, file string
		moves      bool
	}{
		{"armed-code-edit", "cmd/a.go", true},
		{"md-edit-ignored", "cmd/README.md", false},
		{"planning-edit-ignored", ".planning/x.md", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			dir := tempRepo(t)
			before := phase32CodeDigest(t, dir)
			if sh := phase32ShDigest(t, dir); sh != before {
				t.Fatalf("before edit: sh %s != go %s", sh, before)
			}
			path := filepath.Join(dir, filepath.FromSlash(row.file))
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, append(content, '!'), 0o644); err != nil {
				t.Fatal(err)
			}
			phase32Git(t, dir, "add", "-A")
			phase32Git(t, dir, "commit", "-qm", "edit")
			after := phase32CodeDigest(t, dir)
			if sh := phase32ShDigest(t, dir); sh != after {
				t.Fatalf("after edit: sh %s != go %s", sh, after)
			}
			if (before != after) != row.moves {
				t.Fatalf("digest moved=%v, want %v", before != after, row.moves)
			}
		})
	}
}

// phase32ReceiptPrograms is the receipt program set (D-09, D-09a). The
// adler32_fixture (APP-10 native vectors) and byte_index_oob (BYTE-02 native
// OOB) programs are receipt programs. 32-08's verify-phase32.sh emits the same
// set.
var phase32ReceiptPrograms = []string{"checksum", "adler_low", "adler32_cli", "adler32_fixture", "byte_index_oob", "sum_to_n", "fizzbuzz"}

// phase32OptimizationMatrixNonClaim is a recorded non-claim (D-09a): receipt
// case rows are default-build (-O0) rows only. The optimization matrix is
// bound by hosted go test lanes, never by receipt rows.
const phase32OptimizationMatrixNonClaim = "Receipt case rows are taken at the default build (-O0). The -O0, -O3 and -O3+LTO matrix is bound by the hosted go test lanes in the same run (TestPhase29AdlerBothEngines, TestPhase29OOBApplicationLanes), not by receipt rows."

var phase32ReceiptHosts = []string{"ubuntu-latest", "macos-latest"}

type phase32ReceiptHost struct {
	Name          string `json:"name"`
	RunnerOS      string `json:"runnerOs"`
	Arch          string `json:"arch"`
	Go            string `json:"go"`
	Clang         string `json:"clang"`
	LogPath       string `json:"logPath"`
	LogSha256     string `json:"logSha256"`
	JobConclusion string `json:"jobConclusion"`
}

type phase32ReceiptProgram struct {
	Name  string            `json:"name"`
	Hosts map[string]string `json:"hosts"`
}

type phase32Receipt struct {
	Schema      string                  `json:"schema"`
	State       string                  `json:"state"`
	RunID       string                  `json:"runId"`
	HeadSha     string                  `json:"headSha"`
	Codedigest  string                  `json:"codedigest"`
	DigestPaths []string                `json:"digestPaths"`
	Hosts       []phase32ReceiptHost    `json:"hosts"`
	Programs    []phase32ReceiptProgram `json:"programs"`
	NonClaims   []string                `json:"nonClaims"`
}

var (
	phase32BoundMarker    = regexp.MustCompile(`^M006 final hosted receipts: bound to codedigest ([0-9a-f]{64}) \(run ([0-9]+)\)\.?$`)
	phase32ReopenedMarker = regexp.MustCompile(`^M006 final hosted receipts: reopened ([0-9]{4}-[0-9]{2}-[0-9]{2}) \(historical codedigest ([0-9a-f]{64})\)\.?$`)
	phase32Hex64          = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

const phase32MarkerPrefix = "M006 final hosted receipts:"

type phase32ReceiptInputs struct {
	Docs           map[string]string // living doc name -> text
	Receipt        []byte            // nil when no 32-RECEIPT.json exists
	Logs           map[string][]byte // host name -> committed log bytes
	ComputedDigest string
}

// phase32LogRow returns the key=value fields of the first log line that starts
// with prefix (values never contain spaces on these rows), and the row count.
func phase32LogRows(log []byte, prefix string) []map[string]string {
	var rows []map[string]string
	for _, line := range strings.Split(string(log), "\n") {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		fields := map[string]string{}
		for _, token := range strings.Fields(strings.TrimPrefix(line, prefix)) {
			if key, value, ok := strings.Cut(token, "="); ok {
				fields[key] = value
			}
		}
		rows = append(rows, fields)
	}
	return rows
}

// phase32ReceiptProblems is pure: it returns sorted, de-duplicated named
// reasons. No receipt and no marker is the freeze state and has no problems
// (D-08c); a digest that merely drifted never fails on its own (D-10a).
func phase32ReceiptProblems(in phase32ReceiptInputs) []string {
	problems := map[string]bool{}
	add := func(format string, args ...any) { problems[fmt.Sprintf(format, args...)] = true }

	var bound, reopened int
	var boundDigests []string
	docNames := make([]string, 0, len(in.Docs))
	for name := range in.Docs {
		docNames = append(docNames, name)
	}
	slices.Sort(docNames)
	for _, name := range docNames {
		for _, line := range strings.Split(in.Docs[name], "\n") {
			if !strings.HasPrefix(line, phase32MarkerPrefix) {
				continue
			}
			switch {
			case phase32BoundMarker.MatchString(line):
				bound++
				boundDigests = append(boundDigests, phase32BoundMarker.FindStringSubmatch(line)[1])
			case phase32ReopenedMarker.MatchString(line):
				reopened++
			default:
				add("malformed_marker")
			}
		}
	}

	if in.Receipt == nil {
		if bound+reopened > 0 {
			add("claim_without_receipt")
		}
		return phase32SortedProblems(problems)
	}

	var receipt phase32Receipt
	decoder := json.NewDecoder(bytes.NewReader(in.Receipt))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil || receipt.Schema != "schway.m006-receipt/1" || !phase32Hex64.MatchString(receipt.Codedigest) {
		add("schema_invalid")
		return phase32SortedProblems(problems)
	}

	for _, digest := range boundDigests {
		if digest != receipt.Codedigest || digest != in.ComputedDigest {
			add("stale_bound_claim")
		}
	}

	if !slices.Contains(receipt.NonClaims, phase32OptimizationMatrixNonClaim) {
		add("nonclaim_missing_optimization_matrix")
	}

	byName := map[string]phase32ReceiptHost{}
	for _, host := range receipt.Hosts {
		byName[host.Name] = host
	}
	for _, name := range phase32ReceiptHosts {
		host, ok := byName[name]
		if !ok {
			add("host_missing")
			continue
		}
		if host.JobConclusion != "success" {
			add("job_not_success_%s", name)
		}
		log, ok := in.Logs[name]
		if !ok {
			add("log_missing_%s", name)
			continue
		}
		sum := sha256.Sum256(log)
		if hex.EncodeToString(sum[:]) != host.LogSha256 {
			add("log_sha256_mismatch_%s", name)
		}
		identities := phase32LogRows(log, "phase32 host identity ")
		if len(identities) != 1 {
			add("identity_row_count_%s", name)
		} else {
			if identities[0]["codedigest"] != receipt.Codedigest {
				add("log_digest_differs_receipt_%s", name)
			}
			if identities[0]["tree"] != "clean" {
				add("tree_not_clean_%s", name)
			}
		}
		results := phase32LogRows(log, "phase32 host result ")
		if len(results) != 1 || results[0]["status"] != "pass" {
			add("host_result_not_passing_%s", name)
		}
		covered := map[string]bool{}
		for _, row := range phase32LogRows(log, "phase32 case ") {
			if row["status"] == "pass" {
				covered[row["program"]] = true
			}
		}
		for _, program := range phase32ReceiptPrograms {
			if !covered[program] {
				add("program_missing_%s", program)
			}
		}
	}
	for _, host := range receipt.Hosts {
		if !slices.Contains(phase32ReceiptHosts, host.Name) {
			add("unexpected_host_%s", host.Name)
		}
	}

	declared := map[string]phase32ReceiptProgram{}
	for _, program := range receipt.Programs {
		declared[program.Name] = program
	}
	for _, program := range phase32ReceiptPrograms {
		entry, ok := declared[program]
		if !ok {
			add("program_missing_%s", program)
			continue
		}
		for _, host := range phase32ReceiptHosts {
			if entry.Hosts[host] != "pass" {
				add("program_not_passing_%s_%s", program, host)
			}
		}
	}
	for name := range declared {
		if !slices.Contains(phase32ReceiptPrograms, name) {
			add("unexpected_program_%s", name)
		}
	}
	return phase32SortedProblems(problems)
}

func phase32SortedProblems(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for problem := range set {
		out = append(out, problem)
	}
	slices.Sort(out)
	return out
}

// phase32GoodReceipt builds a known-good receipt and two literal logs
// independent of any producer (the logs are written here, not echoed).
func phase32GoodReceipt() (phase32Receipt, map[string][]byte) {
	digest := strings.Repeat("ab", 32)
	receipt := phase32Receipt{
		Schema: "schway.m006-receipt/1", State: "current", RunID: "123456789",
		HeadSha: strings.Repeat("c", 40), Codedigest: digest,
		DigestPaths: phase32DigestPaths,
		NonClaims:   []string{phase32OptimizationMatrixNonClaim},
	}
	logs := map[string][]byte{}
	for _, name := range phase32ReceiptHosts {
		var lines []string
		lines = append(lines, "phase32 host identity host="+name+" arch=x86_64 go=linux/amd64 clang_target=x86_64-pc-linux-gnu headsha="+receipt.HeadSha+" codedigest="+digest+" tree=clean")
		lines = append(lines, `phase32 toolchain go="go version go1.24.0" clang="clang version"`)
		for _, program := range phase32ReceiptPrograms {
			lines = append(lines, "phase32 case program="+program+" input=1 expected_exit=0 actual_exit=0 stdout=exact stderr=exact status=pass")
		}
		lines = append(lines, fmt.Sprintf("phase32 host result host=%s codedigest=%s tree=clean cases=%d controls=0 status=pass", name, digest, len(phase32ReceiptPrograms)))
		logs[name] = []byte(strings.Join(lines, "\n") + "\n")
		receipt.Hosts = append(receipt.Hosts, phase32ReceiptHost{Name: name, LogPath: "host-" + name + ".log", JobConclusion: "success"})
	}
	for _, program := range phase32ReceiptPrograms {
		hosts := map[string]string{}
		for _, name := range phase32ReceiptHosts {
			hosts[name] = "pass"
		}
		receipt.Programs = append(receipt.Programs, phase32ReceiptProgram{Name: program, Hosts: hosts})
	}
	phase32Rehash(&receipt, logs)
	return receipt, logs
}

func phase32Rehash(receipt *phase32Receipt, logs map[string][]byte) {
	for i := range receipt.Hosts {
		sum := sha256.Sum256(logs[receipt.Hosts[i].Name])
		receipt.Hosts[i].LogSha256 = hex.EncodeToString(sum[:])
	}
}

func phase32Marshal(t *testing.T, receipt phase32Receipt) []byte {
	t.Helper()
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestPhase32ReceiptLivingClaims(t *testing.T) {
	root := testsupport.ProjectPath()
	t.Run("real-checkout", func(t *testing.T) {
		docs := map[string]string{}
		for _, name := range []string{"LANGUAGE-MATURITY.md", "PRODUCT-ROADMAP.md", "STATE.md"} {
			data, err := os.ReadFile(filepath.Join(root, ".planning", name))
			if err != nil {
				t.Fatal(err)
			}
			docs[name] = string(data)
		}
		matches, err := phaseArtifactGlob("32-*", "32-RECEIPT.json")
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) > 1 {
			t.Fatalf("more than one 32-RECEIPT.json: %v", matches)
		}
		in := phase32ReceiptInputs{Docs: docs, ComputedDigest: phase32CodeDigest(t, root)}
		if len(matches) == 1 {
			receiptBytes, err := os.ReadFile(matches[0])
			if err != nil {
				t.Fatal(err)
			}
			in.Receipt = receiptBytes
			var receipt phase32Receipt
			if err := json.Unmarshal(receiptBytes, &receipt); err == nil {
				in.Logs = map[string][]byte{}
				for _, host := range receipt.Hosts {
					if filepath.IsAbs(host.LogPath) || strings.Contains(host.LogPath, "..") {
						t.Fatalf("log path escapes the receipt directory: %s", host.LogPath)
					}
					log, err := os.ReadFile(filepath.Join(filepath.Dir(matches[0]), filepath.FromSlash(host.LogPath)))
					if err != nil {
						t.Fatal(err)
					}
					in.Logs[host.Name] = log
				}
			}
		}
		if problems := phase32ReceiptProblems(in); len(problems) != 0 {
			t.Fatalf("receipt problems: %v", problems)
		}
	})

	boundLine := func(digest string) string {
		return "M006 final hosted receipts: bound to codedigest " + digest + " (run 123456789)"
	}
	reopenedLine := "M006 final hosted receipts: reopened 2026-10-12 (historical codedigest " + strings.Repeat("ab", 32) + ")"
	good := strings.Repeat("ab", 32)
	other := strings.Repeat("cd", 32)

	rows := []struct {
		name   string
		build  func(t *testing.T, receipt *phase32Receipt, logs map[string][]byte, in *phase32ReceiptInputs)
		expect []string
	}{
		{"freeze-state-no-receipt-no-marker", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
		}, nil},
		{"bound-marker-without-receipt", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.Docs["STATE.md"] = boundLine(good)
		}, []string{"claim_without_receipt"}},
		{"reopened-marker-without-receipt", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.Docs["STATE.md"] = reopenedLine
		}, []string{"claim_without_receipt"}},
		{"malformed-marker", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.Docs["STATE.md"] = "M006 final hosted receipts: bound to commit abc"
		}, []string{"malformed_marker"}},
		{"receipt-without-marker", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {}, nil},
		{"receipt-with-matching-bound-marker", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.Docs["STATE.md"] = "text\n" + boundLine(good) + "\n"
		}, nil},
		{"bound-marker-stale-against-computed", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.Docs["STATE.md"] = boundLine(good)
			in.ComputedDigest = other
		}, []string{"stale_bound_claim"}},
		{"bound-marker-differs-from-receipt", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.Docs["STATE.md"] = boundLine(other)
			in.ComputedDigest = other
		}, []string{"stale_bound_claim"}},
		{"reopened-marker-passes-when-digest-drifted", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.Docs["STATE.md"] = reopenedLine
			in.ComputedDigest = other
		}, nil},
		{"digest-drift-alone-passes", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.ComputedDigest = other
		}, nil},
		{"log-sha256-mismatch", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			l["macos-latest"] = append(l["macos-latest"], '\n')
		}, []string{"log_sha256_mismatch_macos-latest"}},
		{"log-digest-differs-from-receipt", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			l["ubuntu-latest"] = bytes.Replace(l["ubuntu-latest"], []byte("host identity host=ubuntu-latest arch=x86_64 go=linux/amd64 clang_target=x86_64-pc-linux-gnu headsha="+r.HeadSha+" codedigest="+good), []byte("host identity host=ubuntu-latest arch=x86_64 go=linux/amd64 clang_target=x86_64-pc-linux-gnu headsha="+r.HeadSha+" codedigest="+other), 1)
			phase32Rehash(r, l)
		}, []string{"log_digest_differs_receipt_ubuntu-latest"}},
		{"tree-modified", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			l["macos-latest"] = bytes.Replace(l["macos-latest"], []byte("tree=clean"), []byte("tree=modified"), 1)
			phase32Rehash(r, l)
		}, []string{"tree_not_clean_macos-latest"}},
		{"host-missing", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			r.Hosts = r.Hosts[:1]
		}, []string{"host_missing"}},
		{"host-result-not-passing", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			l["ubuntu-latest"] = bytes.Replace(l["ubuntu-latest"], []byte("controls=0 status=pass"), []byte("controls=0 status=fail"), 1)
			phase32Rehash(r, l)
		}, []string{"host_result_not_passing_ubuntu-latest"}},
		{"program-missing-fizzbuzz-in-receipt", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			r.Programs = slices.DeleteFunc(r.Programs, func(p phase32ReceiptProgram) bool { return p.Name == "fizzbuzz" })
		}, []string{"program_missing_fizzbuzz"}},
		{"program-missing-byte-index-oob-in-receipt", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			r.Programs = slices.DeleteFunc(r.Programs, func(p phase32ReceiptProgram) bool { return p.Name == "byte_index_oob" })
		}, []string{"program_missing_byte_index_oob"}},
		{"program-case-rows-missing-in-log", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			var kept []string
			for _, line := range strings.Split(string(l["macos-latest"]), "\n") {
				if !strings.HasPrefix(line, "phase32 case program=adler32_fixture ") {
					kept = append(kept, line)
				}
			}
			l["macos-latest"] = []byte(strings.Join(kept, "\n"))
			phase32Rehash(r, l)
		}, []string{"program_missing_adler32_fixture"}},
		{"nonclaim-missing-optimization-matrix", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			r.NonClaims = nil
		}, []string{"nonclaim_missing_optimization_matrix"}},
		{"schema-wrong", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			r.Schema = "schway.m006-receipt/2"
		}, []string{"schema_invalid"}},
		{"schema-unknown-field", func(t *testing.T, r *phase32Receipt, l map[string][]byte, in *phase32ReceiptInputs) {
			in.Receipt = bytes.Replace(phase32Marshal(t, *r), []byte("{"), []byte(`{"bogus":1,`), 1)
		}, []string{"schema_invalid"}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			receipt, logs := phase32GoodReceipt()
			in := phase32ReceiptInputs{Docs: map[string]string{"STATE.md": "no claim here\n"}, ComputedDigest: good}
			row.build(t, &receipt, logs, &in)
			// Rows named *no-receipt* / *without-receipt* model an absent 32-RECEIPT.json.
			noReceipt := strings.Contains(row.name, "no-receipt") || strings.HasSuffix(row.name, "without-receipt")
			if !noReceipt && in.Receipt == nil {
				in.Receipt = phase32Marshal(t, receipt)
			}
			in.Logs = logs
			got := phase32ReceiptProblems(in)
			if !slices.Equal(got, row.expect) {
				t.Fatalf("problems = %v, want %v", got, row.expect)
			}
		})
	}
}
