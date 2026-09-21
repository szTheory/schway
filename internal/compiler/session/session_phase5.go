package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

// phase16FrozenEmitterEvidence is the narrow, refusal-first bridge for the
// M004 families cut from public native emission in 0607486.  It is deliberately
// local to verification controls: callers always try cgen.EmitNative first and
// only receive historical C after the current, family-specific refusal and both
// immutable digests have been verified.
type phase16FrozenEmitterEvidence struct {
	refusal        string
	fixtureSHA256  string
	artifact       string
	artifactSHA256 string
}

type phase16GeneratedFrozenEvidence struct {
	Schema    string
	Generator string
	Records   []struct {
		ID             string
		ProgramSHA256  string
		Artifact       string
		ArtifactSHA256 string
	}
}

type phase16FileFrozenEvidence struct {
	Schema  string
	Records []struct {
		Fixture, FixtureSHA256, Artifact, ArtifactSHA256 string
	}
}

var phase16FrozenEmitterEvidenceByFixture = map[string]phase16FrozenEmitterEvidence{
	"testdata/phase4/acquire_three_success.lang": {
		refusal:        "multi-function foreign-call bodies are not supported",
		fixtureSHA256:  "be62a50ecc048174f220cc3775c7abe34dcb856d835377e1059413d345ddbe75",
		artifact:       "testdata/phase16/historical/acquire_three_success.c",
		artifactSHA256: "b86133b1b92d05ac6b95a492a59d5e60e2ff4530a86d8650827a8c3d6253c1f2",
	},
	"testdata/phase4/foreign_acquire_one.lang": {
		refusal:        "multi-function foreign-call bodies are not supported",
		fixtureSHA256:  "f66c55e4cf0b1dab45408c805c131848acf5f5ae1e11fa8ba2746181be409130",
		artifact:       "testdata/phase16/historical/foreign_acquire_one.c",
		artifactSHA256: "621f23618ad0e849ae2c3e0d483ea16ce40af11333ac1d0a8176e4b798890998",
	},
	"testdata/phase4/acquire_three_fail_second.lang": {
		refusal:        "multi-function foreign-call bodies are not supported",
		fixtureSHA256:  "ce63c2a910af3ae6f45154fea2cea3af1aee1772293f372c09351113afd1d596",
		artifact:       "testdata/phase16/historical/acquire_three_fail_second.c",
		artifactSHA256: "ba992be43a9aab13663508d665daeca1ff8d8a2461eddc824cf54ebd2a2ecff2",
	},
	"testdata/phase4/acquire_three_fail_third.lang": {
		refusal:        "multi-function foreign-call bodies are not supported",
		fixtureSHA256:  "f6958321765b0793ad919875c2528c21b6dfe7b5ff916b8943e478d883d319dd",
		artifact:       "testdata/phase16/historical/acquire_three_fail_third.c",
		artifactSHA256: "4ebbf4228aa898242f7d1fb802c63b040a69fd13fb97b1ccee25a16c667e5841",
	},
	"testdata/phase4/nonlocal_exit_probe.lang": {
		refusal:        "multi-function foreign-call bodies are not supported",
		fixtureSHA256:  "92bd1a98a22fc50d5a5c07127187d0d1c4292f60508321d269ab30db595c36ae",
		artifact:       "testdata/phase16/historical/nonlocal_exit_probe.c",
		artifactSHA256: "685d4b88f67533cde6f40dbf9cb1762609fcdaa0174fd9d8a4fb2c99b17f5f14",
	},
	"testdata/phase5/retained_pointer.lang": {
		refusal:        "multi-function foreign-call bodies are not supported",
		fixtureSHA256:  "5136960f513341e8a5167bb80ee0de1228887b038e42f6befa5faf763e853957",
		artifact:       "testdata/phase16/historical/retained_pointer.c",
		artifactSHA256: "be02f1de9184b629e27deb121c260c39457bf301f25ba94e5a1915eb5598f1f7",
	},
	"testdata/phase5/allocator_mismatch.lang": {
		refusal:        "multi-function foreign-call bodies are not supported",
		fixtureSHA256:  "526b3a791ff0870aafc64a56efdd0d98a562a1236d634eacf3251c1ae33f34ef",
		artifact:       "testdata/phase16/historical/allocator_mismatch.c",
		artifactSHA256: "323175965d2fc24de70eebeb979a638f1cd71bbefc4f64d8c4f38b377ab00185",
	},
	"testdata/phase5/inline_across_foreign.lang": {
		refusal:        "multi-function foreign-call bodies are not supported",
		fixtureSHA256:  "7d2ff2d799196b1a8b28025fc8081bd0675591c38c5485a998744c1fc47d95f4",
		artifact:       "testdata/phase16/historical/inline_across_foreign.c",
		artifactSHA256: "6d5fccd190b68bd98c2f0f8bb73ee45307dec297e11ba5c26edbfb0b369abc14",
	},
	"testdata/phase5/restrict_borrow.lang": {
		refusal:        "by-pointer bodies are not supported",
		fixtureSHA256:  "7a12f2a640947cc83c34aafaea84358a1d8e4d3358e424e30ea62fdff71ee1fe",
		artifact:       "testdata/phase16/historical/restrict_borrow.c",
		artifactSHA256: "05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0",
	},
}

