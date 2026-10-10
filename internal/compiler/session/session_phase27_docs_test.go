package session_test

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase27EvidenceRevision = "6224f8b2c9c15f54270a232fab2bce12628c7f19"

var phase27PinnedGateCommand = regexp.MustCompile(`PHASE27_EXPECTED_REVISION=([0-9a-f]{40}) sh scripts/check-phase27-receipts\.sh (\S+host-macos\.receipt) (\S+host-linux\.receipt)`)

func environmentWithout(name string) []string {
	var out []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, name+"=") {
			out = append(out, entry)
		}
	}
	return out
}

func TestPhase27GuideContract(t *testing.T) {
	root, err := os.ReadFile(testsupport.ProjectPath("README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(root), "(examples/phase27/README.md)") != 1 {
		t.Fatal("root README must link the Phase 27 guide exactly once")
	}
	guideBytes, err := os.ReadFile(testsupport.ProjectPath("examples", "phase27", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	guide := string(guideBytes)
	for _, required := range []string{
		"examples/fizzbuzz.schway", "examples/sum_to_n.schway",
		"examples/phase27/fizzbuzz-15.out", "examples/phase27/fizzbuzz-1000.out",
		"go build -o ./schway ./cmd/schway", "./schway build", "./schway app run",
		"500500", "55", "1001", "0` through `1000", "65536-byte",
		"exit status 65", "schway app: input exceeds 1000",
		"schway: application output limit exceeded", "prefix",
		"Unit", "auto-printed", "type.scalar_remainder_zero",
		"schway: U64 remainder by zero",
		"scripts/verify-phase27.sh", "scripts/check-phase27-receipts.sh",
		"unavailable", "27-HOST-EVIDENCE.md", "Historical Phase 25 and Phase 26",
		"PHASE27_EXPECTED_REVISION", "wrong_revision", phase27EvidenceRevision,
		".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/host-macos.receipt",
		".planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/host-linux.receipt",
		"source_hash_mismatch", "fixture_hash_mismatch", "consistency", "authenticity",
		"host-macos.receipt", "host-linux.receipt", "phase27-receipt-macos-latest-", "14 days",
	} {
		if !strings.Contains(guide, required) {
			t.Errorf("Phase 27 guide omits %q", required)
		}
	}
	for _, path := range []string{
		"examples/fizzbuzz.schway", "examples/sum_to_n.schway",
		"examples/phase27/fizzbuzz-15.out", "examples/phase27/fizzbuzz-1000.out",
		"scripts/verify-phase27.sh", "scripts/check-phase27-receipts.sh",
	} {
		if _, err := os.Stat(testsupport.ProjectPath(strings.Split(path, "/")...)); err != nil {
			t.Errorf("guide names missing path %s: %v", path, err)
		}
	}
	matches := phase27PinnedGateCommand.FindAllStringSubmatch(guide, -1)
	if len(matches) != 1 {
		t.Fatalf("guide must contain exactly one pinned gate command, found %d", len(matches))
	}
	group := matches[0]
	if group[1] != phase27EvidenceRevision {
		t.Fatalf("pinned revision %s, want %s", group[1], phase27EvidenceRevision)
	}
	macPath := testsupport.ProjectPath(strings.Split(group[2], "/")...)
	linuxPath := testsupport.ProjectPath(strings.Split(group[3], "/")...)
	script := testsupport.ProjectPath("scripts", "check-phase27-receipts.sh")
	ctx := context.Background()
	pinnedEnv := append(environmentWithout("PHASE27_EXPECTED_REVISION"), "PHASE27_EXPECTED_REVISION="+group[1])
	result, err := testsupport.RunCLIErr(ctx, "sh", pinnedEnv, script, macPath, linuxPath)
	if err != nil || result.Exit != 0 || !strings.Contains(string(result.Stdout), "status=pass hosts=macOS,Linux revision="+group[1]) {
		t.Fatalf("documented pinned command did not pass: err=%v exit=%d\nstdout=%s\nstderr=%s\n"+
			"a source_hash_mismatch or fixture_hash_mismatch means a declared Phase 27 input changed since 6224f8b "+
			"and must be bundled with a deliberate re-evidence run, not fixed by editing the guide",
			err, result.Exit, result.Stdout, result.Stderr)
	}
	head, err := testsupport.RunCLIErr(ctx, "git", os.Environ(), "-C", testsupport.ProjectPath(), "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(head.Stdout)) == phase27EvidenceRevision {
		t.Log("HEAD equals the pinned revision; skipping bare-command negative control")
		return
	}
	bare, err := testsupport.RunCLIErr(ctx, "sh", environmentWithout("PHASE27_EXPECTED_REVISION"), script, macPath, linuxPath)
	if err != nil {
		t.Fatal(err)
	}
	combined := string(bare.Stdout) + string(bare.Stderr)
	if bare.Exit == 0 || !strings.Contains(combined, "reason=wrong_revision") ||
		!strings.Contains(combined, "phase27 receipt gate: hint:") ||
		!strings.Contains(combined, "PHASE27_EXPECTED_REVISION") || strings.Contains(combined, "status=pass") {
		t.Fatalf("bare command must fail wrong_revision with hint at a later HEAD: exit=%d\n%s", bare.Exit, combined)
	}
}
