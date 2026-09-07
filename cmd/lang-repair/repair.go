// Package main implements lang-repair, DX-04's standalone repair driver
// (D-06-28). It is architecturally its own tier, walled off from the
// compiler internals by a structural import-boundary lint
// (import_boundary_test.go) rather than by convention: it spawns the
// shipped `lang` binary as a subprocess and repairs source using ONLY the
// `--json check FILE` protocol's structured repairs[] channel -- never the
// human-readable message or per-cause/per-repair detail prose, and it never
// imports any package under internal/.
//
// The driver is deliberately generic: check.go already attaches a complete
// Span/Replacement/MachineApplicable repair to every mechanically-fixable
// diagnostic (add_missing_arm, use_transfer_target, insert_take,
// narrow_to_shared_borrow), so the driver never needs a hardcoded
// kind-to-edit mapping -- it splices whatever replacement text the wire
// document hands it into whatever span the wire document names, for
// whichever repair kind happens to be MachineApplicable. This is the point
// of D-06-28's "protocol consumer, not library consumer" framing: the
// driver's correctness does not depend on knowing what "add_missing_arm" or
// "use_transfer_target" MEAN, only on applying Span/Replacement literally.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// maxStdoutBytes bounds the `lang --json check` stdout capture the driver
// reads (T-06-BOUNDARY-05). The driver may not import
// internal/compiler/testsupport to reuse its own MaxCLIStreamBytes
// constant (D-06-28), so this is declared locally at the same 1 MiB value.
const maxStdoutBytes = 1 << 20

// subprocessTimeout bounds every `lang` invocation the driver makes,
// mirroring the bounded subprocess discipline used elsewhere in the tree.
const subprocessTimeout = 30 * time.Second

// jsonSpan, jsonRepair, jsonDiagnostic, and checkResult are the driver's own
// minimal local re-declarations of the `lang.command/1` document shape it
// consumes (D-06-28: the driver may not import
// internal/compiler/protocol or internal/compiler/diagnostic). They
// deliberately omit every prose field -- the diagnostic's own top-level
// "message" and each cause/repair's "detail" -- so a field that is never
// decoded cannot be scraped (T-06-BOUNDARY-03,
// TestRepairDriverDecodesNoProseFields).
type jsonSpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type jsonRepair struct {
	Kind          string    `json:"kind"`
	Span          *jsonSpan `json:"span,omitempty"`
	Replacement   string    `json:"replacement,omitempty"`
	Applicability string    `json:"applicability,omitempty"`
}

type jsonDiagnostic struct {
	Code    string       `json:"code"`
	Repairs []jsonRepair `json:"repairs,omitempty"`
}

type checkResult struct {
	Status      string           `json:"status"`
	Diagnostics []jsonDiagnostic `json:"diagnostics"`
}

const (
	statusPass                     = "pass"
	applicabilityMachineApplicable = "MachineApplicable"
)

// DriverError is the driver's own typed-failure shape, matching the
// {Code}-plus-wrapped-Err convention this project already uses
// (native.ToolError, evidence.ValidationError, session.InjectorError).
type DriverError struct {
	Code string
	Err  error
}

func (e *DriverError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *DriverError) Unwrap() error { return e.Err }

// Typed failure codes the driver returns. These are internal to
// cmd/lang-repair -- they never need to line up with any internal/ package's
// own vocabulary, precisely because the driver never imports one.
const (
	CodeStdoutCapExceeded = "repair.stdout_cap_exceeded"
	CodeSubprocessFailed  = "repair.subprocess_failed"
	CodeDecodeFailed      = "repair.decode_failed"
	CodeSpanOutOfRange    = "repair.span_out_of_range"
	CodeSourceReadFailed  = "repair.source_read_failed"
	CodeSourceWriteFailed = "repair.source_write_failed"
)

