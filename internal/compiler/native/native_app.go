package native

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/szTheory/schway/internal/compiler/execution"
)

const (
	ApplicationBuildSchema      = "schway.app-build/1"
	ApplicationEvidenceSchema   = "schway.app-evidence/1"
	applicationCaptureSchema    = "schway.app-capture/1"
	MaxApplicationArgumentBytes = 4096
	maxApplicationReceiptBytes  = 64 * 1024
)

type BuildReceipt struct {
	BuildID             string               `json:"build_id,omitempty"`
	InputID             string               `json:"input_id,omitempty"`
	Identity            *BuildIdentityInputs `json:"identity,omitempty"`
	Bindings            *BindingManifest     `json:"bindings,omitempty"`
	OutputPath          string               `json:"output_path,omitempty"`
	Schema              string               `json:"schema"`
	ExecutableDigest    string               `json:"executable_digest"`
	SourceDigest        string               `json:"source_digest"`
	EmittedCDigest      string               `json:"emitted_c_digest"`
	Compiler            string               `json:"compiler"`
	CompilerVersion     string               `json:"compiler_version"`
	Target              string               `json:"target"`
	Flags               []string             `json:"flags"`
	RuntimeDependencies []string             `json:"runtime_dependencies"`
	DependencyClosure   string               `json:"dependency_closure"`
	Cacheable           bool                 `json:"cacheable"`
}

type RunOutcomeKind string

const (
	RunExited      RunOutcomeKind = "exit"
	RunSignaled    RunOutcomeKind = "signal"
	RunTimedOut    RunOutcomeKind = "timeout"
	RunLaunchError RunOutcomeKind = "launch_error"
)

// RunOutcome keeps child process results separate from tool and receipt errors.
// ExitCode is meaningful for RunExited; Signal and SignalNumber for
// RunSignaled; Diagnostic for timeout and launch-error outcomes.
type RunOutcome struct {
	Kind         RunOutcomeKind `json:"kind"`
	ExitCode     int            `json:"exit_code,omitempty"`
	Signal       string         `json:"signal,omitempty"`
	SignalNumber int            `json:"signal_number,omitempty"`
	Diagnostic   string         `json:"diagnostic,omitempty"`
}

type EvidenceMode string

const (
	EvidenceDisabled EvidenceMode = "disabled"
	EvidenceEvents   EvidenceMode = "events"
)

type EvidenceCaptureStatus string

const (
	EvidenceStatusDisabled          EvidenceCaptureStatus = "disabled"
	EvidenceStatusComplete          EvidenceCaptureStatus = "complete"
	EvidenceStatusIncomplete        EvidenceCaptureStatus = "incomplete"
	EvidenceStatusCapacityExhausted EvidenceCaptureStatus = "capacity_exhausted"
)

// ApplicationEvidenceReport is a status-bearing record of one app process.
// A complete capture records structure and identity; it is never a differential
// semantic verdict.
type ApplicationEvidenceReport struct {
	Schema         string                `json:"schema"`
	BuildID        string                `json:"build_id"`
	InputID        string                `json:"input_id"`
	ArtifactDigest string                `json:"artifact_digest"`
	SourceDigest   string                `json:"source_digest"`
	InputDigest    string                `json:"input_digest"`
	ProcessOutcome RunOutcome            `json:"process_outcome"`
	CaptureStatus  EvidenceCaptureStatus `json:"capture_status"`
	Execution      *execution.Execution  `json:"execution,omitempty"`
	Verified       bool                  `json:"verified"`
}

type applicationCapture struct {
	Schema    string          `json:"schema"`
	Status    string          `json:"status"`
	Execution json.RawMessage `json:"execution,omitempty"`
}

// evidenceLimit is a test seam for capacity exhaustion controls. Production
// reports use the fixed execution.MaxApplicationEvidenceBytes limit.
func (r Runner) evidenceCapacityLimit() int {
	if r.evidenceLimit > 0 {
		return r.evidenceLimit
	}
	return execution.MaxApplicationEvidenceBytes
}

