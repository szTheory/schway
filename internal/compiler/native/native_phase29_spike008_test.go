package native_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// spike008RecordedCap is the byte view cap the spike 008 budget comparison
// scales to (D-12 / D-13, research Decision 5). It is the single place a D-13
// lowering is written in code: after a lowering decision set it to the new cap
// and strict mode passes at it. See .planning/spikes/008-byteview-cap/README.md.
const spike008RecordedCap = 65536

const (
	spike008MeasuredBytes = 65536 // the workload the budgets are defined at (case 9)

	// D-13 budgets: 10% of the real deadlines (native 30 s, interpreter 2 min).
	spike008NativeBudget = 3 * time.Second
	spike008InterpBudget = 12 * time.Second

	// Default-mode ceilings: half the real deadlines, never a timing oracle.
	spike008NativeCeiling = 15 * time.Second
	spike008InterpCeiling = 60 * time.Second
)

type spike008Mode int

const (
	spike008Default spike008Mode = iota
	spike008Record
	spike008Strict
)

// spike008ModeFromEnv selects the mode from SCHWAY_SPIKE008 only. CI is never
// consulted, so no push or pull-request run can reach the 10% budget.
func spike008ModeFromEnv(t *testing.T) spike008Mode {
	t.Helper()
	switch v := os.Getenv("SCHWAY_SPIKE008"); v {
	case "":
		return spike008Default
	case "1":
		return spike008Record
	case "strict":
		return spike008Strict
	default:
		t.Fatalf("SCHWAY_SPIKE008=%q is not recognised: use unset (default), 1 (record) or strict", v)
		return spike008Default
	}
}

// spike008Stats drops the first (cold) run and reports median and max of the rest.
func spike008Stats(samples []time.Duration) (median, max time.Duration) {
	kept := append([]time.Duration(nil), samples[1:]...)
	sort.Slice(kept, func(i, j int) bool { return kept[i] < kept[j] })
	return kept[len(kept)/2], kept[len(kept)-1]
}

// spike008MeasureNative times one run of the artifact on one case and returns
// an error unless the checksum is the pinned literal, so a fast wrong answer
// can never be recorded.
func spike008MeasureNative(artifact string, caseID uint64) (time.Duration, error) {
	want, ok := testsupport.Phase29FixtureExpected(caseID)
	if !ok {
		return 0, fmt.Errorf("case %d has no pinned literal", caseID)
	}
	var stdout, stderr bytes.Buffer
	start := time.Now()
	outcome, err := native.RunApplication(context.Background(), artifact, strconv.FormatUint(caseID, 10), &stdout, &stderr)
	elapsed := time.Since(start)
	if err != nil {
		return 0, fmt.Errorf("run case %d: %w", caseID, err)
	}
	expected := strconv.FormatUint(want, 10) + "\n"
	if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != expected || stderr.Len() != 0 {
		return 0, fmt.Errorf("checksum mismatch on case %d: outcome=%+v stdout=%q stderr=%q, want %q", caseID, outcome, stdout.String(), stderr.String(), expected)
	}
	return elapsed, nil
}

func spike008MeasureInterp(run func() (interp.Execution, error), caseID uint64) (time.Duration, error) {
	want, ok := testsupport.Phase29FixtureExpected(caseID)
	if !ok {
		return 0, fmt.Errorf("case %d has no pinned literal", caseID)
	}
	start := time.Now()
	result, err := run()
	elapsed := time.Since(start)
	if err != nil {
		return 0, fmt.Errorf("interpreter case %d: %w", caseID, err)
	}
	if result.Outcome.Kind != execution.OutcomeReturned || result.Outcome.Value != strconv.FormatUint(want, 10) {
		return 0, fmt.Errorf("checksum mismatch on case %d: interpreter outcome=%+v, want returned %d", caseID, result.Outcome, want)
	}
	return elapsed, nil
}

func spike008Toolchain() string {
	clangVersion, clangTriple := "unavailable", "unavailable"
	if out, err := exec.Command("clang", "--version").Output(); err == nil {
		clangVersion = strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	}
	if out, err := exec.Command("clang", "-print-target-triple").Output(); err == nil {
		clangTriple = strings.TrimSpace(string(out))
	}
	return fmt.Sprintf("go=%s clang=%q GOOS/GOARCH=%s/%s target=%s", runtime.Version(), clangVersion, runtime.GOOS, runtime.GOARCH, clangTriple)
}

