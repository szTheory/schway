---
verification_role: proposal
researched: 2026-09-27
scope: M004 delivery economics and adversarial assurance
---

# M004 Pitfalls and Evidence Cadence

M004's selected product witness is an application compiled once and run once,
using a real Lang-owned, malloc-backed buffer through a C adapter. It includes
per-operation foreign bindings, ownership transfer, normal/error cleanup, and
bounded shared/exclusive read-copy pointer families without `restrict`.
Arithmetic and FizzBuzz follow this milestone. These are planning decisions,
not claims that these capabilities already exist.

This review inspected current implementation and archived research. It ran only
the focused mutation and planning-integrity checks listed below. It did not run
the full suite, a new performance study, or cross-host native verification.

## Current blockers that the product witness must cross

| Boundary | Current evidence | Required M004 result |
|---|---|---|
| Application inputs and execution | Public run selects an engine and file; `interpreterInputs` supplies canonical fixture inputs. | An explicit application input/output contract and a single application execution, with verification retaining its own multi-engine behavior. |
| Foreign and pointer emission | `emitProgram` refuses foreign-call blocks, foreign contracts, and the cut pointer families. | Admit only the selected C-adapter shapes after their independent checks and runtime witnesses pass. |
| Resource proof | Phase 21's JSON contract has `contract_only: true` and `production_admission: false`. | Checked obligations plus observable allocation, transfer, and release behavior on each admitted normal/error path. |
| Host evidence | Phase 21's recorded LTO comparison covers one pure-Lang fixture on Darwin arm64. | Exact-shape macOS and Linux evidence for each pointer family whose reopening condition requires it. |

Sources: [CLI](../../../cmd/lang/main.go),
[session input and run paths](../../../internal/compiler/session/session.go),
[program emitter](../../../internal/compiler/cgen/cgen_program.go),
[Phase 21 contract](../../milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESOURCE-DISCHARGE-CONTRACT.json),
[bounded LTO receipt](../../milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-LTO-EVIDENCE.md).

## Critical pitfalls

### A successful fixture harness is mistaken for an application

`RunInterpreter` and the current native route derive canonical inputs;
the native verification path also compares multiple executions. A command
returning a passing verification document does not demonstrate application
argument handling, one process execution, or useful application output.
Give the M004 witness an explicit input and expected result. Observe the number
of application launches and distinguish application output from evidence
serialization. Keep verifier behavior available under its existing contract.

### A malloc-backed owner is represented only by an instrumentation counter

The product must exercise real storage and a real transfer of responsibility.
An empty resource ledger or a balanced counter cannot establish that native
bytes were correctly allocated, preserved, and released. Check observable
buffer content alongside ownership events and native cleanup; seed wrong-byte,
omitted-release, and duplicate-release faults at the affected seams. Define the
bounded allocation size and failure behavior. Keep C-adapter metadata tied to
each operation so multiple foreign calls cannot accidentally share one
function-level contract or release obligation.

This is a prospective risk for M004. Existing Phase 21 contract validation
explicitly does not prove runtime cleanup. See the
[contract validator](../../../internal/compiler/session/session_phase21_contract_test.go)
and [core validator](../../../internal/compiler/corevalidate/corevalidate.go).

### Structural discharge evidence is promoted to a runtime guarantee

Enumerating exits and checking a JSON schema is useful design evidence. Native
admission requires a traceable acquisition/transfer/discharge rule and a native
witness for every exit the selected family admits. Keep opaque nonlocal exits
and cancellation refused, and preserve the explicitly limited guarantee on
defect/process termination. Do not manufacture unwind semantics to broaden this
milestone. Add a later contract version when admission changes; do not rewrite
Phase 21's design-only contract into a claim it never established.

### Shared and exclusive pointer proofs are conflated

Omitting `restrict` avoids one optimizer promise; it does not prove Lang's
ownership or alias rules. Preserve separate shared/exclusive read-copy
contracts and refuse escape, retention, unsupported writes, and unsupported
transfer shapes. One family's native witness cannot discharge another's
prerequisite. Keep macOS/Linux evidence requirements attached to the exact
family and shape. See the historical
[three-family ownership register](../../milestones/M003-phases/16-branch-match-emitter-port/PHASE-16-DEBT.md)
and [public refusal test](../../../internal/compiler/session/session_phase16_production_paths_test.go).

### Engine equality hides a shared wrong result

Independent validators remain source-blind at documented boundaries. Native
and interpreter agreement is insufficient if both serialize the wrong value.
The application witness needs a source-level expected result as well as engine
comparison; mutation controls must establish that the injected fault actually
occurred. The existing
[Phase 19 shared-wrong-result control](../../../internal/compiler/session/session_phase19_test.go)
and [Phase 18 payload controls](../../../internal/compiler/session/session_phase18_payload_test.go)
provide concrete patterns.