// Outcome status vocabulary. OutcomeUnrepairable is returned, never
// silently skipped and never conflated with a pass, when the binary emits
// zero driver-eligible repairs for a genuinely invalid program (FND-04
// empty-input edge, TestUnrepairableDefectFailsTheGate).
const (
	OutcomeAlreadyClean   = "already_clean"
	OutcomeRepaired       = "repaired"
	OutcomeUnrepairable   = "unrepairable"
	OutcomeReverifyFailed = "reverify_failed"
)

// Outcome is the driver's own result envelope, reported on stdout by main.go.
type Outcome struct {
	Status          string `json:"status"`
	DiagnosisCode   string `json:"diagnosis_code,omitempty"`
	RepairKind      string `json:"repair_kind,omitempty"`
	SubprocessCount int    `json:"subprocess_count"`
}

// boundedWriter caps the bytes captured from a subprocess stream at cap+1 --
// the extra byte makes overflow observable without silently truncating into
// a plausible-looking (and wrong) document, mirroring this project's own
// bounded-subprocess-capture convention (testsupport.boundedWriter,
// T-06-BOUNDARY-05). The driver may not import
// internal/compiler/testsupport to reuse its own type (D-06-28), so this is
// declared locally. Using an assigned io.Writer plus cmd.Run() -- never
// StdoutPipe/StderrPipe, CombinedOutput, or Output -- is this project's own
// required bounded-subprocess shape (TestSourceNeverSpawnsUnboundedProcesses).
type boundedWriter struct {
	buffer bytes.Buffer
	total  int
	cap    int
}

func newBoundedWriter(cap int) *boundedWriter { return &boundedWriter{cap: cap} }

func (w *boundedWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := w.cap + 1 - w.buffer.Len()
	if remaining > len(data) {
		remaining = len(data)
	}
	if remaining > 0 {
		_, _ = w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func (w *boundedWriter) overflowed() bool { return w.total > w.cap }
func (w *boundedWriter) bytes() []byte    { return w.buffer.Bytes() }

// runLangCheck spawns `lang --json check FILE` -- the driver's ONLY
// permitted subprocess and ONLY permitted input channel -- with a bounded
// stdout writer and a deadline, and strict-decodes the result into the
// driver's own locally declared structs. A nonzero `lang` exit code alone
// is not a failure here (StatusInvalid legitimately exits nonzero); only an
// oversized stream or an undecodable document is.
func runLangCheck(ctx context.Context, langBinary, sourcePath string) (checkResult, error) {
	ctx, cancel := context.WithTimeout(ctx, subprocessTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, langBinary, "--json", "check", sourcePath)
	stdout := newBoundedWriter(maxStdoutBytes)
	stderr := newBoundedWriter(maxStdoutBytes)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	runErr := cmd.Run()
	if stdout.overflowed() {
		return checkResult{}, &DriverError{Code: CodeStdoutCapExceeded, Err: fmt.Errorf("stdout exceeded the declared %d byte cap", maxStdoutBytes)}
	}
	if runErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) {
			return checkResult{}, &DriverError{Code: CodeSubprocessFailed, Err: fmt.Errorf("%w (stderr: %s)", runErr, stderr.bytes())}
		}
	}
	var result checkResult
	if err := json.Unmarshal(stdout.bytes(), &result); err != nil {
		return checkResult{}, &DriverError{Code: CodeDecodeFailed, Err: fmt.Errorf("decoding lang --json check output: %w (stderr: %s)", err, stderr.bytes())}
	}
	return result, nil
}

// driverEligible mirrors diagnostic.DriverEligible's own rule (06-11,
// extended by 06-14's D-06-27.2 to also require a non-empty Kind) --
// re-declared, not imported (D-06-28): a repair is eligible only when it
// declares MachineApplicable, carries both a span and replacement text,
// AND carries a Kind.
func driverEligible(r jsonRepair) bool {
	return r.Applicability == applicabilityMachineApplicable && r.Kind != "" && r.Span != nil && r.Replacement != ""
}

