---
phase: separate-pointer-successors-and-integrated-utility
verified: "2026-10-02T06:31:02Z"
status: gaps_found
score: "5/6 must-have truths verified; 5/5 roadmap criteria verified"
covered_files:
  - .github/workflows/ci.yml
  - .planning/LANGUAGE-MATURITY.md
  - .planning/PRODUCT-ROADMAP.md
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-01-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-01-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-02-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-02-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-03-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-03-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-04-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-04-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-05-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-05-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-SUMMARY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-CONTEXT.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
  - examples/phase23/adapter.c
  - examples/phase23/adapter.h
  - examples/phase23/file_byte.bindings.json
  - examples/phase24/README.md
  - examples/phase24/error.schway
  - examples/phase24/transfer.schway
  - internal/compiler/check/check.go
  - internal/compiler/check/check_pointer_successor_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_pointer_successor_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go
  - internal/compiler/corevalidate/corevalidate_pointer_successor_test.go
  - internal/compiler/diagnostic/diagnostic_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_pointer_successor_test.go
  - internal/compiler/native/phase25_pointer_successor_test.go
  - internal/compiler/native/phase25_utility_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_pointer_successor_test.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session_admission_divergence_test.go
  - internal/compiler/session/session_pointer_successor_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - scripts/verify-phase25.sh
  - testdata/phase25/exclusive_conflict_reject.schway
  - testdata/phase25/exclusive_copy_accept.schway
  - testdata/phase25/exclusive_escape_reject.schway
  - testdata/phase25/shared_copy_accept.schway
covered_digest: "v1:sha256:ba87522692a3e1526beeb5d7c63881699c8e6f03b3c9e1020be5f5f7ed4738a8"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: "4/5 roadmap success criteria"
  gaps_closed:
    - "Foreign, shared, and exclusive native receipts now cover macOS and Linux across baseline, optimized, and sanitizer lanes."
  gaps_remaining:
    - "Phase 25 validation, security, product-roadmap, and maturity artifacts still report hosted EVD-10 receipts as pending."
  regressions: []
gaps:
  - truth: "Phase 25-owned evidence and living documents reflect the current hosted macOS/Linux family-by-lane receipts and EVD-10 status."
    status: partial
    reason: "The source and hosted CI receipts now satisfy the runtime evidence contract, but four phase-owned documents still say that the hosted matrix was not run or remains pending."
    artifacts:
      - path: ".planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md"
        issue: "The opening note, T-25-10 row, security follow-up, wave-0 note, and approval text still say hosted receipts/EVD-10 are pending (including lines 12, 45, 79, 83, and 99)."
      - path: ".planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md"
        issue: "The T-25-14 closeout and sign-off still state that hosted EVD-10 receipts remain pending."
      - path: ".planning/PRODUCT-ROADMAP.md"
        issue: "The Phase 25 status and dated amendments still say the hosted matrix is pending and EVD-10 remains open (for example lines 50-52 and 214-250)."
      - path: ".planning/LANGUAGE-MATURITY.md"
        issue: "The Phase 25 amendments and capability table still say hosted macOS/Linux receipts are required or pending (for example lines 15-74, 123, and 142)."
    missing:
      - "Add a dated amendment naming workflow run 36971855722, branch head 421b5b94eb867e940a207dfd64d7971dc8198172, merge revision a90c27c5b432ef6fc59fbafaa68b50a1374ae138, both native hosts, all family/lane results, and the source/history evidence distinction."
      - "Re-run goal-backward verification after these artifacts are refreshed so their covered-file fingerprint matches the final documents."
advisory: []
---

# Phase 25: Separate Pointer Successors and Integrated Utility — Verification Report

