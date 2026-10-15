---
status: resolved
trigger: "Phase36 Plan36-06 native ByteView caller helper composition refuses before execution"
created: 2026-10-15
updated: 2026-10-15
---

## Current Focus

hypothesis: Native dynamic invocation support is selected only for scalar-formal function bodies and does not compose with the existing owning wrapper to ByteView route.
test: Run the named native caller reproduction, then inspect emitter applicability and the existing ByteView/native evidence path.
expecting: Existing ByteView caller evaluates byte_at once before entering its U64 scalar helper; valid index enters the helper, OOB exits 65 with zero helper calls.
bug_class: Bohrbug (deterministic applicability refusal)
next_action: None for this debug correction; parent resumes halted Plan36-06 execution and owns phase completion decisions.
reasoning_checkpoint:
  hypothesis: hasScalarValueCalls disables dynamic support for the entire graph when it encounters the established owner-to-ByteView call, so ByteView scalar helper calls reach the CFG emitter without dynamic support.
  confirming_evidence:
    - hasScalarValueCalls returns false on every non-scalar call and requires all caller places to be U64/Bool.
    - hasScalarOperations already routes ByteView byte_at and byte_len bodies to emitProgramScalarFunction, whose OpCall requires dynamic support.
  falsification_test: If enabling only eligible scalar calls alongside validated legacy owner calls does not remove the refusal, the applicability diagnosis is insufficient.
  fix_rationale: Select dynamic support for checked scalar value calls independently of unchanged legacy call emitters; allow borrowed ByteView caller bodies without adding any ByteView scalar argument signature.
  blind_spots: Native loop invocation identity, independent peer behavior, frozen generated C and capture exhaustion still require fresh execution.
  candidate_causes:
    - code: all-call scalar classifier is stricter than the promised composition boundary
    - config/environment: binding manifest or Clang could prevent application construction
    - data: byte index equal to length intentionally faults, but cannot explain build-time refusal
  and_gate: no; the global applicability predicate alone explains the deterministic refusal before Clang execution

## Symptoms

expected: Existing owning wrapper to ByteView functions can call the new U64/Bool scalar helpers with scalar expressions; native O0/O3/O3-LTO agree with interpreter argument order, values, faults, and invocation attribution.
actual: Interpreter caller boundary passes, but session.BuildApplication refuses before a native executable exists.
errors: operation "s1:examples.phase29.adler32_fixture:fn:adler32:op:4": scalar call emission requires dynamic invocation support
timeline: Exposed during Phase36 Plan36-06 after scalar native call support in Plan36-05; no claim that this previously worked.
reproduction: internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeByteViewCallerArgumentRefusal, using the Phase29 fixture binding manifest and identity(byte_at(bytes, byte_len(bytes))). This test currently expects the refusal only to preserve a concrete blocker, not semantic success.

## Evidence

- timestamp: 2026-10-15
  checked: Manager independent rerun after final production/test edits; /tmp/phase36-byteview-manager-consumer.log, /tmp/phase36-byteview-manager-adjacent.log, /tmp/phase36-byteview-manager-peers.log, /tmp/phase36-byteview-manager-frozen.log
  found: All four commands exit 0. Four new consumer/control roots execute 21 native consumer lanes plus three semantic argument-order mutant kills with clean/restored controls; eight adjacent roots and twelve selected independent-peer/native-capacity roots pass. Both frozen baseline and control roots pass after the authorized three-row ledger metadata repair. No skips, zero selectors, refreshed goldens, whole-suite run, or human UAT claim.
  implication: Required native composition behavior and scoped compatibility gates are independently verified; correction committed as 5d874384.

- timestamp: 2026-10-15
  checked: /tmp/phase36-byteview-final-consumer.log (exit 0, 19.776s)
  found: Four roots pass, including 21 consumer lanes (five fixture boundaries plus fixture loop plus real copied FileBytes/imported ByteView loop, each O0/O3/O3-LTO), and three native argument-order mutants. Each fixture lane executes an unmodified session.BuildApplication application and a separate observed application; FileBytes lanes use unmodified session.BuildApplicationFile. All successful captures pass executionpeer.Validate; the FileBytes loop returns 928 with nine helper calls.
  implication: Both owning routes compose with scalar helpers; no file CLI, backend, schema, dependency, foreign contract or ownership rule changed.

- timestamp: 2026-10-15
  checked: /tmp/phase36-byteview-order-mutation.log and /tmp/phase36-byteview-final-consumer.log
  found: O0/O3/O3-LTO early-helper mutants execute real native helpers before the OOB byte_at and fail exactly at native helper invocation events=1 want 0. Each has a fresh no-op positive and a restored positive. Three semantic kills, zero skips or compile/refusal kills. Existing Go overlay harness reused; no Stryker or new mutation dependency.
  implication: The zero-entry oracle discriminates argument evaluation order independently of exit65 and the unchanged optional capture result.

