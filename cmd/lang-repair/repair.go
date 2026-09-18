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
	"strings"
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
	statusPass = "pass"

	// The closed applicability vocabulary, re-declared locally (D-06-28)
	// from diagnostic.Applicability* rather than imported -- only
	// MachineApplicable was needed before this change; the other two are
	// added now so BestApplicability can rank every value the wire
	// document can carry, never just the one the driver already acted on.
	applicabilityMachineApplicable    = "MachineApplicable"
	applicabilityRequiresConfirmation = "RequiresConfirmation"
	applicabilityUnspecified          = "Unspecified"
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

// Decline reason vocabulary (D-14-42, DX-09). Closed and coded -- every
// value here is a local constant, never a copy of diagnostic prose, so
// lorem-ipsum-scrambling the underlying diagnostics leaves these values
// unchanged (T-14-15).
const (
	// DeclineNoneOffered: the program is invalid, carries at least one
	// diagnostic, and zero repairs[] appear anywhere in the document --
	// this is the spiral's value (D-14-42): "a cycle has no local,
	// mechanical edit" becomes a stated fact instead of silence.
	DeclineNoneOffered = "repair.none_offered"
	// DeclineNoneEligible: repairs are present somewhere in the document,
	// but none is driver-eligible (driverEligible returns false for all
	// of them).
	DeclineNoneEligible = "repair.none_eligible"
	// DeclineNoDiagnostics: the program is invalid but reports zero
	// diagnostics -- a fail-closed anomaly. This must still read as
	// unrepairable, never a pass (T-14-17).
	DeclineNoDiagnostics = "repair.no_diagnostics"
)

// diagnosisCodeList is a comparable, JSON-array-marshaling ordered list of
// diagnostic codes (D-14-42). It is backed by a "|"-joined string rather
// than a plain []string because several existing antitheater_test.go
// checks compare two Outcome values with plain `!=`
// (TestProseScrambleLeavesRepairBehaviourIdentical), and antitheater_test.go
// must stay byte-unchanged by this task -- a slice-typed field would make
// Outcome uncomparable and break that comparison at compile time. "|" is a
// safe separator: every diagnostic code on the wire is a dotted
// lowercase/underscore identifier (e.g. "syntax.expected_rbrace") and never
// contains "|".
type diagnosisCodeList string

func newDiagnosisCodeList(codes []string) diagnosisCodeList {
	return diagnosisCodeList(strings.Join(codes, "|"))
}

func (d diagnosisCodeList) codes() []string {
	if d == "" {
		return nil
	}
	return strings.Split(string(d), "|")
}

func (d diagnosisCodeList) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.codes())
}

func (d *diagnosisCodeList) UnmarshalJSON(data []byte) error {
	var codes []string
	if err := json.Unmarshal(data, &codes); err != nil {
		return err
	}
	*d = newDiagnosisCodeList(codes)
	return nil
}

