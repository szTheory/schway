// Package measure is a dependency-free leaf package (mirroring
// internal/compiler/reduce and internal/compiler/cache's convention): it
// implements the declared-machine probe (D-06-17) and the sample-statistics
// helper (D-06-19) that ratify Phase 6's cost budgets. It imports nothing
// from internal/compiler/protocol, internal/compiler/session, or
// internal/compiler/diagnostic (see TestMeasureImportsStayIndependent).
package measure

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os/exec"
	"runtime"
	"time"
)

// MaxProbeBytes bounds a single spawned tool probe's stdout/stderr,
// mirroring evidence.MaxToolProbeBytes / cache.MaxProbeBytes's 64
// KiB-plus-one bounding mechanism. measure is a dependency-free leaf
// package and therefore duplicates this bounding shape rather than
// importing it -- not a shared helper, and never a second, weaker bound.
const MaxProbeBytes = 64 * 1024

// MachineFacts is D-06-17's closed six-field declared-machine identity.
// NO hostname, serial number, MAC address, username, home directory, or
// any absolute host path may ever be added here -- host fingerprints are a
// needless leak (T-06-MEASURE-01). See FactFieldNames for the enforced
// closed set and TestMachineIDExcludesHostFingerprints for the structural
// and runtime assertions that back this comment.
type MachineFacts struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	CPUModel     string `json:"cpu_model"`
	LogicalCores int    `json:"logical_cores"`
	GoVersion    string `json:"go_version"`
	ClangVersion string `json:"clang_version"`
}

// FactFieldNames returns the closed six-name set MachineFacts declares, in
// stable order, so a seventh field cannot be added to the struct without
// also updating this list -- TestMachineFactFieldsAreClosed pins the two
// together via reflection.
func FactFieldNames() []string {
	return []string{"os", "arch", "cpu_model", "logical_cores", "go_version", "clang_version"}
}

// Error is this package's stable typed failure, matching the {Code}-only
// shape evidence.ValidationError / debugmap.Error / cache.Error already
// use, so callers can dispatch on Code identically.
type Error struct{ Code string }

func (e *Error) Error() string { return e.Code }

// commandFactory is the injectable subprocess-launch seam ProbeMachine uses
// in production (exec.CommandContext) and machine_test.go overrides for
// deterministic, bounded fixture behaviour (mirroring evidence's and
// cache's own commandFactory / runToolProbe test seam).
type commandFactory func(ctx context.Context, name string, args ...string) *exec.Cmd

type boundedProbeWriter struct {
	buffer bytes.Buffer
	total  int
}

func (w *boundedProbeWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := MaxProbeBytes + 1 - w.buffer.Len()
	if remaining > len(data) {
		remaining = len(data)
	}
	if remaining > 0 {
		w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func runBoundedProbe(parent context.Context, command commandFactory, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cmd := command(ctx, name, args...)
	var stdout, stderr boundedProbeWriter
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, &Error{Code: "measure.probe_timeout"}
	}
	if stdout.total > MaxProbeBytes {
		return nil, &Error{Code: "measure.probe_stdout_truncated"}
	}
	if stderr.total > MaxProbeBytes {
		return nil, &Error{Code: "measure.probe_stderr_truncated"}
	}
	if err != nil {
		return nil, &Error{Code: "measure.probe_failed"}
	}
	return append([]byte(nil), stdout.buffer.Bytes()...), nil
}

// probeClangVersion returns clang's bounded --version output, trimmed to
// its first line (matching evidence.defaultFacts's own trim shape).
func probeClangVersion(ctx context.Context, command commandFactory, clangPath string) (string, error) {
	if clangPath == "" {
		clangPath = "clang"
	}
	output, err := runBoundedProbe(ctx, command, clangPath, "--version")
	if err != nil {
		return "", err
	}
	trimmed := bytes.TrimSpace(bytes.SplitN(output, []byte("\n"), 2)[0])
	return string(trimmed), nil
}

// probeCPUModel returns a bounded platform probe of the CPU model string.
// On darwin/linux this shells out to a small stdlib-available probe; the
// probe is bounded and typed-error-failing identically to the Clang probe
// above -- there is no separate, weaker bounding mechanism for this one.
func probeCPUModel(ctx context.Context, command commandFactory) (string, error) {
	name, args := cpuModelProbeCommand()
	if name == "" {
		return "unavailable", nil
	}
	output, err := runBoundedProbe(ctx, command, name, args...)
	if err != nil {
		return "", err
	}
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return "unavailable", nil
	}
	return string(trimmed), nil
}

// cpuModelProbeCommand returns the bounded platform probe command for the
// current GOOS, or an empty name when no stdlib-available probe exists for
// this platform -- an unavailable probe reports "unavailable" (the house
// style), never a fabricated value.
func cpuModelProbeCommand() (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		return "sysctl", []string{"-n", "machdep.cpu.brand_string"}
	case "linux":
		return "sh", []string{"-c", "grep -m1 'model name' /proc/cpuinfo | cut -d: -f2"}
	default:
		return "", nil
	}
}

// ProbeMachine produces a real MachineFacts for the running host, end to
// end: os/arch/logical_cores/go_version come from the Go runtime directly
// (no probe needed), cpu_model and clang_version come from bounded
// subprocess probes sharing the same 64 KiB-plus-one / 5-second discipline.
func ProbeMachine(ctx context.Context) (MachineFacts, error) {
	return probeMachine(ctx, exec.CommandContext)
}

func probeMachine(ctx context.Context, command commandFactory) (MachineFacts, error) {
	clangVersion, err := probeClangVersion(ctx, command, "")
	if err != nil {
		return MachineFacts{}, err
	}
	cpuModel, err := probeCPUModel(ctx, command)
	if err != nil {
		return MachineFacts{}, err
	}
	return MachineFacts{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		CPUModel:     cpuModel,
		LogicalCores: runtime.NumCPU(),
		GoVersion:    runtime.Version(),
		ClangVersion: clangVersion,
	}, nil
}

// MachineID returns D-06-17's short, leak-free machine identifier: a
// content hash of facts's canonical JSON, prefixed "machine:" followed by
// 12 hex characters (6 bytes of a SHA-256 digest) -- the same short-hash
// encoding convention evidence.manifestID and debugmap's identity hashes
// already use.
func MachineID(facts MachineFacts) string {
	encoded, _ := json.Marshal(facts)
	sum := sha256.Sum256(encoded)
	return "machine:" + hex.EncodeToString(sum[:6])
}