var applicationFlags = []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O0"}

func applicationReceiptPath(artifactPath string) string {
	return artifactPath + ".schway-build.json"
}

// BuildApplication compiles the generated translation unit and publishes a
// retained executable plus an adjacent content-bound receipt. It never starts
// the resulting application.
func BuildApplication(ctx context.Context, source []byte, cSource, outputPath string, manifestPath ...string) (BuildReceipt, error) {
	return DefaultRunner().BuildApplication(ctx, source, cSource, outputPath, manifestPath...)
}

func (r Runner) BuildApplication(parent context.Context, source []byte, cSource, outputPath string, manifestPath ...string) (BuildReceipt, error) {
	if r.ClangPath == "" {
		r.ClangPath = "clang"
	}
	if r.Timeout <= 0 {
		r.Timeout = defaultSubprocessTimeout
	}
	if len(manifestPath) > 1 || (len(manifestPath) == 1 && manifestPath[0] == "") {
		return BuildReceipt{}, bindingError("expected one nonempty manifest path")
	}
	var bindings *ResolvedBindings
	if len(manifestPath) == 1 {
		resolved, err := ResolveBindings(manifestPath[0])
		if err != nil {
			return BuildReceipt{}, err
		}
		bindings = &resolved
	}
	if strings.TrimSpace(outputPath) == "" {
		return BuildReceipt{}, &ToolError{Code: "native.output_required", Err: errors.New("an output artifact path is required")}
	}
	artifactPath, err := filepath.Abs(outputPath)
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.output_invalid", Err: err}
	}
	outputDirectory := filepath.Dir(artifactPath)
	if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.output_directory_failed", Err: err}
	}
	if info, statErr := os.Stat(artifactPath); statErr == nil && info.IsDir() {
		return BuildReceipt{}, &ToolError{Code: "native.output_invalid", Err: errors.New("output artifact path names a directory")}
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return BuildReceipt{}, &ToolError{Code: "native.output_invalid", Err: statErr}
	}
	clang, err := exec.LookPath(r.ClangPath)
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.tool_missing", Err: err}
	}
	clang, err = filepath.Abs(clang)
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.tool_missing", Err: err}
	}
	version, ok := r.probe(parent, clang, "--version")
	if !ok || strings.TrimSpace(version) == "" {
		return BuildReceipt{}, &ToolError{Code: "native.compiler_probe_failed", Err: errors.New("unable to determine compiler version")}
	}
	target, ok := r.probe(parent, clang, "-dumpmachine")
	if !ok || strings.TrimSpace(target) == "" {
		return BuildReceipt{}, &ToolError{Code: "native.compiler_probe_failed", Err: errors.New("unable to determine compiler target")}
	}
	compilerDigest, err := digestFile(clang)
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.compiler_probe_failed", Err: err}
	}

	stageDirectory, err := os.MkdirTemp(outputDirectory, ".schway-build-")
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.temp_failed", Err: err}
	}
	preserveStage := false
	defer func() {
		if !preserveStage {
			_ = os.RemoveAll(stageDirectory)
		}
	}()
	// EvalSymlinks keeps dependency comparisons consistent on /var -> /private/var.
	stageDirectory, err = filepath.EvalSymlinks(stageDirectory)
	if err != nil {
		return BuildReceipt{}, err
	}
	sourcePath := filepath.Join(stageDirectory, "program.c")
	stagedArtifact := filepath.Join(stageDirectory, "program")
	if err := os.WriteFile(sourcePath, []byte(cSource), 0o600); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.temp_failed", Err: err}
	}
	identity := BuildIdentityInputs{
		SourceDigest: digestBytes(source), EmittedCDigest: digestBytes([]byte(cSource)),
		CompilerDigest: compilerDigest, CompilerVersion: strings.TrimSpace(version), Target: strings.TrimSpace(target),
		HostABI: runtime.GOOS + "/" + runtime.GOARCH, Flags: append([]string(nil), applicationFlags...),
		RuntimeDependencies: []string{"platform-c-runtime"},
	}
	var objects []string
	if bindings != nil {
		var err error
		objects, identity.Commands, identity.IncludedHeaders, identity.ABIProbeDigest, err = r.compileBindings(parent, clang, stageDirectory, *bindings)
		if err != nil {
			return BuildReceipt{}, err
		}
		identity.ManifestDigest = bindings.ManifestDigest
		identity.Inventory = bindings.Inventory
		identity.Symbols = bindings.Manifest.Symbols
		identity.RuntimeDependencies = bindings.Manifest.RuntimeDependencies
	}
	arguments := append([]string(nil), applicationFlags...)
	arguments = append(arguments, "program.c")
	arguments = append(arguments, objects...)
	arguments = append(arguments, "-o", "program")
	identity.Commands = append(identity.Commands, arguments)
	if _, err := r.applicationCommand(parent, clang, stageDirectory, arguments...); err != nil {
		if bindings != nil {
			return BuildReceipt{}, fmt.Errorf("link declared symbols %v: %w", bindings.Manifest.Symbols, err)
		}
		return BuildReceipt{}, err
	}
	if err := os.Chmod(stagedArtifact, 0o755); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.artifact_publish_failed", Err: err}
	}
	executableDigest, err := digestFile(stagedArtifact)
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.artifact_read_failed", Err: err}
	}
	flags := append([]string(nil), applicationFlags...)
	receipt := BuildReceipt{
		Schema:              ApplicationBuildSchema,
		ExecutableDigest:    executableDigest,
		SourceDigest:        digestBytes(source),
		EmittedCDigest:      digestBytes([]byte(cSource)),
		Compiler:            clang,
		CompilerVersion:     strings.TrimSpace(version),
		Target:              strings.TrimSpace(target),
		Flags:               flags,
		RuntimeDependencies: []string{"platform-c-runtime"},
		// The installed driver does not provide portable runtime/SDK closure
		// discovery here. Keep the artifact usable while refusing to claim
		// complete provenance or cacheability.
		DependencyClosure: "incomplete",
		Cacheable:         false,
	}
	identity.ExecutableDigest = executableDigest
	receipt.Identity = &identity
	receipt.BuildID, receipt.InputID = identity.ID(), identity.InputID()
	receipt.OutputPath = artifactPath
	if bindings != nil {
		receipt.Bindings = &bindings.Manifest
	}
	receiptBytes, err := json.Marshal(receipt)
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_encode_failed", Err: err}
	}
	receiptBytes = append(receiptBytes, '\n')
	if len(receiptBytes) > maxApplicationReceiptBytes {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_encode_failed", Err: errors.New("build receipt exceeds size limit")}
	}
	stagedReceipt := filepath.Join(stageDirectory, "program.schway-build.json")
	if err := os.WriteFile(stagedReceipt, receiptBytes, 0o600); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_write_failed", Err: err}
	}
	rename := r.publishRename
	if rename == nil {
		rename = os.Rename
	}
	if err := publishApplicationPair(stagedArtifact, stagedReceipt, artifactPath, stageDirectory, rename); err != nil {
		var recoveryError *applicationPublicationRecoveryError
		if errors.As(err, &recoveryError) {
			preserveStage = true
		}
		var toolError *ToolError
		if errors.As(err, &toolError) {
			return BuildReceipt{}, err
		}
		return BuildReceipt{}, &ToolError{Code: "native.artifact_publish_failed", Err: err}
	}
	return receipt, nil
}

