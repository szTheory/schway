package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// MaxProbeBytes bounds a single spawned tool probe's stdout/stderr,
// mirroring evidence.MaxToolProbeBytes's 64 KiB-plus-one bounding mechanism
// (D-06-07). cache is a dependency-free leaf package (see cache.go's
// package doc) and therefore does not import evidence; this is a
// deliberate, minimal duplicate of the same bounding shape -- not a shared
// helper, and never a second, weaker bounding mechanism.
const MaxProbeBytes = 64 * 1024

// DeclaredInputNames returns the eight names this package declares, in
// this EXACT stable order: the original seven D-06-07 names, byte-identical
// and in their original order, with the eighth (cgen_source) APPENDED by
// D-11-41 -- never inserted, renamed or reordered among the first seven.
// This list is DECLARED, not COMPLETE: anything not on it is an escape by
// construction. D-06-13 records five such holes; the fifth (below) is the
// only one closed by this package:
//
//  1. undeclared environment -- locale, ulimit, filesystem case-sensitivity;
//  2. a Clang change that does not alter its reported version string (the
//     ccache __TIME__-class footgun) -- mitigated but not closed by
//     cschway_identity below being a probed digest of the binary's own bytes,
//     not merely its --version string;
//  3. any future nondeterministic codegen silently breaking the "same
//     inputs implies same artifact" premise;
//  4. a hand-edited or partially deleted cache directory being
//     indistinguishable from a cold one, since there is no integrity check
//     beyond content-hash lookup (T-06-CACHE-01).
//  5. (D-11-41, CLOSED by cgen_source below) internal/compiler/cgen/*.go --
//     Phase 11's own code generator's source -- was not a declared input.
//     Editing cgen and re-running against an UNCHANGED .schway fixture could
//     serve a binary compiled by the OLD cgen against the NEW interpreter:
//     the same ccache __TIME__-class footgun as hole (2), but for this
//     repo's own generator rather than an external toolchain. Confirmed
//     reproducing end-to-end (Phase 11 plan 11-01's Q-02 spike,
//     BRANCH A -- session_phase6_cache_hole_test.go's
//     TestQ02StaleCgenServesReusedArtifact) before cgen_source closed it.
func DeclaredInputNames() []string {
	return []string{
		"fixture_source",
		"build_flags",
		"cschway_identity",
		"runtime_identity",
		"foreign_translation_unit",
		"mutation_runner_source",
		"go_toolchain",
		"cgen_source",
	}
}

// CgenSourceDigest hashes the concatenated bytes of every *.go file
// (including its own _test.go siblings -- this input's job is "did
// anything under this directory change", not "did production behavior
// change") directly under cgenDir, sorted by filename for a deterministic
// preimage, using ONLY the crypto/sha256 and os facilities this package
// already imports. cache never imports core, corevalidate or
// originvalidate to obtain a digest -- it is a dependency-free leaf (see
// cache.go's package doc), and asking another package for this digest
// would destroy the structural discharge QLT-06b's import-scan tests rest
// on (D-11-39).
//
// An EMPTY file set -- a hypothetical or mis-rooted cgenDir -- is refused
// as Error{Code: "cache.input_undeclared"} rather than silently hashed
// into an empty digest: an empty digest would make every key computed from
// it collide, which is strictly worse than the hole this input closes.
func CgenSourceDigest(cgenDir string) (string, error) {
	entries, err := os.ReadDir(cgenDir)
	if err != nil {
		return "", &Error{Code: "cache.input_undeclared"}
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		names = append(names, entry.Name())
	}
	if len(names) == 0 {
		return "", &Error{Code: "cache.input_undeclared"}
	}
	sort.Strings(names)

	hasher := sha256.New()
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(cgenDir, name))
		if err != nil {
			return "", &Error{Code: "cache.input_undeclared"}
		}
		hasher.Write(data)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
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
// every field maps to one of DeclaredInputNames()'s eight declared inputs,
// plus the Kind Consult classifies against ArtifactKinds(). RuntimeIdentity
// is the empty-input error when a lane needs it but cannot compute it --
// callers for whom no ASan/UBSan/libc++ runtime is linked must still supply
// an explicit non-empty declared value (e.g. "none"), since an empty string
// here is refused as an undeclared input (D-06-11), never silently treated
// as "not applicable." CgenSourceDir (D-11-41, the eighth declared input)
// is a caller-supplied directory path -- cache itself has no notion of
// repo layout -- pointing at internal/compiler/cgen; it is hashed via
// CgenSourceDigest.
type ArtifactSpec struct {
	Kind                     string
	FixtureSource            []byte
	BuildFlags               string
	ClangPath                string
	RuntimeIdentity          string
	ForeignTranslationUnit   []byte
	MutationRunnerSourcePath string
	GoToolchain              string
	CgenSourceDir            string
}

// InputsFor assembles the seven declared inputs for spec, in
// DeclaredInputNames() order. It accepts a context because computing
// cschway_identity requires spawning a bounded probe subprocess
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
		return nil, &Error{Code: "cache.input_undeclared", Cause: err}
	}
	byName["cschway_identity"] = clangDigest

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

	if spec.CgenSourceDir == "" {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	cgenDigest, err := CgenSourceDigest(spec.CgenSourceDir)
	if err != nil {
		return nil, &Error{Code: "cache.input_undeclared"}
	}
	byName["cgen_source"] = cgenDigest

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
// writer and finite 30-second deadline every other tool-identity probe in this
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
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
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
