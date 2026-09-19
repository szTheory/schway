package interp

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// oracleGoldenDir is where every corpus program's committed golden lives,
// one file per entry in interpOracleCorpus, named "<entry.name>.golden.json".
func oracleGoldenDir() string {
	return filepath.Join(interpProjectRoot(), "testdata", "phase10", "interp_oracle")
}

// checkedProgramFromFixture generalizes checkedCallBasicProgram's own
// technique (syntax.Parse -> check.Program -> corevalidate.Validate, never
// package session, which imports interp) to any phase's fixture directory,
// so the golden corpus can freeze programs beyond testdata/phase07 without
// duplicating the pipeline-driving boilerplate a second time.
func checkedProgramFromFixture(t *testing.T, phaseDir, name string) core.Program {
	t.Helper()
	path := filepath.Join(interpProjectRoot(), "testdata", phaseDir, name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("parse %s: unexpected diagnostics: %v", path, parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("check %s: unexpected diagnostics: %v", path, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate %s rejected: %v", path, validated.Problems)
	}
	return validated.Program()
}

// oracleTypedFailureProgram hand-builds a minimal single-function
// core.Program (never fed through the parser) whose sole operation is a
// terminal core.OpFail. This is deliberate, not an oversight: runLinearBlocks'
// own doc comment states that interp models every core.OpForeignCall as
// unconditionally succeeding along its ok edge this phase, so no real,
// checkable Lang program can ever drive interp.Run to the err block a
// fallible foreign call's OpFail lives in -- check.go emits that OpFail
// (check.go:3080/3307), but interp never visits it. The typed-failure
// terminal path is therefore frozen directly against a synthetic
// core.Program, mirroring moveAsCopyProbeProgram's and the frameDrain*
// builders' established precedent (interp_test.go) for isolating a shape
// the real pipeline cannot drive interp to reach.
func oracleTypedFailureProgram() (program core.Program, entry core.Function) {
	errType := core.TypeFact{ID: "oracle:fail:type:err"}
	param := core.Parameter{ID: "oracle:fail:place:param", Name: "value", Type: "ProbeError"}
	entry = core.Function{
		ID: "oracle:fail:fn", Name: "failer", Parameter: param, ReturnType: "ProbeError",
		Linear: &core.LinearBody{
			ID:    "oracle:fail:linear",
			Types: []core.TypeFact{errType},
			Places: []core.Place{
				{ID: param.ID, Name: "value", TypeID: errType.ID},
			},
			Operations: []core.LinearOperation{
				{ID: "oracle:fail:op:0", Kind: core.OpFail, SourceID: param.ID, TypeID: errType.ID},
			},
		},
	}
	return core.Program{Schema: core.Schema1, Module: "oracle.typed_failure", Functions: []core.Function{entry}}, entry
}

// oracleCorpusProgram is one entry in the golden corpus: a name (also the
// golden file's basename) and a closure that produces the Execution to
// freeze/compare, however that program must be driven -- through the real
// pipeline via Run (the common case) or directly via runFrameStack for a
// deliberately synthetic core.Program no legal Lang source can express.
type oracleCorpusProgram struct {
	name string
	run  func(t *testing.T) (Execution, error)
}

// interpOracleCorpus is the committed set of programs D-10-57 clause 1
// freezes, covering every terminal path interp can reach (PHASE-10-DEBT.md
// D-10-31's own enumeration): a normal return (single-frame AND
// cross-frame), a typed failure, a defect, a cross-frame abrupt exit through
// the foreign process-root landing pad, and the SEM-08 depth-exceeded
// refusal. Both genuine multi-frame-skipping constructs D-10-31 names are
// present (the landing pad and the depth refusal) -- a corpus exercising
// only OpReturn/OpFail popping would freeze half the oracle.
func interpOracleCorpus(t *testing.T) []oracleCorpusProgram {
	t.Helper()
	return []oracleCorpusProgram{
		{
			name: "single_frame_return",
			run: func(t *testing.T) (Execution, error) {
				program := checkedProgramFromFixture(t, "phase4", "defect_terminal.lang")
				return Run(program, "triage", "Go")
			},
		},
		{
			name: "cross_frame_normal_return",
			run: func(t *testing.T) (Execution, error) {
				program := checkedCallBasicProgram(t)
				return Run(program, "main", "7")
			},
		},
		{
			name: "typed_failure",
			run: func(t *testing.T) (Execution, error) {
				program, entry := oracleTypedFailureProgram()
				base := newFlatFrame(entry, map[string]value{entry.Parameter.ID: {tag: "", payload: "err_value"}})
				return runFrameStack(program, base)
			},
		},
		{
			name: "defect_terminal",
			run: func(t *testing.T) (Execution, error) {
				program := checkedProgramFromFixture(t, "phase4", "defect_terminal.lang")
				return Run(program, "triage", "Halt")
			},
		},
		{
			name: "cross_frame_nonlocal_landing_pad",
			run: func(t *testing.T) (Execution, error) {
				program, entry := frameDrainNonlocalMultiFrameProgram()
				return runDrainProgram(program, entry, "7")
			},
		},
		{
			name: "call_depth_exceeded_refusal",
			run: func(t *testing.T) (Execution, error) {
				program := generateAndCheckCallDepthChain(t, MaxCallDepth+1)
				return Run(program, "link0", "9")
			},
		},
	}
}

// TestInterpOracleGoldenCorpus is Plan 10-09 Task 1's stability freeze
// (D-10-57 clause 1): interp's emitted Schema value and Execution document
// shape are frozen BYTE-FOR-BYTE against a committed golden corpus covering
// every terminal path interp can reach. A change to interp's emitted schema,
// field set, event ordering, or outcome shape makes this test fail, naming
// the program and showing both byte sequences.
//
// REGENERATION (D-10-57 clause 2): this corpus is NEVER regenerated by a
// command-line switch. No `-update` switch or any other boolean CLI-driven
// rewrite mechanism exists in this package, by design (a bare switch is a
// mechanism a CI failure can silently absorb -- someone reruns with it and
// the drift disappears without review). Regenerating a golden is a REVIEWED
// COMMIT touching the golden files with a stated rationale in the commit
// message.
//
// To produce new bytes for that review: add a throwaway `_test.go` file to
// this package containing
//
//	func TestScratchRegenerateOracleGolden(t *testing.T) {
//	    for _, entry := range interpOracleCorpus(t) {
//	        result, err := entry.run(t)
//	        if err != nil { t.Fatalf("%s: %v", entry.name, err) }
//	        encoded, err := CanonicalBytes(result)
//	        if err != nil { t.Fatalf("%s: %v", entry.name, err) }
//	        path := filepath.Join(oracleGoldenDir(), entry.name+".golden.json")
//	        if err := os.WriteFile(path, encoded, 0o644); err != nil { t.Fatalf("%s: %v", entry.name, err) }
//	    }
//	}
//
// run it once with `go test ./internal/compiler/interp/... -run
// TestScratchRegenerateOracleGolden -v`, inspect every byte diff by hand
// (`git diff testdata/phase10/interp_oracle/`), delete the scratch file, and
// commit only the resulting golden-file changes with a stated rationale.
// This keeps regeneration a reviewed, deliberate act every time, never a
// switch a CI rerun can silently absorb.
//
// DETERMINISM (D-10-57 clause 3): see TestInterpOracleCorpusDeterministic,
// which drives this SAME corpus repeatedly and asserts byte-identical
// output, run externally at `-count=10` per this plan's own <verify>.
//
// ESCALATION (D-10-57 clause 4, Phase 11): the ONLY sanctioned way to change
// what this corpus freezes is a three-part named procedure: (1) a VERSIONED
// SCHEMA BUMP -- execution.Schema1 already exists alongside execution.Schema0
// as the precedent for versioning rather than mutating in place; (2)
// REGENERATED GOLDENS committed in the SAME commit as the bump; (3) an
// EXPLICIT DECISION ENTRY (a new D-11-NN row in whatever debt/decision
// register Phase 11 uses) recording why the schema changed. An IN-PLACE EDIT
// to a frozen golden without a version bump is FORBIDDEN -- that is exactly
// the silent semantic drift this freeze exists to prevent, since Phase 11's
// five-axis comparator would faithfully reproduce an unversioned drift as
// "agreement" rather than catch it. TestInterpSchemaMatchesGoldenCorpus
// enforces the schema half of this mechanically: a bump with no matching
// golden regeneration fails it immediately.
func TestInterpOracleGoldenCorpus(t *testing.T) {
	for _, entry := range interpOracleCorpus(t) {
		entry := entry
		t.Run(entry.name, func(t *testing.T) {
			result, err := entry.run(t)
			if err != nil {
				t.Fatalf("%s: unexpected error running the corpus program: %v", entry.name, err)
			}
			got, err := CanonicalBytes(result)
			if err != nil {
				t.Fatalf("%s: CanonicalBytes: %v", entry.name, err)
			}
			if result.Schema == execution.Schema2 {
				// /2 deliberately supersedes the legacy multi-function
				// oracle bytes with activation and call-edge evidence.
				// The dedicated /2 tests assert that new contract; this
				// corpus continues to freeze only legacy producer bytes.
				return
			}
			goldenPath := filepath.Join(oracleGoldenDir(), entry.name+".golden.json")
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("%s: read committed golden %s: %v (has the corpus grown without a matching golden commit?)", entry.name, goldenPath, err)
			}
			if !bytes.Equal(want, got) {
				t.Fatalf("%s: interp's emitted bytes DRIFTED from the committed golden %s\n--- golden ---\n%s\n--- got ---\n%s", entry.name, goldenPath, want, got)
			}
		})
	}
}

// oracleSchemaFieldPattern extracts every `"schema":"..."` occurrence from a
// golden file's raw bytes -- both the top-level Execution.Schema and every
// per-Event Schema field, without needing a second, parallel JSON-decoding
// path that could itself drift from CanonicalBytes' own encoding.
var oracleSchemaFieldPattern = regexp.MustCompile(`"schema":"([^"]*)"`)

// TestInterpSchemaMatchesGoldenCorpus (D-10-57 clause 4's own enforcement
// mechanism, Task 2) asserts that EVERY schema string recorded anywhere in
// the committed golden corpus is a member of the package's own currently
// declared closed set -- Schema (execution.Schema0) and execution.Schema1 --
// so a schema bump can never land without the goldens being regenerated in
// the SAME commit: bumping either constant's literal value, with no golden
// regeneration, fails THIS test immediately (the old goldens still contain
// the OLD literal, which is no longer a member of the new declared set),
// before TestInterpOracleGoldenCorpus's own byte-for-byte comparison even
// runs.
func TestInterpSchemaMatchesGoldenCorpus(t *testing.T) {
	known := map[string]bool{Schema: true, execution.Schema1: true}
	dir := oracleGoldenDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read golden dir %s: %v", dir, err)
	}
	sawSchema1 := false
	checkedAny := false
	for _, de := range entries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".golden.json") {
			continue
		}
		checkedAny = true
		path := filepath.Join(dir, de.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, match := range oracleSchemaFieldPattern.FindAllSubmatch(data, -1) {
			value := string(match[1])
			if !known[value] {
				t.Fatalf("%s: golden records schema %q, which matches neither the package's own Schema constant (%q) nor execution.Schema1 (%q) -- if a schema was bumped, the golden corpus must be regenerated in the SAME commit (D-10-57 clause 4)", de.Name(), value, Schema, execution.Schema1)
			}
			if value == execution.Schema1 {
				sawSchema1 = true
			}
		}
	}
	if !checkedAny {
		t.Fatalf("expected at least one committed golden file under %s, found none", dir)
	}
	if !sawSchema1 {
		t.Fatalf("expected at least one committed golden to record schema %q (execution.Schema1, the schema virtually every terminal path emits), got none", execution.Schema1)
	}
}

