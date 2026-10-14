package session

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
)

const (
	ReplayCasesSchema        = "lang.replay-cases/1"
	ApplicationVerifySchema  = "lang.app-verification/2"
	MaxReplayCasesBytes      = 64 * 1024
	MaxReplayCaseCount       = 16
	MaxReplayCaseIDBytes     = 128
	ReplayCaseSource         = "source"
	ReplayCaseVerifierModel  = "verifier_model_only"
	ReplayStatusPass         = "pass"
	ReplayStatusFail         = "fail"
	ReplayStatusUnsupported  = "unsupported"
	ReplayEvidenceSource     = "source_replay_conformance"
	ReplayEvidenceModel      = "verifier_model_only"
	verifierModelOperation   = "fixture.value"
	verifierModelOutcomeType = "U64"
)

// ReplayCases is the closed, versioned input contract for the explicit
// application verifier. Expected values are authored independently of every
// execution tier.
type ReplayCases struct {
	Schema string       `json:"schema"`
	Cases  []ReplayCase `json:"cases"`
}

type ReplayCase struct {
	ID              string                 `json:"id"`
	Kind            string                 `json:"kind"`
	Input           string                 `json:"input,omitempty"`
	Operation       string                 `json:"operation,omitempty"`
	Expected        ReplayExpected         `json:"expected"`
	ForeignOutcomes []ReplayForeignOutcome `json:"foreign_outcomes"`
}