- timestamp: 2026-10-15
  checked: /tmp/phase36-byteview-revert.log and /tmp/phase36-byteview-file-revert.log
  found: Overlaying only HEAD's pre-fix cgen_program.go reintroduces the exact scalar call emission requires dynamic invocation support refusal for both authored fixture roots and the file wrapper root. Normal production source passes those same roots. This is causal revert/reconfirm evidence, not a native semantic mutation kill.
  implication: The single production revision corrects applicability; no compensating interpreter or owner-emitter change was needed.

- timestamp: 2026-10-15
  checked: /tmp/phase36-byteview-focused.log, /tmp/phase36-byteview-adjacent-final.log and /tmp/phase36-byteview-peers.log
  found: Existing helper answers/optimization, UBSan, interpreter caller boundary, legacy ByteView bounds, legacy fallible/FileOwner peer routes, native capture separation, dynamic path storage/ordinal/capacity and raw/origin/execution peer controls all pass. git diff --check passes. Host Go1.24.0 darwin/arm64; Apple Clang21.0.0 arm64-apple-darwin25.6.0.
  implication: No observed semantic, owner/lifetime, attribution/storage, capture exhaustion or signature-boundary regression in the authorized focused sample.

- timestamp: 2026-10-15
  checked: /tmp/phase36-byteview-frozen-final.log (exit 1, 1.329s)
  found: Both frozen no-call roots fail before corpus comparison because existing crc_step/importer/public ledger rows lack the expected source_sha256 key. Errors are requires path, source hash, and reason; malformed hash; hash differs from working tree. Source fixtures and ledger are outside this debugger's edits; parent/manager owns the narrow metadata repair. No frozen C golden was refreshed.
  implication: Frozen gate remains pending independent metadata correction; do not claim the whole debug session resolved from semantic results alone.

- timestamp: 2026-10-15
  checked: /tmp/phase36-byteview-green.log
  found: Both authored roots passed all 18 native lanes. OOB at length, length+1 and empty view exits 65 with exact stderr and zero helper invocation events; valid first/last returns 90/100 with one helper call; the loop returns 928 with nine distinct helper invocation returns. Successful captures pass independent executionpeer validation.
  implication: Native semantics now compose; production changes are confined to scalar-call applicability and borrowed-ByteView caller classification. Unmodified session BuildApplication parity is now part of each lane as well as the test-only exit observer.

- timestamp: 2026-10-15
  checked: authored native boundary and loop selectors, /tmp/phase36-byteview-red.log
  found: Both roots fail at EmitApplication with scalar call emission requires dynamic invocation support. All five first/last/length/length+1/empty boundary cases reach that refusal; the repeated-loop case does too. Independent Python byte arithmetic pins Wikipedia sum 919 plus nine increments to 928.
  implication: Honest RED reproduces the exact applicability failure before production changes. Initial test compilation mistakes and a corrected arithmetic typo are not counted as semantic RED or mutation kills.

- timestamp: 2026-10-15
  checked: knowledge-base fallback, project skill discovery, cgen applicability and owner emitters
  found: No project skills exist. Prior legacy-call knowledge warns against moving ownership routes onto scalar emitters. The current failure occurs before Clang; the owner emitters already preserve static child invocation identities and release pairing.
  implication: Retain legacy emitters and new-call-only gating; use a mixed graph with scalar calls rather than widening the scalar call signature. SBFL skipped because no per-test coverage is available; no shrinking framework is configured, so use authored boundary neighbors.

- Plan36-06 terminal status is halted, recorded in 36-06-SUMMARY.md and commit 57060d8e.
- Independent scalar helper answers, public route, legacy ByteView callee bounds, and engine mutation controls pass according to the executor's focused receipts; manager must produce fresh verification for its correction.

## Eliminated

- hypothesis: binding manifest or installed Clang causes the original build failure
  evidence: the original failure occurs in EmitApplication before native construction; unchanged manifests and compiler successfully build all 21 corrected consumer lanes
  timestamp: 2026-10-15
- hypothesis: the intentional out-of-bounds byte index causes build-time refusal
  evidence: valid first/last and repeated-loop cases reproduce the same pre-fix applicability refusal; corrected code preserves runtime OOB exit65 and zero helper events
  timestamp: 2026-10-15

## Resolution

