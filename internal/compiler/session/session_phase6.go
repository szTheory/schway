package session

import (
	"context"
	"os"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cache"
	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

// Phase6RequiredControls is the complete Phase 6 required-control list
// (plan 06-15, mirroring Phase5RequiredControls' shape verbatim): a flat
// slice of exact identifier strings, one per line, no computation.
// scripts/verify-phase6.sh duplicates this list VERBATIM, and
// TestPhase6RequiredControlsMatchScript asserts the two are set-equal in
// both directions, so the gate cannot silently shrink by dropping a
// control from either copy.
//
// The first two identifiers are 06-05's risk-lane registry audit
// (session_phase6_risklanes.go). The next four are 06-16's QLT-02
// budget-manifest audit (session_phase6_budget.go). control:interpreter-o0-o3
// is 06-04's artifact-cache-backed native differential
// (verifyPhase6NativeDifferentialLane, session_phase6_verify.go), driven
// here over testdata/phase1 (the corpus carrying the shared toggle.lang
// fixture). The final five are DX-04's defect-injection exercise
// (session_phase6_injectors.go), one control per injector class, each
// proving that class's injector fires and produces a genuinely rejected
// mutation -- never a silent unmutated pass.
func Phase6RequiredControls() []string {
	return []string{
		ControlRiskLaneStaleLaneReference,
		ControlRiskLaneUndeclaredLane,
		ControlQLT02ManifestEmpty,
		ControlQLT02UnknownMachine,
		ControlQLT02DuplicateRow,
		ControlQLT02IneligibleHardGate,
		"control:interpreter-o0-o3",
		ControlDefectMatchInjection,
		ControlDefectMoveInjection,
		ControlDefectBorrowInjection,
		ControlDefectCleanupInjection,
		ControlDefectStaleEvidenceInjection,
	}
}

// Control identifiers for the DX-04 defect-injection exercise (D-06-25,
// D-06-27), one per injector class, following the risklanes/qlt02
// control-naming convention (control:<area>.<specific>).
const (
	ControlDefectMatchInjection         = "control:defect.match_injection"
	ControlDefectMoveInjection          = "control:defect.move_injection"
	ControlDefectBorrowInjection        = "control:defect.borrow_injection"
	ControlDefectCleanupInjection       = "control:defect.cleanup_injection"
	ControlDefectStaleEvidenceInjection = "control:defect.stale_evidence_injection"
)

// phase6NativeDifferentialCorpus is the corpus this gate drives
// control:interpreter-o0-o3 against: testdata/phase1, the corpus already
// carrying the shared toggle.lang fixture VerifyPhase6ChangedRisk's own
// lane reads (session_phase6_verify.go).
const phase6NativeDifferentialCorpus = "testdata/phase1"

// phase6CleanupFixture is the cleanup injector's source fixture: an
// already-shipped Phase 4 fixture whose generated C carries a real
// lang:release-site marker (README, testdata/phase6).
const phase6CleanupFixture = "testdata/phase4/acquire_three_success.lang"

// phase6DefaultRunner mirrors phase5DefaultRunner's shape for this file's
// own native-differential lane.
func phase6DefaultRunner() native.Runner {
	return native.DefaultRunner()
}

// VerifyPhase6ControlsAndWork is Phase 6's own control-and-work gate,
// mirroring VerifyPhase5ControlsAndWork's shape: it runs the risk-lane
// registry audit, the QLT-02 budget-manifest audit, the cache-backed
// native differential over a real corpus, and DX-04's five defect
// injectors -- collecting every control's fired status and
// RecomputedWork, and never folding an unreached required control into a
// silent pass.
func VerifyPhase6ControlsAndWork(ctx context.Context) (protocol.Result, error) {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	result.ExpectedEscapes = append([]string{}, Phase6ExpectedEscapes()...)

	addLane := func(id, status string, controls []string, work int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		})
		result.Metrics.RecomputedWork += work
	}
	markFail := func(status string) {
		if result.Status == protocol.StatusPass {
			result.Status = status
		}
	}

	// Lane: the risk-lane registry audit (D-06-10).
	riskStarted := time.Now()
	riskLane, riskErr := VerifyRiskLaneRegistry(ctx)
	if riskErr != nil {
		addLane(LaneRiskLaneRegistryAudit, protocol.StatusOperational, nil, 1, riskStarted)
		markFail(protocol.StatusOperational)
	} else {
		addLane(riskLane.ID, riskLane.Status, riskLane.Controls, riskLane.RecomputedWork, riskStarted)
		if riskLane.Status != protocol.StatusPass {
			markFail(riskLane.Status)
		}
	}

	// Lane: the QLT-02 budget-manifest audit (D-06-16).
	budgetStarted := time.Now()
	budgetLane, budgetErr := VerifyQLT02BudgetManifest(ctx)
	if budgetErr != nil {
		addLane(LaneQLT02BudgetAudit, protocol.StatusOperational, nil, 1, budgetStarted)
		markFail(protocol.StatusOperational)
	} else {
		addLane(budgetLane.ID, budgetLane.Status, budgetLane.Controls, budgetLane.RecomputedWork, budgetStarted)
		if budgetLane.Status != protocol.StatusPass {
			markFail(budgetLane.Status)
		}
	}

	// Lane: control:interpreter-o0-o3 (D-06-06), the cache-backed native
	// differential, driven here over testdata/phase1's toggle.lang.
	nativeStarted := time.Now()
	runner := phase6DefaultRunner()
	store, storeErr := phase6Store()
	if storeErr != nil {
		addLane(phase6NativeDifferentialLane, protocol.StatusOperational, nil, 1, nativeStarted)
		markFail(protocol.StatusOperational)
	} else {
		source, readErr := os.ReadFile(nat03CorpusPath(phase6NativeDifferentialCorpus + "/toggle.lang"))
		if readErr != nil {
			addLane(phase6NativeDifferentialLane, protocol.StatusOperational, nil, 1, nativeStarted)
			markFail(protocol.StatusOperational)
		} else {
			lane, _, laneErr := verifyPhase6NativeDifferentialLane(ctx, source, runner, store)
			if laneErr != nil {
				addLane(phase6NativeDifferentialLane, protocol.StatusOperational, nil, 1, nativeStarted)
				markFail(protocol.StatusOperational)
			} else {
				lane.ElapsedNS = time.Since(nativeStarted).Nanoseconds()
				result.Lanes = append(result.Lanes, lane)
				result.Metrics.RecomputedWork += lane.RecomputedWork
				// CR-01: this counter is a phase-wide aggregate of how many
				// artifacts this invocation REUSED, so it must gate on the
				// reused status specifically -- a non-empty CacheStatus also
				// covers artifact_recomputed / not_cacheable / unavailable,
				// none of which is a reuse. The prior `+= 0` made the counter
				// structurally zero in this function, which is the one
				// `lang verify testdata/phase6` actually runs, so FND-04's
				// cache-status reporting was false here while the sibling
				// VerifyPhase6ChangedRisk reported it correctly.
				if lane.CacheStatus == string(cache.StatusArtifactReused) {
					result.Metrics.CacheInputsReusedCount++
				}
				if lane.Status != protocol.StatusPass {
					markFail(lane.Status)
				}
			}
		}
	}

	// Lanes: DX-04's five defect injectors (D-06-25/D-06-27), each proving
	// its own class's injector fires against its held-out fixture and
	// produces a genuinely rejected mutation.
	matchStarted := time.Now()
	if status, control, work := phase6RunMatchInjectionLane(); status == protocol.StatusPass {
		addLane("lane:defect-match-injection", protocol.StatusPass, []string{control}, work, matchStarted)
	} else {
		addLane("lane:defect-match-injection", status, nil, work, matchStarted)
		markFail(status)
	}

	moveStarted := time.Now()
	if status, control, work := phase6RunMoveInjectionLane(); status == protocol.StatusPass {
		addLane("lane:defect-move-injection", protocol.StatusPass, []string{control}, work, moveStarted)
	} else {
		addLane("lane:defect-move-injection", status, nil, work, moveStarted)
		markFail(status)
	}

	borrowStarted := time.Now()
	if status, control, work := phase6RunBorrowInjectionLane(); status == protocol.StatusPass {
		addLane("lane:defect-borrow-injection", protocol.StatusPass, []string{control}, work, borrowStarted)
	} else {
		addLane("lane:defect-borrow-injection", status, nil, work, borrowStarted)
		markFail(status)
	}

	cleanupStarted := time.Now()
	if status, control, work := phase6RunCleanupInjectionLane(); status == protocol.StatusPass {
		addLane("lane:defect-cleanup-injection", protocol.StatusPass, []string{control}, work, cleanupStarted)
	} else {
		addLane("lane:defect-cleanup-injection", status, nil, work, cleanupStarted)
		markFail(status)
	}

	staleStarted := time.Now()
	if status, control, work := phase6RunStaleEvidenceInjectionLane(ctx); status == protocol.StatusPass {
		addLane("lane:defect-stale-evidence-injection", protocol.StatusPass, []string{control}, work, staleStarted)
	} else {
		addLane("lane:defect-stale-evidence-injection", status, nil, work, staleStarted)
		markFail(status)
	}

	for _, required := range Phase6RequiredControls() {
		if !hasControl(result.Lanes, required) {
			markFail(protocol.StatusInvalid)
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("verify.control_missing", diagnostic.Span{}, required))
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			markFail(protocol.StatusInvalid)
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("verify.zero_work", diagnostic.Span{}, lane.ID))
		}
	}

	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize(), nil
}

