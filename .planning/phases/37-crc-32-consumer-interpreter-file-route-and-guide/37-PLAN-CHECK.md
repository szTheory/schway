# Phase 37 independent plan check

## Round 1 — 2026-10-15

Typed checker `/root/check37` returned ISSUES FOUND: two blockers and two
warnings, no information-only findings. All four requirements and 24 decisions
were covered; eight plan structures, 19 tasks and five-wave dependencies passed.
No application tests ran. Canonical command-path probe was not_applicable for
26 forms; failure-direction probe was ok for all 26.

```yaml
issues:
  - plan: null
    dimension: research_resolution
    severity: blocker
    required_property: "RESEARCH.md carries no unresolved open question before plan acceptance"
    description: "The dated amendment resolves the old D-10 conflict, but Open Questions 2 and 3 remain unmarked: Go nonblocking-open behavior across macOS/Linux and the maximum-case timeout budget."
    fix_hint: "Record the host-test strategy and measure-before-tightening timeout rule as resolved, retaining the stated host and latency limits."
  - plan: "37-01"
    dimension: key_links_planned
    severity: blocker
    required_property: "Every tracked Schway source is either in the frozen corpus or registered in its closed fixture ledger"
    description: "Plan 37-01 adds examples/crc32/lib/crc32.schway, examples/crc32/crc32.schway, and examples/crc32/crc_low.schway, but no plan updates testdata/phase36/new-fixture-ledger.json. TestPhase 36NoCallGeneratedCByteBaseline enumerates all tracked .schway files and rejects any not present in the corpus or ledger, so it will fail once these sources are tracked."
    task: 1
    fix_hint: "Register the new sources in the existing fixture ledger with source hashes and reasons while keeping the frozen C corpus unchanged."
  - plan: "37-01"
    dimension: scope_sanity
    severity: warning
    required_property: "Each plan stays within the per-plan file-count budget"
    description: "Plan 37-01 lists 13 modified files, exceeding the 10-file warning threshold."
    fix_hint: "Consider moving a coherent deliverable to another plan if it reduces execution scope without splitting the runnable tracer."
  - plan: "37-07"
    dimension: scope_sanity
    severity: warning
    required_property: "Each plan stays within the per-plan file-count budget"
    description: "Plan 37-07 lists 10 modified files, meeting the file-count warning threshold."
    fix_hint: "Consider separating surface facts from route and exit documentation if that reduces execution scope."
```

The required properties and evidence bind; hints are examples. Parent inspection
found a further constraint on the ledger hint: the Phase 34 consumer pins the
exact three Phase 36 rows. Appending Phase 37 sources there alone breaks that
consumer. The revision must compose an explicit finite current-source addition
set in both consumers, keep prior corpus bytes and historical fixture rows
unchanged, and arm omission/hash/duplicate/false-exemption controls. It must
also order source creation/registration so intermediate task commits verify.

Revision iteration 1/3: four issues before revision. No checker result has been
overridden and no unexecuted implementation is marked passed.

## Revision 1 response — 2026-10-15

Both blockers have targeted plan changes. RESEARCH marks all three original
questions RESOLVED while preserving discovery history and unexecuted host/latency
limits. Plan37-01 creates all three complete sources in Task01, then registers
actual final source hashes in a separate exact-three-path Phase37 ledger consumed
by both historical owners. The original Phase36 ledger rows/schema/hashes and
frozen C corpora remain unchanged. Both consumers receive reached omission,
tracked-source, hash/reason, duplicate/order, unsafe-path and false-old-exemption
controls; initial C25 and final C26 prove the inventory transition at execution.
Task02 owns only the local binding triple/dual-consumer proof and reruns C26, so
the first task gate references no absent later source or adapter.

Both warnings have explicit bounded scope dispositions in their plans and in
37-SOURCE-AUDIT.md. Plan01's16 files retain the runnable tracer plus its required
inventory safety and early canonical proof; Plan07's10 files remain three small
public usage/exit/surface tasks. No ninth plan, new subsystem or generic ledger
framework was added. The updated Plan01 estimate is52k with the inherited factor1
and low/sample-count0 calibration. These are accepted warning rationales for the
next independent checker to assess, not a claim that file counts dropped below
the threshold.

