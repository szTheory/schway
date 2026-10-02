---
id: 261001-wmx
phase: quick
plan: 261001-wmx
type: execute
mode: quick-full
status: planned
wave: 1
depends_on: []
files_modified:
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
  - internal/compiler/session/session_pointer_successor_test.go
  - internal/compiler/session/session.go
  - internal/compiler/originvalidate/originvalidate.go
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VERIFICATION.md
  - .planning/PRODUCT-ROADMAP.md
  - .planning/LANGUAGE-MATURITY.md
autonomous: true
must_haves:
  truths:
    - "T-25-05/T-25-08: separately named path-oracle tests reject a real live shared/exclusive overlap and a real borrowed-result escape in the straight-line Phase 25 helper shape, even with empty or forged declared endpoints; the command gate and each of its three peers reject both mutations independently while accepting ended shared then exclusive access and compatible shared readers."
    - "The actual `schway check` command for exclusive and shared escape sources reports `core.origin_omitted` with a nonempty primary span on the escaping result, a source-located borrow-origin cause, stable diagnostic schema, deterministic multi-function attribution, and no safety-weakening repair."
    - "T-25-09/T-25-12: structured diagnostic code, schema, stable identity rules, and refusal precedence remain valid, with no unsafe repair suggestion."
    - "Phase 25 evidence claims describe only controls and host receipts supported by current code and executed checks; hosted EVD-10 remains open until its own receipts exist."
  artifacts:
    - path: internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
      provides: "Separate oracle-level tests for live overlap and a genuine borrowed-result escape, each failing if the independent refusal is removed"
    - path: internal/compiler/session/session_pointer_successor_test.go
      provides: "Command-gate and separately asserted three-peer mutations for overlap and escape; source-byte, cause, schema, identity, and repair assertions"
    - path: internal/compiler/session/session.go
      provides: "Source-attributed peer refusal projection on the command path"
  key_links:
    - from: internal/compiler/pathoracle/pathoracle.go
      to: internal/compiler/session/session.go
      via: "Phase 25 application admission invokes independent helper-path validation before native emission"
    - from: internal/compiler/originvalidate/originvalidate.go
      to: internal/compiler/session/session.go
      via: "Structured offending-function identity lets the command project the actual escape to checked source without parsing problem prose"
    - from: internal/compiler/session/session_pointer_successor_test.go
      to: testdata/phase25/exclusive_escape_reject.schway
      via: "CheckCommandFile asserts refusal code, exact source bytes at primary/cause spans, and absent unsafe repairs"
---

<objective>
Close Phase 25 security audit findings T-25-05, T-25-08, T-25-09, and T-25-12 with independent path-oracle overlap/escape controls and source-attributed actual escape diagnostics.

Purpose: Make the published borrow-safety and diagnostic evidence match what the independent validator and user-facing command demonstrably enforce.
Output: Focused code, adversarial controls, and evidence-accurate Phase 25 records; no new dependency.
</objective>

<execution_context>
@/Users/jon/.codex/gsd-core/workflows/execute-plan.md
@/Users/jon/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/STATE.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-CONTEXT.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-02-PLAN.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-03-PLAN.md
@internal/compiler/pathoracle/pathoracle.go
@internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
@internal/compiler/session/session.go
@internal/compiler/session/session_pointer_successor_test.go
@internal/compiler/originvalidate/originvalidate.go
@testdata/phase3/sequential_shared_then_exclusive_accept.schway
@testdata/phase3/shared_shared_accept.schway
@testdata/phase25/exclusive_escape_reject.schway
</context>

<tasks>