// RunApplication checks the adjacent build receipt and executable digest,
// then starts exactly one child with one opaque argv token and direct streams.
func RunApplication(ctx context.Context, artifactPath, input string, stdout, stderr io.Writer) (RunOutcome, error) {
	return DefaultRunner().RunApplication(ctx, artifactPath, input, stdout, stderr)
}

func (r Runner) RunApplication(parent context.Context, artifactPath, input string, stdout, stderr io.Writer) (RunOutcome, error) {
	outcome, _, err := r.runApplication(parent, artifactPath, input, "", stdout, stderr)
	return outcome, err
}

// RunApplicationWithEvidence runs the retained application once, optionally
// captures its compiler events through a private file, and atomically publishes
// a separate report. Incomplete or exhausted requested evidence is a tool
// error even though the child process outcome remains recorded and unchanged.
func RunApplicationWithEvidence(ctx context.Context, artifactPath, input, reportPath string, mode EvidenceMode, stdout, stderr io.Writer) (RunOutcome, ApplicationEvidenceReport, error) {
	return DefaultRunner().RunApplicationWithEvidence(ctx, artifactPath, input, reportPath, mode, stdout, stderr)
}

func (r Runner) RunApplicationWithEvidence(parent context.Context, artifactPath, input, reportPath string, mode EvidenceMode, stdout, stderr io.Writer) (RunOutcome, ApplicationEvidenceReport, error) {
	if strings.TrimSpace(reportPath) == "" {
		return RunOutcome{}, ApplicationEvidenceReport{}, &ToolError{Code: "native.evidence_report_required", Err: errors.New("a report path is required")}
	}
	if mode != EvidenceDisabled && mode != EvidenceEvents {
		return RunOutcome{}, ApplicationEvidenceReport{}, &ToolError{Code: "native.evidence_mode_invalid", Err: errors.New("evidence mode must be disabled or events")}
	}

	captureDirectory := ""
	capturePath := ""
	if mode == EvidenceEvents {
		var err error
		captureDirectory, err = os.MkdirTemp("", "lang-app-evidence-")
		if err != nil {
			return RunOutcome{}, ApplicationEvidenceReport{}, &ToolError{Code: "native.evidence_capture_setup_failed", Err: err}
		}
		defer os.RemoveAll(captureDirectory)
		capturePath = filepath.Join(captureDirectory, "capture.json")
	}

	outcome, receipt, err := r.runApplication(parent, artifactPath, input, capturePath, stdout, stderr)
	if err != nil {
		return outcome, ApplicationEvidenceReport{}, err
	}
	report := ApplicationEvidenceReport{
		Schema: ApplicationEvidenceSchema, BuildID: receipt.BuildID, InputID: receipt.InputID,
		ArtifactDigest: receipt.ExecutableDigest, SourceDigest: receipt.SourceDigest,
		InputDigest: digestBytes([]byte(input)), ProcessOutcome: outcome,
		CaptureStatus: EvidenceStatusDisabled, Verified: false,
	}
	if mode == EvidenceEvents {
		report.CaptureStatus = EvidenceStatusIncomplete
		data, readErr := os.ReadFile(capturePath)
		limit := r.evidenceCapacityLimit()
		switch {
		case readErr == nil && len(data) >= limit:
			report.CaptureStatus = EvidenceStatusCapacityExhausted
		case readErr == nil:
			capture, captureErr := decodeApplicationCapture(data, limit)
			if captureErr == nil && capture.Status == EvidenceStatusCapacityExhausted {
				report.CaptureStatus = EvidenceStatusCapacityExhausted
			} else if captureErr == nil && capture.Status == EvidenceStatusComplete {
				report.Execution = &capture.Execution
				report.CaptureStatus = EvidenceStatusComplete
			}
		}
	}

	encoded, encodeErr := json.Marshal(report)
	if encodeErr != nil {
		return outcome, report, &ToolError{Code: "native.evidence_report_encode_failed", Err: encodeErr}
	}
	if len(encoded)+1 > execution.MaxApplicationEvidenceBytes {
		report.CaptureStatus = EvidenceStatusCapacityExhausted
		report.Execution = nil
	}
	if err := writeApplicationEvidenceReport(reportPath, report); err != nil {
		return outcome, report, err
	}
	if report.CaptureStatus == EvidenceStatusIncomplete {
		return outcome, report, &ToolError{Code: "native.evidence_incomplete", Err: errors.New("requested application evidence is missing or invalid")}
	}
	if report.CaptureStatus == EvidenceStatusCapacityExhausted {
		return outcome, report, &ToolError{Code: "native.evidence_capacity_exhausted", Err: errors.New("application evidence exceeds the capture limit")}
	}
	return outcome, report, nil
}