Fresh artifact checks after the revision: all8 frontmatter schemas valid, all8
plan structures valid with no structure warnings, decision coverage24/24,
command-path probe not_applicable with0 blocker/0 warning across29 command
occurrences, failure-direction probe ok across the same29 occurrences. A local
planning consistency check found8 plans/19 tasks/5 waves, all4 requirement IDs,
26 unique registered commands, no unregistered task command, no same-wave file
overlap and no duplicate non-SC threat ID. New threat T-37-30 belongs only to
Plan01's additive inventory safety obligation. No application/host/ledger
implementation check was run; parent owns the fresh independent checker and
canonical commit/state/roadmap updates.

Changed planning files:37-01-PLAN.md,37-07-PLAN.md,37-RESEARCH.md,
37-VALIDATION.md,37-SOURCE-AUDIT.md and this check record. Edge/prohibition
projections remain unchanged because no categories/prohibitions were added;
their unresolved classifier/descriptorless history remains preserved.

## Parent proof handoff clarification before recheck

Plan01 Task03 now explicitly owns its PLAN file for post-proof descriptor wiring. A successful canonical producer request must be projected into those same two verifier-consumed prohibition items and compared field-for-field; the raw unresolved projection remains historical. This avoids a valid receipt leaving descriptorless must-haves unverified. Plan01 has 17 bounded paths after this metadata addition, with unchanged 19 tasks/eight plans/five waves.

## Round 2 — 2026-10-15

Independent recheck after revision iteration 1: **VERIFICATION PASSED**. Eight
plans were checked: 19 tasks across five waves. All eight `verify.plan-structure`
results were valid with zero structural errors or warnings. All four Phase 37
requirements are claimed and have concrete task coverage. The dependency graph
is acyclic and wave-consistent; same-wave file sets are disjoint, and no
undeclared same-wave mutable-resource dependency or incompatible shared data
transformation was found. Key source-to-module, interpreter-to-reader,
CLI-to-session/native, oracle-to-matrix and guide-to-fixture links are executable
plan actions.

The revised research section marks all three open questions resolved as planning
strategy and retains the old discovery history. It claims no Linux, FIFO,
descriptor or maximum-case timing execution that has not occurred. The current
CONTEXT D-10 empty-path tuple is `(65, "", "")`; plans preserve it. Plan01's
initial/final inventory sequence creates the complete three-source set before
ledger closure and names both historical owners, their exact historical bytes,
and the reached controls. Plan01 Task03 explicitly plans post-proof projection
of the real descriptors into the two verifier-consumed prohibition items, with
field-for-field request comparison and preserved raw history; Plan02 extends
outcome proof to interpreter execution and Plan08 reruns the final bound proof.

The two prior scope findings now have bounded, task-level dispositions and are
accepted rather than left as unresolved findings. Plan01's 17 listed paths are
three tasks with 8/4/6 task file sets, a 52,000-token estimate against a
100,000-token budget, and a named runnable-tracer/inventory/prohibition-proof
cohesion rationale. Plan07's 10 paths are three small public-fact tasks with
4/2/4 task file sets and a rationale tied to one finite discoverability concern.
All eight smart-zone checks are under budget (38–54%); confidence is `low` with
sample_count 0, so those estimates are not calibrated project measurements and
the task/file review carries more weight. These scope observations do not
require another plan or a split.

The canonical verify-command-path probe reports 29/29 `not_applicable` with
severity `none`; this is the supported-form verdict, not proof that Go test
selectors exist. The canonical failing-direction probe reports 29/29 `ok` with
severity `none`. New roots are source-grounded in their owning tasks and
existing package owners. No application tests, suite, native matrix, hosted
command or production code were run during this review.

```yaml
issues: []
```

Plans are ready for execution. This result validates plan coverage only; it does
not claim that the proposed implementations or tests have passed.
