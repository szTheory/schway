package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// payloadCorpusPhaseDirs is the pre-Phase-12 corpus (D-12-19): every
// testdata/phaseN directory that existed BEFORE this phase. Phase 12's own
// fixture directory is DELIBERATELY excluded from this list -- it did not
// exist before this phase, so it carries no pre-widening baseline to
// characterize.
var payloadCorpusPhaseDirs = []string{
	"phase1", "phase2", "phase3", "phase4", "phase5", "phase6",
	"phase07", "phase08", "phase10", "phase11",
}

// payloadCorpusFixtures globs every .schway fixture under payloadCorpusPhaseDirs,
// following the phaseArtifactGlob precedent (session_test.go) rather than a
// hand-enumerated table (D-12-19's ratified one-globbing-test shape,
// 12-02-SUMMARY.md).
func payloadCorpusFixtures(t *testing.T) []string {
	t.Helper()
	var found []string
	for _, phaseDir := range payloadCorpusPhaseDirs {
		matches, err := filepath.Glob(testsupport.ProjectPath("testdata", phaseDir, "*.schway"))
		if err != nil {
			t.Fatalf("glob testdata/%s: %v", phaseDir, err)
		}
		for _, match := range matches {
			found = append(found, "testdata/"+phaseDir+"/"+filepath.Base(match))
		}
	}
	sort.Strings(found)
	return found
}

