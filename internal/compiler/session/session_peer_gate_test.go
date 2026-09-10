package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// peerDivergenceExpected is 07-10's named, commented, hand-maintained
// register of every fixture under testdata/ where `check` ADMITS (zero
// checked.Diagnostics) but corevalidate.Validate REFUSES independently.
// TestNoUndeclaredCheckPeerDivergenceAcrossCorpus asserts this is the
// EXACT set, in both directions (T-07-10-05): an undeclared new divergence
// fails, and a stale entry that no longer diverges also fails. Keyed by
// project-relative path (forward slashes); value is the peer's reported
// Problems[0].Code.
var peerDivergenceExpected = map[string]string{
	// D-03-02's interprocedural half (PVG-04/CR-02) was carried here as
	// PHASE-07-DEBT.md D-07-49: check.computeLoanLastUses had no "call"
	// case, so the exclusive loan on `buffer` was treated as ending before
	// `relay`'s call was ever reached and `escort` checked clean, while
	// corevalidate's independent OpCall replay refused with
	// core.move_while_borrowed. CLOSED in Phase 08 (Task 2/3): check now
	// independently refuses this fixture too (check.interprocedural_loan_liveness),
	// so it is no longer a check-admits/peer-refuses divergence and this
	// entry is retired -- see
	// TestRelayEscortWitnessRefusesInterproceduralLiveness (check package)
	// and TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed
	// (corevalidate package), which now assert BOTH sides refuse
	// independently, via different codes, for the same program.
	//
	// WR-01's user-visible half (07-REVIEW.md; check-side half carried as
	// debt, PHASE-07-DEBT.md D-07-50): two `fn helper` declarations share
	// one semanticID; check's buildCalleeContracts silently resolves the
	// call to the LAST declaration, so check itself stays clean, but
	// corevalidate's function-ID uniqueness check refuses independently.
	"testdata/phase07/duplicate_function_name.lang": "core.duplicate_function_id",
}