<task type="tracer" tdd="true">
  <name>Task 1: Refuse live pointer overlap and escape in the independent path oracle</name>
  <files>internal/compiler/pathoracle/pathoracle.go, internal/compiler/pathoracle/pathoracle_pointer_successor_test.go, internal/compiler/session/session.go, internal/compiler/session/session_pointer_successor_test.go</files>
  <behavior>
    - A copied U64 use ends a shared loan before a following exclusive borrow; simultaneous shared readers remain valid.
    - `TestPhase25PointerPathLiveOverlap` moves the first loan's actual last use after the incompatible second borrow; pathoracle refuses the resulting real overlap with both empty and forged LoanEndpoints.
    - `TestPhase25PointerPathBorrowedResultEscape` changes the terminal result to an existing, valid borrowed place created by the helper's borrow operation; pathoracle refuses this real escape with both empty and forged LoanEndpoints, while the copied U64 result remains valid.
    - `TestPhase25IndependentPeerMutations` applies both mutations to checked core and asserts distinct corevalidate, originvalidate, and pathoracle refusals by direct calls; its two positive programs pass each peer separately.
    - `TestPhase25CheckCommandPeerGate` proves the actual command gate invokes all three independent checks on accepted positive sources and rejects the live-overlap and borrowed-result-escape sources without reaching a pass or native emission.
  </behavior>
  <action>Start with failing `TestPhase25PointerPathLiveOverlap` and `TestPhase25PointerPathBorrowedResultEscape` in pathoracle_pointer_successor_test.go. Build them from accepted Phase 3/25 checked core: for overlap move a real first-loan use after the incompatible second borrow; for escape return an already-created borrowed place with a valid loan origin, never an unknown place. For each, challenge pathoracle directly with empty and forged LoanEndpoints, and retain positive compatible shared readers, ended shared then exclusive, and exact U64-copy helpers. `RecomputeEndpoints` currently returns immediately for straight-line functions because its documented job is CFG endpoint synthesis; preserve that endpoint contract. Add a bounded independent straight-line pointer-loan path in pathoracle.go that derives loan ancestry, last uses, access families, and terminal escape from operations and place facts without reading checker-provided LoanEndpoints or importing another validator. Integrate it into `ValidateLocalOwnerPaths` for application admission before C emission. In session_pointer_successor_test.go add `TestPhase25IndependentPeerMutations`: separately call corevalidate.Validate, originvalidate.ValidatePublished, and pathoracle.ValidateLocalOwnerPaths for each positive program and each overlap/escape core mutation; assert each peer's own expected verdict and code, not mere agreement. Add `TestPhase25CheckCommandPeerGate` using accepted and rejected source cases to prove `CheckCommandFile` calls corevalidate, originvalidate, and pathoracle as separate command gates; wire pathoracle into that command after the existing core and origin decisions, retaining refusal precedence, then assert neither mutation returns pass and neither can reach native emission. Keep branch endpoint synthesis, Phase 24 owner-transfer behavior, and path/composition caps intact. This closes T-25-05/T-25-08 and implements D-25-04/D-25-07 within the standard-library boundary D-25-08.</action>
  <verify><automated>go test -count=1 -run '^(TestPhase25PointerPathLiveOverlap|TestPhase25PointerPathBorrowedResultEscape|TestPhase25PointerPath|TestPhase25UtilityOwnerTransfer|TestPhase25IndependentPeerMutations|TestPhase25CheckCommandPeerGate|TestPhase25IndependentPeer|TestPhase25FamilyConflict)$' ./internal/compiler/pathoracle ./internal/compiler/session</automated></verify>
  <done>Two named oracle-level mutation tests independently reject a genuine live overlap and genuine borrowed-result escape; three separately called peers and the actual command gate reject each mutation; both positive cases pass, and application admission refuses the mutations before emission.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Preserve source attribution on the actual escape refusal</name>
  <files>internal/compiler/originvalidate/originvalidate.go, internal/compiler/session/session.go, internal/compiler/session/session_pointer_successor_test.go</files>
  <behavior>
    - `TestPhase25EscapeDiagnosticSourceAttribution` checks the real `CheckCommandFile` result for exclusive and shared escape sources: invalid status, `core.origin_omitted`, schema, exact source bytes at a nonempty escaping-result primary span and borrow-origin cause span, and no unsafe repair.
    - `TestPhase25EscapeDiagnosticFunctionIdentity` checks deterministic offending-function selection in a multi-function source even when peer problem prose changes.
    - `TestPhase25FamilyConflict` retains `ownership.borrow_conflict` at the offending borrow, distinct from unsupported-shape and escape refusals.
  </behavior>
  <action>Add `TestPhase25EscapeDiagnosticSourceAttribution` for actual `CheckCommandFile` exclusive and inline shared escape sources, asserting exact source substrings for the nonempty primary escaping-result span and borrow-origin cause span, invalid status, `core.origin_omitted`, the current wire schema, and an empty repair list. Keep `TestPhase25FamilyEscape` as the existing refusal control. Add `TestPhase25EscapeDiagnosticFunctionIdentity` with a multi-function source whose non-offending function has a similar return, asserting the structured function identity chooses the correct result and that changing peer problem prose cannot redirect source selection. Preserve `TestPhase25FamilyConflict` and assert its distinct conflict code and offending-borrow span. The current originvalidate `Problem` carries only Code/Detail, and `CheckCommandFile` renders its refusal through `diagnostic.Error(..., Span{}, ...)`; give origin problems a structured offending function/return identity needed for deterministic source projection, while keeping the independent predicate and existing problem codes. At the session boundary, project that identity through the already checked source/token and core function facts to the terminal escaping use and originating borrow, then emit `diagnostic.Error` with causal spans. Never parse `Problem.Detail` to locate source, invent a position for unavailable facts, alter the diagnostic wire schema, or change refusal precedence. Keep `peerRefusalDiagnostic`'s existing fallback for unrelated core refusals. This closes T-25-09/T-25-12 and enforces D-25-05.</action>
  <verify><automated>go test -count=1 -run '^(TestPhase25EscapeDiagnosticSourceAttribution|TestPhase25EscapeDiagnosticFunctionIdentity|TestPhase25FamilyEscape|TestPhase25FamilyConflict|TestPhase25UnsupportedPointerShape|TestPhase25StructuredDiagnosticWireSchema|TestPhase25OwnershipDiagnosticBoundary)$' ./internal/compiler/session ./internal/compiler/diagnostic</automated></verify>
  <done>Both escape families have command diagnostics whose exact primary and cause bytes, code, schema, function identity, and empty repair list are asserted by named tests; conflicts retain their distinct refusal and source span.</done>