// payloadCorpusBaseline pins the SHA-256 digest of every pre-Phase-12
// fixture's interp execution-document(s), as committed golden bytes -- never
// a recomputed self-comparison (D-12-19/D-12-41's own anti-vacuity
// prohibition). Honesty note (flagged_assumption 3, 12-04-PLAN.md): plan 02's
// interp value-widening (D-12-17) already landed in this tree BEFORE this
// baseline was captured, so this digest set is a POST-widening snapshot, not
// a true pre-widening one -- a pre-widening baseline can no longer be
// captured from this tree. TestPayloadCorpusCharacterizationReplayMutationKilled
// below is what keeps this replay load-bearing despite that: it proves that
// if the widening's own evidence-invisibility guarantee (D-12-18) had NOT
// held, at least one of these digests would differ, which is exactly the
// property a byte-identity replay must be able to detect. See
// 12-04-SUMMARY.md for the full accounting.
//
// A fixture present on disk but absent from this map is either an expected
// skip (see payloadCorpusExpectedSkips) or a corpus regression this test
// must fail loudly on -- never silently absorbed.
var payloadCorpusBaseline = map[string]string{
	"testdata/phase07/call_argument_used_once.schway":                "eea9b3674d270953a6eefa0c6968cccd2f21ecc6e89e97211625f511dcf65584",
	"testdata/phase07/call_basic.schway":                             "b8ecdb83da07c8fce3f718aabc8664caf6733107a68d35caf14e7b59b86c0cb0",
	"testdata/phase07/call_fallible_foreign_reach.schway":            "b0876d93032b694670b5f0d0d22a78b05b02728444ef140471fb26d52638b15a",
	"testdata/phase07/call_from_both_match_arms.schway":              "f1d79f86e5c9977446ae6692df547b9318a2614b42c125b8f1bb72b72a38d3d2",
	"testdata/phase07/call_two_fallible_callees_disagree.schway":     "b7e078c4985a6af3fea0e2df7de8b12417c912ebeba84c64bb7259233787a2a7",
	"testdata/phase07/clean_but_unpublishable.schway":                "dce0135c33578c4b1f9070cbfb4e27c522313e3d0f9e72af3dcb3c63fe6d0cc2",
	"testdata/phase07/deep_diamond_acyclic.schway":                   "f1ce314cc0779b7accccb97343775d3eb95558e1652e9a7a0373d009115c295e",
	"testdata/phase08/match_arm_call.schway":                         "7016438cd49e14dbf3f6ae6d1331265f3325b3cc19737bc5185d3b5823d5a8b4",
	"testdata/phase08/relay_depth2_accept.schway":                    "c38134de3bcdc7196f6cdf34393e47968d6bf80e65bca59cdd01ab73c662be16",
	"testdata/phase08/twin_a_accept.schway":                          "b3ba6c8a4b2900f42e3bf1ee0d1ba8de70cd24fe0ced6f64646624ebfb64c93f",
	"testdata/phase1/comments.schway":                                "c56e5a35829531d2be88882b74bfe436d678c5fe0787ce72d415a54af1768c58",
	"testdata/phase1/toggle.schway":                                  "c56e5a35829531d2be88882b74bfe436d678c5fe0787ce72d415a54af1768c58",
	"testdata/phase10/compose_per_path_borrow_callee_accept.schway":  "bdbc01678fb4f8864bdeba713b99f0e99da972b614989392d61570768de1a49d",
	"testdata/phase10/compose_per_path_borrow_caller_accept.schway":  "a4eec88f060400b559d877b7509ae5f2233976b4b8bf73b72c17588c9cf6192f",
	"testdata/phase10/relay_depth3_accept.schway":                    "44b7480d43d0300de8d794c4edef7865a30fd8e991ae42ea428302c97c96af16",
	"testdata/phase11/multi_function_diamond_call.schway":            "0ed146c144f1e4677d370081cb6c8bf02d187ca58a5777e6d16ee03b04ac7355",
	"testdata/phase11/multi_function_entry_basic.schway":             "0b94304e30b9b36b61a8ce9658dde4128613ff86166357ad788c0c8d327e9cde",
	"testdata/phase11/multi_function_forward_callee.schway":          "860db8d74e0b524f3bd7d4dc4013355fc12d5f556ee074cb980ad345e15cd864",
	"testdata/phase11/multi_function_gate_corpus.schway":             "9ebb4ee7313ffeee3d30c4154e44590d53ef44a8a06e845d356da96c39b2ac02",
	"testdata/phase11/multi_function_reduce_gate.schway":             "5af27de09fbb5524adcdd193424b0176cb6d370b439cb9b2203c738ee5fd9cf7",
	"testdata/phase11/multi_function_relay_depth2.schway":            "197d2985f85259bbb14ed3b6c99653267a1d93740841b5f41b1a714a8fc2b404",
	"testdata/phase11/multi_function_unreachable.schway":             "cb3b7772161a0112f757f3a8ed3deddc41b78e88356189d96a59b04c4dd759dc",
	"testdata/phase2/implicit_copy.schway":                           "25a3a30cbee7d5c23cde7b06483478e28d61fc9dafaae58cd4a87beedba612f3",
	"testdata/phase2/owned_transfer.schway":                          "e9f2bd41d6670ae79dbd6b194e39d511ff3d2fe1097855a2640b5f0e12207228",
	"testdata/phase3/borrowed_view.schway":                           "34446ca36cdf3e19b59b002098894c85d0a009c93e898dfc135c81e004116a15",
	"testdata/phase3/branch_one_arm_shared_accept.schway":            "a42ec278e1ab84d35937c4c3acd07f70141d240b1aa552ed81f6b310b72d5648",
	"testdata/phase3/branch_view.schway":                             "235de743e26d03e5939fd6ed9b11a9b656559c7153d47a4c0ddc923829cc8108",
	"testdata/phase3/public_view.schway":                             "3fae5012db2aecbab9c4062010e331833bf7cf0a6706a71902976c7c91c9d66e",
	"testdata/phase3/public_view_impossible.schway":                  "8ce622b2c206dce5a1c918dc2161af5fda0ea3b5903987acb733773493319cc0",
	"testdata/phase3/public_view_mixed_access.schway":                "49a6d9655d5eb0a8474f537e7475f9ed4768148cebad81d2e04a6a534a33ed09",
	"testdata/phase3/public_view_multi_arm_access_conflict.schway":   "4db2c9dc9b895216c8f817411f32249da59d0363584c08dc927c6fa5d7096284",
	"testdata/phase3/public_view_multi_arm_omitted.schway":           "a277e9c262048dbe34905d995a1e26405a0113b301dcbb758ac10253a92146c1",
	"testdata/phase3/public_view_omitted.schway":                     "29637d6eb9cbba377323b7ca1da1434e4ec022f732c926f00ac0f404651cf9c3",
	"testdata/phase3/public_view_understated.schway":                 "90b1f67def6323b89f18b6d36f2aed5c2314cb002b7db534a855d683c1b4f1ed",
	"testdata/phase3/sequential_shared_then_exclusive_accept.schway": "a7a4da2bb40d411a4e19b050255c805c81a27aa5dbb84cebea2d5732ea6118e4",
	"testdata/phase3/shared_shared_accept.schway":                    "c46923382c6fb6de54d0b0a8a4f83b28502c0dbeaa19ad86553b1d54ef7d92b2",
	"testdata/phase4/acquire_three_fail_second.schway":               "b96bc1efbc8bfa26486bb6484bb0522b91c434e9d5a2af47b599431eea3d7a39",
	"testdata/phase4/acquire_three_fail_third.schway":                "66e4ae512ac3952a6c3460b69728551468911fb19ca1da8a08c836d47c36b3a1",
	"testdata/phase4/acquire_three_success.schway":                   "1be050cd813aa7daea06bb93dbde7723fb5b04d7dbeacb47ab7616c665904306",
	"testdata/phase4/defect_terminal.schway":                         "c901ae4e7d8f51b123e2f24347ff18c78232f59155ea9980c9850922f797d702",
	"testdata/phase4/discard_because.schway":                         "83ad11846d5e2071640a4ce818b46c2854185f8afe2ae20b73158f179453545d",
	"testdata/phase4/foreign_acquire_one.schway":                     "8f25baba4ad87a1451ddf11ea8fdb5e2f3d51d0a3cb5af43c4537bffd60d0763",
	"testdata/phase4/foreign_origin_omitted.schway":                  "283b185adf82a5aadd432bff4aa1ae31b06da137cb56a12e4ba752103f53dc9e",
	"testdata/phase4/nonlocal_exit_probe.schway":                     "15901e3fd303beac7c3004d4625751d40fc0d396a435e837b3defa50bb677acf",
	"testdata/phase5/allocator_mismatch.schway":                      "631468f7abf7b52ea8a48c05ae2ac12bcd2a5b598f42be9d6825fced17a6f2fb",
	"testdata/phase5/coordinated_lie.schway":                         "a398d7742f92ecfbcd1c4f99f59ec446e05bb12839a0e4119592ad757168cb15",
	"testdata/phase5/dead_store_unused_acquire.schway":               "bf3fb69bd2e7770e437d1b7560e787824540b40c69a62a0ff6c17eccb30ff555",
	"testdata/phase5/defect_dies_by_signal.schway":                   "175956eedbd91176eb09003d7cae170e6a9f034245c83d8df6bd1d4160ac22dd",
	"testdata/phase5/false_restrict_hoist.schway":                    "faae54d0f7ee7219ee010aaea28255e5a46712e45f39d02b774a454901d52177",
	"testdata/phase5/inline_across_foreign.schway":                   "409f0367d55d78814a052a707f3f8e01b63572f3a650e801c0e3a0d95532a5a5",
	"testdata/phase5/reorder_two_events.schway":                      "d978fe761da19fd5fd5da801fd6abeef74ace7823ff025553eb3719b0d781632",
	"testdata/phase5/restrict_borrow.schway":                         "bacfe22b2f1015ea845feff5c5cec776bcf3e65d2b793e568c9eab99e33c06b5",
	"testdata/phase5/retained_pointer.schway":                        "f9a698d7e8ab3b34a19b75997e6a40807bd6eaae4807689e0aedc54002647c84",
	"testdata/phase5/tail_collapse_release_ladder.schway":            "982b8efbc87095597fc658907db890a46e9eb290ab777fb17c9355c02b3fce4b",
	"testdata/phase5/typed_failure_truncated_stdout.schway":          "b0674de93c832db9562c0d69c2a1b2d36fa3837d7d3c59927f04992d21f21129",
	"testdata/phase6/derivation_borrow_defect.schway":                "445ecaee7c936364a24802766f29d423a790ed537d8a3e67e62a34dbc03484ab",
	"testdata/phase6/derivation_match_defect.schway":                 "3d31038add620c9f3d6ec0bd32d677a83436f9ebe5a878aa2d3066f3ac8e52eb",
	"testdata/phase6/derivation_move_defect.schway":                  "d875445991a257506e5eece10afac4c3036fae8db670b374c7c9e374bec5e981",
	"testdata/phase6/heldout_borrow_defect.schway":                   "bb20c5bf97e624e6d993920b40d81cba4b93c7cd9beed0c36e13fbe492125e4c",
	"testdata/phase6/heldout_match_defect.schway":                    "3e054420dbb9907f7245d33bfe6e80a8b14f20932846dda5135f599929eb1d7d",
	"testdata/phase6/heldout_move_defect.schway":                     "91eb15080636142a8264eab31ae26efd2720638c070de0b2f34efdd64a2650c7",
	"testdata/phase6/stale_evidence_subject.schway":                  "d0c214b7fbf9a97f5c9fbc6b0b311877ae8544a4e306e5b1421cb50c65197cd8",
}

