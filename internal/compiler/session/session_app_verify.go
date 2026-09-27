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

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

const (
	ReplayCasesSchema        = "lang.replay-cases/1"
	ApplicationVerifySchema  = "lang.app-verification/1"
	MaxReplayCasesBytes      = 64 * 1024
	MaxReplayCaseCount       = 16
	ReplayCaseSource         = "source"
	ReplayCaseVerifierModel  = "verifier_model_only"
	ReplayStatusPass         = "pass"
	ReplayStatusFail         = "fail"
	ReplayStatusUnsupported  = "unsupported"
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
	ModeledOutcome  *ReplayForeignOutcome          `json:"modeled_outcome,omitempty"`
	Diagnostic      string                         `json:"diagnostic,omitempty"`
	ActualHostIO    bool                           `json:"actual_host_io"`
	PhysicalCleanup bool                           `json:"physical_cleanup"`
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
	source, err := readBoundedFile(sourcePath, syntax.MaxSourceBytes)
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
	report, verifyErr := VerifyApplicationCases(ctx, source, cases, runner)
	if err := writeApplicationVerificationReport(reportPath, report); err != nil {
		return report, err
	}
	return report, verifyErr
}

// VerifyApplicationCases checks one source/core, drives independent explicit
// inputs through the interpreter and generated-C conformance shell at O0/O3,
// and compares all tiers to the independently authored expected value.
func VerifyApplicationCases(ctx context.Context, source []byte, cases ReplayCases, runner native.Runner) (ReplayReport, error) {
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
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return report, fmt.Errorf("source is not accepted by the checker: %s", checked.Diagnostics[0].Code)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return report, errors.New("source core validation failed")
	}
	program := validated.Program()
	inputSet := make([]string, 0, len(cases.Cases))
	indices := make([]int, 0, len(cases.Cases))
	interpreted := make(map[int]execution.Execution)
	for _, replayCase := range cases.Cases {
		caseReport := ReplayCaseReport{
			ID: replayCase.ID, Kind: replayCase.Kind, Status: ReplayStatusFail,
			Expected: replayCase.Expected, ActualHostIO: false, PhysicalCleanup: false,
		}
		if replayCase.Kind == ReplayCaseVerifierModel {
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
			entry, entryErr := callgraph.EntryFunction(program)
			if entryErr != nil || entry.Match != nil || entry.Linear == nil || entry.Parameter.Type != "U64" || entry.ReturnType != "U64" {
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
		o0, err := runner.Run(ctx, cSource, "-O0", validInputs)
		if err != nil || len(o0.Pairs) != len(validInputs) {
			markReplayFailures(&report, validIndices, "-O0 generated-C execution failed")
			return finalizeReplayReport(report), fmt.Errorf("run generated C at -O0: %w", replayErr(err, len(o0.Pairs), len(validInputs)))
		}
		o3, err := runner.Run(ctx, cSource, "-O3", validInputs)
		if err != nil || len(o3.Pairs) != len(validInputs) {
			markReplayFailures(&report, validIndices, "-O3 generated-C execution failed")
			return finalizeReplayReport(report), fmt.Errorf("run generated C at -O3: %w", replayErr(err, len(o3.Pairs), len(validInputs)))
		}
		for pairIndex, reportIndex := range validIndices {
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
		return fmt.Errorf("application verification report exceeds %d bytes", MaxReplayCasesBytes)
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".lang-app-verification-*")
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