func (r Runner) runApplication(parent context.Context, artifactPath, input, capturePath string, stdout, stderr io.Writer) (RunOutcome, BuildReceipt, error) {
	if len(input) > MaxApplicationArgumentBytes {
		return RunOutcome{}, BuildReceipt{}, &ToolError{Code: "native.input_too_long", Err: fmt.Errorf("argument is %d bytes; limit is %d", len(input), MaxApplicationArgumentBytes)}
	}
	if strings.IndexByte(input, '\x00') >= 0 {
		return RunOutcome{}, BuildReceipt{}, &ToolError{Code: "native.input_contains_nul", Err: errors.New("argument contains an embedded NUL byte")}
	}
	if artifactPath == "" {
		return RunOutcome{}, BuildReceipt{}, &ToolError{Code: "native.artifact_required", Err: errors.New("an executable artifact path is required")}
	}
	artifactPath, err := filepath.Abs(artifactPath)
	if err != nil {
		return RunOutcome{}, BuildReceipt{}, &ToolError{Code: "native.artifact_invalid", Err: err}
	}
	receipt, err := readApplicationReceipt(applicationReceiptPath(artifactPath))
	if err != nil {
		return RunOutcome{}, BuildReceipt{}, err
	}
	actualDigest, err := digestFile(artifactPath)
	if err != nil {
		return RunOutcome{}, BuildReceipt{}, &ToolError{Code: "native.artifact_unreadable", Err: err}
	}
	if actualDigest != receipt.ExecutableDigest {
		return RunOutcome{}, BuildReceipt{}, &ToolError{Code: "native.artifact_digest_mismatch", Err: errors.New("executable bytes do not match the build receipt")}
	}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = defaultSubprocessTimeout
	}
	runCtx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	command := r.commandContext(runCtx, artifactPath, input)
	command.Stdout, command.Stderr = stdout, stderr
	// Evidence capture is an explicit runner capability. Never let an inherited
	// environment value authorize the generated application to write a file.
	if capturePath == "" {
		command.Env = removeEnvironment(command.Env, "SCHWAY_APP_EVIDENCE_PATH")
	} else {
		command.Env = replaceEnvironment(command.Env, "SCHWAY_APP_EVIDENCE_PATH", capturePath)
	}
	err = command.Run()
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return RunOutcome{Kind: RunTimedOut, Diagnostic: "application exceeded the run timeout"}, receipt, nil
	}
	if err == nil {
		return RunOutcome{Kind: RunExited, ExitCode: 0}, receipt, nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if adjudication, known := adjudicateExit(err); known && adjudication.Signaled {
			signal := adjudication.Signal
			return RunOutcome{Kind: RunSignaled, Signal: signal.String(), SignalNumber: int(signal)}, receipt, nil
		}
		return RunOutcome{Kind: RunExited, ExitCode: exitError.ExitCode()}, receipt, nil
	}
	return RunOutcome{Kind: RunLaunchError, Diagnostic: err.Error()}, receipt, nil
}

