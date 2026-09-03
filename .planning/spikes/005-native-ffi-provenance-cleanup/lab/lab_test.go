package lab

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	once      sync.Once
	cached    Report
	cachedErr error
)

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source")
	}
	return filepath.Dir(filepath.Dir(file))
}

func runReport(t *testing.T) Report {
	t.Helper()
	once.Do(func() { cached, cachedErr = Run(projectRoot(t)) })
	if cachedErr != nil {
		t.Fatal(cachedErr)
	}
	return cached
}

func TestNativeAndOracleEventsAgreeAtO0AndO3(t *testing.T) {
	report := runReport(t)
	if len(report.Builds) != 2 {
		t.Fatalf("builds=%d, want 2", len(report.Builds))
	}
	for _, build := range report.Builds {
		if build.Events != len(expected) {
			t.Fatalf("%s events=%d want=%d", build.Optimization, build.Events, len(expected))
		}
		if build.BinaryBytes == 0 {
			t.Fatalf("%s binary size missing", build.Optimization)
		}
	}
}

func TestAllHostileMutationsAreObserved(t *testing.T) {
	report := runReport(t)
	if report.DetectedMutations != 7 || report.TotalMutations != 7 {
		t.Fatalf("mutation score=%d/%d: %+v", report.DetectedMutations, report.TotalMutations, report.Mutations)
	}
	for _, mutation := range report.Mutations {
		if !mutation.Detected {
			t.Errorf("mutation escaped: %+v", mutation)
		}
	}
}

func TestOptimizerProbeDemonstratesContractSensitivity(t *testing.T) {
	report := runReport(t)
	for _, mutation := range report.Mutations {
		if mutation.ID == "unsound-noalias" {
			if mutation.Evidence != "safe O0/O3=2/2; restrict O0/O3=2/1" {
				t.Fatalf("unexpected optimizer evidence: %s", mutation.Evidence)
			}
			return
		}
	}
	t.Fatal("unsound-noalias mutation missing")
}

func TestSourceDigestIsStableAndBounded(t *testing.T) {
	first, err := sourceDigest(projectRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	second, err := sourceDigest(projectRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if first != second || !strings.HasPrefix(first, "sha256:") || len(first) != len("sha256:")+64 {
		t.Fatalf("unstable digest: %q %q", first, second)
	}
}