// payloadCorpusPhase15Schema2Baseline is the explicit successor ledger for
// pre-Phase-12 fixtures whose multi-function interpreter documents Phase 15
// intentionally migrated from /1 to /2. Keeping this separate from
// payloadCorpusBaseline preserves the historical D-12-19 snapshot while
// making every authorized byte movement named, schema-checked, and pinned.
// A fixture absent here must still match its original legacy digest exactly.
var payloadCorpusPhase15Schema2Baseline = map[string]string{
	"testdata/phase07/call_argument_used_once.schway":               "39fcf77d9d3215d4d4cfb6873ebca0e8ac83d5aed3f7ee8e16f86f941d38fa57",
	"testdata/phase07/call_basic.schway":                            "c62209ed506b2fcdb9b1f6f78030cea715c8695400656ab5d1b5ee2c900a70ee",
	"testdata/phase07/call_fallible_foreign_reach.schway":           "ab8193bf320aeb9b1721231932a3ad042379fa70fa33724f2a5baa5be2fcf6f0",
	"testdata/phase07/call_from_both_match_arms.schway":             "735bd0ed068d637ed7446a2bf1b68587d3e53e32755615335eb6dcc3fb833119",
	"testdata/phase07/call_two_fallible_callees_disagree.schway":    "8f3caa3c5f55a608b07abf1a80a3368a858a34fc1aa7b4cccb9e9edcf689ec06",
	"testdata/phase07/deep_diamond_acyclic.schway":                  "06a0d63aa3e0c58e868de8bd5a228402c062d6b6fb5e22164b87e0d7bede45dc",
	"testdata/phase08/match_arm_call.schway":                        "b69308cd8dcaa9ccab60fcacce0a43a6c6a1815e423e98ea0cc9c1e29fb90429",
	"testdata/phase08/relay_depth2_accept.schway":                   "57e99ce7fd7e728606d99ff561091579576867d32b1aa88d36d083a45cf51c7c",
	"testdata/phase08/twin_a_accept.schway":                         "55e7c68aa48d6438bde0c3b945eb7b60db7fd2efcf80bfa96d5c40afd796fc1b",
	"testdata/phase10/compose_per_path_borrow_caller_accept.schway": "c5e72c8bc6897cb3bbc67079aa2914e3cc0739dc0b22a2d7405dbd11bc3edf44",
	"testdata/phase10/relay_depth3_accept.schway":                   "9f90f91d18c959dd556d7cf41a7c0446b8c9fa0e11f9651cbf219f3821020d79",
	"testdata/phase11/multi_function_diamond_call.schway":           "61e2d14a6cfe432a6b1fbeec4f469faa4777f903c569f5661bbf6ff23e7f40b6",
	"testdata/phase11/multi_function_entry_basic.schway":            "2f83a2fb76e88c003e2851283860b7d4d7df3ca969c09639aee4bd54f4e27407",
	"testdata/phase11/multi_function_forward_callee.schway":         "246585a63c13f9055bbf38b68379bd70443c5c41c890ad8a99239f7838449488",
	"testdata/phase11/multi_function_gate_corpus.schway":            "3d56cdc6832d970f7f5e8c2a8dc825b60248c6615947227fda6a760094467871",
	"testdata/phase11/multi_function_reduce_gate.schway":            "cff57dd44f306a348858b5a08b3bd3ce01c00e6eaf12e77cb95a48ee6b563537",
	"testdata/phase11/multi_function_relay_depth2.schway":           "c1c3010c1e9c9f20a23057a3e2f415a8fc74f6135e123294ec45aa2444f71579",
	"testdata/phase11/multi_function_unreachable.schway":            "421cc70866dcf72fd74b15ae55dc211dce578ef9283d8c1dfe8d5cb85eb626ed",
}