func replaceEnvironment(environment []string, key, value string) []string {
	if environment == nil {
		environment = os.Environ()
	}
	result := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if name != key {
			result = append(result, entry)
		}
	}
	return append(result, key+"="+value)
}

// publishApplicationPair stages the old artifact and receipt out of the way
// before publishing a new generation. If publishing the receipt fails, it
// removes the new executable and restores the previous pair.
func publishApplicationPair(stagedArtifact, stagedReceipt, artifactPath, stageDirectory string, rename func(string, string) error) error {
	receiptPath := applicationReceiptPath(artifactPath)
	for _, destination := range []string{artifactPath, receiptPath} {
		info, err := os.Lstat(destination)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return &ToolError{Code: "native.artifact_publish_failed", Err: err}
		}
		if err == nil && !info.Mode().IsRegular() {
			code := "native.artifact_publish_failed"
			if destination == receiptPath {
				code = "native.receipt_publish_failed"
			}
			return &ToolError{Code: code, Err: fmt.Errorf("publication destination %q is not a regular file", destination)}
		}
	}
	oldArtifact := filepath.Join(stageDirectory, "previous-program")
	oldReceipt := filepath.Join(stageDirectory, "previous-program.schway-build.json")
	_, artifactErr := os.Lstat(artifactPath)
	_, receiptErr := os.Lstat(receiptPath)
	hadArtifact := artifactErr == nil
	hadReceipt := receiptErr == nil
	if hadArtifact {
		if err := rename(artifactPath, oldArtifact); err != nil {
			return &ToolError{Code: "native.artifact_publish_failed", Err: err}
		}
	}
	if hadReceipt {
		if err := rename(receiptPath, oldReceipt); err != nil {
			if hadArtifact {
				if restoreErr := rename(oldArtifact, artifactPath); restoreErr != nil {
					return &applicationPublicationRecoveryError{directory: stageDirectory, cause: errors.Join(
						&ToolError{Code: "native.receipt_publish_failed", Err: err},
						fmt.Errorf("restoring prior artifact: %w", restoreErr),
					)}
				}
			}
			return &ToolError{Code: "native.receipt_publish_failed", Err: err}
		}
	}
	restore := func() error {
		var restoreErr error
		if hadArtifact {
			if err := rename(oldArtifact, artifactPath); err != nil {
				restoreErr = errors.Join(restoreErr, err)
			}
		}
		if hadReceipt {
			if err := rename(oldReceipt, receiptPath); err != nil {
				restoreErr = errors.Join(restoreErr, err)
			}
		}
		return restoreErr
	}
	if err := rename(stagedArtifact, artifactPath); err != nil {
		publishErr := &ToolError{Code: "native.artifact_publish_failed", Err: err}
		if restoreErr := restore(); restoreErr != nil {
			return &applicationPublicationRecoveryError{directory: stageDirectory, cause: errors.Join(publishErr, restoreErr)}
		}
		return publishErr
	}
	if err := rename(stagedReceipt, receiptPath); err != nil {
		removeErr := os.Remove(artifactPath)
		if errors.Is(removeErr, os.ErrNotExist) {
			removeErr = nil
		}
		publishErr := errors.Join(&ToolError{Code: "native.receipt_publish_failed", Err: err}, removeErr)
		if restoreErr := restore(); restoreErr != nil {
			return &applicationPublicationRecoveryError{directory: stageDirectory, cause: errors.Join(publishErr, restoreErr)}
		}
		return publishErr
	}
	return nil
}

