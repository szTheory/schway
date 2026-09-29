package session

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
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
	"inline_across_foreign.schway",
	"dead_store_unused_acquire.schway",
	"reorder_two_events.schway",
	"tail_collapse_release_ladder.schway",
	"typed_failure_truncated_stdout.schway",
	"defect_dies_by_signal.schway",
}

// phase5AdversarialTargetHeader is the exact header comment form every
// Phase 5 adversarial fixture must open with.
func phase5AdversarialTargetHeader(target string) string {
	return fmt.Sprintf("// adversarial-target: %s", target)
}

// Phase5MilestoneCorpus returns the union corpus D-05-18 defines: every
// `.schway` fixture under testdata/phase1, testdata/phase2, testdata/phase3,
// testdata/phase4, and testdata/phase5, named by path relative to corpusRoot.
// Prior-phase members are included BY PATH, never copied or rewritten
// (D-05-18's own "the union grows, it never rewrites" must-have).
func Phase5MilestoneCorpus(corpusRoot string) ([]string, error) {
	var paths []string
	for _, phase := range []string{"phase1", "phase2", "phase3", "phase4", "phase5"} {
		dir := filepath.Join(corpusRoot, phase)
		matches, err := filepath.Glob(filepath.Join(dir, "*.schway"))
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
	// Phase 11 (11-GUARD-LEDGER.md): KEPT. D-05-18b's own enumerated
	// closure grammar is scoped to "one parameter" straight-line/branch
	// programs by definition -- this is a single-function generator, not
	// a guard that could admit the multi-function corpus this phase
	// widens, so there is nothing to widen here.
	if len(program.Functions) != 1 {
		return core.Program{}, false
	}
	return program, true
}

// phase5ChainKinds is the ownership-operation alphabet D-05-18b's
// enumerated closure draws from for its straight-line candidates:
// borrow/take are two of the four "only fallible consumers" D-05-18b names
// (the other two, try/discard, are exercised separately by
// phase5ForeignChainSource). `borrow mut` is included as the exclusive
// counterpart to `borrow`'s shared loan, matching the same alphabet
// restrict_borrow.schway's own chain uses (05-01-SUMMARY.md).
var phase5ChainKinds = []string{"borrow", "borrow mut", "take"}

// phase5OwnershipChainSource generates one straight-line candidate source
// text: a chain of exactly length bindings over paramType, each binding's
// kind selected by decoding encoded as a base-len(phase5ChainKinds) integer
// (the same deterministic base-N encoding idiom check_test.go's own
// TestOwnershipSequenceExhaustive/generatedOwnershipBody uses), so the same
// (paramType, length, encoded) triple always yields byte-identical source.
// Every step chains onto the immediately preceding binding, never back to
// the parameter or an earlier binding, so the checker alone -- never this
// generator -- decides which chains are legal (e.g. `take` of an
// already-borrowed, not-owned place is refused, not filtered out here).
// touchParamAgain, when true, appends one further binding that re-reads
// the ORIGINAL parameter (`let extra = borrow x`) after the whole chain,
// before the result line. This is the generator's own genuine-rejection
// axis: if the chain's first step moved x away (kind "take"), re-touching
// x here is a real ownership.use_after_move the checker must refuse, so
// EnumeratePhase5Closure's rejected count is never vacuously zero
// (T-05-17) -- a candidate is refused by the checker for a real, checkable
// reason, not by any filter this generator applies itself.
func phase5OwnershipChainSource(paramType string, length, encoded int, touchParamAgain bool) string {
	var body strings.Builder
	fmt.Fprintf(&body, "module phase5.enum_chain_%s_l%d_e%d_t%v\n\n", strings.ToLower(paramType), length, encoded, touchParamAgain)
	body.WriteString("export {\n  fn f\n}\n\n")
	fmt.Fprintf(&body, "fn f(x: %s) -> %s {\n", paramType, paramType)
	prev := "x"
	remaining := encoded
	for step := 0; step < length; step++ {
		kind := phase5ChainKinds[remaining%len(phase5ChainKinds)]
		remaining /= len(phase5ChainKinds)
		name := fmt.Sprintf("v%d", step)
		fmt.Fprintf(&body, "  let %s = %s %s\n", name, kind, prev)
		prev = name
	}
	if touchParamAgain {
		body.WriteString("  let extra = borrow x\n")
	}
	fmt.Fprintf(&body, "  %s\n}\n", prev)
	return body.String()
}

// phase5ForeignChainShapes and phase5ForeignAlternativeCounts are the two
// enumeration axes phase5ForeignChainSource combines: a fixed set of
// try/discard chain shapes (peering testdata/phase4's own acquire_three_*
// and discard_because.schway shapes) crossed with a declared failure ADT of
// either 1 or 2 alternatives (D-05-18b's "at most 2 ADT alternatives").
// Only single-call shapes are enumerated here: D-05-18b's own grammar
// bounds a candidate to "0 or 1 foreign call" per function, and try2/try3/
// mixed multi-stage chains (each with 2+ OpForeignCall operations) belong
// to the hand-written adversarial subset's tail_collapse_release_ladder.schway
// and reorder_two_events.schway instead (Task 1), not this bounded closure.
var phase5ForeignChainShapes = []string{"try1", "discard1"}
var phase5ForeignAlternativeCounts = []int{1, 2}

// phase5ForeignChainSource generates one resource-lifecycle candidate: a
// single declared foreign symbol called through the named shape's
// try/discard sequence, with the function returning its own parameter --
// checkFallibleLinear's resource-lifecycle shape (D-05-18b's "0 or 1
// foreign call" axis). The declared symbol is `schway_res_open` -- the SAME
// real, frozen, byte-identical production symbol every Phase 4 foreign
// fixture declares -- rather than a made-up name, because
// native.ForeignSourcePathForSymbol only resolves a fixed, closed set of
// real frozen translation units (D-04-10/D-04-17): a synthetic symbol name
// would check-admit but never link natively, which would make Task 3's
// three-engine agreement run over this closure vacuous for every foreign
// candidate.
func phase5ForeignChainSource(shape string, alternativeCount int) string {
	var body strings.Builder
	fmt.Fprintf(&body, "module phase5.enum_foreign_%s_alt%d\n\n", shape, alternativeCount)
	body.WriteString("export {\n  fn main\n}\n\n")
	body.WriteString("foreign C {\n\n  fn schway_res_open(request: Byte) -> Byte {\n    unwind: forbidden\n    nonlocal_exit: forbidden\n    allocator: \"libc_malloc\"\n    fails: AcquireError\n  }\n}\n\n")
	body.WriteString("data AcquireError =\n  | OpenFailed\n")
	if alternativeCount >= 2 {
		body.WriteString("  | OpenFailedAgain\n")
	}
	body.WriteString("\nfn main(request: Byte) -> Byte {\n")
	switch shape {
	case "try1":
		body.WriteString("  let a = try schway_res_open(request)\n")
	case "discard1":
		body.WriteString("  discard schway_res_open(request) because \"advisory\"\n")
	}
	body.WriteString("  request\n}\n")
	return body.String()
}

// enumeratePhase5Closure is the single computation both EnumeratePhase5Closure
// and EnumeratePhase5ClosureRejected read from: deterministic (no
// randomness, no map iteration in the generation path), so two calls in the
// same process, or two calls across separate `go test` runs, produce
// byte-identical accepted-program sequences and an identical rejected
// count (TestPhase5EnumerationIsDeterministic).
func enumeratePhase5Closure() ([]core.Program, int) {
	var accepted []core.Program
	rejected := 0
	for _, paramType := range []string{"Byte", "Buffer"} {
		for length := 0; length <= Phase5EnumerationMaxDepth; length++ {
			combos := 1
			for step := 0; step < length; step++ {
				combos *= len(phase5ChainKinds)
			}
			for encoded := 0; encoded < combos; encoded++ {
				for _, touchParamAgain := range []bool{false, true} {
					source := phase5OwnershipChainSource(paramType, length, encoded, touchParamAgain)
					if program, ok := admitPhase5Candidate(source); ok {
						accepted = append(accepted, program)
					} else {
						rejected++
					}
				}
			}
		}
	}
	for _, alternativeCount := range phase5ForeignAlternativeCounts {
		for _, shape := range phase5ForeignChainShapes {
			source := phase5ForeignChainSource(shape, alternativeCount)
			if program, ok := admitPhase5Candidate(source); ok {
				accepted = append(accepted, program)
			} else {
				rejected++
			}
		}
	}
	return accepted, rejected
}

// EnumeratePhase5Closure is D-05-18b's bounded enumerated closure: every
// admissible candidate this file's two generators produce, over exactly
// the grammar slice D-05-18b names (one parameter, Byte | Buffer, at most 2
// ADT alternatives, straight-line or single-level-branch body, 0 or 1
// foreign call, borrow/take/try/discard as the only fallible consumers)
// and bounded by Phase5EnumerationMaxDepth/Phase5EnumerationMaxStatements.
// A candidate the checker or corevalidate refuses is never included here --
// see EnumeratePhase5ClosureRejected for its count.
func EnumeratePhase5Closure() []core.Program {
	accepted, _ := enumeratePhase5Closure()
	return accepted
}

// EnumeratePhase5ClosureRejected reports how many generated candidates
// EnumeratePhase5Closure's own admission gate refused -- exposed
// specifically so a silently-empty (or silently-all-rejected) enumeration
// can never be mistaken for a clean run (T-05-17, D-05-18b).
func EnumeratePhase5ClosureRejected() int {
	_, rejected := enumeratePhase5Closure()
	return rejected
}