// payloadCorpusExpectedSkips names every pre-Phase-12 fixture that does NOT
// check cleanly, or checks cleanly but cannot run through interp for a
// reason unrelated to Phase 12 (a negative-control fixture, a deliberately
// malformed source, an intentionally cyclic/ambiguous-entry program, etc.).
// This is the explicit "named and why" ledger D-12-19 requires: a fixture
// silently dropping out of both this set and payloadCorpusBaseline is a
// corpus regression the test below must catch, not absorb.
var payloadCorpusExpectedSkips = map[string]bool{
	"testdata/phase07/call_argument_used_twice.schway":       true,
	"testdata/phase07/call_type_mismatch.schway":             true,
	"testdata/phase07/call_uncallable_callee.schway":         true,
	"testdata/phase07/cycle_indirect.schway":                 true,
	"testdata/phase07/cycle_mutual.schway":                   true,
	"testdata/phase07/cycle_self.schway":                     true,
	"testdata/phase07/cycle_through_match_arm.schway":        true,
	"testdata/phase07/cycle_unreachable.schway":              true,
	"testdata/phase07/duplicate_function_name.schway":        true,
	"testdata/phase07/foreign_symbol_shadowing.schway":       true,
	"testdata/phase07/relay_escort_witness.schway":           true,
	"testdata/phase08/negative_control_fails.schway":         true,
	"testdata/phase08/negative_control_infallible.schway":    true,
	"testdata/phase08/relay_depth2_refuse.schway":            true,
	"testdata/phase08/twin_a_refuse.schway":                  true,
	"testdata/phase08/twin_b_accept.schway":                  true,
	"testdata/phase08/twin_b_refuse.schway":                  true,
	"testdata/phase1/malformed.schway":                       true,
	"testdata/phase1/non_exhaustive.schway":                  true,
	"testdata/phase10/relay_depth3_refuse.schway":            true,
	"testdata/phase11/multi_function_zero_call.schway":       true,
	"testdata/phase2/ability_shapes.schway":                  true,
	"testdata/phase2/implicit_noncopy.schway":                true,
	"testdata/phase2/move_while_borrowed.schway":             true,
	"testdata/phase2/reborrow_while_moved.schway":            true,
	"testdata/phase2/use_after_move.schway":                  true,
	"testdata/phase3/branch_one_arm_shared_reject.schway":    true,
	"testdata/phase3/exclusive_exclusive_reject.schway":      true,
	"testdata/phase3/exclusive_move_reject.schway":           true,
	"testdata/phase3/shared_exclusive_reject.schway":         true,
	"testdata/phase4/fallible_call_unconsumed.schway":        true,
	"testdata/phase4/foreign_call_target_not_foreign.schway": true,
	"testdata/phase4/foreign_policy_value_injection.schway":  true,
	"testdata/phase4/foreign_unwind_undeclared.schway":       true,
	"testdata/phase5/explain_use_after_move.schway":          true,
}

