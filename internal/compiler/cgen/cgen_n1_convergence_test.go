package cgen

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
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
// A passing run before Plan 16-09 means the two families have NOT converged:
// emitProgram now admits match and defect bodies as well as ordinary linear
// programs, but it serializes lang.execution/2 while the legacy N=1 routes
// serialize lang.execution/1. The byte difference is therefore measured
// provenance, not an unfinished identity assertion. Plan 16-09 owns the
// atomic dispatch flip after which the relation must become byte identity.
// Foreign-call and pointer-specialized bodies remain named refusals under the
// separately-owned M004 cut.
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
		fixture            string
		legacyEmitSchema   string
		legacyNativeSchema string
		// legacyOK/programOK record whether each path succeeds (true) or
		// returns a non-nil error (false) on this fixture.
		legacyOK  bool
		programOK bool
		// preCutSchemaDivergence requires legacy /1 and direct-program /2
		// provenance. It deliberately refuses byte identity until Plan 16-09.
		preCutSchemaDivergence bool
	}

	table := []expectation{
		// match-only: emitProgram admits this through the schema-2 program
		// writer, while the legacy output remains different.
		{fixture: "testdata/phase1/toggle.lang", legacyNativeSchema: "lang.execution/0", legacyOK: true, programOK: true, preCutSchemaDivergence: true},
		// plain linear: BOTH paths succeed, but their outputs DIFFER --
		// this is the one shape where a byte-level comparison is even
		// meaningful today.
		{fixture: "testdata/phase2/owned_transfer.lang", legacyEmitSchema: "lang.execution/1", legacyNativeSchema: "lang.execution/1", legacyOK: true, programOK: true, preCutSchemaDivergence: true},
		// branch/match+linear: emitProgram now admits arm-local operations
		// through the shared schema-2 event writer. The legacy /1 document
		// still differs, so this remains a convergence characterization.
		{fixture: "testdata/phase3/borrowed_view.lang", legacyEmitSchema: "lang.execution/1", legacyNativeSchema: "lang.execution/1", legacyOK: true, programOK: true, preCutSchemaDivergence: true},
		// foreign-call blocks: emitProgram refuses -- multi-function
		// foreign-call bodies are not supported by native emission this
		// phase.
		{fixture: "testdata/phase4/foreign_acquire_one.lang", legacyOK: true, programOK: false},
		// branch with a defect terminator: the program writer records its
		// schema-2 defect terminal before aborting, so it is admitted but
		// differs from the legacy document.
		{fixture: "testdata/phase4/defect_terminal.lang", legacyEmitSchema: "lang.execution/1", legacyNativeSchema: "lang.execution/1", legacyOK: true, programOK: true, preCutSchemaDivergence: true},
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

			modes := []struct {
				name         string
				legacy       func() (string, error)
				directJSON   bool
				legacySchema string
			}{
				{name: "Emit", legacy: func() (string, error) { return Emit(program) }, legacySchema: row.legacyEmitSchema},
				{name: "EmitNative", legacy: func() (string, error) { return EmitNative(program) }, directJSON: true, legacySchema: row.legacyNativeSchema},
			}
			for _, mode := range modes {
				mode := mode
				t.Run(mode.name, func(t *testing.T) {
					legacyOutput, legacyErr := mode.legacy()
					legacyOK := legacyErr == nil
					if legacyOK != row.legacyOK {
						t.Fatalf("fixture=%s mode=%s legacy success=%v (err=%v), expected success=%v", row.fixture, mode.name, legacyOK, legacyErr, row.legacyOK)
					}

					programOutput, programErr := emitProgram(program, mode.directJSON)
					programOK := programErr == nil
					if programOK != row.programOK {
						t.Fatalf("fixture=%s mode=%s direct emitProgram success=%v (err=%v), expected success=%v", row.fixture, mode.name, programOK, programErr, row.programOK)
					}

					if !row.legacyOK || !row.programOK {
						return
					}
					legacyDigest := sha256.Sum256([]byte(legacyOutput))
					directDigest := sha256.Sum256([]byte(programOutput))
					if row.preCutSchemaDivergence {
						legacyProvenance := strings.Contains(legacyOutput, mode.legacySchema)
						if mode.legacySchema == "" {
							legacyProvenance = !strings.Contains(legacyOutput, "lang.execution/")
						}
						if legacyOutput == programOutput || !legacyProvenance || !strings.Contains(programOutput, "lang.execution/2") {
							reason := "legacy source-mode plain output vs direct lang.execution/2"
							if mode.legacySchema != "" {
								reason = "legacy " + mode.legacySchema + " vs direct lang.execution/2"
							}
							t.Fatalf("fixture=%s mode=%s: expected pre-cut %s byte divergence; legacy_sha256=%s direct_sha256=%s\n--- legacy ---\n%s\n--- direct emitProgram ---\n%s", row.fixture, mode.name, reason, hex.EncodeToString(legacyDigest[:]), hex.EncodeToString(directDigest[:]), legacyOutput, programOutput)
						}
					}
				})
			}
		})
	}
}