// spike008FitCap is the largest power of two whose linearly scaled 65,536 max
// fits the budget (1 when even a single byte would not).
func spike008FitCap(max, budget time.Duration) int {
	for c := spike008MeasuredBytes; c >= 1; c /= 2 {
		if time.Duration(float64(max)*float64(c)/float64(spike008MeasuredBytes)) <= budget {
			return c
		}
	}
	return 1
}

type spike008Lane struct {
	name      string
	native    bool
	maxAt64Ki time.Duration
	budget    time.Duration
	ceiling   time.Duration
	line      string
}

// TestSpike008ByteViewCap is the Stage-B D-12 / D-14 cap record, measured through
// the real D-05 fixture path. Modes are selected only by SCHWAY_SPIKE008: unset
// (default, case 9, 3 runs, half-deadline ceiling), 1 (record, full measurement,
// reports OVER BUDGET), strict (full measurement, an overrun fails).
func TestSpike008ByteViewCap(t *testing.T) {
	phase29RequireClang(t)
	mode := spike008ModeFromEnv(t)
	runs, cases := 3, []uint64{9}
	if mode != spike008Default {
		runs, cases = 7, []uint64{7, 8, 9}
	}
	t.Logf("spike008 mode=%d runs=%d cases=%v cap=%d %s", mode, runs, cases, spike008RecordedCap, spike008Toolchain())

	t.Run("armed-fault", spike008ArmedFault)

	source := testsupport.ProjectPath(phase29FixtureSource)
	var lanes []spike008Lane
	var table []string
	medians := map[string]map[uint64]time.Duration{}

	record := func(lane string, caseID uint64, samples []time.Duration, extra string) (time.Duration, time.Duration) {
		median, max := spike008Stats(samples)
		bytesLen := 0
		if data, ok := testsupport.Phase29FixtureBytes(caseID); ok {
			bytesLen = len(data)
		}
		want, _ := testsupport.Phase29FixtureExpected(caseID)
		line := fmt.Sprintf("lane=%s case=%d bytes=%d checksum=%d runs_kept=%d median=%s max=%s%s", lane, caseID, bytesLen, want, len(samples)-1, median, max, extra)
		table = append(table, line)
		t.Log(line)
		if medians[lane] == nil {
			medians[lane] = map[uint64]time.Duration{}
		}
		medians[lane][caseID] = median
		return median, max
	}

	for _, lane := range phase29Lanes() {
		var compileSamples []time.Duration
		var artifact string
		for i := 0; i < runs; i++ {
			start := time.Now()
			artifact = phase29BuildLane(t, source, lane.flags)
			compileSamples = append(compileSamples, time.Since(start))
		}
		info, err := os.Stat(artifact)
		if err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(artifact)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(content)
		cMedian, cMax := spike008Stats(compileSamples)
		compileLine := fmt.Sprintf("lane=%s compile median=%s max=%s artifact_bytes=%d sha256=%s flags=%v", lane.name, cMedian, cMax, info.Size(), hex.EncodeToString(digest[:]), lane.flags)
		table = append(table, compileLine)
		t.Log(compileLine)

		entry := spike008Lane{name: lane.name, native: true, budget: spike008NativeBudget, ceiling: spike008NativeCeiling}
		for _, caseID := range cases {
			var samples []time.Duration
			for i := 0; i < runs; i++ {
				elapsed, err := spike008MeasureNative(artifact, caseID)
				if err != nil {
					t.Fatalf("%s: %v", lane.name, err)
				}
				samples = append(samples, elapsed)
			}
			_, max := record(lane.name, caseID, samples, "")
			if caseID == 9 {
				entry.maxAt64Ki = max
			}
		}
		lanes = append(lanes, entry)
	}

	// Interpreter lane: check once, run repeatedly.
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(data)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	interpName := "interpreter"
	if phase29RaceEnabled {
		interpName = "interpreter-race"
	}
	entry := spike008Lane{name: interpName, budget: spike008InterpBudget, ceiling: spike008InterpCeiling}
	for _, caseID := range cases {
		var samples []time.Duration
		for i := 0; i < runs; i++ {
			caseID := caseID
			elapsed, err := spike008MeasureInterp(func() (interp.Execution, error) {
				return interp.RunWithFixtureBytes(checked.Program, "main", strconv.FormatUint(caseID, 10), testsupport.Phase29FixtureBytes)
			}, caseID)
			if err != nil {
				t.Fatalf("%s: %v", interpName, err)
			}
			samples = append(samples, elapsed)
		}
		_, max := record(interpName, caseID, samples, "")
		if caseID == 9 {
			entry.maxAt64Ki = max
		}
	}
	lanes = append(lanes, entry)

	if mode != spike008Default {
		for _, name := range func() []string {
			var n []string
			for lane := range medians {
				n = append(n, lane)
			}
			sort.Strings(n)
			return n
		}() {
			m := medians[name]
			t.Logf("scaling lane=%s median(16384)/median(4096)=%.2f median(65536)/median(4096)=%.2f (linear would be 4.00 and 16.00)", name, float64(m[8])/float64(m[7]), float64(m[9])/float64(m[7]))
		}
	}

	tableText := strings.Join(table, "\n")
	var overruns []string
	for _, lane := range lanes {
		scaled := time.Duration(float64(lane.maxAt64Ki) * float64(spike008RecordedCap) / float64(spike008MeasuredBytes))
		fit := spike008FitCap(lane.maxAt64Ki, lane.budget)
		switch mode {
		case spike008Default:
			if phase29RaceEnabled {
				t.Logf("lane=%s max(65536)=%s ceiling=%s (not asserted under the race detector)", lane.name, lane.maxAt64Ki, lane.ceiling)
			} else if lane.maxAt64Ki > lane.ceiling {
				t.Errorf("lane=%s max(65536)=%s exceeds the default ceiling %s", lane.name, lane.maxAt64Ki, lane.ceiling)
			}
		default:
			if scaled > lane.budget {
				msg := fmt.Sprintf("lane=%s OVER BUDGET: max(65536)=%s scaled to cap %d = %s > %s; largest fitting cap %d", lane.name, lane.maxAt64Ki, spike008RecordedCap, scaled, lane.budget, fit)
				t.Log(msg)
				overruns = append(overruns, msg)
			} else {
				t.Logf("lane=%s within budget: max(65536)=%s scaled to cap %d = %s <= %s (largest fitting cap %d)", lane.name, lane.maxAt64Ki, spike008RecordedCap, scaled, lane.budget, fit)
			}
		}
	}
	if mode == spike008Strict && len(overruns) > 0 {
		t.Fatalf("spike 008 strict budget overrun at cap %d:\n%s\n\nlane table:\n%s", spike008RecordedCap, strings.Join(overruns, "\n"), tableText)
	}
}

