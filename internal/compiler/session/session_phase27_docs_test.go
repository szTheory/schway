package session_test

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

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
}
