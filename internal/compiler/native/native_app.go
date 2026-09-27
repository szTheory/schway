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
)

const (
	ApplicationBuildSchema      = "lang.app-build/1"
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

var applicationFlags = []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O0"}

func applicationReceiptPath(artifactPath string) string {
	return artifactPath + ".lang-build.json"
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

	stageDirectory, err := os.MkdirTemp(outputDirectory, ".lang-build-")
	if err != nil {
		return BuildReceipt{}, &ToolError{Code: "native.temp_failed", Err: err}
	}
	defer os.RemoveAll(stageDirectory)
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
