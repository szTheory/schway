package native

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SanitizerOptimization is D-05-12's pinned optimization level for every
// sanitizer-lane build: -O1, not -O0 (too far from shipped codegen), and
// not -O3 (false positives and inlining obscure blame). This flag is NEVER
// toggled onto the differential's own build (D-05-11/D-05-12) -- it only
// ever drives a separate, distinctly-pathed binary.
const SanitizerOptimization = "-O1"

// SanitizerCompileFlags is the exact, host-verified sanitizer compile
// invocation (05-RESEARCH.md Pattern 3, verified this phase). It is used
// ONLY by RunSanitized/ProbeSanitizerRuntime's own compile step, never
// merged into native.Runner.Run's differential build.
var SanitizerCompileFlags = []string{
	"-std=c17",
	SanitizerOptimization,
	"-g",
	"-fno-omit-frame-pointer",
	"-fsanitize=address,undefined",
	"-fno-sanitize-recover=all",
}

// ASanOptions and UBSanOptions are D-05-13's harness-pinned sanitizer
// runtime options. They are set explicitly in the sanitizer child process's
// own environment on every invocation (sanitizedEnviron) -- NEVER read from
// the contributor's environment -- so a hostile or merely default parent
// ASAN_OPTIONS/UBSAN_OPTIONS can never silently disable a check this
// control depends on.
//
// alloc_dealloc_mismatch=1 is explicit because it defaults OFF on Apple
// Clang/arm64 (05-RESEARCH.md Pitfall 1 -- [CITED], not independently
// [VERIFIED] this session, which is precisely why the pin is mandatory
// rather than belt-and-braces). detect_leaks=0 is explicit because leak
// detection is unreliable-to-absent on Apple clang arm64. symbolize=0 is
// what keeps absolute paths and ASLR-influenced addresses out of the
// classified report (D-05-16) -- raw stderr is still captured, bounded,
// separately, in SanitizerReport.TruncatedStderr.
const (
	ASanOptions  = "halt_on_error=1:abort_on_error=1:symbolize=0:detect_leaks=0:detect_odr_violation=0:alloc_dealloc_mismatch=1"
	UBSanOptions = "halt_on_error=1:print_stacktrace=0"
)

// SanitizerReport is D-05-16's stable classification tuple plus its bounded
// raw artifact. It is a DISTINCT type from native.Result, by design
// (D-05-11): there is no conversion function between the two anywhere in
// this package, and the equivalence comparator's signature cannot accept
// this type. A shared table with a boolean flag would be one refactor away
// from laundering sanitizer noise as a semantic mismatch; a type the
// comparator cannot consume is not.
type SanitizerReport struct {
	// Sanitizer is "address" or "undefined", or empty when no finding was
	// classified.
	Sanitizer string
	// CheckKind is the normalized diagnostic class (e.g.
	// "heap-use-after-free", "alloc-dealloc-mismatch", "runtime error"), or
	// "unclassified_abort" for a nonzero exit with no matching signature
	// (D-05-14: abort-without-report is an operational failure, never a
	// detection), or empty for a clean (zero-exit) run.
	CheckKind string
	// ReportSignature is the exact matched, enumerated substring -- never
	// the full symbolized trace, never a path, never an address (D-05-16).
	ReportSignature string
	ExitCode        int
	// TruncatedStderr is the raw captured stderr, bounded by the same 64
	// KiB-plus-one bound the rest of this package's stream capture uses.
	TruncatedStderr string
	// TruncationCode is the stable stage/stream truncation code set when
	// TruncatedStderr was actually truncated, empty otherwise.
	TruncationCode string
}

// SanitizerReportSignatures is D-05-14's classification table -- the ONLY
// place diagnostic substrings this control keys on appear in this file.
// Adding a signature is a reviewable, single-line act here, never
// duplicated at a call site.
var SanitizerReportSignatures = []struct{ Sanitizer, CheckKind, Substring string }{
	{Sanitizer: "address", CheckKind: "heap-use-after-free", Substring: "ERROR: AddressSanitizer: heap-use-after-free"},
	{Sanitizer: "address", CheckKind: "alloc-dealloc-mismatch", Substring: "ERROR: AddressSanitizer: alloc-dealloc-mismatch"},
	{Sanitizer: "undefined", CheckKind: "runtime error", Substring: "runtime error:"},
}