// phase6RunMatchInjectionLane, phase6RunMoveInjectionLane, and
// phase6RunBorrowInjectionLane each drive RunDefectInjectionExercise
// against their own held-out fixture (README, testdata/phase6): a
// mutated source that Check refuses is the proof this injector's marker
// convention still targets a reachable defect, never a silent unmutated
// pass (D-06-27.3's fail-closed discipline, exercised here rather than
// merely asserted by go test).
func phase6RunMatchInjectionLane() (status, control string, work int) {
	source, err := os.ReadFile(nat03CorpusPath("testdata/phase6/heldout_match_defect.lang"))
	if err != nil {
		return protocol.StatusOperational, "", 1
	}
	exercise := RunDefectInjectionExercise(MatchInjector{}, source)
	if exercise.Status != "pass" {
		return protocol.StatusInvalid, "", 1
	}
	return protocol.StatusPass, ControlDefectMatchInjection, 1
}

func phase6RunMoveInjectionLane() (status, control string, work int) {
	source, err := os.ReadFile(nat03CorpusPath("testdata/phase6/heldout_move_defect.lang"))
	if err != nil {
		return protocol.StatusOperational, "", 1
	}
	exercise := RunDefectInjectionExercise(MoveInjector{}, source)
	if exercise.Status != "pass" {
		return protocol.StatusInvalid, "", 1
	}
	return protocol.StatusPass, ControlDefectMoveInjection, 1
}

