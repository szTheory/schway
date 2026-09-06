package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// MaxProbeBytes bounds a single spawned tool probe's stdout/stderr,
// mirroring evidence.MaxToolProbeBytes's 64 KiB-plus-one bounding mechanism
// (D-06-07). cache is a dependency-free leaf package (see cache.go's
// package doc) and therefore does not import evidence; this is a
// deliberate, minimal duplicate of the same bounding shape -- not a shared
// helper, and never a second, weaker bounding mechanism.
const MaxProbeBytes = 64 * 1024

// DeclaredInputNames returns EXACTLY the seven names D-06-07 lists, in this
// EXACT stable order. This list is DECLARED, not COMPLETE: anything not on
// it is an escape by construction. D-06-13 records four such holes, none
// closed by this package:
//
//  1. undeclared environment -- locale, ulimit, filesystem case-sensitivity;
//  2. a Clang change that does not alter its reported version string (the
//     ccache __TIME__-class footgun) -- mitigated but not closed by
//     clang_identity below being a probed digest of the binary's own bytes,
//     not merely its --version string;
//  3. any future nondeterministic codegen silently breaking the "same
//     inputs implies same artifact" premise;
//  4. a hand-edited or partially deleted cache directory being
//     indistinguishable from a cold one, since there is no integrity check
//     beyond content-hash lookup (T-06-CACHE-01).
func DeclaredInputNames() []string {
	return []string{
		"fixture_source",
		"build_flags",
		"clang_identity",
		"runtime_identity",
		"foreign_translation_unit",
		"mutation_runner_source",
		"go_toolchain",
	}
}

// Declared artifact kinds (Task 3's closed classification): a spec whose
// Kind is not one of these is refused by Consult with StatusNotCacheable,
// never guessed at (D-06-11).
const (
	KindCompiledBinary     = "compiled_binary"
	KindInstrumentedBinary = "instrumented_binary"
	KindMutantBinary       = "mutant_binary"
)

// ArtifactKinds returns the closed set of artifact kinds this cache
// classifies. A Kind outside this set is, by construction, not exhaustively
// classified.
func ArtifactKinds() []string {
	return []string{KindCompiledBinary, KindInstrumentedBinary, KindMutantBinary}
}

// ArtifactSpec is the caller-supplied description of one artifact build:
// every field maps to one of DeclaredInputNames()'s seven declared inputs,
// plus the Kind Consult classifies against ArtifactKinds(). RuntimeIdentity
// is the empty-input error when a lane needs it but cannot compute it --
// callers for whom no ASan/UBSan/libc++ runtime is linked must still supply
// an explicit non-empty declared value (e.g. "none"), since an empty string
// here is refused as an undeclared input (D-06-11), never silently treated
// as "not applicable."
type ArtifactSpec struct {
	Kind                     string
	FixtureSource            []byte
	BuildFlags               string
	ClangPath                string
	RuntimeIdentity          string
	ForeignTranslationUnit   []byte
	MutationRunnerSourcePath string
	GoToolchain              string
}

// InputsFor assembles the seven declared inputs for spec, in
// DeclaredInputNames() order. It accepts a context because computing
// clang_identity requires spawning a bounded probe subprocess
// (ProbeClangIdentity). Any input that cannot be computed -- including a
// failed Clang probe -- returns Error{Code: "cache.input_undeclared"}
// naming the failure by refusing outright: ambiguity and silence always
// resolve to "run it," never to "skip it" (D-06-11).
func InputsFor(ctx context.Context, spec ArtifactSpec) ([]Input, error) {
	byName := make(map[string]string, len(DeclaredInputNames()))

	if len(spec.FixtureSource) == 0 {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	byName["fixture_source"] = hashBytes(spec.FixtureSource)

	if spec.BuildFlags == "" {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	byName["build_flags"] = hashString(spec.BuildFlags)

	clangDigest, err := ProbeClangIdentity(ctx, spec.ClangPath)
	if err != nil {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	byName["clang_identity"] = clangDigest

	if spec.RuntimeIdentity == "" {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	byName["runtime_identity"] = hashString(spec.RuntimeIdentity)

	if len(spec.ForeignTranslationUnit) == 0 {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	byName["foreign_translation_unit"] = hashBytes(spec.ForeignTranslationUnit)

	if spec.MutationRunnerSourcePath == "" {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	runnerBytes, err := os.ReadFile(spec.MutationRunnerSourcePath)
	if err != nil {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	byName["mutation_runner_source"] = hashBytes(runnerBytes)

	toolchain := spec.GoToolchain
	if toolchain == "" {
		toolchain = runtime.Version()
	}
	byName["go_toolchain"] = hashString(toolchain)

	names := DeclaredInputNames()
	inputs := make([]Input, 0, len(names))
	for _, name := range names {
		inputs = append(inputs, Input{Name: name, Digest: byName[name]})
	}
	return inputs, nil
}

// commandFactory is the injectable subprocess-launch seam ProbeClangIdentity
// uses in production (exec.CommandContext) and probe_test.go overrides for
// deterministic, bounded fixture behaviour (mirroring evidence's own
// commandFactory / runToolProbe test seam).
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

// ProbeClangIdentity returns a PROBED DIGEST of clangPath's own identity --
// never merely its --version string. The digest combines the SHA-256 of
// the resolved clang binary's own bytes with the bounded "--version" probe
// output, so a Clang change that leaves the reported version string
// unchanged (D-06-13 hole 2, the ccache __TIME__-class footgun) still moves
// the resulting digest. The probe itself reuses the 64 KiB-plus-one bounded
// writer and 5-second deadline every other tool-identity probe in this
// project uses (evidence.runToolProbe's structure, duplicated rather than
// imported per this package's dependency-free-leaf convention).
func ProbeClangIdentity(ctx context.Context, clangPath string) (string, error) {
	return probeClangIdentityWithCommand(ctx, clangPath, exec.CommandContext)
}

func probeClangIdentityWithCommand(ctx context.Context, clangPath string, command commandFactory) (string, error) {
	if clangPath == "" {
		clangPath = "clang"
	}
	binaryBytes, err := resolveClangBinaryBytes(clangPath)
	if err != nil {
		return "", err
	}
	output, err := runBoundedProbe(ctx, command, clangPath, "--version")
	if err != nil {
		return "", err
	}
	binarySum := sha256.Sum256(binaryBytes)
	combined := append(append([]byte(nil), binarySum[:]...), output...)
	sum := sha256.Sum256(combined)
	return hex.EncodeToString(sum[:]), nil
}

func resolveClangBinaryBytes(clangPath string) ([]byte, error) {
	if data, err := os.ReadFile(clangPath); err == nil {
		return data, nil
	}
	resolved, err := exec.LookPath(clangPath)
	if err != nil {
		return nil, &Error{Code: "cache.probe_unresolved"}
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, &Error{Code: "cache.probe_unresolved"}
	}
	return data, nil
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
		return nil, &Error{Code: "cache.probe_timeout"}
	}
	if stdout.total > MaxProbeBytes {
		return nil, &Error{Code: "cache.probe_stdout_truncated"}
	}
	if stderr.total > MaxProbeBytes {
		return nil, &Error{Code: "cache.probe_stderr_truncated"}
	}
	if err != nil {
		return nil, &Error{Code: "cache.probe_failed"}
	}
	return append([]byte(nil), stdout.buffer.Bytes()...), nil
}
