package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestReceiptGate(t *testing.T) {
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
	manifestBytes, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	producer, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase32.sh"))
	if err != nil {
		t.Fatal(err)
	}
	rows := func(s string) string {
		var out []string
		begin := strings.Index(s, "# phase32-cases-begin")
		end := strings.Index(s, "# phase32-cases-end")
		if begin >= 0 || end >= 0 {
			if begin < 0 || end <= begin {
				t.Fatal("producer case table markers are malformed")
			}
			s = s[begin:end]
		}
		for _, line := range strings.Split(s, "\n") {
			if strings.HasPrefix(line, "# case ") {
				out = append(out, line)
			}
		}
		return strings.Join(out, "\n")
	}
	manifestLines := string(manifestBytes)
	if rows(manifestLines) != rows(string(producer)) {
		t.Fatal("manifest case rows differ from producer")
	}
	programLine := func(s string) string {
		for _, line := range strings.Split(s, "\n") {
			if strings.HasPrefix(line, "phase32_programs=") {
				return strings.TrimPrefix(line, "phase32_programs=")
			}
		}
		return ""
	}
	manifestProgram := ""
	for _, line := range strings.Split(manifestLines, "\n") {
		if strings.HasPrefix(line, "programs=") {
			manifestProgram = strings.TrimPrefix(line, "programs=")
			break
		}
	}
	if manifestProgram != programLine(string(producer)) {
		t.Fatal("manifest programs differ from producer")
	}
	t.Run("manifest-equals-producer-table", func(t *testing.T) {
		if rows(manifestLines) != rows(string(producer)) {
			t.Fatal("manifest case rows differ from producer")
		}
		if manifestProgram != programLine(string(producer)) {
			t.Fatal("manifest programs differ from producer")
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

func TestReceiptGateDifferential(t *testing.T) {
	root := testsupport.ProjectPath()
	oldScript := testsupport.ProjectPath("scripts", "check-phase32-receipts.sh")
	newScript := "scripts/check-receipts.sh"
	manifest := testsupport.ProjectPath("testdata", "receipts", "M006.manifest")
	mac := phase32ReceiptLog(t, "macOS", phase32GateDigest)
	linux := phase32ReceiptLog(t, "Linux", phase32GateDigest)
	type row struct {
		name, want string
		a, b       string
		args       int
	}
	rows := []row{
		{"usage", "usage_two_logs_required", mac, "", 1},
		{"host_missing", "host_missing_", mac, "\x00missing", 2},
		{"log_empty", "log_empty_", mac, "", 2},
		{"failed_or_unavailable_row", "failed_or_unavailable_row_", mac, strings.Replace(linux, "status=pass", "status=fail", 1), 2},
		{"identity_row_count", "identity_row_count_0_", mac, dropReceiptRows(t, linux, "phase32 host identity "), 2},
		{"result_row_count", "result_row_count_2_", mac, linux + "phase32 host result host=Linux\n", 2},
		{"identity_not_first", "identity_not_first_", mac, "phase32 toolchain go=\"g\" clang=\"c\"\n" + linux, 2},
		{"result_not_last", "result_not_last_", mac, linux + "phase32 toolchain go=\"g\" clang=\"c\"\n", 2},
		{"unknown_phase32_row", "unknown_phase32_row_", mac, strings.Replace(linux, "phase32 host result ", "phase32 forged row\nphase32 host result ", 1), 2},
		{"toolchain_row_count", "toolchain_row_count_", mac, dropReceiptRows(t, linux, "phase32 toolchain "), 2},
		{"toolchain_row_malformed", "toolchain_row_malformed_", mac, strings.Replace(linux, "go=\"go version go1.24.0\"", "go=\"\"", 1), 2},
		{"unknown_host", "unknown_host_", mac, strings.Replace(strings.Replace(linux, "host=Linux", "host=Plan9", 1), "host=Linux", "host=Plan9", 1), 2},
		{"arch_missing", "arch_missing_", mac, strings.Replace(linux, " arch=x86_64", "", 1), 2},
		{"target_missing", "target_missing_", mac, strings.Replace(linux, " clang_target=x86_64-pc-linux-gnu", "", 1), 2},
		{"go_host_missing", "go_host_missing_", mac, strings.Replace(linux, " go=linux/amd64", "", 1), 2},
		{"headsha_missing", "headsha_missing_", mac, strings.Replace(linux, " headsha="+phase32GateHeadSha, "", 1), 2},
		{"digest_mismatch", "digest_mismatch_", strings.Replace(mac, "codedigest="+phase32GateDigest, "codedigest="+strings.Repeat("cd", 32), 1), strings.Replace(linux, "codedigest="+phase32GateDigest, "codedigest="+strings.Repeat("cd", 32), 1), 2},
		{"tree_not_clean", "tree_not_clean_", mac, strings.Replace(linux, "tree=clean", "tree=dirty", 1), 2},
		{"result_host_mismatch", "result_host_mismatch_", mac, strings.Replace(linux, "phase32 host result host=Linux", "phase32 host result host=macOS", 1), 2},
		{"result_digest_mismatch", "result_digest_mismatch_", mac, strings.Replace(linux, "phase32 host result host=Linux codedigest="+phase32GateDigest, "phase32 host result host=Linux codedigest="+strings.Repeat("cd", 32), 1), 2},
		{"result_tree_not_clean", "result_tree_not_clean_", mac, strings.Replace(linux, "phase32 host result host=Linux codedigest="+phase32GateDigest+" tree=clean", "phase32 host result host=Linux codedigest="+phase32GateDigest+" tree=dirty", 1), 2},
		{"result_not_pass", "result_not_pass_", mac, strings.Replace(linux, "programs="+strings.Join(phase32ReceiptPrograms, ",")+" status=pass", "programs="+strings.Join(phase32ReceiptPrograms, ",")+" status=skipped", 1), 2},
		{"result_programs_mismatch", "result_programs_mismatch_", mac, strings.Replace(linux, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs=other", 1), 2},
		{"program_missing", "program_missing_adler_low_", mac, dropReceiptRows(t, linux, "phase32 case program=adler_low "), 2},
		{"result_control_count", "result_control_count_", mac, strings.Replace(linux, "controls=4", "controls=3", 1), 2},
		{"result_case_count", "result_case_count_", mac, strings.Replace(linux, "cases=39", "cases=38", 1), 2},
		{"case_row_count", "case_row_count_", mac, strings.Replace(linux, "phase32 host result ", strings.Split(linux, "\n")[2]+"\nphase32 host result ", 1), 2},
		{"case_missing_or_altered", "case_missing_or_altered_checksum_empty_", mac, strings.Replace(linux, "stdout=exact", "stdout=wrong", 1), 2},
		{"control_row_count", "control_row_count_", mac, dropReceiptRows(t, linux, "phase32 control name=codedigest "), 2},
		{"control_missing_or_altered", "control_missing_or_altered_guide_", mac, strings.Replace(linux, "name=guide package=./cmd/schway", "name=guide package=", 1), 2},
		{"log_digest_differs", "log_digest_differs", mac, phase32ReceiptLog(t, "Linux", strings.Repeat("cd", 32)), 2},
		{"duplicate_host", "duplicate_host_macOS", mac, mac, 2},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			paths := []string{filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")}
			if err := os.WriteFile(paths[0], []byte(tc.a), 0o600); err != nil {
				t.Fatal(err)
			}
			if tc.b != "\x00missing" {
				if err := os.WriteFile(paths[1], []byte(tc.b), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			oldEnv := envWith("PHASE32_EXPECTED_CODEDIGEST", phase32GateDigest)
			newEnv := envWith("EXPECTED_CODEDIGEST", phase32GateDigest)
			oldArgs := paths[:tc.args]
			newArgs := append([]string{manifest}, paths[:tc.args]...)
			old := phase33RunAtEnv(t, root, oldScript, oldEnv, oldArgs...)
			got := phase33RunAtEnv(t, root, newScript, newEnv, newArgs...)
			if old != got {
				t.Fatalf("old/new differ\nold=%#v\nnew=%#v", old, got)
			}
			if !strings.Contains(got.stderr, "reason="+tc.want) {
				t.Fatalf("wanted reason=%s, got %#v", tc.want, got)
			}
		})
	}
	for _, name := range []string{"case_table_empty", "programs_line_missing", "program_without_table_entry"} {
		t.Run(name, func(t *testing.T) {
			// Both implementations read their own table source; make equivalent variants.
			temp := t.TempDir()
			producer := mustRead(t, testsupport.ProjectPath("scripts", "verify-phase32.sh"))
			old := mustRead(t, oldScript)
			digest := mustRead(t, testsupport.ProjectPath("scripts", "codedigest.sh"))
			m := mustRead(t, manifest)
			p, mText := string(producer), string(m)
			switch name {
			case "case_table_empty":
				p = stripCaseRows(p)
				mText = stripCaseRows(mText)
			case "programs_line_missing":
				p = strings.Replace(p, "phase32_programs="+strings.Join(phase32ReceiptPrograms, ","), "", 1)
				mText = strings.Replace(mText, "programs="+strings.Join(phase32ReceiptPrograms, ","), "", 1)
			case "program_without_table_entry":
				p = strings.Replace(p, "phase32_programs="+strings.Join(phase32ReceiptPrograms, ","), "phase32_programs="+strings.Join(phase32ReceiptPrograms, ",")+",bogus", 1)
				mText = strings.Replace(mText, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs="+strings.Join(phase32ReceiptPrograms, ",")+",bogus", 1)
				mac = strings.Replace(mac, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs="+strings.Join(phase32ReceiptPrograms, ",")+",bogus", 1)
				linux = strings.Replace(linux, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs="+strings.Join(phase32ReceiptPrograms, ",")+",bogus", 1)
			}
			for path, data := range map[string][]byte{"scripts/verify-phase32.sh": []byte(p), "scripts/check-phase32-receipts.sh": old, "scripts/codedigest.sh": digest, "scripts/check-receipts.sh": mustRead(t, testsupport.ProjectPath("scripts", "check-receipts.sh")), "testdata/receipts/variant.manifest": []byte(mText)} {
				dst := filepath.Join(temp, filepath.FromSlash(path))
				if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(dst, data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			phase32Git(t, temp, "init", "-q")
			phase32Git(t, temp, "add", "-A")
			phase32Git(t, temp, "commit", "-qm", "fixture")
			logsDir := t.TempDir()
			a, b := filepath.Join(logsDir, "a.log"), filepath.Join(logsDir, "b.log")
			if err := os.WriteFile(a, []byte(mac), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(b, []byte(linux), 0o600); err != nil {
				t.Fatal(err)
			}
			oldResult := phase33RunAtEnv(t, temp, filepath.Join(temp, "scripts/check-phase32-receipts.sh"), envWith("PHASE32_EXPECTED_CODEDIGEST", phase32GateDigest), a, b)
			newResult := phase33RunAtEnv(t, root, newScript, envWith("EXPECTED_CODEDIGEST", phase32GateDigest), filepath.Join(temp, "testdata/receipts/variant.manifest"), a, b)
			if oldResult != newResult {
				t.Fatalf("old/new differ: %#v %#v", oldResult, newResult)
			}
			want := name
			if name == "program_without_table_entry" {
				want = "program_without_table_entry_bogus"
			}
			if !strings.Contains(oldResult.stderr, "reason="+want) {
				t.Fatalf("wanted %s, got %#v", want, oldResult)
			}
		})
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