root_cause: hasScalarValueCalls used an all-call scalar-only predicate, so the existing owner-to-ByteView call disabled dynamic invocation support for a ByteView body that already used the scalar CFG emitter.
fix: Select eligible scalar calls independently of legacy calls, and admit only the existing ByteView parameter place in borrowed-view scalar caller bodies. Scalar call argument and result signatures remain U64/Bool; legacy owner emitters and lifetime checks remain unchanged.
oracle_type: specified values and exit/entry ordering; independently derived byte-sum literal 928
verification:
  target_test: { result: pass, consumer_lanes: 21, roots: [TestPhase36NativeByteViewCallerArgumentBoundary, TestPhase36NativeByteViewLoopHelpers, TestPhase36NativeFileByteViewLoopHelpers] }
  mutation_check: { result: pass, mechanism: existing Go source overlays, semantic_mutants_killed: 3, no_op_positives: 3, restored_positives: 3, applicability_refusals_counted_as_semantic_kills: 0 }
  no_op_deletion: { result: pass, deletion_justified_by_rca: true, detail: removes the erroneous all-call veto while retaining scalar signature and legacy owner checks }
  adjacent_tests: { result: pass, detail: manager independently reran focused scalar/helper/UBSan/capture/legacy/core/origin/execution/cgen paths and both frozen corpus roots after authorized ledger metadata repair }
  revert_and_reconfirm: { result: pass, bug_returned_on_revert: true, fixed_on_reapply: true, mechanism: HEAD production-file overlay without touching workspace edits }
  guardrail_verdict: pass
  cycle_counts: { production_fix_revisions: 1, honest_red_then_green: 1, semantic_order_mutation_lanes: 3 }
files_changed: [internal/compiler/cgen/cgen_program.go, internal/compiler/session/session_phase36_calls_test.go, internal/compiler/session/session_phase36_byteview_test.go, testdata/phase36/new-fixture-ledger.json]

### Exact focused commands and receipts

All commands run from /Users/jon/projects/schway. Every receipt contains root RUN/PASS (or intentional RED root FAIL); native skips and zero selectors are not accepted.