## Process pitfalls that consume delivery time

| Pitfall observed | Consequence | Bounded response |
|---|---|---|
| Maturity census numbers pass while capability prose stays stale. | Counts imply freshness without checking the feature claim; adding a fixture demands prose maintenance. | Refresh capability claims from source witnesses. Preserve the current census guard for this planning change; reconsider census maintenance separately. |
| A historical ownership test reads the mutable live milestone charter. | Formal milestone kickoff breaks an already-settled historical assignment. | Read the archived M003 roadmap, preserve the family/owner/prerequisite laws and four seeded mutations. Current milestones may change status and charter. |
| An old probe name is interpreted as its current assertion. | D-12-43 remained labelled unconstructible after Phase 18 made the wrong-slot fault observable. | Keep the compatibility name, correct the authored disposition with a dated amendment, then derive the view from that row. |
| CI runs full and race suites in two jobs per host. | Repeated work increases feedback cost without a separately identified claim. | Give expensive lanes one recurring owner per host; preserve historical scripts and scope any CI change separately. |
| Precise future milestone numbers substitute for feature triggers. | Closed debt is promised again and dependencies become stale promises. | Maintain next/following/later capability horizons and a concrete source/input/output witness for promoted work. |

Sources: [maturity guard](../../../internal/compiler/session/self_describing_docs_test.go),
[ownership and debt tests](../../../internal/compiler/session/session_test.go),
[claims derivation](../../../internal/compiler/session/witness_registry_test.go),
[CI](../../../.github/workflows/ci.yml),
[historical Phase 6 gate](../../../scripts/verify-phase6.sh).

The original estimates are historical forecasts: M003's adversarial synthesis
suggested 48–57 plans and an 8–10-plan emitter phase; closeout records 85 and 26.
Those differences justify scope review, not a claim that any particular
assurance mechanism caused the expansion. Remove uncalibrated language-maturity
percentages and redesign probabilities from current planning. Sources:
[M003 synthesis](../M003/ADVERSARIAL-SYNTHESIS.md),
[M003 audit](../../milestones/M003-MILESTONE-AUDIT.md),
[M003 roadmap](../../milestones/M003-ROADMAP.md).

## Evidence cadence and stopping rules

- At milestone kickoff, phase planning, and phase completion, the agent reviews
  the living product roadmap and maturity snapshot against actual witnesses.
  It proposes the top three useful capabilities, each with a blocker, checker
  change trigger, and evidence needed. This is agent-executed workflow behavior,
  not an unattended background service.
- Each feature phase names one observable application result and the failure
  controls needed for its changed semantics. A second consecutive enabling
  phase without a newly runnable capability triggers scope reassessment.
- Use focused checks while editing and required integration lanes at the
  boundary. Repeat expensive measurements for changed inputs or unresolved
  concerns, not merely a stale report fingerprint.
- Preserve negative controls, independent validation, and explicit refusal
  boundaries. Avoid a new general assurance framework when the selected
  witness needs only a bounded extension to existing mechanisms.
- Report cold and warm samples with their scope. Phase 20 recorded three
  full-suite pairs spanning 129–466 seconds cold and 120–382 seconds warm;
  these do not establish a dependable p95. See the
  [dated timing receipt](../../milestones/M003-phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-CLOSURE-TIMING.md).

## Focused evidence executed during this review

On 2026-09-27, `TestD1243ControlIsUnconstructible` and
`TestPhase18WrongSlotMutation` passed against the current implementation. Both
require positive wrong-slot injection and exact terminal-outcome divergence.
The canonical D-12-43 row now records `CLOSED(d5ddb04)` and `MUTATION-KILLED`;
its original Phase 12 measurement and ratification text remain preserved.

After the planning repairs, the following checks passed: debt-register shape,
global debt ID uniqueness, current Phase 20 unowned population and cap rule,
Phase 16 archived owner/family/prerequisite validation including its four
mutations, the canonical unreachable-claims view, and that view's hand-edit
and stale-entry controls. The current qualified unowned population is three;
the ten-ID historical audit cohort and 13-ID starting population are unchanged.

The claims view was refreshed from the exact output of
`TestUnreachableClaimsViewIsCurrent`, then checked green. No validation command
cells or execution corpora were changed. Existing Phase 16 and Phase 21
verification reports cover changed files, so their fingerprints need separate
current-report reconciliation; these focused checks do not claim to refresh
their complete historical verification or UAT.