// Phase16ControlNativeC returns generated C for an admitted program, or the
// digest-bound historical C for a cut M004 verification control. It is not a
// general emitter: an unrecognised fixture always receives cgen's error.
func Phase16ControlNativeC(program core.Program, fixture string) (string, error) {
	if filepath.IsAbs(fixture) {
		relative, relErr := filepath.Rel(nat03ProjectRoot(), fixture)
		if relErr != nil {
			return "", relErr
		}
		fixture = filepath.ToSlash(relative)
	}
	generated, err := cgen.EmitNative(program)
	if err == nil {
		return generated, nil
	}
	evidence, cut := phase16FrozenEmitterEvidenceByFixture[fixture]
	if !cut {
		if fixture != "" {
			return phase16FileControlNativeC(fixture, err)
		}
		return phase16GeneratedControlNativeC(program, err)
	}
	if !strings.Contains(err.Error(), evidence.refusal) {
		return "", fmt.Errorf("%s: public native refusal changed: got %q, want %q", fixture, err, evidence.refusal)
	}
	fixtureBytes, readErr := os.ReadFile(nat03CorpusPath(fixture))
	if readErr != nil {
		return "", fmt.Errorf("%s: reading frozen-evidence fixture: %w", fixture, readErr)
	}
	fixtureDigest := sha256.Sum256(fixtureBytes)
	if hex.EncodeToString(fixtureDigest[:]) != evidence.fixtureSHA256 {
		return "", fmt.Errorf("%s: frozen-evidence fixture digest changed", fixture)
	}
	artifactPath := filepath.Join(nat03ProjectRoot(), filepath.FromSlash(evidence.artifact))
	artifact, readErr := os.ReadFile(artifactPath)
	if readErr != nil {
		return "", fmt.Errorf("%s: reading frozen evidence artifact: %w", fixture, readErr)
	}
	artifactDigest := sha256.Sum256(artifact)
	if hex.EncodeToString(artifactDigest[:]) != evidence.artifactSHA256 {
		return "", fmt.Errorf("%s: frozen evidence artifact digest changed", fixture)
	}
	return string(artifact), nil
}