// sanitizerSmokeFixtureSource is D-05-15's two-line ASan+UBSan availability
// probe (05-RESEARCH.md Pattern 3, verified working on this host). It is
// embedded as a Go string constant so ProbeSanitizerRuntime works even when
// the repository's own testdata directory is absent.
const sanitizerSmokeFixtureSource = "#include <stdlib.h>\nint main(void){ char *p = malloc(1); free(p); return p[0]; }\n"

// sanitizedEnviron builds the child process environment for every
// sanitizer-lane invocation: every inherited ASAN_OPTIONS/UBSAN_OPTIONS
// entry is stripped, and the harness-pinned constants are appended in their
// place (D-05-13). A contributor's (or an attacker's) parent-process
// ASAN_OPTIONS/UBSAN_OPTIONS therefore never reaches the sanitizer binary.
func sanitizedEnviron() []string {
	inherited := os.Environ()
	env := make([]string, 0, len(inherited)+2)
	for _, keyValue := range inherited {
		if strings.HasPrefix(keyValue, "ASAN_OPTIONS=") || strings.HasPrefix(keyValue, "UBSAN_OPTIONS=") {
			continue
		}
		env = append(env, keyValue)
	}
	env = append(env, "ASAN_OPTIONS="+ASanOptions, "UBSAN_OPTIONS="+UBSanOptions)
	return env
}

// compileSanitized compiles cSource (plus foreignSources) with compileFlags
// into a DISTINCT output path -- suffixed ".sanitize", in its own temp
// directory, never reusing the differential's own output path (D-05-11) --
// and returns that binary path plus a cleanup function the caller MUST
// call. This deliberately duplicates native.Runner.Run/CompileOnly's
// compile-only steps rather than sharing code with them, so a change to the
// differential's own build can never accidentally change what the
// sanitizer lane builds (mirrors CompileOnly's own stated rationale).
func (r Runner) compileSanitized(parent context.Context, cSource string, compileFlags []string, foreignSources []string) (string, func(), *ToolError) {
	if r.ClangPath == "" {
		r.ClangPath = "clang"
	}
	if r.Timeout <= 0 {
		r.Timeout = 5 * time.Second
	}
	directory, err := os.MkdirTemp("", "lang-native-sanitize-")
	if err != nil {
		return "", func() {}, &ToolError{Code: "native.temp_failed", Err: err}
	}
	cleanup := func() { os.RemoveAll(directory) }

	sourcePath := filepath.Join(directory, "program.c")
	binaryPath := filepath.Join(directory, "program.sanitize")
	if err := os.WriteFile(sourcePath, []byte(cSource), 0o600); err != nil {
		cleanup()
		return "", func() {}, &ToolError{Code: "native.temp_failed", Err: err}
	}

	objectPaths := make([]string, 0, len(foreignSources))
	for index, foreignSource := range foreignSources {
		objectPath := filepath.Join(directory, fmt.Sprintf("foreign_%d.sanitize.o", index))
		foreignCtx, foreignCancel := context.WithTimeout(parent, r.Timeout)
		arguments := append(append([]string{}, compileFlags...), "-c", foreignSource, "-o", objectPath)
		foreignCommand := r.commandContext(foreignCtx, r.ClangPath, arguments...)
		var foreignStdout, foreignStderr boundedWriter
		foreignCommand.Stdout = &foreignStdout
		foreignCommand.Stderr = &foreignStderr
		foreignErr := foreignCommand.Run()
		deadlineExceeded := errors.Is(foreignCtx.Err(), context.DeadlineExceeded)
		foreignCancel()
		if deadlineExceeded {
			cleanup()
			return "", func() {}, &ToolError{Code: "native.timeout", Err: foreignCtx.Err()}
		}
		if foreignErr != nil {
			cleanup()
			code := "native.compile_failed"
			if errors.Is(foreignErr, exec.ErrNotFound) || errors.Is(foreignErr, os.ErrNotExist) {
				code = "native.tool_missing"
			}
			return "", func() {}, &ToolError{Code: code, Err: withStderr(foreignErr, foreignStderr.bytes())}
		}
		objectPaths = append(objectPaths, objectPath)
	}

	ctx, cancel := context.WithTimeout(parent, r.Timeout)
	defer cancel()
	arguments := append(append([]string{}, compileFlags...), sourcePath)
	arguments = append(arguments, objectPaths...)
	// -lc++ is linked unconditionally on every sanitizer-lane build
	// (Phase 5 plan 05-08, D-05-08): native/schway_foreign_arena.c's dynamic
	// allocator-mismatch defect calls the Itanium-mangled operator-new/
	// operator-delete entry points (_Znwm/_ZdlPv) directly from plain C to
	// reach a genuinely ASan-distinguishable allocator identity -- verified
	// empirically that no purely-libc allocator pairing (malloc, calloc,
	// realloc, posix_memalign, aligned_alloc, memalign, valloc) produces
	// ASan's alloc-dealloc-mismatch diagnostic on this host, since every
	// one of them shares ASan's single FROM_MALLOC allocation-type bucket.
	// Harmless for every other sanitizer-lane build: an unreferenced
	// library adds no behavior.
	arguments = append(arguments, "-lc++")
	arguments = append(arguments, "-o", binaryPath)
	command := r.commandContext(ctx, r.ClangPath, arguments...)
	var stdout, stderr boundedWriter
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		cleanup()
		return "", func() {}, &ToolError{Code: "native.timeout", Err: ctx.Err()}
	}
	if stdout.overflowed() {
		cleanup()
		return "", func() {}, &ToolError{Code: "native.compile_stdout_truncated", Err: fmt.Errorf("process stream exceeded %d bytes", MaxStreamBytes)}
	}
	if stderr.overflowed() {
		cleanup()
		return "", func() {}, &ToolError{Code: "native.compile_stderr_truncated", Err: fmt.Errorf("process stream exceeded %d bytes", MaxStreamBytes)}
	}
	if err != nil {
		cleanup()
		code := "native.compile_failed"
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			code = "native.tool_missing"
		}
		return "", func() {}, &ToolError{Code: code, Err: withStderr(err, stderr.bytes())}
	}
	return binaryPath, cleanup, nil
}