// phase07SpotCheckRegression is 07-VERIFICATION.md's own pre-07-10
// Behavioral Spot-Checks table, restated here as a mechanical assertion
// (must_haves backstop truth): every fixture below is NOT a declared
// divergence entry, and its `lang check` status/code must be byte-for-byte
// unchanged by this plan.
var phase07SpotCheckRegression = []struct {
	fixture string
	status  string
	code    string // empty when status is pass (no diagnostics expected)
}{
	{"call_basic.lang", protocol.StatusPass, ""},
	{"call_from_both_match_arms.lang", protocol.StatusPass, ""},
	{"deep_diamond_acyclic.lang", protocol.StatusPass, ""},
	{"call_type_mismatch.lang", protocol.StatusInvalid, "check.call_argument_type_mismatch"},
	{"call_uncallable_callee.lang", protocol.StatusInvalid, "core.callee_not_callable"},
	{"clean_but_unpublishable.lang", protocol.StatusInvalid, "core.origin_omitted"},
	{"cycle_indirect.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"cycle_mutual.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"cycle_self.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"cycle_through_match_arm.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"cycle_unreachable.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
	{"foreign_symbol_shadowing.lang", protocol.StatusInvalid, "core.call_graph_cycle"},
}

// TestCheckCommandSurfacesPeerRefusal is Task 1's core proof: CheckCommandFile
// now takes the refusing union of `check`'s own diagnostics and
// corevalidate.Validate's independent verdict, in fixed precedence order
// (07-10 checkpoint, SEM-04), closing 07-REVIEW.md CR-04 / PVG-03.
func TestCheckCommandSurfacesPeerRefusal(t *testing.T) {
	// Test 1: relay_escort_witness.lang reported the PEER's own
	// core.move_while_borrowed refusal at the CLI from 07-10 through Phase
	// 07 (before 07-10 it reported status: pass, exit-0-equivalent
	// StatusPass). Phase 08 closes the interprocedural half of D-03-02:
	// check ITSELF now refuses this fixture, via its own new
	// check.interprocedural_loan_liveness law
	// (TestRelayEscortWitnessRefusesInterproceduralLiveness, check
	// package), which runs and reports BEFORE CheckCommandFile ever
	// consults the peer -- the peer's own core.move_while_borrowed refusal
	// (TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed,
	// corevalidate package) still independently exists, it is simply no
	// longer the code surfaced at the CLI for this fixture.
	t.Run("relay_escort_witness flips to the peer's refusal", func(t *testing.T) {
		result, err := session.CheckCommandFile(phase07Fixture(t, "relay_escort_witness.lang"))
		if err != nil {
			t.Fatalf("CheckCommandFile returned an error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %s, want %s", result.Status, protocol.StatusInvalid)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
			t.Fatalf("diagnostics = %+v, want exactly one check.interprocedural_loan_liveness", result.Diagnostics)
		}
	})

	// Test 2: duplicate_function_name.lang (WR-01's user-visible half)
	// refuses with the peer's own core.duplicate_function_id.
	t.Run("duplicate_function_name refuses with the peer's code", func(t *testing.T) {
		result, err := session.CheckCommandFile(phase07Fixture(t, "duplicate_function_name.lang"))
		if err != nil {
			t.Fatalf("CheckCommandFile returned an error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %s, want %s", result.Status, protocol.StatusInvalid)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("diagnostics = %+v, want exactly one core.duplicate_function_id", result.Diagnostics)
		}
	})

	// Test 3 (no regression): every fixture NOT in the declared divergence
	// register still reports exactly the status/code
	// 07-VERIFICATION.md's Behavioral Spot-Checks table recorded pre-07-10.
	t.Run("no regression across the rest of the phase07 corpus", func(t *testing.T) {
		for _, spot := range phase07SpotCheckRegression {
			spot := spot
			t.Run(spot.fixture, func(t *testing.T) {
				result, err := session.CheckCommandFile(phase07Fixture(t, spot.fixture))
				if err != nil {
					t.Fatalf("CheckCommandFile returned an error: %v", err)
				}
				if result.Status != spot.status {
					t.Fatalf("status = %s, want %s", result.Status, spot.status)
				}
				if spot.code == "" {
					if len(result.Diagnostics) != 0 {
						t.Fatalf("expected zero diagnostics, got %+v", result.Diagnostics)
					}
					return
				}
				// Edge 1 (adjacency): whether the shared code came from
				// check's own diagnostics or the peer, the union never
				// duplicates it -- exactly one diagnostic is reported.
				if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != spot.code {
					t.Fatalf("diagnostics = %+v, want exactly one %s", result.Diagnostics, spot.code)
				}
			})
		}
	})

	// Test 4 (edge 3, ordering): call_type_mismatch.lang has BOTH a check
	// diagnostic (check.call_argument_type_mismatch) and would separately
	// be refused by the peer's own OpCall replay -- the reported code is
	// check's, because check's diagnostics are reported before the peer
	// ever runs (precedence is asserted, not assumed).
	t.Run("edge 3: check diagnostics take precedence over the peer", func(t *testing.T) {
		result, err := session.CheckCommandFile(phase07Fixture(t, "call_type_mismatch.lang"))
		if err != nil {
			t.Fatalf("CheckCommandFile returned an error: %v", err)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "check.call_argument_type_mismatch" {
			t.Fatalf("diagnostics = %+v, want exactly one check.call_argument_type_mismatch (check's own code, not a core.* peer code)", result.Diagnostics)
		}
	})

	// Test 5 (edge 2, empty/fail-closed): !Valid with an EMPTY Problems
	// slice is structurally impossible from any real program, so it is
	// driven through a hand-built corevalidate.Result. The fallback must
	// never report an empty code and must never behave as StatusPass.
	t.Run("edge 2: empty Problems fails closed to the named fallback code", func(t *testing.T) {
		diag := session.PeerRefusalDiagnosticForTest(corevalidate.Result{Valid: false}, "phase07.example")
		if diag.Code != session.PeerRefusalUnnamedCodeForTest {
			t.Fatalf("code = %q, want the fail-closed fallback %q", diag.Code, session.PeerRefusalUnnamedCodeForTest)
		}
		if diag.Code == "" {
			t.Fatal("fail-closed fallback code must never be empty")
		}
	})
}

// TestDuplicateFunctionDeclarationRefusedAtCLI is the standing negative
// control for WR-01's user-visible half (D-07-50): before 07-10, this
// program reported status: pass, exit 0. It is now refused, incidentally,
// once CheckCommandFile consults the peer -- named as its own test per
// 07-10's <artifacts_this_phase_produces>, distinct from the table-driven
// assertion above so a future regression on this specific fixture names
// itself unambiguously in test output.
func TestDuplicateFunctionDeclarationRefusedAtCLI(t *testing.T) {
	result, err := session.CheckCommandFile(phase07Fixture(t, "duplicate_function_name.lang"))
	if err != nil {
		t.Fatalf("CheckCommandFile returned an error: %v", err)
	}
	if result.Status != protocol.StatusInvalid {
		t.Fatalf("status = %s, want %s (WR-01's user-visible half must be refused, not silently admitted)", result.Status, protocol.StatusInvalid)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
		t.Fatalf("diagnostics = %+v, want exactly one core.duplicate_function_id", result.Diagnostics)
	}
}

// TestNoUndeclaredCheckPeerDivergenceAcrossCorpus is Task 1 Test 7
// (T-07-10-05): walks every .lang file under testdata/, computes both
// verdicts (check's own diagnostics, and corevalidate.Validate on the
// checked core.Program when check itself admits), and asserts the set of
// files where check admits but the peer refuses equals peerDivergenceExpected
// EXACTLY, in both directions. An undeclared new divergence fails this test
// rather than passing silently; a stale entry that no longer diverges also
// fails it.
func TestNoUndeclaredCheckPeerDivergenceAcrossCorpus(t *testing.T) {
	root := testsupport.ProjectPath("testdata")
	found := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".lang" {
			return nil
		}
		checked, checkErr := session.CheckFile(path)
		if checkErr != nil {
			t.Fatalf("CheckFile(%s): %v", path, checkErr)
		}
		if len(checked.Diagnostics) > 0 {
			// check itself refuses -- not a check-admits/peer-refuses
			// divergence, regardless of what the peer would say.
			return nil
		}
		validated := corevalidate.Validate(checked.Program)
		if validated.Valid {
			return nil
		}
		code := session.PeerRefusalUnnamedCodeForTest
		if len(validated.Problems) > 0 {
			code = validated.Problems[0].Code
		}
		relative, relErr := filepath.Rel(testsupport.ProjectPath("."), path)
		if relErr != nil {
			t.Fatalf("filepath.Rel: %v", relErr)
		}
		found[filepath.ToSlash(relative)] = code
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}

	for path, wantCode := range peerDivergenceExpected {
		gotCode, ok := found[path]
		if !ok {
			t.Errorf("declared divergence %s no longer diverges (check admits, peer no longer refuses, or the fixture is missing) -- stale entry in peerDivergenceExpected", path)
			continue
		}
		if gotCode != wantCode {
			t.Errorf("%s: peer code = %q, want declared %q", path, gotCode, wantCode)
		}
	}
	for path, gotCode := range found {
		if _, declared := peerDivergenceExpected[path]; !declared {
			t.Errorf("UNDECLARED divergence: %s (check admits, corevalidate refuses with %s) is not in peerDivergenceExpected", path, gotCode)
		}
	}
}

// TestInterfaceCommandsReportPeerRefusalAsInvalid is Task 2's core proof: a
// program corevalidate.Validate refuses is reported by both
// InterfaceExportCommandFile and InterfaceCoreCommandFile as a
// protocol.StatusInvalid result carrying the peer's own code, with a NIL
// error -- never the pre-07-10 `fmt.Errorf("core validation failed: …")`
// that reached the CLI as tool.operation_failed / exit 3 with the code
// discarded (07-REVIEW.md CR-04's `interface` half).
func TestInterfaceCommandsReportPeerRefusalAsInvalid(t *testing.T) {
	fixture := phase07Fixture(t, "duplicate_function_name.lang")

	// Test 1: InterfaceCoreCommandFile.
	t.Run("InterfaceCoreCommandFile reports StatusInvalid with a nil error", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "duplicate_function_name.core.json")
		result, err := session.InterfaceCoreCommandFile(fixture, outPath)
		if err != nil {
			t.Fatalf("InterfaceCoreCommandFile returned a non-nil error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %s, want %s", result.Status, protocol.StatusInvalid)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("diagnostics = %+v, want exactly one core.duplicate_function_id", result.Diagnostics)
		}
		// Test 3: no output artifact is written on a refusal.
		if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
			t.Fatalf("expected %s to not exist after a refusal, stat err = %v", outPath, statErr)
		}
	})

	// Test 2: InterfaceExportCommandFile, the same shape.
	t.Run("InterfaceExportCommandFile reports StatusInvalid with a nil error", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "duplicate_function_name.summary.json")
		result, err := session.InterfaceExportCommandFile(fixture, outPath)
		if err != nil {
			t.Fatalf("InterfaceExportCommandFile returned a non-nil error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %s, want %s", result.Status, protocol.StatusInvalid)
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("diagnostics = %+v, want exactly one core.duplicate_function_id", result.Diagnostics)
		}
		if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
			t.Fatalf("expected %s to not exist after a refusal, stat err = %v", outPath, statErr)
		}
	})

	// Test 4: an accepting program behaves exactly as before -- same
	// StatusPass, artifact written, and a working round-trip through
	// InterfaceCheckCommandFile (the existing CoreDigest-bound contract).
	t.Run("accepting program: unchanged StatusPass, written artifacts, and a working round-trip", func(t *testing.T) {
		accepting := phase07Fixture(t, "call_basic.lang")
		dir := t.TempDir()
		summaryPath := filepath.Join(dir, "call_basic.summary.json")
		corePath := filepath.Join(dir, "call_basic.core.json")

		exportResult, err := session.InterfaceExportCommandFile(accepting, summaryPath)
		if err != nil {
			t.Fatalf("InterfaceExportCommandFile: %v", err)
		}
		if exportResult.Status != protocol.StatusPass {
			t.Fatalf("export status = %s, want %s", exportResult.Status, protocol.StatusPass)
		}
		if exportResult.Interface == nil {
			t.Fatal("expected a non-nil Interface projection on an accepting program")
		}

		coreResult, err := session.InterfaceCoreCommandFile(accepting, corePath)
		if err != nil {
			t.Fatalf("InterfaceCoreCommandFile: %v", err)
		}
		if coreResult.Status != protocol.StatusPass {
			t.Fatalf("core status = %s, want %s", coreResult.Status, protocol.StatusPass)
		}

		checkResult, err := session.InterfaceCheckCommandFile(summaryPath, corePath)
		if err != nil {
			t.Fatalf("InterfaceCheckCommandFile: %v", err)
		}
		if checkResult.Status != protocol.StatusPass {
			t.Fatalf("round-trip check status = %s, want %s (the digest bound at export time must still match the bytes InterfaceCoreCommandFile wrote)", checkResult.Status, protocol.StatusPass)
		}
	})

	// Test 5 (edge 2, fail-closed): !Valid with an empty Problems slice
	// yields StatusInvalid with the fallback code, on both paths -- driven
	// through the shared peerRefusalDiagnostic helper, since no real
	// program produces an empty Problems slice.
	t.Run("edge 2: empty Problems fails closed on both interface paths too", func(t *testing.T) {
		diag := session.PeerRefusalDiagnosticForTest(corevalidate.Result{Valid: false}, "phase07.example")
		if diag.Code != session.PeerRefusalUnnamedCodeForTest {
			t.Fatalf("code = %q, want the fail-closed fallback %q", diag.Code, session.PeerRefusalUnnamedCodeForTest)
		}
	})
}

