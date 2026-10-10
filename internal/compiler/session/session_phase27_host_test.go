package session_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase27GateRevision = "0123456789abcdef0123456789abcdef01234567"

func phase27GitHash(t *testing.T, relative string) string {
	t.Helper()
	path := testsupport.ProjectPath(strings.Split(relative, "/")...)
	output, err := exec.Command("git", "hash-object", path).Output()
	if err != nil {
		t.Fatalf("git hash-object %s: %v", relative, err)
	}
	return relative + "=" + strings.TrimSpace(string(output))
}

// phase27ReceiptLog builds a literal known-good receipt for one host. The
// values are independent of scripts/verify-phase27.sh: only the file hashes are
// measured, so the gate cannot pass by echoing its own producer.
func phase27ReceiptLog(t *testing.T, host string) string {
	t.Helper()
	sources := phase27GitHash(t, "examples/fizzbuzz.schway") + "," + phase27GitHash(t, "examples/sum_to_n.schway")
	fixtures := phase27GitHash(t, "examples/phase27/fizzbuzz-15.out") + "," + phase27GitHash(t, "examples/phase27/fizzbuzz-1000.out")
	var lines []string
	lines = append(lines, "phase27 host identity host="+host+" arch=x86_64 go=linux/amd64 clang_target=x86_64-pc-linux-gnu revision="+phase27GateRevision+" tree=clean sources="+sources+" fixtures="+fixtures+" inputs=fizzbuzz:0,1,15,1000,1001;sum_to_n:0,1,15,1000,1001")
	lines = append(lines, `phase27 toolchain go="go version go1.24.0" clang="clang version"`)
	for _, program := range []string{"fizzbuzz", "sum_to_n"} {
		for _, input := range []string{"0", "1", "15", "1000", "1001"} {
			code := "0"
			if input == "1001" {
				code = "65"
			}
			lines = append(lines, "phase27 case program="+program+" input="+input+" expected_exit="+code+" actual_exit="+code+" stdout=exact stderr=exact status=pass")
		}
	}
	lines = append(lines,
		"phase27 control name=dynamic-zero package=./internal/compiler/session tests=TestPhase27DynamicZero status=pass",
		"phase27 control name=output-limit package=./internal/compiler/cgen tests=TestPhase27OutputExactLimit TestPhase27OutputCrossingWrite status=pass",
		"phase27 control name=output-limit-public package=./cmd/schway tests=TestPhase27OutputPublicStatus status=pass",
		"phase27 host result host="+host+" revision="+phase27GateRevision+" tree=clean cases=10 controls=3 status=pass",
	)
	return strings.Join(lines, "\n") + "\n"
}