type applicationPublicationRecoveryError struct {
	directory string
	cause     error
}

func (e *applicationPublicationRecoveryError) Error() string {
	return fmt.Sprintf("application publication rollback failed; recovery files retained at %q: %v", e.directory, e.cause)
}

func (e *applicationPublicationRecoveryError) Unwrap() error { return e.cause }

func decodeApplicationCapture(data []byte, limit int) (struct {
	Status    EvidenceCaptureStatus
	Execution execution.Execution
}, error) {
	result := struct {
		Status    EvidenceCaptureStatus
		Execution execution.Execution
	}{}
	if len(data) == 0 || len(data) >= limit {
		return result, errors.New("application capture is empty or at capacity")
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return result, err
	}
	var envelope applicationCapture
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return result, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return result, errors.New("trailing application capture document")
	}
	if envelope.Schema != applicationCaptureSchema {
		return result, errors.New("unsupported application capture schema")
	}
	if envelope.Status == string(EvidenceStatusCapacityExhausted) {
		result.Status = EvidenceStatusCapacityExhausted
		return result, nil
	}
	if envelope.Status != string(EvidenceStatusComplete) || len(envelope.Execution) == 0 {
		return result, errors.New("application capture is not complete")
	}
	executionValue, err := decodeExecution(envelope.Execution, ExpectValue)
	if err != nil {
		return result, err
	}
	if executionValue.Schema != execution.Schema2 {
		return result, errors.New("application capture must contain a schema-2 execution")
	}
	result.Status = EvidenceStatusComplete
	result.Execution = executionValue
	return result, nil
}