**Phase Goal:** A developer can use separately checked shared and exclusive read-copy pointer helpers in the documented native utility, with reproducible evidence for each admitted family.
**Verified:** 2026-10-02T06:31:02Z
**Status:** gaps_found
**Re-verification:** Yes — the prior hosted-evidence gap was closed; this pass found that the phase-owned evidence documents have not yet been refreshed from the new receipts.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A source consumer invokes the bounded shared read/copy helper through a real pointer-parameter C ABI, while independent checking rejects incompatible access and escape. | ✓ VERIFIED | Source inspection: testdata/phase25/shared_copy_accept.schway, checker recognizer in checkLocalFileByteTransferCaller and family-specific tests. checkedSharedPointerABIFact emits const uint64_t *. C/manifest tests and native TestPhase25UtilityNative/TestPhase25UtilityFamilyControls cover actual behavior and distinct conflict/escape refusals. Named checks passed during this re-verification. |
| 2 | A distinct source consumer invokes the bounded exclusive read/copy helper through a real pointer-parameter C ABI, while independent checking rejects conflicting access and escape. | ✓ VERIFIED | Source inspection: testdata/phase25/exclusive_copy_accept.schway, checkedExclusivePointerABIFact and the integrated caller. The emitter derives uint64_t *; TestPhase25UtilityNative, TestPhase25NativeFamilyWrongResult, TestPhase25FamilyConflict, and TestPhase25FamilyEscape pass. The independent core/origin/path peers have positive, conflict, and escape controls. |
| 3 | Emitted C and manifests agree without unsupported alias/capture/alignment promises; unsupported mutation, forwarding, retention, callbacks, nonlocal exits, and wider shapes fail before serialization. | ✓ VERIFIED | pointerABIFact supplies declaration/body/call/manifest facts; TestPhase25SharedPointerCopyABI, TestPhase25ExclusivePointerABI, TestPhase25ExclusivePointerManifest, and TestPhase25ExclusivePointerRefusal pass. The integrated manifests assert exact C types and no unsupported emitted attributes. Source and serializer guard inspection confirms the bounded subset. |
| 4 | Foreign, shared, and exclusive families each have native receipts on macOS and Linux for every applicable optimizer/sanitizer lane. | ✓ VERIFIED | Hosted run [36971855722](https://github.com/szTheory/schway/actions/runs/36971855722) passed both current evidence aggregates at merge revision a90c27c5b432ef6fc59fbafaa68b50a1374ae138, corresponding to branch head 421b5b94eb867e940a207dfd64d7971dc8198172. Across the two native jobs, all 18 family×host×lane rows passed; details are below. |
| 5 | A clean checkout can follow the public utility commands, observe 65/66 and the typed 0x43 failure, and locate the C inputs and lifetime evidence; unsupported ownership uses have stable source-attributed diagnostics. | ✓ VERIFIED | The previous verifier verified the README smoke and 0x43 error-before-helper flow; its files remain unchanged since that pass and received the required quick regression check. This pass freshly ran native utility, conflict/escape, and emitter named tests. TestPhase25EscapeDiagnosticSourceAttribution, TestPhase25EscapeDiagnosticFunctionIdentity, and diagnostic wire-schema tests pin codes, primary/cause spans, and absence of unsafe repairs. |
| 6 | Phase-owned validation/security/roadmap/maturity documents distinguish the new hosted receipts from source inspection, local checks, and historical receipts while retaining the three ranked next capabilities. | ✗ FAILED | The documents have the required capability candidates and evidence categories, but they still report the new run as pending/unobserved. Exact stale locations are in the YAML gaps list. |

**Score:** 5/6 must-have truths verified. The five roadmap success criteria are all satisfied; the remaining gap is the current evidence record, not implementation or hosted behavior.

### Re-verification of Previous Gap

The prior 4/5 report's only gap was EVD-10's missing hosted family-by-lane matrix. That gap is closed: both native hosts executed the same PR merge revision, and all nine family/lane rows on each host reported expected and actual values. The earlier four passing roadmap truths received a quick regression check; the local source and public utility artifacts remain present and substantive, and the three named re-verification commands below pass.

### Six-Plan Cross-Check

The six plans and summaries were read and their must-haves were checked against source, tests, CI, and current receipts:

| Plan | Verified implementation contract |
|---|---|
| 25-01 | Shared U64 borrow/copy source fact selects an exact const uint64_t * C ABI. Declaration, address-taking at call sites, copied value, and manifest use one checked ABI fact. Wider shapes are refused. |
| 25-02 | corevalidate, originvalidate, and pathoracle derive family, origin, conflict, and lifetime facts independently; valid readers and ended-shared-then-exclusive use pass, while overlap, escape, wrong-family, and forged-source cases fail. |
| 25-03 | A separately named exclusive helper is read/copy-only. The fixed caller admits only owner acquire/transfer, typed use, shared copy, exclusive copy, and return; positive, conflict, and escape cases remain distinct. |
| 25-04 | Exclusive pointer lowering and its manifest are checked, modeled success returns 65/66, typed 0x43 failure precedes either helper, owner discharge remains checked, and structured diagnostic identity/schema/source spans are pinned. |
| 25-05 | The actual-C utility and both independent peers accept the exact result chain while preserving the direct Phase 24 route; wrong-result controls are reached and malformed or reordered shapes refuse. |
| 25-06 | README, script, and existing dual-host CI wiring are substantive and connected. The only unmet plan-level item is synchronizing the four evidence/living documents with the newly completed hosted receipts, recorded as truth 6 above. |

No dependency was added: go.mod and go.sum are unchanged from the Phase 25 starting revision, and the six summaries declare no added dependencies.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| testdata/phase25/shared_copy_accept.schway, testdata/phase25/exclusive_copy_accept.schway | Distinct source witnesses | ✓ VERIFIED | Both are non-stub source helpers consumed by the checker, emitter, and native evidence tests. |
| internal/compiler/check/check.go and family fixtures/tests | Exact caller and helper admission | ✓ VERIFIED | The caller requires the exact ordered flow; family checker tests assert source facts and reject conflicts/escapes. |
| internal/compiler/{corevalidate,originvalidate,pathoracle} | Independent peer checks | ✓ VERIFIED | Each package has its own derivation and mutation tests; session tests call the validators independently over positive and modified cores. |
| internal/compiler/cgen/{cgen.go,cgen_program.go} | Pointer ABI C lowering, manifest, and refusal boundary | ✓ VERIFIED | ABI facts select pointer parameter types and call-site address-taking; tests pin exact fields and refusal controls. |
| internal/compiler/native/phase25_utility_test.go | Executable family, error-order, wrong-result, conflict, and escape witnesses | ✓ VERIFIED | Tests compile and run emitted C with fixed independent expected results. |
| examples/phase24/README.md and explicit Phase 23 binding inputs | Clean-checkout commands and evidence index | ✓ VERIFIED | Commands name both source and binding manifest; test confirms family/host/lane index. Previous verifier recorded the command smoke; unchanged since. |
| scripts/verify-phase25.sh and .github/workflows/ci.yml | Fail-closed receipts owned by the existing aggregate | ✓ VERIFIED | Script verifies each local family/lane and emits other-host rows as incomplete; CI runs it exactly once in each Ubuntu/macOS evidence aggregate. Both passed on the same merge revision. |
| 25-VALIDATION.md, 25-SECURITY.md, PRODUCT-ROADMAP.md, LANGUAGE-MATURITY.md | Current, evidence-calibrated status | ⚠️ PARTIAL | Historical/source/local descriptions are useful, and ranked future capability content remains. Their hosted/EVD-10 statements are stale; see gap 1. 25-SECURITY.md itself remains SECURED, ASVS L1, 15/15 threats closed, zero open. |

### Hosted Native Receipt Matrix

The Phase 25 script covers three families (foreign, shared, exclusive) and three lanes (baseline -O0, optimized -O2, ASan+UBSan). Every listed family/lane completed on each native host:

| Host | Native target and compiler | Passing rows | Cold min/median/max | Warm min/median/max | Receipt identity |
|---|---|---:|---|---|---|
| Linux x86_64 | GO linux/amd64; x86_64-pc-linux-gnu; Ubuntu Clang 18.1.3 | 9/9 | 9.350 / 9.540 / 9.980 s | 0.390 / 0.490 / 0.550 s | a90c27c5b432ef6fc59fbafaa68b50a1374ae138; clean; pass |
| macOS arm64 | GO darwin/arm64; arm64-apple-darwin25.6.0; Apple Clang 21 | 9/9 | 12.470 / 12.670 / 12.930 s | 0.820 / 1.400 / 2.130 s | a90c27c5b432ef6fc59fbafaa68b50a1374ae138; clean; pass |

For all 18 rows the native utility returned 65 for 0x41 and 66 for 0x42. The negative 0x43 control returned UseError.UnsupportedByte / exit 65 before either helper. Both pointer families reached their helper-local wrong-result controls (shared 99; exclusive 98), and each has separate conflict, escape, and manifest controls. The local evidence script also passed at branch head 421b5b94; it correctly marked rows for the other host incomplete because local execution cannot claim a remote-host result.

The same hosted run passed all required CI jobs: Linux and macOS vet, build, full tests, and race tests; and Linux and macOS current evidence aggregates covering Phases 24, 25, 23, 6, and Phase 15 admission/session seams. The optional validation-corpus receipt was skipped by its default condition and is not part of Phase 25 acceptance.

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Shared/exclusive source functions | Checked core.LinearOperation borrow/copy facts | check.go exact source recognizers | WIRED | Fixture tests assert separate OpBorrowShared and OpBorrowExclusive family facts. |
| Checked pointer ABI fact | Generated C signature, body copy, call argument, and foreign manifest | checkedSharedPointerABIFact, checkedExclusivePointerABIFact, emitProgramFunction, manifest builder | WIRED | Exact pointer declarations and manifest fields are asserted together; cgen uses the same checked fact for call-site representation. |
| Admitted source/core | Independent core, origin, and path decisions | Peer-local derivation plus direct mutation tests | WIRED | TestPhase25IndependentPeerMutations invokes validators separately; the path oracle recomputes liveness and keeps the call source as a loan use while deferring call-result ancestry to callee-aware peers. |
| Integrated utility | Actual native outputs and error order | C emitter + native application test | WIRED | Fixed 65/66 outputs and helper-call instrumentation are observed from emitted C. |
| Evidence script | Dual-host evidence-aggregate | One workflow step per matrix host | WIRED | .github/workflows/ci.yml line 98 runs the script once in each aggregate job. GSD's text-only link heuristic could not resolve several module/test links, so those links were traced manually. |
| Living docs | Current hosted receipt | Dated evidence amendment | NOT CURRENT | The evidence run exists, but the dated amendments have not yet been written; this is the sole gap. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| examples/phase24/transfer.schway | file byte → U64 → shared copy → exclusive copy → returned U64 | Explicit file-byte foreign binding and caller-provided files | Yes; native emitted-C utility returns fixed expected values and its error path logs zero helper calls | ✓ FLOWING |

This compiler/tooling phase has no database, API, or rendered UI data path.

### Behavioral Spot-Checks

| Behavior | Command/evidence | Result | Status |
|---|---|---|---|
| Native utility produces the independent success/error behavior | GOCACHE=/private/tmp/schway-gocache go test -count=1 -run '^(TestPhase25UtilityNative|TestPhase25NativeFamilyWrongResult|TestPhase25UtilityFamilyControls)$' ./internal/compiler/native | Exit 0; named emitted-C tests pass | ✓ PASS |
| Checker/session rejects family conflict and escape while peers accept valid sequential use | GOCACHE=/private/tmp/schway-gocache go test -count=1 -run '^(TestPhase25FamilyConflict|TestPhase25FamilyEscape|TestPhase25IndependentPeer.*)$' ./internal/compiler/session | Exit 0; named tests pass | ✓ PASS |
| C ABI and serializer refusal preserve the pointer boundary | GOCACHE=/private/tmp/schway-gocache go test -count=1 -run '^(TestPhase25SharedPointerCopyABI|TestPhase25ExclusivePointerABI|TestPhase25ExclusivePointerRefusal)$' ./internal/compiler/cgen | Exit 0; named tests pass | ✓ PASS |
| Repository regression gate | GOCACHE=/private/tmp/schway-gocache go test -count=1 ./... | Uncached full suite passed in the execute-phase regression gate | ✓ PASS |
| Static/build/race validation | Hosted run 36971855722 on both native hosts | go vet ./..., go build ./..., go test ./..., and go test -race -timeout=20m ./... all passed | ✓ PASS |
| All native family/lane receipts | sh scripts/verify-phase25.sh in hosted aggregate on each native host | 9/9 per host; all 18 rows pass | ✓ PASS |

The three named local re-verification commands completed in under one second each. Existing Phase 25 validation files contain no separately declared probe-* script; verify-phase25.sh is the evidence gate and was executed in the hosted aggregate.

### Decision Coverage

All 8 trackable CONTEXT.md decisions are present in shipped artifacts. The gate reported honored: 8, not_honored: [] (non-blocking decision-coverage check).

### Test Quality Audit

| Test evidence | Linked requirement | Disabled | Circular expected values | Strongest assertion | Verdict |
|---|---|---:|---|---|---|
| internal/compiler/native/phase25_utility_test.go | NAT-11, NAT-12, NAT-13, EVD-10, DX-14 | 0 | No — fixed 65/66 and typed-error answers; os.WriteFile creates temporary source/control inputs, not expected output | Native behavioral/value assertions and reached wrong-result controls | PASS |
| internal/compiler/session/session_pointer_successor_test.go | NAT-11, NAT-12, DX-15 | 0 | No generated expected outputs | Exact semantic diagnostic code, spans, and peer result | PASS |
| internal/compiler/cgen/cgen_pointer_successor_test.go | NAT-11, NAT-12, NAT-13 | 0 | No | Exact generated C, manifest fields, and refusal assertions | PASS |
| internal/compiler/{corevalidate,originvalidate,pathoracle}/*pointer_successor_test.go | NAT-11, NAT-12, NAT-13 | 0 | No | Independent peer outcomes for positive and mutated cores | PASS |

No disabled requirement-linked test or circular oracle was found. The test expectation for each success value is specified directly; the wrong-result controls deliberately mutate generated helper C and assert that the application output changes to 99/98.

### Probe Execution

N/A — no phase plan or summary declares a probe-* path or probe-specific PASS/stage marker contract. The distinct native evidence script is wired to and passed by the existing dual-host CI aggregate.

### Requirements Coverage

| Requirement | Source plan | Description | Status | Evidence |
|---|---|---|---|---|
| NAT-11 | 25-01, 25-02, 25-03, 25-05, 25-06 | Shared read/copy pointer ABI with independent checks and its own positive/conflict/escape evidence | SATISFIED | Shared source witness, pointer C/manifest tests, actual native results, peer mutation controls, hosted rows |
| NAT-12 | 25-02, 25-03, 25-04, 25-05, 25-06 | Separate exclusive read/copy pointer ABI with independent checks and its own positive/conflict/escape evidence | SATISFIED | Exclusive source witness, pointer C/manifest tests, actual native results, peer mutation controls, hosted rows |
| NAT-13 | 25-01, 25-02, 25-04, 25-05, 25-06 | C/manifests claim only checked facts and unsupported shapes fail before serialization | SATISFIED | Cgen tests and emitted manifests; structural refusal tests; CI run |
| EVD-10 | 25-06 | Reproducible per-family macOS/Linux native receipts for applicable lanes | SATISFIED | All 18 hosted family×host×lane receipts in run 36971855722 |
| DX-14 | 25-06 | Clean-checkout public commands for two file inputs, error behavior, bindings, and evidence | SATISFIED | README contract and previous verified clean-checkout smoke, retained without intervening file changes |
| DX-15 | 25-03, 25-04 | Stable source-attributed diagnostics for conflicts, escapes, moved/discarded ownership, and unsupported boundaries | SATISFIED | Named session/diagnostic tests pin code, schema, primary/cause spans, and no unsafe repair |

No additional Phase 25 requirement is mapped outside these six plan declarations. REQUIREMENTS.md continues to show the IDs as pending/gaps until the execute-phase completion transition updates its canonical statuses.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | None in the inspected Phase 25 source and test files | — | Debt-marker scan found no unreferenced TBD, FIXME, or XXX. The XXXXXX match is the standard mktemp template; other scanner-token matches are in the groundedness test itself. |

### Human Verification Required

N/A — this is a compiler/tooling phase with deterministic CLI/native checks and no visual or external-service behavior. No state-transition truth lacks behavioral tests, so no end-of-phase human checkpoint is needed.

### Advisory (New Scope, Unevidenced)

None. The sole issue is not an unevidenced new-scope code observation; it is a directly reproducible mismatch between current CI receipts and named phase-owned documentation statements.

### Gaps Summary

The runtime/source goal is achieved and EVD-10 is now satisfied by hosted native evidence. Security remains SECURED with 15/15 threats closed and no open threats. The phase cannot receive a passing closeout yet because four phase-owned documents still describe the now-passing host matrix as absent or pending. Refresh those dated claims from the run receipt, then re-run verification so the final document digest and status reflect the updated artifacts.

---

_Verified: 2026-10-02T06:31:02Z_
_Verifier: the agent (gsd-verifier)_
