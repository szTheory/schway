package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

// Phase5SanitizeResult is lane:native-sanitize's own verify-lane report,
// the same shape verifyForeignCorpus/VerifyCorpus already use elsewhere in
// this package -- a type alias, not a new report format, so this lane
// composes with the rest of the phase gate without any conversion step.
type Phase5SanitizeResult = protocol.Result

const (
	// ControlSanitizeRetainedPointer is D-05-14's always-on positive
	// control: the deliberately-defective retained-pointer fixture MUST
	// report a genuine heap-use-after-free finding on EVERY verify
	// invocation, not merely once at authoring time. If it does not, the
	// whole lane is red regardless of what else passed -- a clean
	// sanitizer run is never evidence.
	ControlSanitizeRetainedPointer = "control:native.sanitize.retained_pointer"
	// ControlSanitizeUseAfterFree is the diagnostic-class control fired by
	// the SAME retained-pointer positive control above: it names the
	// underlying heap-use-after-free defect class the fixture proves is
	// caught, distinct from ControlSanitizeRetainedPointer's own
	// always-on-gate identity.
	ControlSanitizeUseAfterFree = "control:native.sanitize.use_after_free"
	// ControlSanitizeAllocatorMismatch is D-05-08's dynamic allocator-
	// mismatch control: testdata/phase5/allocator_mismatch.lang must
	// report ASan's own alloc-dealloc-mismatch diagnostic.
	ControlSanitizeAllocatorMismatch = "control:native.sanitize.allocator_mismatch"
	// ControlSanitizeUBSanNoRecover proves -fno-sanitize-recover=all
	// actually took effect (D-05-13/D-05-14) -- what catches a typo'd
	// options string. A UBSan-triggering fixture must report a genuine
	// "runtime error:" finding with a nonzero exit code.
	ControlSanitizeUBSanNoRecover = "control:native.sanitize.ubsan_no_recover"
	// LaneNativeSanitize is this lane's own ID in the phase gate's Lanes
	// list.
	LaneNativeSanitize = "lane:native-sanitize"
	// EscapeCallbackInvocationUnsubjected is D-05-07's named, gate-visible
	// residual: NAT-03 stale-callback-retention is subjected via the
	// retained-pointer lane (a `static`-stashed borrowed pointer
	// dereferenced post-release, detected by ASan). The
	// callback-invocation mechanism itself -- registration, later firing,
	// generation tokens -- has no subject, because M001 has no
	// calls-into-Lang by design (D-04-01). The pointer-lifetime root cause
	// is tested; the invocation-time variant is not. This escape MUST
	// appear in ExpectedEscapes and MUST NEVER appear in any control list.
	EscapeCallbackInvocationUnsubjected = "escape:callback-invocation-unsubjected"
)

// phase5SanitizeProjectRoot mirrors nat03ProjectRoot's own
// runtime.Caller(0)-anchored technique (this file is exactly as deep as
// session_phase5_alias.go) rather than importing the test-only testsupport
// package into production code.
func phase5SanitizeProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func phase5SanitizeFixturePath(relative string) string {
	return filepath.Join(phase5SanitizeProjectRoot(), filepath.FromSlash(relative))
}

// sanitizerSignatureFor looks up D-05-14's classification substring for a
// given check kind from native.SanitizerReportSignatures -- the ONLY place
// this file's controls key their expected substrings from (the plan's own
// instruction: reuse SanitizerReportSignatures, never re-spell the
// substrings here).
func sanitizerSignatureFor(checkKind string) string {
	for _, signature := range native.SanitizerReportSignatures {
		if signature.CheckKind == checkKind {
			return signature.Substring
		}
	}
	return ""
}