func writeApplicationEvidenceReport(path string, report ApplicationEvidenceReport) error {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return &ToolError{Code: "native.evidence_report_write_failed", Err: err}
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		return &ToolError{Code: "native.evidence_report_encode_failed", Err: err}
	}
	encoded = append(encoded, '\n')
	if len(encoded) > execution.MaxApplicationEvidenceBytes {
		return &ToolError{Code: "native.evidence_report_encode_failed", Err: errors.New("application evidence report exceeds the size limit")}
	}
	file, err := os.CreateTemp(filepath.Dir(absolutePath), ".schway-app-evidence-*")
	if err != nil {
		return &ToolError{Code: "native.evidence_report_write_failed", Err: err}
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return &ToolError{Code: "native.evidence_report_write_failed", Err: err}
	}
	if _, err := file.Write(encoded); err != nil {
		_ = file.Close()
		return &ToolError{Code: "native.evidence_report_write_failed", Err: err}
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return &ToolError{Code: "native.evidence_report_write_failed", Err: err}
	}
	if err := file.Close(); err != nil {
		return &ToolError{Code: "native.evidence_report_write_failed", Err: err}
	}
	if err := os.Rename(temporaryPath, absolutePath); err != nil {
		return &ToolError{Code: "native.evidence_report_write_failed", Err: err}
	}
	return nil
}

func readApplicationReceipt(path string) (BuildReceipt, error) {
	file, err := os.Open(path)
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_missing", Err: err}
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxApplicationReceiptBytes+1))
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_read_failed", Err: err}
	}
	if len(data) > maxApplicationReceiptBytes {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: errors.New("build receipt exceeds the size limit")}
	}
	var receipt BuildReceipt
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: err}
	}
	if err := json.Unmarshal(data, &receipt); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: err}
	}
	if receipt.Schema != ApplicationBuildSchema || !validSHA256(receipt.ExecutableDigest) || !validSHA256(receipt.SourceDigest) || !validSHA256(receipt.EmittedCDigest) || receipt.Compiler == "" || receipt.CompilerVersion == "" || receipt.Target == "" || receipt.DependencyClosure != "incomplete" || receipt.Cacheable || len(receipt.RuntimeDependencies) != 1 || receipt.RuntimeDependencies[0] != "platform-c-runtime" {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: errors.New("build receipt is missing required v1 identity facts")}
	}
	if receipt.Identity != nil {
		identity := receipt.Identity
		if identity.ID() != receipt.BuildID || identity.InputID() != receipt.InputID || identity.ExecutableDigest != receipt.ExecutableDigest || identity.SourceDigest != receipt.SourceDigest || identity.EmittedCDigest != receipt.EmittedCDigest || identity.CompilerVersion != receipt.CompilerVersion || identity.Target != receipt.Target || !validSHA256(identity.CompilerDigest) || !equalJSON(identity.Flags, receipt.Flags) || !equalJSON(identity.RuntimeDependencies, receipt.RuntimeDependencies) {
			return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: errors.New("build identity does not match receipt facts")}
		}
		if receipt.Bindings != nil {
			canonical, _ := json.Marshal(receipt.Bindings)
			if digestBytes(canonical) != identity.ManifestDigest {
				return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: errors.New("manifest does not match build identity")}
			}
		} else if identity.ManifestDigest != "" {
			return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: errors.New("manifest is missing from build receipt")}
		}
	} else if receipt.Bindings != nil || receipt.BuildID != "" || receipt.InputID != "" {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: errors.New("build identity is missing")}
	}
	return receipt, nil
}

func equalJSON(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

func digestFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func removeEnvironment(environment []string, key string) []string {
	if environment == nil {
		environment = os.Environ()
	}
	result := make([]string, 0, len(environment))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if name != key {
			result = append(result, entry)
		}
	}
	return result
}
