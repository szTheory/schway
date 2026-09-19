package cgen

import (
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestN1ConvergenceDifferential is D-11-02's Q-05 gate and D-12-31 step (i):
// it drives the legacy single-function family (Emit's own dispatch, lines
// 29-92 of cgen.go) and emitProgram (cgen_program.go) over the SAME checked
// program, for every single-function emitter shape the legacy family
// selects between, and asserts the CURRENT, MEASURED state of each.
//
// A passing run today means the two families have NOT converged: on four of
// five shapes emitProgram refuses outright, and on the fifth (a plain linear
// function with no branch/match, no foreign-call blocks, and no by-pointer
// lowering) both paths succeed but their C output differs -- a call-boundary
// attribute comment and an extra #include shift the preamble, and
// emitProgram writes a static function plus a separate main where emitLinear
// writes one inline main.
//
// When convergence work lands (porting branch bodies, foreign-call block
// bodies, and both by-pointer lowering variants into emitProgram's family,
// converging the preamble, and re-pinning the four frozen golden-C digests),
// this test's expectations FLIP to byte-identity across all five shapes --
// the test is never deleted, only its recorded expectations change, per
// PHASE-11-DEBT.md's `### D-11-02` section (the six single-function
// emitters are not deleted this phase; their deletion is gated on this
// differential going green).
//
// This is an internal test file in package cgen (not cgen_test) so
// emitProgram is directly reachable without exporting it -- mirroring
// export_test.go's existing test-only-seam precedent, but here no new seam
// is needed since emitProgram is already an unexported package-level func.
// It does NOT import internal/compiler/session (an import cycle from this
// package); instead it reproduces session.Check's two-stage pipeline
// directly via syntax.Parse + check.Program, exactly as the planner's own
// throwaway probe did.
func TestN1ConvergenceDifferential(t *testing.T) {
	type expectation struct {
		fixture string
		// legacyOK/programOK record whether each path succeeds (true) or
		// returns a non-nil error (false) on this fixture.
		legacyOK  bool
		programOK bool
		// identical is only meaningful when both legacyOK and programOK are
		// true: it records whether their two C outputs are byte-identical.
		identical bool
	}

	table := []expectation{
		// match-only: emitProgram refuses -- multi-function branch bodies
		// are not supported by native emission this phase
		// (cgen_program.go's core.Match refusal).
		{fixture: "testdata/phase1/toggle.lang", legacyOK: true, programOK: false},
		// plain linear: BOTH paths succeed, but their outputs DIFFER --
		// this is the one shape where a byte-level comparison is even
		// meaningful today.
		{fixture: "testdata/phase2/owned_transfer.lang", legacyOK: true, programOK: true, identical: false},
		// branch/match+linear: emitProgram now admits arm-local operations
		// through the shared schema-2 event writer. The legacy /1 document
		// still differs, so this remains a convergence characterization.
		{fixture: "testdata/phase3/borrowed_view.lang", legacyOK: true, programOK: true, identical: false},
		// foreign-call blocks: emitProgram refuses -- multi-function
		// foreign-call bodies are not supported by native emission this
		// phase.
		{fixture: "testdata/phase4/foreign_acquire_one.lang", legacyOK: true, programOK: false},
		// branch with a defect terminator: emitProgram refuses -- reaches
		// the same core.Match refusal as toggle.lang and borrowed_view.lang
		// (every defect terminator in this codebase is reached through a
		// core.Match arm at this maturity, per PHASE-11-DEBT.md D-11-52).
		{fixture: "testdata/phase4/defect_terminal.lang", legacyOK: true, programOK: false},
	}

	if len(table) != 5 {
		t.Fatalf("expected exactly 5 fixtures in the N=1 convergence table, got %d", len(table))
	}

	for _, row := range table {
		row := row
		t.Run(row.fixture, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(row.fixture))
			if err != nil {
				t.Fatalf("read fixture %s: %v", row.fixture, err)
			}

			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 {
				t.Fatalf("fixture %s failed to parse: %+v", row.fixture, parsed.Diagnostics)
			}

			checkResult := check.Program(parsed.Program)
			if len(checkResult.Diagnostics) != 0 {
				t.Fatalf("fixture %s failed to check: %+v", row.fixture, checkResult.Diagnostics)
			}
			program := checkResult.Program

			legacyOutput, legacyErr := Emit(program)
			legacyOK := legacyErr == nil
			if legacyOK != row.legacyOK {
				t.Fatalf("fixture %s: legacy Emit success=%v (err=%v), expected success=%v", row.fixture, legacyOK, legacyErr, row.legacyOK)
			}

			programOutput, programErr := emitProgram(program, false)
			programOK := programErr == nil
			if programOK != row.programOK {
				t.Fatalf("fixture %s: emitProgram success=%v (err=%v), expected success=%v", row.fixture, programOK, programErr, row.programOK)
			}

			if row.legacyOK && row.programOK {
				identical := legacyOutput == programOutput
				if identical != row.identical {
					t.Fatalf("fixture %s: legacy/emitProgram output identical=%v, expected identical=%v\n--- legacy ---\n%s\n--- emitProgram ---\n%s", row.fixture, identical, row.identical, legacyOutput, programOutput)
				}
			}
		})
	}
}