// TestCheckCommandPeerConsultMutationKilled is Task 3's own observation
// that control:check.peer_consulted fails under its seeded fault (QLT-08):
// a control never seen to fail is not evidence. With checkCommandPeerSeam
// enabled, CheckCommandFile skips the peer entirely and WRONGLY reports
// StatusPass on relay_escort_witness.lang; with the seam restored, it
// correctly refuses with core.move_while_borrowed. Both directions are
// asserted in one test, so a control that is green for the wrong reason
// (e.g. a no-op seam) fails here.
func TestCheckCommandPeerConsultMutationKilled(t *testing.T) {
	// relay_escort_witness.lang was this test's fixture through Phase 07:
	// check itself admitted it, so a peer-consult seam that wrongly skips
	// the peer entirely made CheckCommandFile wrongly report StatusPass,
	// proving the peer consult load-bearing. Phase 08 closes the
	// interprocedural half of D-03-02 (Task 2/3): check now independently
	// refuses this exact fixture via its own new
	// check.interprocedural_loan_liveness law, BEFORE CheckCommandFile ever
	// reaches the peer-consult seam -- so this fixture can no longer
	// demonstrate the seam is load-bearing (it would report StatusInvalid
	// regardless of the seam). duplicate_function_name.lang is the
	// remaining declared divergence (peerDivergenceExpected): check admits
	// it cleanly and only the peer's independent function-ID uniqueness
	// check refuses it, which is exactly the shape this mutation-kill needs.
	fixture := phase07Fixture(t, "duplicate_function_name.lang")

	restore := session.SetCheckCommandPeerSeamForTest(true)
	wronglyPermissive, err := session.CheckCommandFile(fixture)
	restore()
	if err != nil {
		t.Fatalf("CheckCommandFile (seam on) returned an error: %v", err)
	}
	if wronglyPermissive.Status != protocol.StatusPass {
		t.Fatalf("with the seam ON, expected the mutation to wrongly report %s, got %s -- the seam is not load-bearing", protocol.StatusPass, wronglyPermissive.Status)
	}

	correctlyRefusing, err := session.CheckCommandFile(fixture)
	if err != nil {
		t.Fatalf("CheckCommandFile (seam restored) returned an error: %v", err)
	}
	if correctlyRefusing.Status != protocol.StatusInvalid || len(correctlyRefusing.Diagnostics) != 1 || correctlyRefusing.Diagnostics[0].Code != "core.duplicate_function_id" {
		t.Fatalf("with the seam restored, expected exactly one core.duplicate_function_id diagnostic, got status=%s diagnostics=%+v", correctlyRefusing.Status, correctlyRefusing.Diagnostics)
	}
}