// spike008ArmedFault pins the spike CONVENTIONS armed fault: the
// measurement routine applied to a wrong-modulus mutant artifact reports a
// checksum mismatch, so the harness cannot record a fast wrong answer.
func spike008ArmedFault(t *testing.T) {
	phase29RequireClang(t)
	original, err := os.ReadFile(testsupport.ProjectPath(phase29FixtureSource))
	if err != nil {
		t.Fatal(err)
	}
	mutated := phase29ReplaceOnce(t, string(original), "a = a % 65521", "a = a % 65520")
	mutated = phase29ReplaceOnce(t, mutated, "b = b % 65521", "b = b % 65520")
	path := filepath.Join(t.TempDir(), "mutant.schway")
	if err := os.WriteFile(path, []byte(mutated), 0o600); err != nil {
		t.Fatal(err)
	}
	artifact := phase29BuildLane(t, path, phase29Lanes()[0].flags)
	detected := false
	for _, caseID := range []uint64{5, 6} {
		if _, err := spike008MeasureNative(artifact, caseID); err != nil && strings.Contains(err.Error(), "checksum mismatch") {
			detected = true
		}
	}
	if !detected {
		t.Fatal("wrong-modulus mutant artifact was measured without a checksum mismatch")
	}
	good := phase29BuildLane(t, testsupport.ProjectPath(phase29FixtureSource), phase29Lanes()[0].flags)
	if _, err := spike008MeasureNative(good, 5); err != nil {
		t.Fatalf("control: %v", err)
	}
}