// Outcome is the driver's own result envelope, reported on stdout by main.go.
//
// DiagnosisCodes, DeclineReason, and BestApplicability are additive
// omitempty siblings of the pre-existing DiagnosisCode field (D-14-42):
// Outcome carries no schema string, so this is a purely additive protocol
// change -- no existing field's type or JSON tag changes, and a
// repairable-path Outcome is byte-identical before and after this change
// since all three new fields stay unset on that path.
type Outcome struct {
	Status            string            `json:"status"`
	DiagnosisCode     string            `json:"diagnosis_code,omitempty"`
	DiagnosisCodes    diagnosisCodeList `json:"diagnosis_codes,omitempty"`
	DeclineReason     string            `json:"decline_reason,omitempty"`
	BestApplicability string            `json:"best_applicability,omitempty"`
	RepairKind        string            `json:"repair_kind,omitempty"`
	SubprocessCount   int               `json:"subprocess_count"`
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

// declineDecision is the classification computed while selectRepair walks
// the document looking for a driver-eligible repair and finds none (D-14-42).
// Before this change the false-return path discarded everything it had
// already computed; this struct is what the false path now carries out
// instead of throwing away.
type declineDecision struct {
	// diagnosisCode is the first diagnostic's code in document order --
	// deterministic and already specified by the document-order iteration
	// below, never an invented heuristic. Empty only when there are zero
	// diagnostics.
	diagnosisCode string
	// diagnosisCodes lists every diagnostic's code in document order --
	// the honest answer for the multi-diagnostic parse-failure path.
	diagnosisCodes []string
	// declineReason is one of the three closed Decline* constants.
	declineReason string
	// bestApplicability is the maximum applicability observed across all
	// repairs in the document, using the closed vocabulary re-declared
	// above. Left empty when declineReason is DeclineNoneOffered (no
	// repairs anywhere to observe an applicability from).
	bestApplicability string
}

// applicabilityRank orders the closed applicability vocabulary so
// bestApplicability can be computed as a running maximum -- MachineApplicable
// ranks highest, an empty/unrecognised value ranks lowest.
func applicabilityRank(a string) int {
	switch a {
	case applicabilityMachineApplicable:
		return 3
	case applicabilityRequiresConfirmation:
		return 2
	case applicabilityUnspecified:
		return 1
	default:
		return 0
	}
}

// maxApplicability returns whichever of a, b ranks higher under
// applicabilityRank, so accumulating it across every repair in the
// document yields the single best-observed value.
func maxApplicability(a, b string) string {
	if applicabilityRank(b) > applicabilityRank(a) {
		return b
	}
	return a
}

// selectRepair returns the first driver-eligible repair in document order
// (diagnostics order, then each diagnostic's own repairs order) -- the
// driver's specified, stable tie-break rule for a diagnostic that carries
// two equally applicable repairs (FND-04 ordering edge,
// TestRepairSelectionIsSpecifiedOnTies). Document order is itself
// deterministic on the wire: check.go emits diagnostics in traversal order,
// and diagnostic.ErrorWithRepairs canonically sorts each diagnostic's own
// repairs by Kind then Detail before they ever reach JSON.
//
// Signature deliberately left unchanged (D-14-42): both antitheater_test.go
// and repair_test.go call this exact three-value form at several sites, and
// those files must stay byte-unchanged by this task. classifyDecline below
// is the sibling that widens the false path instead -- Repair calls it only
// when selectRepair's own ok is false, so the classification computed on
// that path is carried out to the Outcome rather than discarded, without
// touching selectRepair's own return shape or its existing callers.
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

// classifyDecline walks the same document selectRepair just found no
// driver-eligible repair in, and reconstructs the classification that was
// previously computed inline and discarded on that path (D-14-42): every
// diagnostic code seen in document order, the closed-vocabulary decline
// reason, and the best applicability observed. Callers must only invoke
// this after selectRepair has already returned ok == false for the same
// result -- it does not itself re-check driver-eligibility as a fast exit,
// since the caller already knows none exists.
func classifyDecline(result checkResult) declineDecision {
	var codes []string
	anyRepairs := false
	bestApplicability := ""
	for _, d := range result.Diagnostics {
		codes = append(codes, d.Code)
		for _, r := range d.Repairs {
			anyRepairs = true
			bestApplicability = maxApplicability(bestApplicability, r.Applicability)
		}
	}
	decision := declineDecision{diagnosisCodes: codes}
	if len(codes) > 0 {
		decision.diagnosisCode = codes[0]
	}
	switch {
	case len(codes) == 0:
		// Invalid with zero diagnostics -- a fail-closed anomaly that must
		// still read as unrepairable, never a pass (T-14-17). There is no
		// diagnostic to name as the first element in document order, but
		// diagnosis_code must still never be empty on an unrepairable
		// outcome (D-14-42's own global truth): the decline reason itself
		// is the only honest non-empty value available here, and it is
		// still a closed-vocabulary code, never invented prose.
		decision.declineReason = DeclineNoDiagnostics
		decision.diagnosisCode = DeclineNoDiagnostics
	case !anyRepairs:
		decision.declineReason = DeclineNoneOffered
	default:
		decision.declineReason = DeclineNoneEligible
		decision.bestApplicability = bestApplicability
	}
	return decision
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
		// pass (FND-04 empty-input edge). The decline now carries the
		// classification classifyDecline computes instead of discarding it
		// (D-14-42): every value below is a code from a closed vocabulary
		// or read off an already-decoded jsonDiagnostic.Code -- no prose
		// field is decoded anywhere in this path (T-14-15).
		decision := classifyDecline(first)
		return Outcome{
			Status:            OutcomeUnrepairable,
			DiagnosisCode:     decision.diagnosisCode,
			DiagnosisCodes:    newDiagnosisCodeList(decision.diagnosisCodes),
			DeclineReason:     decision.declineReason,
			BestApplicability: decision.bestApplicability,
			SubprocessCount:   1,
		}, nil
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
