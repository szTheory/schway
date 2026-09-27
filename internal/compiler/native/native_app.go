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
	"strings"
)

const (
	ApplicationBuildSchema      = "lang.app-build/1"
	MaxApplicationArgumentBytes = 4096
	maxApplicationReceiptBytes  = 64 * 1024
)

type BuildReceipt struct {
	Schema              string   `json:"schema"`
	ExecutableDigest    string   `json:"executable_digest"`
	SourceDigest        string   `json:"source_digest"`
	EmittedCDigest      string   `json:"emitted_c_digest"`
	Compiler            string   `json:"compiler"`
	CompilerVersion     string   `json:"compiler_version"`
	Target              string   `json:"target"`
	Flags               []string `json:"flags"`
	RuntimeDependencies []string `json:"runtime_dependencies"`
	DependencyClosure   string   `json:"dependency_closure"`
	Cacheable           bool     `json:"cacheable"`
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

var applicationFlags = []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O0"}

func applicationReceiptPath(artifactPath string) string {
	return artifactPath + ".lang-build.json"
}

// BuildApplication compiles the generated translation unit and publishes a
// retained executable plus an adjacent content-bound receipt. It never starts
// the resulting application.
func BuildApplication(ctx context.Context, source []byte, cSource, outputPath string) (BuildReceipt, error) {
	return DefaultRunner().BuildApplication(ctx, source, cSource, outputPath)
}

func (r Runner) BuildApplication(parent context.Context, source []byte, cSource, outputPath string) (BuildReceipt, error) {
	if r.ClangPath == "" {
		r.ClangPath = "clang"
	}
	if r.Timeout <= 0 {
		r.Timeout = defaultSubprocessTimeout
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

	stageDirectory, err := os.MkdirTemp(outputDirectory, ".lang-build-")
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.temp_failed", Err: err}
	}
	defer os.RemoveAll(stageDirectory)
	sourcePath := filepath.Join(stageDirectory, "program.c")
	stagedArtifact := filepath.Join(stageDirectory, "program")
	if err := os.WriteFile(sourcePath, []byte(cSource), 0o600); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.temp_failed", Err: err}
	}
	arguments := append([]string(nil), applicationFlags...)
	arguments = append(arguments, sourcePath, "-o", stagedArtifact)
	compileCtx, cancel := context.WithTimeout(parent, r.Timeout)
	command := r.commandContext(compileCtx, clang, arguments...)
	var stdout, stderr boundedWriter
	command.Stdout, command.Stderr = &stdout, &stderr
	compileErr := command.Run()
	timedOut := errors.Is(compileCtx.Err(), context.DeadlineExceeded)
	cancel()
	if timedOut {
		return BuildReceipt{}, &ToolError{Code: "native.timeout", Err: context.DeadlineExceeded}
	}
	if stdout.overflowed() || stderr.overflowed() {
		return BuildReceipt{}, streamError("native.compile_diagnostics_truncated")
	}
	if compileErr != nil {
		code := "native.compile_failed"
		if errors.Is(compileErr, exec.ErrNotFound) || errors.Is(compileErr, os.ErrNotExist) {
			code = "native.tool_missing"
		}
		return BuildReceipt{}, &ToolError{Code: code, Err: withStderr(compileErr, stderr.bytes())}
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
	receiptBytes, err := json.Marshal(receipt)
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_encode_failed", Err: err}
	}
	receiptBytes = append(receiptBytes, '\n')
	stagedReceipt := filepath.Join(stageDirectory, "program.lang-build.json")
	if err := os.WriteFile(stagedReceipt, receiptBytes, 0o600); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_write_failed", Err: err}
	}
	if err := os.Rename(stagedArtifact, artifactPath); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.artifact_publish_failed", Err: err}
	}
	if err := os.Rename(stagedReceipt, applicationReceiptPath(artifactPath)); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_publish_failed", Err: err}
	}
	return receipt, nil
}

// RunApplication checks the adjacent build receipt and executable digest,
// then starts exactly one child with one opaque argv token and direct streams.
func RunApplication(ctx context.Context, artifactPath, input string, stdout, stderr io.Writer) (RunOutcome, error) {
	return DefaultRunner().RunApplication(ctx, artifactPath, input, stdout, stderr)
}

func (r Runner) RunApplication(parent context.Context, artifactPath, input string, stdout, stderr io.Writer) (RunOutcome, error) {
	if len(input) > MaxApplicationArgumentBytes {
		return RunOutcome{}, &ToolError{Code: "native.input_too_long", Err: fmt.Errorf("argument is %d bytes; limit is %d", len(input), MaxApplicationArgumentBytes)}
	}
	if artifactPath == "" {
		return RunOutcome{}, &ToolError{Code: "native.artifact_required", Err: errors.New("an executable artifact path is required")}
	}
	artifactPath, err := filepath.Abs(artifactPath)
	if err != nil {
		return RunOutcome{}, &ToolError{Code: "native.artifact_invalid", Err: err}
	}
	receipt, err := readApplicationReceipt(applicationReceiptPath(artifactPath))
	if err != nil {
		return RunOutcome{}, err
	}
	actualDigest, err := digestFile(artifactPath)
	if err != nil {
		return RunOutcome{}, &ToolError{Code: "native.artifact_unreadable", Err: err}
	}
	if actualDigest != receipt.ExecutableDigest {
		return RunOutcome{}, &ToolError{Code: "native.artifact_digest_mismatch", Err: errors.New("executable bytes do not match the build receipt")}
	}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = defaultSubprocessTimeout
	}
	runCtx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	command := r.commandContext(runCtx, artifactPath, input)
	command.Stdout, command.Stderr = stdout, stderr
	err = command.Run()
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return RunOutcome{Kind: RunTimedOut, Diagnostic: "application exceeded the run timeout"}, nil
	}
	if err == nil {
		return RunOutcome{Kind: RunExited, ExitCode: 0}, nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if adjudication, known := adjudicateExit(err); known && adjudication.Signaled {
			signal := adjudication.Signal
			return RunOutcome{Kind: RunSignaled, Signal: signal.String(), SignalNumber: int(signal)}, nil
		}
		return RunOutcome{Kind: RunExited, ExitCode: exitError.ExitCode()}, nil
	}
	return RunOutcome{Kind: RunLaunchError, Diagnostic: err.Error()}, nil
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
	if err := json.Unmarshal(data, &receipt); err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: err}
	}
	if receipt.Schema != ApplicationBuildSchema || !validSHA256(receipt.ExecutableDigest) || !validSHA256(receipt.SourceDigest) || !validSHA256(receipt.EmittedCDigest) || receipt.Compiler == "" || receipt.CompilerVersion == "" || receipt.Target == "" || receipt.DependencyClosure != "incomplete" || receipt.Cacheable || len(receipt.RuntimeDependencies) != 1 || receipt.RuntimeDependencies[0] != "platform-c-runtime" {
		return BuildReceipt{}, &ToolError{Code: "native.receipt_invalid", Err: errors.New("build receipt is missing required v1 identity facts")}
	}
	return receipt, nil
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