// classifySanitizerRun is D-05-14's pure classifier: it decides ONLY from a
// completed process's exit code and captured stderr, never from a live
// process, which is what makes it directly unit-testable on synthetic
// input. A zero exit produces no detection (D-05-14: a clean sanitizer run
// is never evidence). A nonzero exit is classified on signature substring
// match; a nonzero exit with NO matching signature is "unclassified_abort"
// -- an OPERATIONAL failure (returned as a *ToolError), never a detected
// finding, since abort-without-report is exactly the false-green mode this
// control exists to catch.
func classifySanitizerRun(exitCode int, stderrText string) (SanitizerReport, *ToolError) {
	if exitCode == 0 {
		return SanitizerReport{ExitCode: exitCode}, nil
	}
	for _, signature := range SanitizerReportSignatures {
		if strings.Contains(stderrText, signature.Substring) {
			return SanitizerReport{
				Sanitizer:       signature.Sanitizer,
				CheckKind:       signature.CheckKind,
				ReportSignature: signature.Substring,
				ExitCode:        exitCode,
			}, nil
		}
	}
	report := SanitizerReport{CheckKind: "unclassified_abort", ExitCode: exitCode}
	return report, &ToolError{Code: "native.sanitizer_unclassified_abort", Err: fmt.Errorf("sanitizer process aborted with exit code %d but produced no recognized signature", exitCode)}
}

// runSanitizedBinary runs binaryPath once with input as its sole argv (or
// no argument when input is empty), with ASAN_OPTIONS/UBSAN_OPTIONS pinned
// explicitly in the child environment via sanitizedEnviron (D-05-13), and
// classifies the result via classifySanitizerRun. A genuine nonzero exit
// (including a signalled SIGABRT termination, which Go's ExitCode reports
// as -1) is expected and routed to the classifier, never treated as a Go
// error on its own -- only a process that failed to even start (tool
// missing, timeout) is a *ToolError from THIS function's own perspective;
// classifySanitizerRun may still return its own *ToolError for an
// unclassified abort.
func (r Runner) runSanitizedBinary(parent context.Context, binaryPath, input string) (SanitizerReport, *ToolError) {
	if r.Timeout <= 0 {
		r.Timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, r.Timeout)
	defer cancel()
	var arguments []string
	if input != "" {
		arguments = []string{input}
	}
	command := r.commandContext(ctx, binaryPath, arguments...)
	command.Env = sanitizedEnviron()
	var stdout, stderr boundedWriter
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return SanitizerReport{}, &ToolError{Code: "native.timeout", Err: ctx.Err()}
	}
	if runErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) {
			code := "native.run_failed"
			if errors.Is(runErr, exec.ErrNotFound) || errors.Is(runErr, os.ErrNotExist) {
				code = "native.tool_missing"
			}
			return SanitizerReport{}, &ToolError{Code: code, Err: withStderr(runErr, stderr.bytes())}
		}
	}
	exitCode := -1
	if command.ProcessState != nil {
		exitCode = command.ProcessState.ExitCode()
	}
	report, toolErr := classifySanitizerRun(exitCode, string(stderr.bytes()))
	report.TruncatedStderr = string(stderr.bytes())
	if stderr.overflowed() {
		report.TruncationCode = "native.run_stderr_truncated"
	}
	return report, toolErr
}