// compilePhase5SanitizeFixture reads, checks, independently validates, and
// lowers a testdata/phase5 fixture to generated C, for use as one of this
// lane's controls.
func compilePhase5SanitizeFixture(relative string) (string, error) {
	source, err := os.ReadFile(phase5SanitizeFixturePath(relative))
	if err != nil {
		return "", fmt.Errorf("%s: %w", relative, err)
	}
	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		return "", fmt.Errorf("%s: fixture failed to check: %+v", relative, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return "", fmt.Errorf("%s: fixture rejected by corevalidate: %+v", relative, validated.Problems)
	}
	cSource, err := Phase16ControlNativeC(validated.Program(), relative)
	if err != nil {
		return "", fmt.Errorf("%s: cgen failed: %w", relative, err)
	}
	return cSource, nil
}

// ubsanTriggerFixtureSource is D-05-14's UBSan-triggering control: a
// genuine signed-integer-overflow, verified on this host to produce
// UBSan's "runtime error:" report with a nonzero exit code under
// -fno-sanitize-recover=all. Embedded as a Go string constant, exactly
// like sanitize.go's own sanitizerSmokeFixtureSource, so this control
// never depends on the repository's own testdata directory.
const ubsanTriggerFixtureSource = "#include <limits.h>\nint main(void){ int x = INT_MAX; x = x + 1; return x; }\n"

// Phase5RetainedPointerFixtureLoaderForTest resolves the retained-pointer
// positive control's compiled C. Exported as a mutable seam ONLY for
// TestSanitizeLaneCleanPositiveControlFailsTheLane's mutation-kill
// (D-05-14): the test substitutes it with a benign, non-defective
// compiled C to prove VerifyPhase5SanitizeLane goes RED when the positive
// control's own fixture source stops producing a genuine defect --
// without forking VerifyPhase5SanitizeLane's control flow or widening its
// plan-specified signature. Production callers must never set this.
var Phase5RetainedPointerFixtureLoaderForTest = func() (string, error) {
	return compilePhase5SanitizeFixture("testdata/phase5/retained_pointer.lang")
}