1. Honest RED and corrected fixture consumer: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NativeByteViewCallerArgumentBoundary|TestPhase36NativeByteViewLoopHelpers)$' -count=1 -v -timeout=180s` — RED `/tmp/phase36-byteview-red.log`; GREEN `/tmp/phase36-byteview-green.log`.
2. Final consumers and argument-order controls: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NativeByteViewCallerArgumentBoundary|TestPhase36NativeByteViewLoopHelpers|TestPhase36NativeFileByteViewLoopHelpers|TestPhase36ByteViewArgumentOrderMutationControls)$' -count=1 -v -timeout=300s` — PASS `/tmp/phase36-byteview-final-consumer.log`.
3. Original scalar/native and legacy sample: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NativeByteViewCallerArgumentBoundary|TestPhase36NativeByteViewLoopHelpers|TestPhase36HelperAnswersAndDefects|TestPhase36HelperOptimizationMatrix|TestPhase36NativeCallCaptureOutcomeSeparation|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase36LegacyFallibleCallRoutes|TestPhase36LegacyFileOwnerPeerRoutes|TestPhase36LegacyByteViewCalleeBounds)$' -count=1 -v -timeout=300s` — semantic roots PASS, frozen roots FAIL for ledger metadata `/tmp/phase36-byteview-focused.log`.
4. Focused independent peers and native storage/capacity: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/executionpeer ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^(TestPhase36NativeDynamicInvocationPaths|TestPhase36CallEventCapacityBoundary|TestPhase36NativeOrdinalAndPathBoundaries|TestPhase36EmitterCallInventoryControls|TestPhase36DynamicCallOrdinalControls|TestPhase36CallCausalAttributionControls|TestPhase36CallCFGApplicability|TestPhase36LegacyFileOwnerCalls|TestPhase36RawCallShapeAndFreshness|TestPhase36RawCallRefusalRemovalControls|TestPhase36OriginCallFreshnessAndTransfer|TestPhase36OriginCallRefusalRemovalControls)$' -count=1 -v -timeout=300s` — PASS `/tmp/phase36-byteview-peers.log`.
5. Argument-order control alone: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36ByteViewArgumentOrderMutationControls$' -count=1 -v -timeout=300s` — PASS `/tmp/phase36-byteview-order-mutation.log`.
6. Revert/reconfirm: `GOCACHE=/tmp/ai-lang-verification-gocache go test -overlay=/tmp/phase36-byteview-revert/overlay.json ./internal/compiler/session -run '^(TestPhase36NativeByteViewCallerArgumentBoundary|TestPhase36NativeByteViewLoopHelpers)$' -count=1 -v -timeout=180s` — original refusal returns `/tmp/phase36-byteview-revert.log`; same overlay with `-run '^TestPhase36NativeFileByteViewLoopHelpers$'` returns original file consumer refusal `/tmp/phase36-byteview-file-revert.log`. Normal-source restored consumer command is item2.
7. Final adjacent sample: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36HelperUBSan|TestPhase36ByteViewCallerArgumentBoundary|TestPhase36LegacyByteViewCalleeBounds|TestPhase36LegacyFallibleCallRoutes|TestPhase36LegacyFileOwnerPeerRoutes|TestPhase36NativeCallCaptureOutcomeSeparation)$' -count=1 -v -timeout=180s` — PASS `/tmp/phase36-byteview-adjacent-final.log`.
8. Frozen gate retry: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=180s` — FAIL only on pre-existing ledger schema `/tmp/phase36-byteview-frozen-final.log`.

### Prevention and proposed review

No gate previously combined owner-wrapper-to-ByteView and scalar-helper calls; refusal-only coverage encoded the gap. The replacement boundary root plus both repeated-loop consumers now require actual native values and attribution at three optimization lanes, and the argument-order control proves the zero-entry assertion discriminates a semantic defect. Review the mixed-call predicate against unchanged isDynamicScalarCall (U64/Bool only), legacy owner emitters/release pairing, and the exit observer's isolation from execution/capture semantics. Manager owns fresh independent rerun, final archival and exact-file commit; no metadata/roadmap completion change is authorized in this subtask.

### Manager resolution amendment — 2026-10-15

Independent review confirms the 15-line production diff changes only mixed-graph scalar applicability and the existing borrowed ByteView parameter allowance. The scalar call signature predicate remains U64/Bool; ownership-bearing functions retain legacy emitters. The test-only exit observer runs after program termination and reads the actual recorder without changing application exit, streams, event construction, or optional capture behavior. Unmodified session.BuildApplication binaries independently establish process outcomes before the observed binary runs. Success traces pass the independent causal peer. Specialist hint: go; no matching specialist dispatch skill is required by the manager table.

Separate metadata deviation: parent explicitly extended this debug scope to the three new Plan36-06 fixture-ledger rows. Their actual source hashes and reasons were checked; only `sha256` became `source_sha256` and the required `sha256:` prefix was added. Entry order, sources, old corpus rows, comparator, and frozen C goldens remain unchanged. The initial key-only repair still failed the explicit malformed-hash gate; adding the required prefix made both frozen roots pass. This is metadata correction, not a semantic mutation kill.

Manager fresh commands (all exit 0, no skipped roots):

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NativeByteViewCallerArgumentBoundary|TestPhase36NativeByteViewLoopHelpers|TestPhase36NativeFileByteViewLoopHelpers|TestPhase36ByteViewArgumentOrderMutationControls)$' -count=1 -v -timeout=300s` — `/tmp/phase36-byteview-manager-consumer.log`, 19.379s.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36HelperAnswersAndDefects|TestPhase36HelperOptimizationMatrix|TestPhase36HelperUBSan|TestPhase36ByteViewCallerArgumentBoundary|TestPhase36LegacyByteViewCalleeBounds|TestPhase36LegacyFallibleCallRoutes|TestPhase36LegacyFileOwnerPeerRoutes|TestPhase36NativeCallCaptureOutcomeSeparation)$' -count=1 -v -timeout=300s` — `/tmp/phase36-byteview-manager-adjacent.log`, 3.160s.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/executionpeer ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^(TestPhase36NativeDynamicInvocationPaths|TestPhase36CallEventCapacityBoundary|TestPhase36NativeOrdinalAndPathBoundaries|TestPhase36EmitterCallInventoryControls|TestPhase36DynamicCallOrdinalControls|TestPhase36CallCausalAttributionControls|TestPhase36CallCFGApplicability|TestPhase36LegacyFileOwnerCalls|TestPhase36RawCallShapeAndFreshness|TestPhase36RawCallRefusalRemovalControls|TestPhase36OriginCallFreshnessAndTransfer|TestPhase36OriginCallRefusalRemovalControls)$' -count=1 -v -timeout=300s` — `/tmp/phase36-byteview-manager-peers.log`, all four packages pass.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=180s` — `/tmp/phase36-byteview-manager-frozen.log`, 1.989s.

The correction is commit `5d874384`; this dated debug record is archived only after the independent gates pass. Phase36 state, roadmap, plan completion and halted summary remain parent-owned and unchanged. No human UAT occurred. Prevention: why not caught: no mixed owner-wrapper/ByteView/scalar-helper native gate existed; guard: executable boundary and fixture/FileBytes repeated-loop roots plus semantic argument-order controls.