func phase16FileControlNativeC(fixture string, refusal error) (string, error) {
	if !strings.Contains(refusal.Error(), "foreign-call bodies are not supported") && !strings.Contains(refusal.Error(), "by-pointer bodies are not supported") {
		return "", refusal
	}
	bytes, err := os.ReadFile(nat03CorpusPath("testdata/phase16/file-frozen-evidence.json"))
	if err != nil {
		return "", err
	}
	var manifest phase16FileFrozenEvidence
	if err := json.Unmarshal(bytes, &manifest); err != nil || manifest.Schema != "phase16.file-frozen-evidence/1" {
		return "", fmt.Errorf("invalid file frozen evidence manifest")
	}
	source, err := os.ReadFile(nat03CorpusPath(fixture))
	if err != nil {
		return "", err
	}
	sourceSum := sha256.Sum256(source)
	for _, r := range manifest.Records {
		if r.Fixture == fixture && r.FixtureSHA256 == hex.EncodeToString(sourceSum[:]) {
			artifact, err := os.ReadFile(nat03CorpusPath(r.Artifact))
			if err != nil {
				return "", err
			}
			sum := sha256.Sum256(artifact)
			if hex.EncodeToString(sum[:]) != r.ArtifactSHA256 {
				return "", fmt.Errorf("file frozen artifact digest changed for %s", fixture)
			}
			return string(artifact), nil
		}
	}
	return "", fmt.Errorf("file-backed cut fixture %q lacks digest-bound frozen evidence", fixture)
}

func phase16GeneratedControlNativeC(program core.Program, refusal error) (string, error) {
	if !strings.Contains(refusal.Error(), "by-pointer bodies are not supported") {
		return "", refusal
	}
	manifestBytes, err := os.ReadFile(nat03CorpusPath("testdata/phase16/generated-frozen-evidence.json"))
	if err != nil {
		return "", fmt.Errorf("generated frozen evidence manifest: %w", err)
	}
	var manifest phase16GeneratedFrozenEvidence
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "", fmt.Errorf("decode generated frozen evidence manifest: %w", err)
	}
	if manifest.Schema != "phase16.generated-frozen-evidence/1" || manifest.Generator != "EnumeratePhase5Closure/v1" {
		return "", fmt.Errorf("generated frozen evidence manifest has unsupported schema or generator")
	}
	canonical, err := json.Marshal(program)
	if err != nil {
		return "", fmt.Errorf("canonical generated program: %w", err)
	}
	digest := sha256.Sum256(canonical)
	want := hex.EncodeToString(digest[:])
	for _, record := range manifest.Records {
		if record.ID != program.Module || record.ProgramSHA256 != want {
			continue
		}
		artifact, err := os.ReadFile(nat03CorpusPath(record.Artifact))
		if err != nil {
			return "", fmt.Errorf("generated frozen artifact: %w", err)
		}
		artifactDigest := sha256.Sum256(artifact)
		if hex.EncodeToString(artifactDigest[:]) != record.ArtifactSHA256 {
			return "", fmt.Errorf("generated frozen artifact digest changed for %s", program.Module)
		}
		return string(artifact), nil
	}
	return "", fmt.Errorf("generated cut program %q lacks a digest-bound frozen artifact", program.Module)
}

// Phase5RequiredControls is the complete Phase 5 required-control list as
// of plan 05-09 (D-05-17): every control this gate's own lanes fire,
// copying Phase4RequiredControls' shape verbatim -- a flat slice of exact
// identifier strings, one per line, no computation.
// scripts/verify-phase5.sh (this plan) duplicates this list VERBATIM, and
// TestPhase5RequiredControlsMatchScript asserts the two are set-equal in
// both directions, so the gate cannot silently shrink by dropping a
// control from either copy (T-05-32).
//
// control:foreign.no_unproven_attributes is plan 05-04's control, narrowed
// (not deleted) by that plan; the first ten identifiers are Phase 5
// controls introduced by plans 05-04 through 05-08. Plan 05-14 (this plan)
// closes the phase by adding the final five: the three reducer vacuity
// controls plan 05-12 built but deliberately left unwired
// (LaneMismatchReduce, session_phase5_mismatch.go), and the two QLT-01
// registry-audit controls plan 05-11 built but deliberately left unwired
// (LaneQLT01RegistryAudit, qlt01.go) -- both plans left the wiring to this
// plan precisely so the Go control set and scripts/verify-phase5.sh's own
// required-control block are extended TOGETHER, in one commit, and
// TestPhase5RequiredControlsMatchScript never observes a transiently
// divergent pair (D-05-17).
func Phase5RequiredControls() []string {
	return []string{
		"control:foreign.no_unproven_attributes",
		"control:alias.false_no_alias",
		"control:interpreter-o0-o3-lto",
		"control:diagnostic.reject_program_id_equivalence",
		"control:compare.field_routing_unrouted",
		"control:native.sanitize.retained_pointer",
		"control:native.sanitize.ubsan_no_recover",
		"control:native.sanitize.allocator_mismatch",
		"control:native.sanitize.use_after_free",
		"control:core.attribute_unjustified",
		"control:reduce.no_progress",
		"control:reduce.predicate_too_loose",
		"control:reduce.nondeterministic",
		"control:qlt01.registry_incomplete",
		"control:qlt01.stale_control_reference",
	}
}