func phase6RunBorrowInjectionLane() (status, control string, work int) {
	source, err := os.ReadFile(nat03CorpusPath("testdata/phase6/heldout_borrow_defect.lang"))
	if err != nil {
		return protocol.StatusOperational, "", 1
	}
	exercise := RunDefectInjectionExercise(BorrowInjector{}, source)
	if exercise.Status != "pass" {
		return protocol.StatusInvalid, "", 1
	}
	return protocol.StatusPass, ControlDefectBorrowInjection, 1
}

// phase6RunCleanupInjectionLane compiles phase6CleanupFixture's checked,
// validated core to real generated C (the artifact carrying a genuine
// lang:release-site marker), drives CleanupInjector.Inject against it, and
// requires a real mutation -- distinct bytes from the unmutated source --
// proving the shared release-omission mutation runner (D-06-25's "reuse
// verbatim" instruction) still finds and removes a real release site.
func phase6RunCleanupInjectionLane() (status, control string, work int) {
	source, err := os.ReadFile(nat03CorpusPath(phase6CleanupFixture))
	if err != nil {
		return protocol.StatusOperational, "", 1
	}
	checked := Check(source)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		return protocol.StatusInvalid, "", 1
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return protocol.StatusInvalid, "", 1
	}
	cSource, emitErr := cgen.EmitNative(validated.Program())
	if emitErr != nil {
		return protocol.StatusOperational, "", 1
	}
	mutated, injectErr := CleanupInjector{}.Inject([]byte(cSource))
	if injectErr != nil {
		return protocol.StatusInvalid, "", 1
	}
	if string(mutated) == cSource {
		return protocol.StatusInvalid, "", 1
	}
	return protocol.StatusPass, ControlDefectCleanupInjection, 1
}

// phase6RunStaleEvidenceInjectionLane builds a real evidence manifest for
// stale_evidence_subject.lang (evidence.Build, the same construction
// `lang evidence` itself uses), re-touches the source with
// StaleEvidenceInjector, and requires evidence.Validate to report a
// mismatch against the ORIGINAL manifest -- the same locator
// ValidateEvidenceCommandFile's own mismatch report uses (session.go),
// proving the stale-evidence defect class is genuinely detectable through
// already-shipped surface, per D-06-25's own instruction.
func phase6RunStaleEvidenceInjectionLane(ctx context.Context) (status, control string, work int) {
	source, err := os.ReadFile(nat03CorpusPath("testdata/phase6/stale_evidence_subject.lang"))
	if err != nil {
		return protocol.StatusOperational, "", 1
	}
	facts, factsErr := evidence.DefaultFacts(ctx, "clang")
	if factsErr != nil {
		return protocol.StatusOperational, "", 1
	}
	product, diagnostics, buildErr := evidence.Build(source, facts)
	if buildErr != nil || len(diagnostics) != 0 {
		return protocol.StatusInvalid, "", 1
	}
	retouched, injectErr := StaleEvidenceInjector{}.Inject(source)
	if injectErr != nil {
		return protocol.StatusInvalid, "", 1
	}
	validateErr := evidence.Validate(product.Manifest, retouched, facts)
	if validateErr == nil {
		return protocol.StatusInvalid, "", 2
	}
	return protocol.StatusPass, ControlDefectStaleEvidenceInjection, 2
}