</task>

<task type="auto">
  <name>Task 3: Reconcile Phase 25 security and maturity claims with executed evidence</name>
  <files>.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md, .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VERIFICATION.md, .planning/PRODUCT-ROADMAP.md, .planning/LANGUAGE-MATURITY.md</files>
  <action>After Tasks 1–2 pass, reconcile four tightly coupled evidence surfaces: Phase 25 VALIDATION names the decisive controls and commands, Phase 25 VERIFICATION states their observed verdicts and remaining hosted gap, PRODUCT-ROADMAP records the next-capability and assurance boundary, and LANGUAGE-MATURITY records the source witness and receipt provenance required by AGENTS.md. Edit only claims about independent overlap/escape coverage, source-attributed diagnostics, and their evidence; name the newly executed oracle, three-peer, command-gate, and attribution tests by exact identifier. Distinguish source inspection from fresh checks and historical receipts, and retain EVD-10 as open absent hosted macOS/Linux family-by-lane evidence. If any mitigation control still fails, record the corresponding threat as open and narrow the claim instead of declaring it mitigated. Preserve archived Phase 25 decisions and dated amendments, and keep the three existing ranked next capabilities with their witness, blocker, slice, checker change, evidence/debt, owner/action, and reprioritization observation. No dependency, external API, schema migration, or unrelated language capability is added.</action>
  <verify><automated>go test -count=1 -run '^(TestPhase25PointerPathLiveOverlap|TestPhase25PointerPathBorrowedResultEscape|TestPhase25IndependentPeerMutations|TestPhase25CheckCommandPeerGate|TestPhase25EscapeDiagnosticSourceAttribution|TestPhase25EscapeDiagnosticFunctionIdentity|TestPhase25FamilyConflict|TestPhase25StructuredDiagnosticWireSchema|TestLanguageMaturityCountsAreCurrent)$' ./internal/compiler/pathoracle ./internal/compiler/originvalidate ./internal/compiler/session ./internal/compiler/diagnostic</automated></verify>
  <done>All four current documents state the supported mitigation and receipt boundary accurately, with newly executed controls named and hosted EVD-10 still explicitly incomplete.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Checked core to independent path oracle | A forged or misclassified straight-line loan can hide live incompatible access or escape. |
| Independent origin refusal to command diagnostic | A spanless peer refusal conceals the offending source use and its borrow origin from the reviewer. |
| Test receipts to Phase 25 claims | Passing adjacent controls can be mistaken for proof of the missing mitigation or hosted native evidence. |

## STRIDE Threat Register

ASVS level 1; high and critical findings block closure. These four audit findings remain open until the named controls pass on the repaired code.

| Threat ID | Category | Component | Severity | Disposition | Mitigation and decisive evidence |
|-----------|----------|-----------|----------|-------------|---------------------------------|
| T-25-05 | Tampering | Path-sensitive pointer loan validation | high | mitigate | Task 1 adds separately named direct oracle tests for a real live overlap and a real borrowed-result escape, each run with empty and forged endpoint claims. |
| T-25-08 | Tampering | Borrow conflict and escape gate | high | mitigate | Task 1 separately challenges corevalidate, originvalidate, and pathoracle on both mutations and both positive cases, then asserts the command gate refuses both mutations and native application admission cannot emit them. |
| T-25-09 | Repudiation / Information Disclosure | Escape source attribution | medium | mitigate | Task 2 asserts actual command refusal primary/cause spans against source bytes for both families. |
| T-25-12 | Repudiation / Information Disclosure | Structured diagnostic contract | medium | mitigate | Task 2 retains code/schema/refusal precedence and rejects unsafe repair suggestions; Task 3 limits claims to evidence. |
</threat_model>

<verification>Run each task's focused automated command; then run the existing Phase 25 focused package matrix and document gate appropriate to the changed files. A passing local test does not substitute for EVD-10 hosted native receipts.</verification>

<success_criteria>All four audit findings have passing targeted controls, actual escape diagnostics identify their source and cause, and Phase 25 records make no unsupported security or hosted-evidence claim.</success_criteria>

<output>Create `.planning/quick/261001-wmx-close-phase-25-security-audit-findings-a/261001-wmx-SUMMARY.md` after execution.</output>