// Phase5ExpectedEscapes is Phase 5's own declared, gate-visible residual
// set: the callback-invocation half of NAT-03's stale-callback-retention
// mutation (D-05-07), and plan 05-13's coordinated source-to-core false
// claim (D-05-30) -- neither claimed solved, neither ever permitted to
// appear as a detected lane control (TestBothPhase5EscapesAreVisible,
// TestCoordinatedLieEscapeIsNeverDetected).
func Phase5ExpectedEscapes() []string {
	return []string{EscapeCallbackInvocationUnsubjected, EscapeCoordinatedSourceToCoreFalseClaim}
}

// Phase5ClangPathOverrideForTest is a mutable seam ONLY for
// TestPhase5ToolMissingIsNamedObligationNotPass (D-05-17): the test
// substitutes a broken ClangPath to prove the sanitizer controls surface
// as named, gate-visible operational obligations -- never a silent pass --
// without widening VerifyPhase5ControlsAndWork's own plan-specified
// signature. Production callers must never set this. Empty means "use the
// default toolchain resolution", mirroring native.DefaultRunner's own
// default.
var Phase5ClangPathOverrideForTest = ""

func phase5DefaultRunner() native.Runner {
	runner := native.DefaultRunner()
	if Phase5ClangPathOverrideForTest != "" {
		runner.ClangPath = Phase5ClangPathOverrideForTest
	}
	return runner
}