// digestExecutions hashes every interp.Execution's own CanonicalBytes
// projection into one SHA-256 digest, in order -- the multi-execution case
// (a match-bodied fixture driven once per alternative) folds into a single
// per-fixture digest exactly like previousPhaseGoldenCDigests folds one file
// into one digest.
func digestExecutions(executions []interp.Execution) (string, error) {
	h := sha256.New()
	for _, execution := range executions {
		bytes, err := interp.CanonicalBytes(execution)
		if err != nil {
			return "", err
		}
		h.Write(bytes)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// TestPayloadCorpusCharacterizationReplay is D-12-19's own committed
// corpus-characterization replay: for every pre-Phase-12 fixture, either it
// checks cleanly and its interp execution document(s) hash to the pinned
// payloadCorpusBaseline digest, or it is a named, pinned expected skip
// (payloadCorpusExpectedSkips). A fixture in neither set fails loudly --
// this is what discriminates a real regression from a no-op, since it
// exercises the real session.Check/corevalidate/interp path per fixture
// rather than a hand-picked property.
//
// See payloadCorpusBaseline's own doc comment for the honest accounting of
// what "pre-Phase-12 baseline" means here: plan 02's interp widening (D-12-17)
// already landed before this digest set was captured, so this is a
// post-widening snapshot pinned as a forward-looking regression tripwire,
// not literal pre-widening evidence. TestPayloadCorpusCharacterizationReplayMutationKilled
// is what proves this replay is load-bearing despite that honesty gap.
func TestPayloadCorpusCharacterizationReplay(t *testing.T) {
	fixtures := payloadCorpusFixtures(t)
	if len(fixtures) == 0 {
		t.Fatal("payloadCorpusFixtures found zero fixtures -- glob is broken")
	}

	seenBaseline := make(map[string]bool, len(payloadCorpusBaseline))
	seenSkip := make(map[string]bool, len(payloadCorpusExpectedSkips))

	for _, relative := range fixtures {
		relative := relative
		t.Run(relative, func(t *testing.T) {
			parts := strings.Split(relative, "/")
			path := testsupport.ProjectPath(parts...)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", relative, err)
			}

			checked := session.Check(source)
			if len(checked.Diagnostics) > 0 {
				if !payloadCorpusExpectedSkips[relative] {
					t.Fatalf("%s: does not check cleanly (%v) and is not in payloadCorpusExpectedSkips -- if this is expected, add it there and name why; if not, this is a regression", relative, checked.Diagnostics)
				}
				seenSkip[relative] = true
				t.Skipf("D-12-19: does not check cleanly, expected (named in payloadCorpusExpectedSkips): %v", checked.Diagnostics)
				return
			}

			executions, diagnostics, runErr := session.RunInterpreterFile(path)
			if len(diagnostics) > 0 || runErr != nil {
				if !payloadCorpusExpectedSkips[relative] {
					t.Fatalf("%s: checks cleanly but cannot run through interp (diagnostics=%v err=%v) and is not in payloadCorpusExpectedSkips -- if this is expected (e.g. an ambiguous-entry or negative-control shape), add it there and name why; if not, this is a regression", relative, diagnostics, runErr)
				}
				seenSkip[relative] = true
				t.Skipf("D-12-19: checks cleanly but cannot run through interp, expected (named in payloadCorpusExpectedSkips): diagnostics=%v err=%v", diagnostics, runErr)
				return
			}

			got, err := digestExecutions(executions)
			if err != nil {
				t.Fatalf("%s: canonical bytes: %v", relative, err)
			}
			want, known := payloadCorpusBaseline[relative]
			if !known {
				t.Fatalf("%s: produced a clean execution document but is not in payloadCorpusBaseline -- add its pinned digest (got %s)", relative, got)
			}
			if phase15Want, migrated := payloadCorpusPhase15Schema2Baseline[relative]; migrated {
				for index, document := range executions {
					if document.Schema != execution.Schema2 {
						t.Fatalf("%s: Phase 15 migration ledger names execution %d, but its schema is %q, want %q", relative, index, document.Schema, execution.Schema2)
					}
				}
				want = phase15Want
			}
			seenBaseline[relative] = true
			if got != want {
				t.Fatalf("%s: interp execution document moved: got sha256 %s, want current pinned digest %s", relative, got, want)
			}
		})
	}

	for relative := range payloadCorpusBaseline {
		if !seenBaseline[relative] {
			t.Errorf("pinned baseline fixture %s no longer produces a clean execution document (dropped from the corpus, or now refused/errors)", relative)
		}
	}
	for relative := range payloadCorpusExpectedSkips {
		if !seenSkip[relative] {
			t.Errorf("expected-skip fixture %s no longer skips -- its status changed without this test being updated", relative)
		}
	}
}

// TestPayloadCorpusCharacterizationReplayMutationKilled is Task 2's
// mutation-kill beat: with interp's D-12-18 empty-tag serialization special
// case disabled via the unexported, test-only
// interp.SetDisableEmptyTagSerializationForTest seam, a fixture's interp
// execution document must diverge from the PINNED baseline above. Without
// this beat, TestPayloadCorpusCharacterizationReplay would be a control that
// reports pass without ever having exercised the hazard it names -- it could
// pass identically whether or not interp's value model preserved scalar
// evidence-invisibility at all.
func TestPayloadCorpusCharacterizationReplayMutationKilled(t *testing.T) {
	restore := interp.SetDisableEmptyTagSerializationForTest(true)
	defer restore()

	const relative = "testdata/phase2/owned_transfer.schway"
	path := testsupport.ProjectPath(strings.Split(relative, "/")...)
	executions, diagnostics, err := session.RunInterpreterFile(path)
	if len(diagnostics) > 0 || err != nil {
		t.Fatalf("%s: diagnostics=%v err=%v", relative, diagnostics, err)
	}
	got, err := digestExecutions(executions)
	if err != nil {
		t.Fatalf("canonical bytes: %v", err)
	}
	want, known := payloadCorpusBaseline[relative]
	if !known {
		t.Fatalf("%s missing from payloadCorpusBaseline -- fix the fixture reference", relative)
	}
	if got == want {
		t.Fatalf("expected disabling the empty-tag serialization special case to move %s's digest away from the pinned baseline %s, but it matched -- TestPayloadCorpusCharacterizationReplay is not load-bearing", relative, want)
	}
}

// scalarSurvivesSequenceProgram builds, directly at core-IR level (matching
// this project's own synthetic-program test style), a function whose body
// takes a scalar Byte through copy, borrow, and move -- D-12-19's own named
// "narrow property test that a scalar value survives an arbitrary
// copy/move/borrow sequence unchanged" -- without depending on any corpus
// fixture.
func scalarSurvivesSequenceProgram() core.Program {
	functionID := "s1:payload_replay_property:fn:identity"
	typeID := functionID + ":type:0"
	paramID := functionID + ":place:0"
	copiedID := functionID + ":place:1"
	viewID := functionID + ":place:2"
	observedID := functionID + ":place:3"
	movedID := functionID + ":place:4"
	return core.Program{
		Schema: core.Schema1, Module: "payload_replay_property", ModuleID: "s1:payload_replay_property:module",
		Functions: []core.Function{{
			ID: functionID, Name: "identity",
			EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter: core.Parameter{ID: paramID, Name: "input", Type: "Byte"}, ReturnType: "Byte",
			Linear: &core.LinearBody{
				ID: functionID + ":linear",
				Types: []core.TypeFact{{
					ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
					Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
					NegativeWitnesses: []core.AbilityWitness{},
				}},
				Places: []core.Place{
					{ID: paramID, Name: "input", TypeID: typeID},
					{ID: copiedID, Name: "copied", TypeID: typeID},
					{ID: viewID, Name: "view", TypeID: typeID},
					{ID: observedID, Name: "observed", TypeID: typeID},
					{ID: movedID, Name: "moved", TypeID: typeID},
				},
				Operations: []core.LinearOperation{
					{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpCopy, SourceID: paramID, TargetID: copiedID, TypeID: typeID},
					{ID: functionID + ":op:1", PointID: functionID + ":point:linear:1", Kind: core.OpBorrowShared, SourceID: copiedID, TargetID: viewID, LoanID: functionID + ":loan:0", TypeID: typeID},
					{ID: functionID + ":op:2", PointID: functionID + ":point:linear:2", Kind: core.OpCopy, SourceID: viewID, TargetID: observedID, TypeID: typeID},
					{ID: functionID + ":op:3", PointID: functionID + ":point:linear:3", Kind: core.OpMove, SourceID: copiedID, TargetID: movedID, TypeID: typeID},
					{ID: functionID + ":op:4", PointID: functionID + ":point:linear:4", Kind: core.OpReturn, SourceID: movedID, TypeID: typeID},
				},
			},
		}},
	}
}

// TestScalarValueSurvivesCopyMoveBorrowSequence is D-12-19's narrow
// companion property test: an arbitrary copy/borrow/copy/move sequence over
// one scalar Byte value must return that value UNCHANGED. Kept narrow --
// TestPayloadCorpusCharacterizationReplay above is the discriminating
// mechanism, because it exercises the real oracle path over the whole
// corpus; this property test exists only to state the invariant directly,
// independent of any one fixture.
func TestScalarValueSurvivesCopyMoveBorrowSequence(t *testing.T) {
	program := scalarSurvivesSequenceProgram()
	const input = "7"
	execution, err := interp.Run(program, "identity", input)
	if err != nil {
		t.Fatalf("interp.Run: %v", err)
	}
	if execution.Outcome.Value != input {
		t.Fatalf("expected the scalar to survive copy/borrow/copy/move unchanged, got outcome value %q, want %q", execution.Outcome.Value, input)
	}
}