// RunSanitized compiles the given generated C plus any configured
// r.ForeignSources with SanitizerCompileFlags into a DISTINCT output path
// and runs it once per input (or once with no argument when inputs is
// empty), with ASAN_OPTIONS/UBSAN_OPTIONS pinned explicitly in the child
// environment, classifying each run. If any input produces a classified
// finding (or an unclassified abort), that report/error is returned
// immediately; otherwise the last (clean) report is returned.
//
// Per D-05-11, this returns SanitizerReport, NOT native.Result: there is
// deliberately no shared table, no boolean "sanitizer-tainted" field, and
// no conversion function anywhere in this package between the two types --
// the equivalence comparator's signature cannot accept a SanitizerReport.
func (r Runner) RunSanitized(parent context.Context, cSource string, inputs []string) (SanitizerReport, error) {
	binaryPath, cleanup, toolErr := r.compileSanitized(parent, cSource, SanitizerCompileFlags, r.ForeignSources)
	if toolErr != nil {
		return SanitizerReport{}, toolErr
	}
	defer cleanup()

	if len(inputs) == 0 {
		report, runToolErr := r.runSanitizedBinary(parent, binaryPath, "")
		if runToolErr != nil {
			return report, runToolErr
		}
		return report, nil
	}
	var last SanitizerReport
	for _, input := range inputs {
		report, runToolErr := r.runSanitizedBinary(parent, binaryPath, input)
		if runToolErr != nil {
			return report, runToolErr
		}
		last = report
		if report.CheckKind != "" {
			return report, nil
		}
	}
	return last, nil
}

// probeOutcome is D-05-15's posture decision over a completed smoke-fixture
// run: only the expected heap-use-after-free ASan finding is a pass. Every
// other shape -- a clean exit, an unclassified abort, a different finding
// -- means the binary was not actually instrumented as expected, and is
// reported as native.sanitizer_inert, never a pass (this is the "silently
// isn't linked against the sanitizer runtime" failure mode 05-RESEARCH.md
// Pitfall 4 names).
func probeOutcome(report SanitizerReport) *ToolError {
	if report.Sanitizer == "address" && report.CheckKind == "heap-use-after-free" {
		return nil
	}
	return &ToolError{Code: "native.sanitizer_inert", Err: fmt.Errorf("sanitizer smoke fixture produced no heap-use-after-free report (exit=%d, check_kind=%q)", report.ExitCode, report.CheckKind)}
}

// ProbeSanitizerRuntime compiles and runs sanitizerSmokeFixtureSource with
// SanitizerCompileFlags and the pinned options, asserting it produces the
// heap-use-after-free signature with a nonzero exit (D-05-15). Posture
// rules, copied verbatim from control:foreign.unwind_forbidden's nm -u
// probe (D-04-19):
//
//   - Clang absent, the sanitizer runtime absent, or the smoke fixture
//     failing to compile/link returns native.tool_missing -- an OPERATIONAL
//     status, never a pass, never a silent skip.
//   - The smoke fixture linking but producing no report means the binary is
//     not actually instrumented; returns native.sanitizer_inert, distinct
//     from tool_missing.
//   - Only a smoke run that produces the expected signature returns
//     success.
func (r Runner) ProbeSanitizerRuntime(parent context.Context) (SanitizerReport, error) {
	binaryPath, cleanup, toolErr := r.compileSanitized(parent, sanitizerSmokeFixtureSource, SanitizerCompileFlags, nil)
	if toolErr != nil {
		code := "native.tool_missing"
		return SanitizerReport{}, &ToolError{Code: code, Err: toolErr}
	}
	defer cleanup()

	report, runToolErr := r.runSanitizedBinary(parent, binaryPath, "")
	if runToolErr != nil && runToolErr.Code != "native.sanitizer_unclassified_abort" {
		// The binary failed to even run (tool missing, timeout) -- never a
		// pass, and distinct from an inert (ran cleanly / aborted without a
		// report) binary.
		return SanitizerReport{}, &ToolError{Code: "native.tool_missing", Err: runToolErr}
	}
	if inertErr := probeOutcome(report); inertErr != nil {
		return report, inertErr
	}
	return report, nil
}