// VerifyPhase5ControlsAndWork is Phase 5's own control-and-work gate
// (D-05-17), peering verifyForeignCorpus's own addLane/fail shape: it runs
// the alias-fact mutation lane, the two attribute-justification lanes
// (positive and negative, narrowing control:foreign.no_unproven_attributes
// per plan 05-04), the reject-program diagnostic-ID equivalence lane, the
// comparator's own fail-closed field-routing check, a real
// interpreter/-O0/-O3/-O3-LTO differential over one of the Phase 5
// adversarial fixtures, and the sanitizer lane -- collecting every
// control's fired status and RecomputedWork.
//
// A `tool_missing`/`sanitizer_inert` sanitizer lane is NEVER folded into a
// silent pass: when VerifyPhase5SanitizeLane itself does not reach every
// one of its own controls (an unavailable or non-instrumented sanitizer
// runtime), this function adds an explicit obligation lane naming all four
// sanitizer controls under the sanitizer lane's own reported status, so
// every declared control is always visible in the result even when it
// could not be exercised (T-05-33).
func VerifyPhase5ControlsAndWork(ctx context.Context) (protocol.Result, error) {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)
	result.ExpectedEscapes = append([]string{}, Phase5ExpectedEscapes()...)

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

	runner := phase5DefaultRunner()

	// Lanes: control:foreign.no_unproven_attributes (positive, narrowed by
	// plan 05-04) and control:core.attribute_unjustified (negative,
	// D-05-03b's own falsifier), both driven off the same checked
	// restrict_borrow.lang fixture so the negative lane's corrupted claim
	// is compared against the identical program the positive lane proved
	// justified.
	attributesStarted := time.Now()
	restrictSource, err := os.ReadFile(nat03CorpusPath("testdata/phase5/restrict_borrow.lang"))
	switch {
	case err != nil:
		addLane("lane:foreign-no-unproven-attributes", protocol.StatusOperational, nil, 1, attributesStarted)
		addLane("lane:attribute-unjustified", protocol.StatusOperational, nil, 1, attributesStarted)
		markFail(protocol.StatusOperational)
	default:
		restrictChecked := Check(restrictSource)
		// Phase 11 (11-GUARD-LEDGER.md): KEPT. Bound to the fixed Phase 5
		// fixture restrict_borrow.lang, genuinely single-function by
		// construction; unrelated to the multi-function corpus this
		// phase widens.
		if len(restrictChecked.Diagnostics) != 0 || len(restrictChecked.Program.Functions) != 1 {
			addLane("lane:foreign-no-unproven-attributes", protocol.StatusInvalid, nil, 1, attributesStarted)
			addLane("lane:attribute-unjustified", protocol.StatusInvalid, nil, 1, attributesStarted)
			markFail(protocol.StatusInvalid)
		} else {
			function := restrictChecked.Program.Functions[0]
			loanID := ""
			if function.Linear != nil && len(function.Linear.Operations) > 0 {
				loanID = function.Linear.Operations[0].LoanID
			}

			validClaim := corevalidate.AttributeClaim{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: loanID}
			if validateErr := corevalidate.ValidateEmittedAttributes(restrictChecked.Program, []corevalidate.AttributeClaim{validClaim}); validateErr != nil {
				addLane("lane:foreign-no-unproven-attributes", protocol.StatusInvalid, nil, 1, attributesStarted)
				markFail(protocol.StatusInvalid)
			} else {
				addLane("lane:foreign-no-unproven-attributes", protocol.StatusPass, []string{"control:foreign.no_unproven_attributes"}, 1, attributesStarted)
			}

			corruptClaim := corevalidate.AttributeClaim{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: "gate-bogus-loan-id"}
			corruptErr := corevalidate.ValidateEmittedAttributes(restrictChecked.Program, []corevalidate.AttributeClaim{corruptClaim})
			var attributeErr *corevalidate.AttributeUnjustifiedError
			if corruptErr == nil || !errors.As(corruptErr, &attributeErr) || attributeErr.Code != "core.attribute_unjustified" {
				addLane("lane:attribute-unjustified", protocol.StatusInvalid, nil, 1, attributesStarted)
				markFail(protocol.StatusInvalid)
			} else {
				addLane("lane:attribute-unjustified", protocol.StatusPass, []string{"control:core.attribute_unjustified"}, 1, attributesStarted)
			}
		}
	}

	// Lane: control:alias.false_no_alias (D-05-05), reusing the exact
	// mutation runner and assertion NAT03Mutations' own row 5 cites.
	aliasStarted := time.Now()
	aliasRunner := NewAliasFactMutationRunner(runner, nat03CorpusPath("testdata/phase5/false_restrict_hoist.lang"))
	if aliasErr := VerifyAliasFalseNoAlias(ctx, aliasRunner); aliasErr != nil {
		addLane("lane:alias-false-no-alias", protocol.StatusMismatch, nil, aliasRunner.RecomputedWork()+1, aliasStarted)
		markFail(protocol.StatusMismatch)
	} else {
		addLane("lane:alias-false-no-alias", protocol.StatusPass, []string{ControlAliasFalseNoAlias}, aliasRunner.RecomputedWork()+1, aliasStarted)
	}

	// Lane: control:diagnostic.reject_program_id_equivalence (D-05-20):
	// two independent Check() runs over the same reject-program must agree
	// on the exact same diagnostic ID.
	diagnosticStarted := time.Now()
	rejectSource, rejectErr := os.ReadFile(nat03CorpusPath("testdata/phase4/foreign_call_target_not_foreign.lang"))
	if rejectErr != nil {
		addLane("lane:diagnostic-reject-program-id-equivalence", protocol.StatusOperational, nil, 1, diagnosticStarted)
		markFail(protocol.StatusOperational)
	} else {
		firstChecked := Check(rejectSource)
		secondChecked := Check(rejectSource)
		if len(firstChecked.Diagnostics) == 0 || len(secondChecked.Diagnostics) == 0 {
			addLane("lane:diagnostic-reject-program-id-equivalence", protocol.StatusInvalid, nil, 1, diagnosticStarted)
			markFail(protocol.StatusInvalid)
		} else {
			diagnostics := map[string]diagnostic.Diagnostic{"check-run-1": firstChecked.Diagnostics[0], "check-run-2": secondChecked.Diagnostics[0]}
			if compareErr := Phase5CompareDiagnosticIDs("foreign_call_target_not_foreign.lang", diagnostics); compareErr != nil {
				addLane("lane:diagnostic-reject-program-id-equivalence", protocol.StatusMismatch, nil, 2, diagnosticStarted)
				markFail(protocol.StatusMismatch)
			} else {
				addLane("lane:diagnostic-reject-program-id-equivalence", protocol.StatusPass, []string{ControlDiagnosticRejectProgramIDEquivalence}, 2, diagnosticStarted)
			}
		}
	}

	// Lane: control:compare.field_routing_unrouted (D-05-21): every field
	// reachable from execution.Execution is routed to either
	// Phase5ComparedComparisonFields or Phase5ExcludedComparisonFields --
	// never left unrouted, and never a stale entry naming a field that no
	// longer exists.
	routingStarted := time.Now()
	actualFields := reachableFieldPaths(reflect.TypeOf(execution.Execution{}), "Execution")
	unrouted := unroutedFields(actualFields, Phase5ComparedComparisonFields, Phase5ExcludedComparisonFields)
	stale := staleFields(actualFields, Phase5ComparedComparisonFields, Phase5ExcludedComparisonFields)
	if len(unrouted) != 0 || len(stale) != 0 {
		addLane("lane:compare-field-routing", protocol.StatusInvalid, nil, len(actualFields)+1, routingStarted)
		markFail(protocol.StatusInvalid)
	} else {
		addLane("lane:compare-field-routing", protocol.StatusPass, []string{"control:compare.field_routing_unrouted"}, len(actualFields)+1, routingStarted)
	}

	// Lane: control:interpreter-o0-o3-lto (D-05-19): a real
	// interpreter/-O0/-O3/-O3-LTO differential over one of the Phase 5
	// adversarial fixtures -- inline_across_foreign.lang, the fixture
	// D-05-18a specifically engineered to give `-flto` a real cross-TU
	// inlining opportunity the non-LTO tiers cannot take.
	ltoStarted := time.Now()
	ltoStatus, ltoControls, ltoWork := phase5RunInterpreterO0O3LTOLane(ctx, runner)
	addLane("lane:interpreter-o0-o3-lto", ltoStatus, ltoControls, ltoWork, ltoStarted)
	if ltoStatus != protocol.StatusPass {
		markFail(ltoStatus)
	}

	// Lane: the sanitizer lane's own four controls (D-05-08/D-05-09/
	// D-05-13/D-05-14). When VerifyPhase5SanitizeLane itself does not
	// pass -- an unavailable sanitizer runtime, or one of its own controls
	// failing to fire -- an explicit obligation lane names all four
	// sanitizer controls under the sanitizer lane's own reported status,
	// so a `tool_missing`/`sanitizer_inert` runtime is a NAMED OBLIGATION,
	// never an implicit pass (T-05-33).
	sanitizeStarted := time.Now()
	sanitizeResult, sanitizeErr := VerifyPhase5SanitizeLane(ctx, runner)
	if sanitizeErr != nil {
		return protocol.Result{}, sanitizeErr
	}
	result.Lanes = append(result.Lanes, sanitizeResult.Lanes...)
	result.Metrics.RecomputedWork += sanitizeResult.Metrics.RecomputedWork
	if sanitizeResult.Status != protocol.StatusPass {
		addLane("lane:native-sanitize-obligations", sanitizeResult.Status,
			[]string{ControlSanitizeRetainedPointer, ControlSanitizeUseAfterFree, ControlSanitizeAllocatorMismatch, ControlSanitizeUBSanNoRecover},
			1, sanitizeStarted)
		markFail(sanitizeResult.Status)
	}

	// Lane: the QLT-01 registry completeness audit (D-05-29), built by
	// plan 05-11 and deliberately left unwired until this plan so the
	// registry's two controls (control:qlt01.registry_incomplete,
	// control:qlt01.stale_control_reference) join Phase5RequiredControls()
	// in the same commit as the shell gate's own extension.
	qlt01Started := time.Now()
	qlt01Lane, qlt01Err := VerifyQLT01Registry(ctx)
	if qlt01Err != nil {
		addLane(LaneQLT01RegistryAudit, protocol.StatusOperational, nil, 1, qlt01Started)
		markFail(protocol.StatusOperational)
	} else {
		addLane(qlt01Lane.ID, qlt01Lane.Status, qlt01Lane.Controls, qlt01Lane.RecomputedWork, qlt01Started)
		if qlt01Lane.Status != protocol.StatusPass {
			markFail(qlt01Lane.Status)
		}
	}

	// Lane: the reducer's three vacuity controls (D-05-27), built by
	// plan 05-12 and deliberately left unwired until this plan
	// (LaneMismatchReduce, session_phase5_mismatch.go). Its own
	// VerifyMismatchReduceLane already returns a full protocol.Result with
	// its own lane and RecomputedWork, so it is merged directly rather
	// than re-derived through addLane.
	mismatchStarted := time.Now()
	mismatchResult, mismatchErr := VerifyMismatchReduceLane(ctx)
	if mismatchErr != nil {
		addLane(LaneMismatchReduce, protocol.StatusOperational, nil, 1, mismatchStarted)
		markFail(protocol.StatusOperational)
	} else {
		result.Lanes = append(result.Lanes, mismatchResult.Lanes...)
		result.Metrics.RecomputedWork += mismatchResult.Metrics.RecomputedWork
		if mismatchResult.Status != protocol.StatusPass {
			markFail(mismatchResult.Status)
		}
	}

	// Lane: the coordinated source-to-core false-claim escape (D-05-30),
	// declared by plan 05-13 and wired here alongside this plan's own
	// extension of Phase5ExpectedEscapes(). A pass here is attributed to
	// the named escape (never a silent absence of checking); an error
	// means the adversarial pair failed to construct as claimed, which is
	// a genuine gate failure, not an escape.
	coordinatedStarted := time.Now()
	coordinatedLane, coordinatedErr := VerifyCoordinatedLieEscape(ctx)
	if coordinatedErr != nil {
		addLane(LaneCoordinatedLieEscape, protocol.StatusOperational, nil, 1, coordinatedStarted)
		markFail(protocol.StatusOperational)
	} else {
		addLane(coordinatedLane.ID, coordinatedLane.Status, coordinatedLane.Controls, coordinatedLane.RecomputedWork, coordinatedStarted)
		if coordinatedLane.Status != protocol.StatusPass {
			markFail(coordinatedLane.Status)
		}
	}

	for _, required := range Phase5RequiredControls() {
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

// phase5RunInterpreterO0O3LTOLane drives inline_across_foreign.lang
// through the interpreter and three native builds -- -O0, -O3, and
// -O3 with LTO -- and compares all four via Phase5CompareEngines. Split
// out of VerifyPhase5ControlsAndWork as its own function so a failure at
// any step degrades to a single reported lane status rather than aborting
// the whole gate.
func phase5RunInterpreterO0O3LTOLane(ctx context.Context, runner native.Runner) (status string, controls []string, work int) {
	const fixture = "inline_across_foreign.lang"
	source, err := os.ReadFile(nat03CorpusPath("testdata/phase5/" + fixture))
	if err != nil {
		return protocol.StatusOperational, nil, 1
	}
	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		return protocol.StatusInvalid, nil, 1
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return protocol.StatusInvalid, nil, 1
	}
	program := validated.Program()
	// Phase 11 (11-GUARD-LEDGER.md): KEPT. Bound to the fixed Phase 5
	// fixture inline_across_foreign.lang, genuinely single-function by
	// construction -- its own cross-TU LTO opportunity is the foreign
	// boundary, not the multi-function corpus this phase widens. The
	// comparator this lane calls (Phase5CompareEngines) is unaffected;
	// only this caller's own admission stays scoped to one function.
	if len(program.Functions) != 1 {
		return protocol.StatusInvalid, nil, 1
	}
	functionName := program.Functions[0].Name

	interpreted, err := interp.Run(program, functionName, "7")
	if err != nil {
		return protocol.StatusOperational, nil, 1
	}
	cSource, err := Phase16ControlNativeC(program, "testdata/phase5/inline_across_foreign.lang")
	if err != nil {
		return protocol.StatusOperational, nil, 1
	}

	nativeRunner := runner
	nativeRunner.Expect = expectForOutcomeKind(interpreted.Outcome.Kind)
	for _, function := range program.Functions {
		if function.Name != functionName || function.ForeignContract == nil {
			continue
		}
		if sourcePath, known := native.ForeignSourcePathForSymbol(function.ForeignContract.Symbol); known {
			nativeRunner.ForeignSources = append(append([]string(nil), nativeRunner.ForeignSources...), sourcePath)
		}
	}

	o0, err := nativeRunner.Run(ctx, cSource, "-O0", []string{"7"})
	if err != nil || len(o0.Pairs) != 1 {
		return protocol.StatusOperational, nil, 1
	}
	o3, err := nativeRunner.Run(ctx, cSource, "-O3", []string{"7"})
	if err != nil || len(o3.Pairs) != 1 {
		return protocol.StatusOperational, nil, 2
	}
	ltoRunner := nativeRunner
	ltoRunner.LTO = true
	o3lto, err := ltoRunner.Run(ctx, cSource, "-O3", []string{"7"})
	if err != nil || len(o3lto.Pairs) != 1 {
		return protocol.StatusOperational, nil, 3
	}

	engines := map[string]execution.Execution{
		"interpreter": interpreted,
		"O0":          o0.Pairs[0].Execution,
		"O3":          o3.Pairs[0].Execution,
		"O3-LTO":      o3lto.Pairs[0].Execution,
	}
	if compareErr := Phase5CompareEngines(fixture, engines); compareErr != nil {
		return protocol.StatusMismatch, nil, 4
	}
	return protocol.StatusPass, []string{"control:interpreter-o0-o3-lto"}, 4
}

