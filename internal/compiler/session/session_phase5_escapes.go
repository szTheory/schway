package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
)

// EscapeCoordinatedSourceToCoreFalseClaim names D-05-30's decisive addition
// over D-04-31 item 1's prose-only record: a real adversarial artifact pair
// (testdata/phase5/coordinated_lie.schway + coordinated_lie.core.json) that
// ACTUALLY constructs matching-but-false source/core artifacts and is
// asserted to pass the gate under this named escape. It reuses
// corevalidate.KnownEscape's own wording for the exact boundary being
// named -- a producer that coordinates a false source claim with
// internally consistent core facts remains outside source-blind
// validation -- applied here at the source<->core trust crossing exactly
// as originvalidate.KnownEscape already applies the identical wording at
// the origin<->summary trust crossing.
//
// This phase makes the escape DEMONSTRABLE, not closed. Demonstrating an
// escape is not closing it, and no control in this repository closes a
// coordinated source-to-core lie: corevalidate.KnownEscape and
// originvalidate.KnownEscape are REFERENCED here, never redefined or
// replaced (D-05-30's explicit reuse requirement).
const EscapeCoordinatedSourceToCoreFalseClaim = "escape:coordinated-source-to-core-false-claim"

// LaneCoordinatedLieEscape identifies VerifyCoordinatedLieEscape's own
// LaneResult, mirroring LaneQLT01RegistryAudit's naming convention
// (qlt01.go).
const LaneCoordinatedLieEscape = "lane:coordinated-lie-escape"

// coordinatedLieSourceCorpusFile and coordinatedLieCoreCorpusFile are the
// D-05-30 adversarial artifact pair's repo-relative paths, resolved via
// nat03CorpusPath (session_phase5_alias.go) exactly as this file's sibling
// Phase 5 lanes resolve their own fixtures.
const (
	coordinatedLieSourceCorpusFile = "testdata/phase5/coordinated_lie.schway"
	coordinatedLieCoreCorpusFile   = "testdata/phase5/coordinated_lie.core.json"
)

// VerifyCoordinatedLieEscape runs the full gate over the D-05-30
// adversarial artifact pair: coordinated_lie.schway independently through
// check + originvalidate.ValidatePublished (exactly the admission path any
// other Lang source takes), and the HAND-AUTHORED coordinated_lie.core.json
// independently through corevalidate.Validate on its own terms. Neither
// validator is ever asked to compare the two artifacts against each
// other -- that absence is corevalidate.KnownEscape's own documented
// boundary, not an oversight here.
//
// If either artifact is refused by its own validator, this is a genuine
// finding (the demonstration failed to construct a reachable escape as
// claimed) and VerifyCoordinatedLieEscape returns an error naming which
// validator refused and why -- never a silent pass. On success, the
// returned LaneResult's ExpectedEscapes names
// EscapeCoordinatedSourceToCoreFalseClaim explicitly: the pass is
// attributed to a declared, gate-visible residual, not to an absence of
// checking.
func VerifyCoordinatedLieEscape(ctx context.Context) (LaneResult, error) {
	_ = ctx

	work := 0

	source, err := os.ReadFile(nat03CorpusPath(coordinatedLieSourceCorpusFile))
	if err != nil {
		return LaneResult{}, fmt.Errorf("reading %s: %w", coordinatedLieSourceCorpusFile, err)
	}
	checked := Check(source)
	work++ // one unit for the source-side check admission
	if len(checked.Diagnostics) != 0 {
		return LaneResult{}, fmt.Errorf("%s: unexpected diagnostics, the adversarial source is not reachable as constructed: %+v", coordinatedLieSourceCorpusFile, checked.Diagnostics)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		return LaneResult{}, fmt.Errorf("%s: originvalidate.ValidatePublished refused the source-derived program, the adversarial source is not reachable as constructed: %+v", coordinatedLieSourceCorpusFile, problems)
	}
	work++ // one unit for the origin admission

	coreBytes, err := os.ReadFile(nat03CorpusPath(coordinatedLieCoreCorpusFile))
	if err != nil {
		return LaneResult{}, fmt.Errorf("reading %s: %w", coordinatedLieCoreCorpusFile, err)
	}
	var coreProgram core.Program
	if err := json.Unmarshal(coreBytes, &coreProgram); err != nil {
		return LaneResult{}, fmt.Errorf("decoding %s: %w", coordinatedLieCoreCorpusFile, err)
	}
	work++ // one unit for decoding the hand-authored core

	validated := corevalidate.Validate(coreProgram)
	work++ // one unit for the core-side corevalidate admission
	if !validated.Valid {
		fired := make(map[string]bool, len(validated.Problems))
		controls := make([]string, 0, len(validated.Problems))
		for _, problem := range validated.Problems {
			fired[problem.Code] = true
			controls = append(controls, problem.Code)
		}
		return LaneResult{
			ID:             LaneCoordinatedLieEscape,
			Status:         "fail",
			Controls:       controls,
			RecomputedWork: work,
			Fired:          fired,
		}, fmt.Errorf("%s: corevalidate refused the hand-authored core, the adversarial pair is not reachable as constructed: %+v", coordinatedLieCoreCorpusFile, validated.Problems)
	}

	// No control fired on either half of the pair (both artifacts passed
	// their own independent validator on their own terms). The pass is
	// attributed to the named escape, never left as a silent absence of
	// checking.
	return LaneResult{
		ID:              LaneCoordinatedLieEscape,
		Status:          "pass",
		Controls:        []string{},
		RecomputedWork:  work,
		Fired:           map[string]bool{},
		ExpectedEscapes: []string{EscapeCoordinatedSourceToCoreFalseClaim},
	}, nil
}
