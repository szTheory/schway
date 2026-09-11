package interp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
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
				base := newFlatFrame(entry, map[string]string{entry.Parameter.ID: "err_value"})
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