// assertAllocatorMismatchMovesAxis proves control:native.sanitize.allocator_mismatch
// against its cited fixture (testdata/phase5/allocator_mismatch.lang):
// ASan's own alloc-dealloc-mismatch report is a genuine divergence from a
// clean terminal outcome so extreme no comparable execution.Execution
// document is ever produced at all -- mapped onto axis:terminal-outcome,
// the same mapping assertLayoutMismatchMovesAxis uses for an equally
// total refusal (a normal return vs. an ASan-terminated process).
func assertAllocatorMismatchMovesAxis(ctx context.Context, mutation NAT03Mutation) error {
	cSource, err := compilePhase5SanitizeFixture("testdata/phase5/allocator_mismatch.lang")
	if err != nil {
		return err
	}
	runner := phase5DefaultRunner()
	runner.ForeignSources = []string{native.ForeignArenaSourcePath()}
	report, runErr := runner.RunSanitized(ctx, cSource, []string{"7"})
	wantSignature := sanitizerSignatureFor("alloc-dealloc-mismatch")
	if runErr != nil || report.ExitCode == 0 || report.ReportSignature != wantSignature {
		return fmt.Errorf("%s produced no divergence: allocator-mismatch fixture did not report %q (exit=%d, check_kind=%q, err=%v)",
			mutation.ControlID, wantSignature, report.ExitCode, report.CheckKind, runErr)
	}
	return requireAxis(mutation, AxisTerminalOutcome)
}
