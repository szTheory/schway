package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase35PeerTransferMutationControls changes the arithmetic in temporary
// implementation copies, then runs each value-deriving peer's named behavioral
// sink under Go's source overlay. A passing no-op overlay for every peer proves
// the child commands and selectors are live before the 15 wrong-result runs.
func TestPhase35PeerTransferMutationControls(t *testing.T) {
	const childSelector = "^TestPhase35ScalarTransferMutants$"
	peers := []phase35TransferPeer{
		{
			name: "check", packagePath: "./internal/compiler/check",
			sourcePath: "internal/compiler/check/scalar.go", transferSiteCount: 1,
			failureText: "did not derive expected",
			tests: []phase35TransferMutationCase{
				{name: "xor", childName: "xor_high-bit_pattern", before: "source.u64 ^ right.u64", after: "source.u64 & right.u64"},
				{name: "and", childName: "and_high-bits_to_all-zero", before: "source.u64 & right.u64", after: "source.u64 | right.u64"},
				{name: "or", childName: "or_overlapping_high-bit", before: "source.u64 | right.u64", after: "source.u64 ^ right.u64"},
				{name: "shift-left", childName: "left_shift_all-zero", before: "source.u64 << uint(amount)", after: "source.u64 >> uint(amount)"},
				{name: "shift-right", childName: "right_shift_high-bit", before: "source.u64 >> uint(amount)", after: "source.u64 << uint(amount)"},
			},
		},
		{
			name: "corevalidate", packagePath: "./internal/compiler/corevalidate",
			sourcePath: "internal/compiler/corevalidate/corevalidate.go", transferSiteCount: 2,
			failureText: "transfer failed to reach the static zero-remainder sink",
			tests: []phase35TransferMutationCase{
				{name: "xor", childName: "xor_zero", before: "source.u64 ^ right.u64", after: "source.u64 & right.u64"},
				{name: "and", childName: "and_zero", before: "source.u64 & right.u64", after: "source.u64 | right.u64"},
				{name: "or", childName: "or_overlapping_bit", before: "source.u64 | right.u64", after: "source.u64 ^ right.u64"},
				{name: "shift-left", childName: "shift_left_zero", before: "source.u64 << amount", after: "source.u64 >> amount"},
				{name: "shift-right", childName: "shift_right_one", before: "source.u64 >> amount", after: "source.u64 << amount"},
			},
		},
		{
			name: "originvalidate", packagePath: "./internal/compiler/originvalidate",
			sourcePath: "internal/compiler/originvalidate/originvalidate.go", transferSiteCount: 1,
			failureText: "transfer missed its static zero-remainder sink",
			tests: []phase35TransferMutationCase{
				{name: "xor", childName: "xor_zero", before: "source.u64 ^ right.u64", after: "source.u64 & right.u64"},
				{name: "and", childName: "and_zero", before: "source.u64 & right.u64", after: "source.u64 | right.u64"},
				{name: "or", childName: "or_overlapping_bit", before: "source.u64 | right.u64", after: "source.u64 ^ right.u64"},
				{name: "shift-left", childName: "shift_left_zero", before: "source.u64 << amount", after: "source.u64 >> amount"},
				{name: "shift-right", childName: "shift_right_one", before: "source.u64 >> amount", after: "source.u64 << amount"},
			},
		},
	}

	root := testsupport.ProjectPath()
	for _, peer := range peers {
		peer := peer
		t.Run(peer.name, func(t *testing.T) {
			implementationPath := testsupport.ProjectPath(peer.sourcePath)
			original, err := os.ReadFile(implementationPath)
			if err != nil {
				t.Fatalf("read %s implementation: %v", peer.name, err)
			}
			sourceHash := sha256.Sum256(original)
			hashText := hex.EncodeToString(sourceHash[:])

			t.Run("no-op-overlay", func(t *testing.T) {
				output, err := phase35RunTransferChild(t, root, peer, implementationPath, original, childSelector)
				if err != nil {
					t.Fatalf("no-op overlay should pass the named transfer suite: %v\n%s", err, output)
				}
				phase35RequireTransferChildRun(t, output, peer.tests, "PASS")
				t.Logf("no-op overlay PASS peer=%s source_sha256=%s selector=%s", peer.name, hashText, childSelector)
			})

			for _, mutation := range peer.tests {
				mutation := mutation
				t.Run(mutation.name, func(t *testing.T) {
					mutated, count := phase35ReplaceTransferLaw(string(original), mutation.before, mutation.after)
					if count != peer.transferSiteCount {
						t.Fatalf("%s %s mutation matched %d transfer sites, want %d; needle=%q", peer.name, mutation.name, count, peer.transferSiteCount, mutation.before)
					}
					if mutated == string(original) {
						t.Fatalf("%s %s mutation did not change its implementation copy", peer.name, mutation.name)
					}

					output, err := phase35RunTransferChild(t, root, peer, implementationPath, []byte(mutated), childSelector)
					if err == nil {
						t.Fatalf("%s %s wrong-result transfer mutant survived the named test\n%s", peer.name, mutation.name, output)
					}
					childName := "TestPhase35ScalarTransferMutants/" + mutation.childName
					if !strings.Contains(output, "=== RUN   "+childName) || !strings.Contains(output, "--- FAIL: "+childName) || !strings.Contains(output, peer.failureText) {
						t.Fatalf("%s %s child failed outside its reached behavioral assertion; wanted run/fail for %q and %q\n%s", peer.name, mutation.name, childName, peer.failureText, output)
					}
					t.Logf("mutant KILLED peer=%s operator=%s sites=%d source_sha256=%s named_failure=%s", peer.name, mutation.name, count, hashText, childName)
				})
			}
		})
	}
}

