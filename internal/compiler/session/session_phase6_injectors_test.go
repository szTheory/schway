package session

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func phase6Fixture(t *testing.T, name string) []byte {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase6", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	return source
}

// hasDriverEligibleRepair reports whether any diagnostic in diagnostics
// carries a diagnostic.DriverEligible repair.
func hasDriverEligibleRepair(diagnostics []diagnostic.Diagnostic) bool {
	for _, problem := range diagnostics {
		for _, repair := range problem.Repairs {
			if diagnostic.DriverEligible(repair) {
				return true
			}
		}
	}
	return false
}

// injectorError, when non-nil, unwraps err to *InjectorError or returns nil.
func injectorError(err error) *InjectorError {
	var typed *InjectorError
	if errors.As(err, &typed) {
		return typed
	}
	return nil
}

// TestMatchDefectInjectorProducesExactlyOneDefect is Task 1's tracer
// end-to-end assertion: a real held-out fixture, mutated by a real
// injector, checked by the real checker, produces a real repair-bearing
// diagnostic.
func TestMatchDefectInjectorProducesExactlyOneDefect(t *testing.T) {
	source := phase6Fixture(t, "heldout_match_defect.lang")

	clean := Check(source)
	if len(clean.Diagnostics) != 0 {
		t.Fatalf("heldout_match_defect.lang must check clean unmutated: %+v", clean.Diagnostics)
	}

	mutated, err := MatchInjector{}.Inject(source)
	if err != nil {
		t.Fatalf("MatchInjector.Inject failed on an eligible fixture: %v", err)
	}
	if bytes.Equal(mutated, source) {
		t.Fatal("MatchInjector.Inject returned the source unmutated")
	}

	checked := Check(mutated)
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "match.non_exhaustive" {
		t.Fatalf("mutated fixture did not reject with match.non_exhaustive: %+v", checked.Diagnostics)
	}
	if !hasDriverEligibleRepair(checked.Diagnostics) {
		t.Fatalf("match.non_exhaustive diagnostic carries no DriverEligible repair: %+v", checked.Diagnostics)
	}

	noTarget := bytes.ReplaceAll(source, []byte(matchTargetMarker), []byte(""))
	_, err = MatchInjector{}.Inject(noTarget)
	typed := injectorError(err)
	if typed == nil || typed.Code != InjectorTargetMissingCode {
		t.Fatalf("MatchInjector.Inject on a fixture with no eligible arm did not refuse with %s: %v", InjectorTargetMissingCode, err)
	}
}

// TestPhase6DefectCorpusIsHeldOut asserts D-06-29's structural split: the
// heldout_ and derivation_ prefix sets are both non-empty and genuinely
// distinct (different file content per class), not merely differently
// named. The per-class distinctness check is driven from whichever
// heldout_*_defect.lang fixtures exist on disk, so it scales automatically
// as later tasks add the move and borrow classes alongside match.
func TestPhase6DefectCorpusIsHeldOut(t *testing.T) {
	dir := testsupport.ProjectPath("testdata", "phase6")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var heldout, derivation []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
			continue
		}
		switch {
		case strings.HasPrefix(entry.Name(), "heldout_"):
			heldout = append(heldout, entry.Name())
		case strings.HasPrefix(entry.Name(), "derivation_"):
			derivation = append(derivation, entry.Name())
		}
	}
	if len(heldout) == 0 {
		t.Fatal("no heldout_*.lang fixtures found in testdata/phase6")
	}
	if len(derivation) == 0 {
		t.Fatal("no derivation_*.lang fixtures found in testdata/phase6")
	}
	seen := make(map[string]bool, len(heldout))
	for _, name := range heldout {
		seen[name] = true
	}
	for _, name := range derivation {
		if seen[name] {
			t.Fatalf("%s appears in both the heldout_ and derivation_ sets", name)
		}
	}
	readmePath := filepath.Join(dir, "README")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("testdata/phase6/README must exist and name both prefixes: %v", err)
	}
	if !strings.Contains(string(readme), "heldout_") || !strings.Contains(string(readme), "derivation_") {
		t.Fatal("testdata/phase6/README does not name both the heldout_ and derivation_ prefixes")
	}

	// Genuine distinctness, not just naming: each class's heldout and
	// derivation fixture must differ in content.
	for _, name := range heldout {
		class := strings.TrimSuffix(strings.TrimPrefix(name, "heldout_"), "_defect.lang")
		derivationName := "derivation_" + class + "_defect.lang"
		heldoutBytes := phase6Fixture(t, name)
		derivationBytes := phase6Fixture(t, derivationName)
		if bytes.Equal(heldoutBytes, derivationBytes) {
			t.Fatalf("%s class: heldout and derivation fixtures are byte-identical", class)
		}
	}
}