// TestInterfacePeerRefusalMutationKilled is Task 3's own observation that
// control:interface.peer_refusal_is_invalid fails under its seeded fault:
// with interfacePeerRefusalSeam enabled, both interface command paths
// restore the pre-07-10 behaviour of a non-nil error with no StatusInvalid
// result (the code is discarded); with it restored, both report
// StatusInvalid carrying the peer's own code. One seam kills the control
// on both paths, asserted here on both.
func TestInterfacePeerRefusalMutationKilled(t *testing.T) {
	fixture := phase07Fixture(t, "duplicate_function_name.lang")

	t.Run("InterfaceCoreCommandFile", func(t *testing.T) {
		restore := session.SetInterfacePeerRefusalSeamForTest(true)
		wrongResult, wrongErr := session.InterfaceCoreCommandFile(fixture, filepath.Join(t.TempDir(), "out.core.json"))
		restore()
		if wrongErr == nil {
			t.Fatalf("with the seam ON, expected a non-nil error (the mutation restores discarding the peer's code), got result=%+v", wrongResult)
		}
		if wrongResult.Status == protocol.StatusInvalid {
			t.Fatal("with the seam ON, a non-nil error must not also carry a StatusInvalid result -- the seam is not load-bearing")
		}

		correctResult, correctErr := session.InterfaceCoreCommandFile(fixture, filepath.Join(t.TempDir(), "out.core.json"))
		if correctErr != nil {
			t.Fatalf("with the seam restored, expected a nil error, got %v", correctErr)
		}
		if correctResult.Status != protocol.StatusInvalid || len(correctResult.Diagnostics) != 1 || correctResult.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("with the seam restored, expected exactly one core.duplicate_function_id diagnostic, got status=%s diagnostics=%+v", correctResult.Status, correctResult.Diagnostics)
		}
	})

	t.Run("InterfaceExportCommandFile", func(t *testing.T) {
		restore := session.SetInterfacePeerRefusalSeamForTest(true)
		wrongResult, wrongErr := session.InterfaceExportCommandFile(fixture, filepath.Join(t.TempDir(), "out.summary.json"))
		restore()
		if wrongErr == nil {
			t.Fatalf("with the seam ON, expected a non-nil error (the mutation restores discarding the peer's code), got result=%+v", wrongResult)
		}
		if wrongResult.Status == protocol.StatusInvalid {
			t.Fatal("with the seam ON, a non-nil error must not also carry a StatusInvalid result -- the seam is not load-bearing")
		}

		correctResult, correctErr := session.InterfaceExportCommandFile(fixture, filepath.Join(t.TempDir(), "out.summary.json"))
		if correctErr != nil {
			t.Fatalf("with the seam restored, expected a nil error, got %v", correctErr)
		}
		if correctResult.Status != protocol.StatusInvalid || len(correctResult.Diagnostics) != 1 || correctResult.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("with the seam restored, expected exactly one core.duplicate_function_id diagnostic, got status=%s diagnostics=%+v", correctResult.Status, correctResult.Diagnostics)
		}
	})
}

// phase07Fixture resolves a testdata/phase07 fixture's absolute path,
// mirroring this package's other phase07 test helpers' precedent.
func phase07Fixture(t testing.TB, name string) string {
	t.Helper()
	if strings.Contains(name, string(filepath.Separator)) || strings.Contains(name, "/") {
		t.Fatalf("phase07Fixture expects a bare filename, got %q", name)
	}
	return testsupport.ProjectPath("testdata", "phase07", name)
}