// VerifyPhase5SanitizeLane is lane:native-sanitize's own verify-lane
// dispatch (D-05-08/D-05-09/D-05-14/D-05-15): it probes sanitizer runtime
// availability, always runs the retained-pointer positive control, and
// runs the allocator-mismatch and UBSan controls, keying every assertion
// on diagnostic substring PLUS exit code (never "process exited nonzero"
// alone). The callback-invocation residual is declared as a named,
// gate-visible expected escape, never claimed as covered.
func VerifyPhase5SanitizeLane(ctx context.Context, runner native.Runner) (Phase5SanitizeResult, error) {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	result.ExpectedEscapes = []string{EscapeCallbackInvocationUnsubjected}

	addLane := func(id, status string, controls []string, work int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		})
		result.Metrics.RecomputedWork += work
	}
	fail := func(status, code, message string) (Phase5SanitizeResult, error) {
		result.Status = status
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(code, diagnostic.Span{}, message))
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
		return result, nil
	}

	// (a) Probe sanitizer runtime availability. A tool that is absent, or
	// a binary that links but is not genuinely instrumented, is NEVER a
	// pass -- the exact posture control:foreign.unwind_forbidden took for
	// `nm` in D-04-19.
	probeStarted := time.Now()
	if _, probeErr := runner.ProbeSanitizerRuntime(ctx); probeErr != nil {
		addLane(LaneNativeSanitize, protocol.StatusOperational, nil, 1, probeStarted)
		return fail(protocol.StatusOperational, "native.sanitizer_runtime_unavailable", fmt.Sprintf("sanitizer runtime probe failed: %v", probeErr))
	}

	// (b) The always-on retained-pointer positive control (D-05-14): a
	// clean run here is not evidence -- it means the lane's own detection
	// path is inert, and the whole lane goes red.
	retainedStarted := time.Now()
	retainedSource, err := Phase5RetainedPointerFixtureLoaderForTest()
	if err != nil {
		return Phase5SanitizeResult{}, err
	}
	retainedRunner := runner
	retainedRunner.ForeignSources = []string{native.ForeignRetainedSourcePath()}
	retainedReport, retainedErr := retainedRunner.RunSanitized(ctx, retainedSource, []string{"7"})
	wantUseAfterFree := sanitizerSignatureFor("heap-use-after-free")
	if retainedErr != nil || retainedReport.ExitCode == 0 || retainedReport.ReportSignature != wantUseAfterFree {
		addLane(LaneNativeSanitize, protocol.StatusMismatch, nil, 1, retainedStarted)
		return fail(protocol.StatusMismatch, "native.sanitizer_positive_control_silent",
			fmt.Sprintf("%s: the always-on positive control did not report %q (exit=%d, check_kind=%q, err=%v)",
				ControlSanitizeRetainedPointer, wantUseAfterFree, retainedReport.ExitCode, retainedReport.CheckKind, retainedErr))
	}
	addLane(LaneNativeSanitize, protocol.StatusPass, []string{ControlSanitizeRetainedPointer, ControlSanitizeUseAfterFree}, 1, retainedStarted)

	// (c) Allocator-mismatch dynamic control (D-05-08). alloc_dealloc_mismatch=1
	// is pinned in sanitize.go's ASanOptions specifically because it
	// defaults OFF on macOS -- a contributor on default options would get
	// a false green with no warning, and 05-RESEARCH.md rates that claim
	// [CITED], not [VERIFIED], which is why the pin is mandatory rather
	// than belt-and-braces. If the fixture does not trip ASan on this
	// host, the lane goes RED and the fixture is treated as broken; the
	// assertion below is never relaxed.
	allocatorStarted := time.Now()
	allocatorSource, err := compilePhase5SanitizeFixture("testdata/phase5/allocator_mismatch.lang")
	if err != nil {
		return Phase5SanitizeResult{}, err
	}
	allocatorRunner := runner
	allocatorRunner.ForeignSources = []string{native.ForeignArenaSourcePath()}
	allocatorReport, allocatorErr := allocatorRunner.RunSanitized(ctx, allocatorSource, []string{"7"})
	wantAllocMismatch := sanitizerSignatureFor("alloc-dealloc-mismatch")
	if allocatorErr != nil || allocatorReport.ExitCode == 0 || allocatorReport.ReportSignature != wantAllocMismatch {
		addLane(LaneNativeSanitize, protocol.StatusMismatch, nil, 1, allocatorStarted)
		return fail(protocol.StatusMismatch, "native.sanitizer_allocator_mismatch_silent",
			fmt.Sprintf("%s: allocator-mismatch fixture did not report %q (exit=%d, check_kind=%q, err=%v)",
				ControlSanitizeAllocatorMismatch, wantAllocMismatch, allocatorReport.ExitCode, allocatorReport.CheckKind, allocatorErr))
	}
	addLane(LaneNativeSanitize, protocol.StatusPass, []string{ControlSanitizeAllocatorMismatch}, 1, allocatorStarted)

	// (d) UBSan no-recover control (D-05-13/D-05-14): proves
	// -fno-sanitize-recover=all actually took effect.
	ubsanStarted := time.Now()
	ubsanReport, ubsanErr := runner.RunSanitized(ctx, ubsanTriggerFixtureSource, nil)
	wantRuntimeError := sanitizerSignatureFor("runtime error")
	if ubsanErr != nil || ubsanReport.ExitCode == 0 || ubsanReport.ReportSignature != wantRuntimeError {
		addLane(LaneNativeSanitize, protocol.StatusMismatch, nil, 1, ubsanStarted)
		return fail(protocol.StatusMismatch, "native.sanitizer_ubsan_no_recover_silent",
			fmt.Sprintf("%s: UBSan fixture did not report %q (exit=%d, check_kind=%q, err=%v)",
				ControlSanitizeUBSanNoRecover, wantRuntimeError, ubsanReport.ExitCode, ubsanReport.CheckKind, ubsanErr))
	}
	addLane(LaneNativeSanitize, protocol.StatusPass, []string{ControlSanitizeUBSanNoRecover}, 1, ubsanStarted)

	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result, nil
}