type phase35TransferPeer struct {
	name              string
	packagePath       string
	sourcePath        string
	transferSiteCount int
	failureText       string
	tests             []phase35TransferMutationCase
}

type phase35TransferMutationCase struct {
	name      string
	childName string
	before    string
	after     string
}

func phase35RunTransferChild(t *testing.T, root string, peer phase35TransferPeer, implementationPath string, source []byte, selector string) (string, error) {
	t.Helper()
	overlayDir := t.TempDir()
	overlaySource := filepath.Join(overlayDir, filepath.Base(implementationPath))
	if err := os.WriteFile(overlaySource, source, 0o600); err != nil {
		return "", fmt.Errorf("write overlay source: %w", err)
	}
	overlayPath := filepath.Join(overlayDir, "overlay.json")
	overlay := struct {
		Replace map[string]string `json:"Replace"`
	}{Replace: map[string]string{implementationPath: overlaySource}}
	encoded, err := json.Marshal(overlay)
	if err != nil {
		return "", fmt.Errorf("encode overlay: %w", err)
	}
	if err := os.WriteFile(overlayPath, encoded, 0o600); err != nil {
		return "", fmt.Errorf("write overlay: %w", err)
	}
	command := exec.Command("go", "test", "-overlay="+overlayPath, peer.packagePath,
		"-run="+selector, "-count=1", "-v", "-timeout=600s")
	command.Dir = root
	command.Env = phase35TransferGoEnv(os.Environ())
	output, runErr := command.CombinedOutput()
	return string(output), runErr
}

func phase35TransferGoEnv(environment []string) []string {
	filtered := make([]string, 0, len(environment)+2)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, "GOCACHE=") && !strings.HasPrefix(entry, "GOTOOLCHAIN=") {
			filtered = append(filtered, entry)
		}
	}
	return append(filtered, "GOCACHE=/tmp/ai-lang-verification-gocache", "GOTOOLCHAIN=local")
}

func phase35RequireTransferChildRun(t *testing.T, output string, cases []phase35TransferMutationCase, status string) {
	t.Helper()
	if !strings.Contains(output, "=== RUN   TestPhase35ScalarTransferMutants") || !strings.Contains(output, "--- "+status+": TestPhase35ScalarTransferMutants") {
		t.Fatalf("child output did not prove the named transfer test ran and ended %s:\n%s", status, output)
	}
	for _, mutation := range cases {
		name := "TestPhase35ScalarTransferMutants/" + mutation.childName
		if !strings.Contains(output, "=== RUN   "+name) || !strings.Contains(output, "--- "+status+": "+name) {
			t.Fatalf("child output did not prove operator subtest %q ran and ended %s:\n%s", name, status, output)
		}
	}
}

func phase35ReplaceTransferLaw(source, needle, replacement string) (string, int) {
	count := strings.Count(source, needle)
	return strings.ReplaceAll(source, needle, replacement), count
}
