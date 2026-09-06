package session_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// coordinatedLieSourcePath/coordinatedLieCorePath name D-05-30's adversarial
// artifact pair (plan 05-13): an ordinary, checker-accepted Lang source
// program and a HAND-AUTHORED core artifact naming the exact same
// module/function identity but claiming a different ownership-transfer
// outcome. Neither file is derived from the other -- that absence of a real
// translation step is the whole point of the demonstration.
func coordinatedLieSourcePath() string {
	return testsupport.ProjectPath("testdata", "phase5", "coordinated_lie.lang")
}

func coordinatedLieCorePath() string {
	return testsupport.ProjectPath("testdata", "phase5", "coordinated_lie.core.json")
}

// loadCoordinatedLieCore reads and decodes the hand-authored core artifact.
// It performs no validation of its own -- callers run corevalidate.Validate
// explicitly, exactly as they would for any other independently-produced
// core.Program.
func loadCoordinatedLieCore(t *testing.T) core.Program {
	t.Helper()
	raw, err := os.ReadFile(coordinatedLieCorePath())
	if err != nil {
		t.Fatalf("reading coordinated_lie.core.json: %v", err)
	}
	var program core.Program
	if err := json.Unmarshal(raw, &program); err != nil {
		t.Fatalf("decoding coordinated_lie.core.json: %v", err)
	}
	return program
}

// checkCoordinatedLieSource runs the SOURCE half of the pair through
// check+originvalidate exactly as any other Lang program would be admitted,
// returning the checker-derived core.Program (the source's own honest
// claim).
func checkCoordinatedLieSource(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(coordinatedLieSourcePath())
	if err != nil {
		t.Fatalf("reading coordinated_lie.lang: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("coordinated_lie.lang: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("coordinated_lie.lang: originvalidate.ValidatePublished rejected the source-derived program: %+v", problems)
	}
	return checked.Program
}

// TestCoordinatedLieArtifactsBothValidate is D-05-30's reachability
// demonstration: coordinated_lie.lang independently passes check and
// originvalidate.ValidatePublished on its own terms, coordinated_lie.core.json
// independently passes corevalidate.Validate on ITS own terms, and the two
// artifacts -- which name the identical module/function identity -- assert
// DIFFERENT ownership-transfer claims about the same buffer parameter: the
// source moves it (full ownership transfer, never returning access to the
// original place), the hand-authored core only ever shares it (a borrow,
// never a move). Neither validator ever compares one artifact against the
// other; asserting that disagreement explicitly here is the whole point of
// the demonstration.
func TestCoordinatedLieArtifactsBothValidate(t *testing.T) {
	sourceProgram := checkCoordinatedLieSource(t)

	coreProgram := loadCoordinatedLieCore(t)
	validated := corevalidate.Validate(coreProgram)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected the hand-authored coordinated_lie.core.json on its own terms: %+v", validated.Problems)
	}

	if len(sourceProgram.Functions) != 1 || sourceProgram.Functions[0].Linear == nil || len(sourceProgram.Functions[0].Linear.Operations) == 0 {
		t.Fatalf("coordinated_lie.lang: expected exactly one straight-line function with at least one operation, got %+v", sourceProgram.Functions)
	}
	if len(coreProgram.Functions) != 1 || coreProgram.Functions[0].Linear == nil || len(coreProgram.Functions[0].Linear.Operations) == 0 {
		t.Fatalf("coordinated_lie.core.json: expected exactly one straight-line function with at least one operation, got %+v", coreProgram.Functions)
	}

	sourceModuleID := sourceProgram.ModuleID
	coreModuleID := coreProgram.ModuleID
	if sourceModuleID != coreModuleID {
		t.Fatalf("expected the pair to name the SAME module identity (that is what makes this a coordinated lie, not just two unrelated programs): source=%q core=%q", sourceModuleID, coreModuleID)
	}

	sourceClaim := sourceProgram.Functions[0].Linear.Operations[0].Kind
	coreClaim := coreProgram.Functions[0].Linear.Operations[0].Kind
	if sourceClaim != core.OpMove {
		t.Fatalf("expected coordinated_lie.lang's own checker-derived claim to be a move (full ownership transfer), got %q", sourceClaim)
	}
	if coreClaim != core.OpBorrowShared {
		t.Fatalf("expected coordinated_lie.core.json's hand-authored claim to be a shared borrow (never a move), got %q", coreClaim)
	}
	if sourceClaim == coreClaim {
		t.Fatalf("expected the source-derived claim and the hand-authored core claim to DIFFER -- both were %q, which is not a coordinated lie, just an accurate translation", sourceClaim)
	}
}
