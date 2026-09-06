package session

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// Phase5CorpusBoundVersion, Phase5EnumerationMaxDepth, and
// Phase5EnumerationMaxStatements are D-05-18's versioned bound constants:
// the exact scope statement the Phase 5 corpus's enumerated closure (part
// (b) of D-05-18) is asserted against. scripts/verify-phase5.sh (plan
// 05-09) duplicates all three verbatim and TestPhase5CorpusBoundConstantsAreExported
// asserts they exist with these exact names, mirroring the
// Phase4RequiredControls()/verify-phase4.sh duplication-plus-equality-test
// pattern (session.go:2004-2019) so corpus scope cannot drift silently any
// more than a control set can.
const (
	Phase5CorpusBoundVersion       = 1
	Phase5EnumerationMaxDepth      = 3
	Phase5EnumerationMaxStatements = 4
)

// Phase5AdversarialTargets names the exact six named `-O3`/LTO
// transformations D-05-18a's hand-written adversarial subset engineers one
// program per, in the exact spelling each fixture's own
// `// adversarial-target: <name>` header comment carries. Plan 05-07's
// D-05-22 assertion and the corpus manifest both key on these exact
// strings.
var Phase5AdversarialTargets = []string{
	"inline-across-foreign-call",
	"dead-store-elimination-of-unused-borrowed-acquire",
	"reordering-of-two-independent-event-emissions",
	"tail-collapse-of-a-release-ladder",
	"typed-failure-with-truncated-stdout",
	"defect-terminating-by-signal",
}

// Phase5AdversarialFixtureFiles is the fixed, deterministic file list
// backing Phase5AdversarialTargets, in the same declared order -- fixture
// N's header must carry Phase5AdversarialTargets[N] verbatim
// (TestPhase5AdversarialSubsetIsComplete asserts this pairing).
var Phase5AdversarialFixtureFiles = []string{
	"inline_across_foreign.lang",
	"dead_store_unused_acquire.lang",
	"reorder_two_events.lang",
	"tail_collapse_release_ladder.lang",
	"typed_failure_truncated_stdout.lang",
	"defect_dies_by_signal.lang",
}

// phase5AdversarialTargetHeader is the exact header comment form every
// Phase 5 adversarial fixture must open with.
func phase5AdversarialTargetHeader(target string) string {
	return fmt.Sprintf("// adversarial-target: %s", target)
}

// Phase5MilestoneCorpus returns the union corpus D-05-18 defines: every
// `.lang` fixture under testdata/phase1, testdata/phase2, testdata/phase3,
// testdata/phase4, and testdata/phase5, named by path relative to corpusRoot.
// Prior-phase members are included BY PATH, never copied or rewritten
// (D-05-18's own "the union grows, it never rewrites" must-have).
func Phase5MilestoneCorpus(corpusRoot string) ([]string, error) {
	var paths []string
	for _, phase := range []string{"phase1", "phase2", "phase3", "phase4", "phase5"} {
		dir := filepath.Join(corpusRoot, phase)
		matches, err := filepath.Glob(filepath.Join(dir, "*.lang"))
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", dir, err)
		}
		paths = append(paths, matches...)
	}
	sort.Strings(paths)
	return paths, nil
}

// admitPhase5Candidate parses and checks source exactly as session.Check
// does, additionally requiring independent corevalidate.Validate agreement
// (Phase4CheckedProgram's own admission bar), returning the validated
// single-function core.Program on success. This is the fail-closed
// admission gate EnumeratePhase5Closure filters every generated candidate
// through: a candidate the checker or corevalidate refuses is counted
// (EnumeratePhase5ClosureRejected), never silently dropped.
func admitPhase5Candidate(source string) (core.Program, bool) {
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		return core.Program{}, false
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		return core.Program{}, false
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return core.Program{}, false
	}
	program := validated.Program()
	if len(program.Functions) != 1 {
		return core.Program{}, false
	}
	return program, true
}