func runPhase27Gate(t *testing.T, first, second string) (string, error) {
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
	command := exec.Command("sh", testsupport.ProjectPath("scripts", "check-phase27-receipts.sh"), paths[0], paths[1])
	command.Env = append(os.Environ(), "PHASE27_EXPECTED_REVISION="+phase27GateRevision)
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestPhase27HostReceiptGate(t *testing.T) {
	mac := phase27ReceiptLog(t, "macOS")
	linux := phase27ReceiptLog(t, "Linux")
	if output, err := runPhase27Gate(t, mac, linux); err != nil || !strings.Contains(output, "status=pass hosts=macOS,Linux") {
		t.Fatalf("known-good pair rejected: err=%v output=%s", err, output)
	}
	if output, err := runPhase27Gate(t, linux, mac); err != nil {
		t.Fatalf("host order must not matter: err=%v output=%s", err, output)
	}

	mutate := func(log, old, replacement string) string {
		if !strings.Contains(log, old) {
			t.Fatalf("mutation anchor %q not present", old)
		}
		return strings.Replace(log, old, replacement, 1)
	}
	lines := strings.Split(strings.TrimSuffix(linux, "\n"), "\n")
	without := func(prefix string) string {
		var kept []string
		dropped := false
		for _, line := range lines {
			if !dropped && strings.HasPrefix(line, prefix) {
				dropped = true
				continue
			}
			kept = append(kept, line)
		}
		if !dropped {
			t.Fatalf("no line with prefix %q", prefix)
		}
		return strings.Join(kept, "\n") + "\n"
	}

	for _, test := range []struct {
		name         string
		first, other string
		reason       string
	}{
		{"one-host-missing-second", mac, "\x00missing", "log_missing"},
		{"same-host-duplicate", mac, mac, "duplicate_host_macOS"},
		{"empty-log", mac, "", "log_empty"},
		{"unavailable-host", mac, mutate(linux, "status=pass\nphase27 host result", "status=unavailable\nphase27 host result"), ""},
		{"failed-result", mac, mutate(linux, "controls=3 status=pass", "controls=3 status=fail"), "failed_or_unavailable_row"},
		{"failed-case", mac, mutate(linux, "program=fizzbuzz input=15 expected_exit=0 actual_exit=0 stdout=exact stderr=exact status=pass", "program=fizzbuzz input=15 expected_exit=0 actual_exit=0 stdout=mismatch stderr=exact status=fail"), "failed_or_unavailable_row"},
		{"wrong-revision", mac, strings.ReplaceAll(linux, phase27GateRevision, strings.Repeat("a", 40)), "wrong_revision"},
		{"mixed-revision-result", mac, mutate(linux, "host result host=Linux revision="+phase27GateRevision, "host result host=Linux revision="+strings.Repeat("a", 40)), "result_revision_mismatch"},
		{"altered-source-hash", mac, mutate(linux, "examples/fizzbuzz.schway=", "examples/fizzbuzz.schway=0000"), "source_hash_mismatch"},
		{"altered-fixture-hash", mac, mutate(linux, "examples/phase27/fizzbuzz-1000.out=", "examples/phase27/fizzbuzz-1000.out=0000"), "fixture_hash_mismatch"},
		{"wrong-input-set", mac, mutate(linux, "inputs=fizzbuzz:0,1,15,1000,1001", "inputs=fizzbuzz:0,1,15,1000"), "input_set_mismatch"},
		{"dirty-tree", mac, mutate(linux, "tree=clean sources", "tree=modified sources"), "tree_not_clean"},
		{"missing-case", mac, without("phase27 case program=sum_to_n input=1000 "), "case_"},
		{"altered-case-exit", mac, mutate(linux, "input=1001 expected_exit=65 actual_exit=65", "input=1001 expected_exit=0 actual_exit=0"), "case_missing_or_altered"},
		{"missing-control", mac, without("phase27 control name=dynamic-zero"), "control_row_count"},
		{"missing-result", mac, without("phase27 host result"), "result_row_count"},
		{"missing-identity", mac, without("phase27 host identity"), "identity_row_count"},
		{"result-not-last", mac, linux + "phase27 toolchain trailing\n", "result_not_last"},
		{"unknown-host", mac, strings.ReplaceAll(linux, "host=Linux", "host=Windows"), "unknown_host"},
		{"malformed-identity", mac, mutate(linux, "arch=x86_64 ", ""), "arch_missing"},
		{"malformed-case-row", mac, mutate(linux, "phase27 case program=fizzbuzz input=0 ", "phase27 case program=fizzbuzz input=0 junk "), "case_missing_or_altered"},
	} {
		t.Run(test.name, func(t *testing.T) {
			output, err := runPhase27Gate(t, test.first, test.other)
			if err == nil || !strings.Contains(output, "status=fail") || !strings.Contains(output, test.reason) {
				t.Fatalf("mutated pair was not rejected with %q: err=%v output=%s", test.reason, err, output)
			}
			if strings.Contains(output, "status=pass") {
				t.Fatalf("rejected pair also reported pass: %s", output)
			}
		})
	}
	if output, err := runPhase27Gate(t, mac, ""); err == nil {
		t.Fatalf("one real host plus empty log passed: %s", output)
	}
}