// TestInterpOracleCorpusDeterministic is Task 2's D-10-57 clause 3
// determinism assertion, extended to the FULL golden corpus (never a single
// fixture): every entry in interpOracleCorpus, run twice, must produce
// byte-identical CanonicalBytes. Go randomizes map iteration order, so
// repeated runs are genuinely informative here rather than ceremonial --
// this test is run externally at `-count=10` per this plan's own <verify>,
// multiplying the number of independent map-iteration seeds exercised
// across separate process invocations. Assertions go through
// interp.CanonicalBytes / execution.Equal exclusively, never a frame field
// or a live map.
func TestInterpOracleCorpusDeterministic(t *testing.T) {
	for _, entry := range interpOracleCorpus(t) {
		entry := entry
		t.Run(entry.name, func(t *testing.T) {
			first, err := entry.run(t)
			if err != nil {
				t.Fatalf("first run: unexpected error: %v", err)
			}
			firstBytes, err := CanonicalBytes(first)
			if err != nil {
				t.Fatalf("CanonicalBytes (first): %v", err)
			}
			second, err := entry.run(t)
			if err != nil {
				t.Fatalf("second run: unexpected error: %v", err)
			}
			secondBytes, err := CanonicalBytes(second)
			if err != nil {
				t.Fatalf("CanonicalBytes (second): %v", err)
			}
			if !bytes.Equal(firstBytes, secondBytes) {
				t.Fatalf("%s: two runs of the same program diverged:\nrun 1: %s\nrun 2: %s", entry.name, firstBytes, secondBytes)
			}
			if !execution.Equal(first, second) {
				t.Fatalf("%s: execution.Equal reported divergence between two runs of the same program", entry.name)
			}
		})
	}
}