type ReplayExpected struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// ReplayForeignOutcome is only consumed by the verifier-owned model adapter.
// It never represents a Lang foreign call or host operation.
type ReplayForeignOutcome struct {
	Operation string `json:"operation"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

type ReplayCaseReport struct {
	ID              string                         `json:"id"`
	Kind            string                         `json:"kind"`
	Status          string                         `json:"status"`
	InputID         string                         `json:"input_id,omitempty"`
	Expected        ReplayExpected                 `json:"expected"`
	Engines         map[string]execution.Execution `json:"engines,omitempty"`
	OutcomeDigest   *ReplayOutcomeDigest           `json:"outcome_digest,omitempty"`
	ModeledOutcome  *ReplayForeignOutcome          `json:"modeled_outcome,omitempty"`
	EvidenceScope   string                         `json:"evidence_scope"`
	Diagnostic      string                         `json:"diagnostic,omitempty"`
	ActualHostIO    bool                           `json:"actual_host_io"`
	PhysicalCleanup bool                           `json:"physical_cleanup"`
}

// ReplayOutcomeDigest replaces embedded interpreter evidence for cases whose
// generated C exhausted its static event capacity (D-13). It is fixed-size, so
// the report no longer grows with the program's event count. EvidenceSHA256 is
// the lowercase hex SHA-256 of the exact json.Marshal bytes of the omitted
// schema-2 interpreter execution and EvidenceBytes is their length.
type ReplayOutcomeDigest struct {
	Engine         string `json:"engine"`
	OutcomeKind    string `json:"outcome_kind"`
	OutcomeValue   string `json:"outcome_value"`
	EventCount     int    `json:"event_count"`
	EvidenceSHA256 string `json:"evidence_sha256"`
	EvidenceBytes  int    `json:"evidence_bytes"`
	ExhaustedLevel string `json:"exhausted_level"`
}

type ReplayReport struct {
	Schema            string             `json:"schema"`
	Status            string             `json:"status"`
	Verified          bool               `json:"verified"`
	SourceDigest      string             `json:"source_digest"`
	EmittedCDigest    string             `json:"emitted_c_digest,omitempty"`
	BuildID           string             `json:"build_id,omitempty"`
	InputSetID        string             `json:"input_set_id"`
	DependencyClosure string             `json:"dependency_closure"`
	Cacheable         bool               `json:"cacheable"`
	ControlledWorld   string             `json:"controlled_world"`
	ActualHostIO      bool               `json:"actual_host_io"`
	PhysicalCleanup   bool               `json:"physical_cleanup"`
	Cases             []ReplayCaseReport `json:"cases"`
}

// ErrApplicationVerificationReportOversize reports that the encoded report
// would exceed MaxReplayCasesBytes. It is defense in depth behind the static
// budget proven by the report-budget test; no file is written when it fires.
var ErrApplicationVerificationReportOversize = errors.New("application verification report exceeds size limit")

// DecodeReplayCases accepts only one bounded JSON value with no unknown
// fields. Missing foreign_outcomes stays nil and is rejected by the case
// validator; ordinary source cases must explicitly declare an empty script.
func DecodeReplayCases(data []byte) (ReplayCases, error) {
	if len(data) == 0 || len(data) > MaxReplayCasesBytes {
		return ReplayCases{}, fmt.Errorf("replay cases must be between 1 and %d bytes", MaxReplayCasesBytes)
	}
	if err := native.RejectDuplicateJSONKeys(data); err != nil {
		return ReplayCases{}, fmt.Errorf("decode replay cases: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var cases ReplayCases
	if err := decoder.Decode(&cases); err != nil {
		return ReplayCases{}, fmt.Errorf("decode replay cases: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ReplayCases{}, errors.New("replay cases contain more than one JSON value")
		}
		return ReplayCases{}, fmt.Errorf("read replay cases trailer: %w", err)
	}
	if cases.Schema != ReplayCasesSchema {
		return ReplayCases{}, fmt.Errorf("replay cases schema must be %q", ReplayCasesSchema)
	}
	if len(cases.Cases) == 0 || len(cases.Cases) > MaxReplayCaseCount {
		return ReplayCases{}, fmt.Errorf("replay case count must be between 1 and %d", MaxReplayCaseCount)
	}
	seen := make(map[string]struct{}, len(cases.Cases))
	for index, replayCase := range cases.Cases {
		if strings.TrimSpace(replayCase.ID) == "" {
			return ReplayCases{}, fmt.Errorf("case %d has an empty id", index)
		}
		if len(replayCase.ID) > MaxReplayCaseIDBytes {
			return ReplayCases{}, fmt.Errorf("case %d id exceeds %d bytes (MaxReplayCaseIDBytes)", index, MaxReplayCaseIDBytes)
		}
		for position := 0; position < len(replayCase.ID); position++ {
			if replayCase.ID[position] < 0x20 || replayCase.ID[position] > 0x7e {
				return ReplayCases{}, fmt.Errorf("case %d id must be printable ASCII (byte %d is not)", index, position)
			}
		}
		if _, exists := seen[replayCase.ID]; exists {
			return ReplayCases{}, fmt.Errorf("duplicate replay case id %q", replayCase.ID)
		}
		seen[replayCase.ID] = struct{}{}
		if replayCase.ForeignOutcomes == nil {
			return ReplayCases{}, fmt.Errorf("case %q must declare foreign_outcomes as an array", replayCase.ID)
		}
		if replayCase.Expected.Kind != execution.OutcomeReturned || !canonicalU64(replayCase.Expected.Value) {
			return ReplayCases{}, fmt.Errorf("case %q expected outcome must be a canonical U64 value", replayCase.ID)
		}
		switch replayCase.Kind {
		case ReplayCaseSource:
			if replayCase.Operation != "" {
				return ReplayCases{}, fmt.Errorf("source case %q cannot declare a modeled operation", replayCase.ID)
			}
			if len(replayCase.Input) == 0 || len(replayCase.Input) > native.MaxApplicationArgumentBytes || !canonicalU64(replayCase.Input) {
				return ReplayCases{}, fmt.Errorf("source case %q input must be a canonical U64 within the application input limit", replayCase.ID)
			}
		case ReplayCaseVerifierModel:
			if replayCase.Input != "" {
				return ReplayCases{}, fmt.Errorf("verifier-model case %q cannot declare a source input", replayCase.ID)
			}
		case "":
			return ReplayCases{}, fmt.Errorf("case %q must declare a kind", replayCase.ID)
		default:
			return ReplayCases{}, fmt.Errorf("case %q has unsupported kind %q", replayCase.ID, replayCase.Kind)
		}
	}
	return cases, nil
}

// VerifyApplicationCasesFile reads bounded source and case inputs, performs
// explicit differential verification, then atomically publishes its report.
// A semantic mismatch is represented in the report and is not a tool error;
// setup, toolchain, and file errors are returned to the CLI.
func VerifyApplicationCasesFile(ctx context.Context, sourcePath, casesPath, reportPath string, runner native.Runner) (ReplayReport, error) {
	if strings.TrimSpace(reportPath) == "" {
		return ReplayReport{}, errors.New("an application verification report path is required")
	}
	// A report left by an earlier run must never outlive a run that refuses,
	// fails, or overflows before publishing its own report: a consumer that
	// reads only the file would otherwise see a stale pass.
	if err := removeStaleVerificationReport(reportPath, sourcePath, casesPath); err != nil {
		return ReplayReport{}, err
	}
	checked, source, err := checkSourcePath(sourcePath)
	if err != nil {
		return ReplayReport{}, fmt.Errorf("read source: %w", err)
	}
	caseBytes, err := readBoundedFile(casesPath, MaxReplayCasesBytes)
	if err != nil {
		return ReplayReport{}, fmt.Errorf("read replay cases: %w", err)
	}
	cases, err := DecodeReplayCases(caseBytes)
	if err != nil {
		return ReplayReport{}, err
	}
	report, verifyErr := verifyApplicationCases(ctx, source, &checked, cases, runner)
	if err := writeApplicationVerificationReport(reportPath, report); err != nil {
		// Keep the verification error visible (errors.As still finds a
		// *SourceRefusalError, so the CLI keeps exit 2) alongside the
		// report-write failure.
		if verifyErr != nil {
			return report, errors.Join(verifyErr, err)
		}
		return report, err
	}
	return report, verifyErr
}

// removeStaleVerificationReport deletes a pre-existing report file. It never
// removes an input: a report path that names the source or cases file is left
// for the later atomic publish to overwrite.
func removeStaleVerificationReport(reportPath, sourcePath, casesPath string) error {
	info, err := os.Lstat(reportPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("inspect stale verification report: %w", err)
	}
	for _, input := range []string{sourcePath, casesPath} {
		if inputInfo, statErr := os.Stat(input); statErr == nil && os.SameFile(info, inputInfo) {
			return nil
		}
	}
	if err := os.Remove(reportPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove stale verification report: %w", err)
	}
	return nil
}

// SourceRefusalError reports that the source was refused by the checker, an
// admission layer or entry resolution (Phase 28 D-21). The CLI maps it to
// exit 2 (exitcode.Invalid); capacity and tool outcomes stay exit 3.
type SourceRefusalError struct {
	Diagnostics  []diagnostic.Diagnostic
	Presentation *protocol.PresentationContext
}

// Code returns the first refusing diagnostic code.
func (e *SourceRefusalError) Code() string {
	if e == nil || len(e.Diagnostics) == 0 {
		return ""
	}
	return e.Diagnostics[0].Code
}

func (e *SourceRefusalError) Error() string { return "source refused: " + e.Code() }

// VerifyApplicationCases checks one source/core, drives independent explicit
// inputs through the interpreter and generated-C conformance shell at O0/O3,
// and compares all tiers to the independently authored expected value.
func VerifyApplicationCases(ctx context.Context, source []byte, cases ReplayCases, runner native.Runner) (ReplayReport, error) {
	return verifyApplicationCases(ctx, source, nil, cases, runner)
}

// verifyApplicationCases is VerifyApplicationCases over an optional
// pre-computed check: the file route hands in the module-aware result, the
// byte-taking API passes nil and checks source itself. source is the ENTRY
// bytes either way (the report SourceDigest stays the entry digest).
func verifyApplicationCases(ctx context.Context, source []byte, precomputed *CheckResult, cases ReplayCases, runner native.Runner) (ReplayReport, error) {
	sourceDigest := replayDigest(source)
	caseBytes, _ := json.Marshal(cases)
	report := ReplayReport{
		Schema: ApplicationVerifySchema, Status: ReplayStatusFail,
		SourceDigest: sourceDigest, InputSetID: replayDigest(caseBytes),
		DependencyClosure: "incomplete", Cacheable: false,
		ControlledWorld: "local interpreter and generated-C conformance; verifier-model outcomes are scripted",
		ActualHostIO:    false, PhysicalCleanup: false,
		Cases: make([]ReplayCaseReport, 0, len(cases.Cases)),
	}
	if cases.Schema != ReplayCasesSchema || len(cases.Cases) == 0 || len(cases.Cases) > MaxReplayCaseCount {
		return report, errors.New("invalid replay case set")
	}
	validatedCases, err := DecodeReplayCases(caseBytes)
	if err != nil {
		return report, err
	}
	cases = validatedCases
	// D-21: the same Check -> admitProgram -> resolveEntry steps as every other
	// command; any source refusal is a typed *SourceRefusalError (exit 2). The
	// fail report is still returned (and written by the file route).
	var checked CheckResult
	if precomputed != nil {
		checked = *precomputed
	} else {
		checked = Check(source)
	}
	if len(checked.Diagnostics) > 0 {
		return report, &SourceRefusalError{Diagnostics: checked.Diagnostics, Presentation: checked.Presentation}
	}
	admitted, admissionDiagnostics := admitProgram(checked)
	if len(admissionDiagnostics) > 0 {
		return report, &SourceRefusalError{Diagnostics: admissionDiagnostics, Presentation: checked.Presentation}
	}
	// D-04/D-18: one entry resolution after admission and before the per-case
	// loop; every engine below sees the pruned program.
	entry, prunedProgram, entryDiagnostics, entryErr := resolveEntry(admitted)
	if len(entryDiagnostics) > 0 {
		return report, &SourceRefusalError{Diagnostics: entryDiagnostics, Presentation: checked.Presentation}
	}
	// A non-diagnostic resolution failure is a tool error on every lane; it
	// must never degrade to per-case "unsupported" results (which a
	// verifier-model-only case set would then never consult).
	if entryErr != nil {
		return report, entryErr
	}
	program := prunedProgram
	inputSet := make([]string, 0, len(cases.Cases))
	indices := make([]int, 0, len(cases.Cases))
	interpreted := make(map[int]execution.Execution)
	for _, replayCase := range cases.Cases {
		caseReport := ReplayCaseReport{
			ID: replayCase.ID, Kind: replayCase.Kind, Status: ReplayStatusFail,
			Expected: replayCase.Expected, ActualHostIO: false, PhysicalCleanup: false,
			EvidenceScope: ReplayEvidenceSource,
		}
		if replayCase.Kind == ReplayCaseVerifierModel {
			caseReport.EvidenceScope = ReplayEvidenceModel
			caseReport = verifyVerifierModelCase(caseReport, replayCase)
			report.Cases = append(report.Cases, caseReport)
			continue
		}
		caseReport.InputID = replayDigest([]byte(replayCase.Input))
		switch {
		case replayCase.Kind != ReplayCaseSource:
			caseReport.Status = ReplayStatusUnsupported
			caseReport.Diagnostic = "unsupported replay case kind"
		case replayCase.ForeignOutcomes == nil:
			caseReport.Status = ReplayStatusUnsupported
			caseReport.Diagnostic = "source replay must explicitly declare an empty foreign_outcomes array"
		case len(replayCase.ForeignOutcomes) != 0:
			caseReport.Status = ReplayStatusUnsupported
			caseReport.Diagnostic = "source replay does not consume foreign outcomes; host C is not executed"
		case replayCase.Expected.Kind != execution.OutcomeReturned || !canonicalU64(replayCase.Expected.Value):
			caseReport.Status = ReplayStatusFail
			caseReport.Diagnostic = "expected outcome is not a canonical U64 value"
		case len(replayCase.Input) == 0 || len(replayCase.Input) > native.MaxApplicationArgumentBytes || !canonicalU64(replayCase.Input):
			caseReport.Status = ReplayStatusFail
			caseReport.Diagnostic = "input is not a canonical U64 within the application input limit"
		default:
			inputSet = append(inputSet, replayCase.Input)
			indices = append(indices, len(report.Cases))
			if entry.Match != nil || entry.Linear == nil || entry.Parameter.Type != "U64" || entry.ReturnType != "U64" {
				caseReport.Status = ReplayStatusUnsupported
				caseReport.Diagnostic = "source replay requires a linear U64-to-U64 entry"
			} else if hasForeignContract(program) {
				caseReport.Status = ReplayStatusUnsupported
				caseReport.Diagnostic = "source replay refuses Lang foreign contracts; no local C is executed"
			} else if len(runner.ForeignSources) != 0 {
				caseReport.Status = ReplayStatusUnsupported
				caseReport.Diagnostic = "source replay refuses configured local C inputs"
			} else {
				actual, runErr := interp.Run(program, entry.Name, replayCase.Input)
				if runErr != nil {
					caseReport.Status = ReplayStatusFail
					caseReport.Diagnostic = "interpreter did not complete the isolated input"
				} else {
					actual, runErr = ProjectExecutionSchema2(program, actual)
					if runErr != nil {
						caseReport.Status = ReplayStatusFail
						caseReport.Diagnostic = "interpreter evidence could not be projected to execution schema 2"
					} else {
						interpreted[len(report.Cases)] = actual
					}
				}
			}
		}
		report.Cases = append(report.Cases, caseReport)
	}

	// Invalid/unsupported ordinary cases are reported individually and never
	// acquire an interpreter or native execution. Valid source cases share one
	// checked program and one generated C body, with a private input per pair.
	validIndices := make([]int, 0, len(indices))
	validInputs := make([]string, 0, len(indices))
	for n, index := range indices {
		if _, ok := interpreted[index]; ok {
			validIndices = append(validIndices, index)
			validInputs = append(validInputs, inputSet[n])
		}
	}
	if len(validIndices) != 0 {
		cSource, err := Phase16ControlNativeC(program, "app-verify")
		if err != nil {
			markReplayFailures(&report, validIndices, "generated-C emission failed")
			return finalizeReplayReport(report), fmt.Errorf("emit conformance program: %w", err)
		}
		report.EmittedCDigest = replayDigest([]byte(cSource))
		buildIdentity, _ := json.Marshal(struct {
			Source   string
			CSource  string
			Compiler string
			Target   string
			Levels   []string
		}{sourceDigest, report.EmittedCDigest, effectiveClangPath(runner), runtime.GOOS + "/" + runtime.GOARCH, []string{"-O0", "-O3"}})
		report.BuildID = replayDigest(buildIdentity)
		runner.ForeignSources = nil
		runner.BuildCache = nil
		runner.Expect = native.ExpectValue
		o0, err := runAppVerifyLevel(ctx, runner, cSource, "-O0", validInputs)
		if err != nil {
			markReplayFailures(&report, validIndices, "-O0 generated-C execution failed")
			return finalizeReplayReport(report), fmt.Errorf("run generated C at -O0: %w", err)
		}
		// Capacity is static and level-independent, so both levels must agree on
		// exhaustion. Run every input at -O3 even when -O0 exhausted, so a
		// one-level exhaustion is detected instead of masked.
		o3, err := runAppVerifyLevel(ctx, runner, cSource, "-O3", validInputs)
		if err != nil {
			markReplayFailures(&report, validIndices, "-O3 generated-C execution failed")
			return finalizeReplayReport(report), fmt.Errorf("run generated C at -O3: %w", err)
		}
		for pairIndex, reportIndex := range validIndices {
			if o0.Exhausted[pairIndex] != o3.Exhausted[pairIndex] {
				level := "-O3"
				if o0.Exhausted[pairIndex] {
					level = "-O0"
				}
				digest, digestErr := replayOutcomeDigest(interpreted[reportIndex], level)
				if digestErr != nil {
					report.Cases[reportIndex].Status = ReplayStatusFail
					report.Cases[reportIndex].Diagnostic = "interpreter evidence could not be digested"
					continue
				}
				report.Cases[reportIndex].OutcomeDigest = digest
				report.Cases[reportIndex].Status = ReplayStatusFail
				report.Cases[reportIndex].Diagnostic = fmt.Sprintf("event capacity exhausted at %s only; the other optimization level completed, so the levels diverge", level)
				continue
			}
			if o0.Exhausted[pairIndex] {
				level := "-O0"
				// An exhausted case has no native evidence: only the interpreter
				// engine is reported, and unsupported is claimed only when the
				// interpreter outcome already equals the expected outcome.
				actual := interpreted[reportIndex].Outcome
				expected := report.Cases[reportIndex].Expected
				digest, digestErr := replayOutcomeDigest(interpreted[reportIndex], level)
				if digestErr != nil {
					report.Cases[reportIndex].Status = ReplayStatusFail
					report.Cases[reportIndex].Diagnostic = "interpreter evidence could not be digested"
					continue
				}
				report.Cases[reportIndex].OutcomeDigest = digest
				if actual.Kind == expected.Kind && actual.Value == expected.Value {
					report.Cases[reportIndex].Status = ReplayStatusUnsupported
					report.Cases[reportIndex].Diagnostic = eventCapacityDiagnostic(level)
				} else {
					report.Cases[reportIndex].Status = ReplayStatusFail
					report.Cases[reportIndex].Diagnostic = fmt.Sprintf("interpreter produced %s:%s, expected %s:%s", actual.Kind, actual.Value, expected.Kind, expected.Value)
				}
				continue
			}
			o0Execution, projectionErr := ProjectExecutionSchema2(program, o0.Pairs[pairIndex].Execution)
			if projectionErr != nil {
				report.Cases[reportIndex].Diagnostic = "-O0 evidence could not be projected to execution schema 2"
				continue
			}
			o3Execution, projectionErr := ProjectExecutionSchema2(program, o3.Pairs[pairIndex].Execution)
			if projectionErr != nil {
				report.Cases[reportIndex].Diagnostic = "-O3 evidence could not be projected to execution schema 2"
				continue
			}
			engineResults := map[string]execution.Execution{
				"interpreter": interpreted[reportIndex], "O0": o0Execution, "O3": o3Execution,
			}
			report.Cases[reportIndex].Engines = engineResults
			compareErr := Phase5CompareProgramEngines("phase22-app-verify-"+report.Cases[reportIndex].ID, program, engineResults)
			expected := report.Cases[reportIndex].Expected
			for _, name := range []string{"interpreter", "O0", "O3"} {
				actual := engineResults[name].Outcome
				if actual.Kind != expected.Kind || actual.Value != expected.Value {
					if compareErr == nil {
						compareErr = fmt.Errorf("%s produced %s:%s, expected %s:%s", name, actual.Kind, actual.Value, expected.Kind, expected.Value)
					}
				}
			}
			if compareErr != nil {
				report.Cases[reportIndex].Status = ReplayStatusFail
				report.Cases[reportIndex].Diagnostic = compareErr.Error()
			} else {
				report.Cases[reportIndex].Status = ReplayStatusPass
			}
		}
	}
	return finalizeReplayReport(report), nil
}

func verifyVerifierModelCase(report ReplayCaseReport, replayCase ReplayCase) ReplayCaseReport {
	if replayCase.Operation != verifierModelOperation {
		report.Status = ReplayStatusUnsupported
		report.Diagnostic = "verifier-model operation is not supported"
		return report
	}
	if len(replayCase.ForeignOutcomes) == 0 {
		report.Diagnostic = "modeled outcome is missing"
		return report
	}
	var match *ReplayForeignOutcome
	matches := 0
	for index := range replayCase.ForeignOutcomes {
		outcome := &replayCase.ForeignOutcomes[index]
		if outcome.Operation == verifierModelOperation {
			matches++
			match = outcome
		}
	}
	if matches > 1 {
		report.Diagnostic = "modeled outcome is duplicated"
		return report
	}
	if matches == 0 {
		report.Diagnostic = "modeled outcome for fixture.value is missing"
		return report
	}
	if len(replayCase.ForeignOutcomes) != 1 {
		report.Diagnostic = "declared modeled outcome was not consumed"
		return report
	}
	report.ModeledOutcome = &ReplayForeignOutcome{Operation: match.Operation, Type: match.Type, Value: match.Value}
	if match.Type != verifierModelOutcomeType || !canonicalU64(match.Value) {
		report.Diagnostic = "modeled fixture.value outcome must be one canonical U64"
		return report
	}
	if replayCase.Expected.Kind != execution.OutcomeReturned || replayCase.Expected.Value != match.Value {
		report.Diagnostic = fmt.Sprintf("modeled fixture.value returned %s:%s, expected %s:%s", match.Type, match.Value, replayCase.Expected.Kind, replayCase.Expected.Value)
		return report
	}
	report.Status = ReplayStatusPass
	return report
}

func hasForeignContract(program core.Program) bool {
	for _, function := range program.Functions {
		if function.ForeignContract != nil {
			return true
		}
	}
	return false
}

func canonicalU64(value string) bool {
	parsed, err := strconv.ParseUint(value, 10, 64)
	return err == nil && strconv.FormatUint(parsed, 10) == value
}

// replayOutcomeDigest summarizes an omitted schema-2 interpreter execution. It
// fails closed: a marshal error is returned so the caller marks the case fail.
func replayOutcomeDigest(document execution.Execution, level string) (*ReplayOutcomeDigest, error) {
	raw, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	return &ReplayOutcomeDigest{
		Engine: "interpreter", OutcomeKind: document.Outcome.Kind, OutcomeValue: document.Outcome.Value,
		EventCount: len(document.Events), EvidenceSHA256: replayDigest(raw), EvidenceBytes: len(raw),
		ExhaustedLevel: level,
	}, nil
}

func replayDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func effectiveClangPath(runner native.Runner) string {
	if runner.ClangPath == "" {
		return "clang"
	}
	return runner.ClangPath
}

// appVerifyLevel holds one optimization level's per-input results. A Pair is
// the zero value where the matching Exhausted entry is true.
type appVerifyLevel struct {
	Pairs     []native.Pair
	Exhausted []bool
}

func isEventCapacityExhausted(err error) bool {
	var toolErr *native.ToolError
	return errors.As(err, &toolErr) && toolErr.Code == native.CodeEventCapacityExhausted
}

func eventCapacityDiagnostic(level string) string {
	return native.CodeEventCapacityExhausted + ": generated C exceeded its static event capacity at " + level + "; outcome not established, not a mismatch"
}

// runAppVerifyLevel runs one batch per level. Only when the batch fails with
// the named capacity code does it re-run each input alone to recover per-input
// results; any other error stays a whole-set operational error.
//
// Cost of the recovery path: runner.BuildCache is nil in app verify, and
// native.Runner.Run compiles its C source on every call. After a batch capacity
// failure, each input therefore costs one extra clang compile and one extra
// subprocess timeout budget at this level. For n inputs that is n+1 compiles
// instead of 1. The cost is paid only after the named capacity code, and app
// verify caps cases at MaxReplayCaseCount (16), so the path stays a fresh
// compile on purpose. Reusing one binary would need a new native.Runner API.
// Revisit if app verify gains larger case sets.
func runAppVerifyLevel(ctx context.Context, runner native.Runner, cSource, level string, inputs []string) (appVerifyLevel, error) {
	if len(inputs) == 0 {
		return appVerifyLevel{}, nil
	}
	result, err := runner.Run(ctx, cSource, level, inputs)
	if err == nil && len(result.Pairs) == len(inputs) {
		return appVerifyLevel{Pairs: result.Pairs, Exhausted: make([]bool, len(inputs))}, nil
	}
	if !isEventCapacityExhausted(err) {
		return appVerifyLevel{}, replayErr(err, len(result.Pairs), len(inputs))
	}
	recovered := appVerifyLevel{Pairs: make([]native.Pair, len(inputs)), Exhausted: make([]bool, len(inputs))}
	for index, input := range inputs {
		single, singleErr := runner.Run(ctx, cSource, level, []string{input})
		if singleErr == nil && len(single.Pairs) == 1 {
			recovered.Pairs[index] = single.Pairs[0]
			continue
		}
		if isEventCapacityExhausted(singleErr) {
			recovered.Exhausted[index] = true
			continue
		}
		return appVerifyLevel{}, replayErr(singleErr, len(single.Pairs), 1)
	}
	return recovered, nil
}

func replayErr(err error, got, want int) error {
	if err != nil {
		return err
	}
	return fmt.Errorf("received %d execution documents, want %d", got, want)
}

func markReplayFailures(report *ReplayReport, indices []int, message string) {
	for _, index := range indices {
		report.Cases[index].Status = ReplayStatusFail
		report.Cases[index].Diagnostic = message
	}
}

func finalizeReplayReport(report ReplayReport) ReplayReport {
	report.Verified = len(report.Cases) != 0
	for _, replayCase := range report.Cases {
		if replayCase.Status != ReplayStatusPass {
			report.Verified = false
			break
		}
	}
	if report.Verified {
		report.Status = ReplayStatusPass
	} else {
		report.Status = ReplayStatusFail
	}
	return report
}

func writeApplicationVerificationReport(path string, report ReplayReport) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("application verification report path is empty")
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("encode application verification report: %w", err)
	}
	if len(encoded)+1 > MaxReplayCasesBytes {
		return fmt.Errorf("%w: %d bytes > %d", ErrApplicationVerificationReportOversize, len(encoded)+1, MaxReplayCasesBytes)
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".schway-app-verification-*")
	if err != nil {
		return fmt.Errorf("create verification report: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("set verification report permissions: %w", err)
	}
	if _, err := temporary.Write(append(encoded, '\n')); err != nil {
		temporary.Close()
		return fmt.Errorf("write verification report: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync verification report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close verification report: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("publish verification report: %w", err)
	}
	return nil
}