// selectRepair returns the first driver-eligible repair in document order
// (diagnostics order, then each diagnostic's own repairs order) -- the
// driver's specified, stable tie-break rule for a diagnostic that carries
// two equally applicable repairs (FND-04 ordering edge,
// TestRepairSelectionIsSpecifiedOnTies). Document order is itself
// deterministic on the wire: check.go emits diagnostics in traversal order,
// and diagnostic.ErrorWithRepairs canonically sorts each diagnostic's own
// repairs by Kind then Detail before they ever reach JSON.
func selectRepair(result checkResult) (jsonRepair, string, bool) {
	for _, d := range result.Diagnostics {
		for _, r := range d.Repairs {
			if driverEligible(r) {
				return r, d.Code, true
			}
		}
	}
	return jsonRepair{}, "", false
}

func spanStart(s *jsonSpan) int {
	if s == nil {
		return -1
	}
	return s.Start
}

func spanEnd(s *jsonSpan) int {
	if s == nil {
		return -1
	}
	return s.End
}

// applyRepair splices repair.Replacement into the source at exactly
// [repair.Span.Start, repair.Span.End) -- the driver's ONLY write path, and
// the only reason it ever opens sourcePath at all
// (TestRepairDriverNeverOpensSourceOutsideSpan). A span outside the
// source's byte bounds is refused with CodeSpanOutOfRange rather than
// clamped: a clamped out-of-range span is exactly how a splice would escape
// its declared range (T-06-BOUNDARY-02).
func applyRepair(sourcePath string, repair jsonRepair) error {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return &DriverError{Code: CodeSourceReadFailed, Err: err}
	}
	span := repair.Span
	if span == nil || span.Start < 0 || span.End < span.Start || span.End > len(source) {
		return &DriverError{Code: CodeSpanOutOfRange, Err: fmt.Errorf(
			"repair span [%d,%d) is outside source bounds [0,%d)", spanStart(span), spanEnd(span), len(source),
		)}
	}
	patched := make([]byte, 0, len(source)-(span.End-span.Start)+len(repair.Replacement))
	patched = append(patched, source[:span.Start]...)
	patched = append(patched, []byte(repair.Replacement)...)
	patched = append(patched, source[span.End:]...)
	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(sourcePath); statErr == nil {
		mode = info.Mode()
	}
	if err := os.WriteFile(sourcePath, patched, mode); err != nil {
		return &DriverError{Code: CodeSourceWriteFailed, Err: err}
	}
	return nil
}

// Repair drives the single-pass repair cycle D-06-30 requires: one
// diagnose, at most one apply, at most one reverify -- never a loop.
// Subprocess counts are therefore always exactly 1 (already_clean or
// unrepairable) or exactly 2 (repaired or reverify_failed); this is the
// structural property TestRepairDriverFixesEveryDefectClassSinglePass
// asserts by counting invocations, and it is asserted BY CONSTRUCTION here
// -- there is no loop construct anywhere in this function.
func Repair(ctx context.Context, langBinary, sourcePath string) (Outcome, error) {
	first, err := runLangCheck(ctx, langBinary, sourcePath)
	if err != nil {
		return Outcome{}, err
	}
	if first.Status == statusPass {
		return Outcome{Status: OutcomeAlreadyClean, SubprocessCount: 1}, nil
	}
	repair, code, ok := selectRepair(first)
	if !ok {
		// A defect for which the binary emits zero driver-eligible repairs
		// is reported unrepairable -- never skipped, never counted as a
		// pass (FND-04 empty-input edge).
		return Outcome{Status: OutcomeUnrepairable, DiagnosisCode: code, SubprocessCount: 1}, nil
	}
	if err := applyRepair(sourcePath, repair); err != nil {
		return Outcome{}, err
	}
	second, err := runLangCheck(ctx, langBinary, sourcePath)
	if err != nil {
		return Outcome{}, err
	}
	if second.Status != statusPass {
		return Outcome{Status: OutcomeReverifyFailed, DiagnosisCode: code, RepairKind: repair.Kind, SubprocessCount: 2}, nil
	}
	return Outcome{Status: OutcomeRepaired, DiagnosisCode: code, RepairKind: repair.Kind, SubprocessCount: 2}, nil
}
