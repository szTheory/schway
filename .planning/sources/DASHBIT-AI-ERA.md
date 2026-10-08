# José Valim, “Evolving programming languages in the AI era”

**Source:** [Dashbit Blog, September 24, 2026](https://dashbit.co/blog/evolving-ai-era)  
**Author:** José Valim  
**Accessed:** 2026-10-08  
**Inspection baseline:** `c9656f0d74e4222b58e9b1041eb28195c84a0aca`  
**Evidence labels:** “Source inspection” means code or planning text was read at the named path; it does not establish runtime behavior. “Historical receipt” means an existing recorded execution result at its stated revision and hosts. A named test found in source is an evidence anchor, not a receipt or a newly passing check. “Newly executed” below refers only to documentation/link/frontmatter checks described in the quick summary; no Go/native behavior checks were run for this note.

## Brief précis

Valim asks how languages and ecosystems should change if agents write more code. He argues that explicit information can improve guarantees, that agent tools should query program facts rather than depend on precise editor positions, and that runtime state should be easier to inspect. He also questions the value of syntax optimized for human typing and considers how cheap generated libraries could weaken collaboration. These are proposals and forecasts, not comparative studies or settled outcomes; the essay itself frames several as opinions that may change.

## Assessment

### Contracts, inference, and human review

In “Stronger guarantees over user constraints,” Valim argues that agents tolerate explicit signatures and that inference can constrain expressiveness. This supports making checked obligations visible, but it does not show that duplicating every derivable fact improves Schway. Schway already pairs explicit boundary contracts with local scalar inference: declarations carry authority across boundaries, while `internal/compiler/ability/ability.go` centralizes structural derivation (`Derive`, `DeriveSealed`, `combineStructural`). The six semantic consumers—`check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, and `cgen`—also mean any new derived guarantee must be checked independently where it is consumed. An SMT or proof backend would enlarge the trusted checker surface without a named consumer or an evidence plan.

“On ergonomics” understates human review when it treats syntax mainly as typing effort. Schway’s canonical source remains an audit and repair surface even if an agent generated it. A future comparison should use the same program and semantic obligations in competing explicit-contract/canonical forms, seed defects and malformed edits, and measure reviewer defect detection and time with actual reviewers. If humans are not involved, call the result unmeasured rather than using AI or formatter proxies. Count tokens as one component of the whole generate → verify → run → observe → repair loop; report cold and warm p50/p95, review effort, and maintenance cost. Formalism or syntax earns its cost only when independent checks show a meaningful guarantee or review gain.

One relevant, already bounded question is widening scalar loops. `internal/compiler/check/scalar.go` has a deterministic work bound in `checkScalarFixedPointDetailed` and preserves refusal through `scalarBackEdgeDiagnostic`. `internal/compiler/session/session_phase27_remainder_test.go` names static-zero, forged-remainder, and dynamic-zero cases (`TestPhase27StaticZero`, `TestPhase27ForgedRemainder`, `TestPhase27DynamicZero`). If a selected resource/loan consumer later requires loop-carried state, bounded state models and independently generated traces fit the existing guarantee map; they are not a separate general-formal-methods initiative.

### Program queries, source locality, and cost

“Program databases over LSPs” describes a plausible agent interface, but its LSP-death forecast and preference for a general database are opinions. Official LSP 3.17 already specifies workspace symbol search and call hierarchy ([workspace/symbol](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.17/workspace/symbol.md), [call hierarchy](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.17/language/callHierarchy.md)); this counters a blanket claim that existing interfaces are inherently position-only. Schway has no claim of implementing LSP here.

Schway’s `cmd/schway/main.go` and `internal/compiler/session/session_phase6_query.go` already provide a stable-ID query seam across five vocabularies, with 64-fact pages, honest `not_captured` answers, and cold source rechecking. `QueryOptions.Depth` is reserved and inert. The gap is narrower: no named symbol or direct-caller vocabulary. A useful first consumer would ask for a helper’s declaration/definition and every direct callsite in one source unit. `internal/compiler/callgraph/callgraph.go` represents deduplicated caller/callee adjacency with a representative operation, so it cannot alone substantiate an “all call occurrences” answer.

Before composing multiple queries, bind cursors to the source revision/query as well as the current vocabulary/span/ID/checksum key in `queryCursorKey`; changed input should fail stale rather than silently continue. Before enabling concurrent repair, inspect the source digest precondition: `cmd/schway-repair/repair.go` rereads and bounds-checks the span in `applyRepair`, but the inspected path does not bind the write to a digest. These are conditional risks from source inspection, not reproduced races or exploit reports. Existing complete-by-abstention caches also matter: `internal/compiler/cache/probe_test.go` names `TestNoClosureDigestInCache`, and the cache uses whole-source keys. Do not add a signature-only `ClosureDigest` that weakens invalidation.

No SQLite, Datalog, persistent session, or broad relations store is justified without a measured consumer. Compare the narrow query with current `rg`/JSON workflows on exact answers, bytes, tool calls, success rate, and cold/warm p50/p95 as input size grows. Negative controls should include repeated call occurrences, name collisions, unknown IDs, malformed/stale cursors, changed source, and truncated pages that never claim completeness. Stop if ordinary search matches it at lower end-to-end cost or if facts cannot be independently checked.

### Runtime observation, security, and authority

“Runtime observability over debuggers” has real overlap with existing Schway evidence. `internal/compiler/native/native_app.go` produces `ApplicationEvidenceReport`, keeps `Verified:false` separate from process outcome, and reports disabled, incomplete, or capacity-exhausted capture. Source-inspected test anchors include `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, `TestPhase22EvidenceMissingPartialAndCapacityControls`, and `TestPhase22EvidenceReportWriteFailurePreservesAppOutcome`. `internal/compiler/session/session_app_verify.go` distinguishes `source_replay_conformance` from `verifier_model_only`, and reports `ActualHostIO` and `PhysicalCleanup` explicitly. Source-inspected observer controls (`TestPhase23ObserverMutationControlsPreservePlausibleEvents`, `TestPhase24ObserverReachedPhysicalDestructorControls`) guard against plausible-looking traces being mistaken for proof. `wiki/debug-evidence-symbolication-and-proof.md` already sketches broader symbolication; `internal/compiler/debugmap/debugmap.go` implements semantic lineage, not native crash capture.

The remaining seed is one named checksum/file utility failure that current capture, observers, and replay cannot explain. A candidate experiment is one optional, local, read-only, build-bound diagnostic that maps to semantic IDs, with explicit missing/truncated/unavailable states and unchanged process outcome. It must not imply a mandatory VM, tracing heap, resident service, automatic upload, or live mutation authority. Source inspection found `os.ReadFile(capturePath)` before the capacity check in `native_app.go`; this means the classification ceiling does not itself bound allocation. That is a hardening concern to resolve before expanding this trust boundary, not an observed exploit. Any added observation can expose secrets: field minimization, retention, consent, and path/identifier sensitivity matter; hashes are not anonymization. Query/runtime strings remain data, not agent instructions. Test wrong-build attribution, missing and oversized capture, stale evidence, redaction, and enabled/disabled overhead; stop if current evidence suffices or the bound/privacy contract cannot be made explicit.

### Portability, supply chain, reuse, and community

The essay’s ecosystem tension is credible in both directions: generated libraries may reduce implementation cost while also reducing shared maintenance. It does not make compatibility, licensing, provenance, security review, and conformance cheap. Prefer a second real consumer sharing a bounded local helper, with independent conformance cases, before broad ecosystem work. The existing roadmap already owns a byte utility plus local module boundary; do not duplicate that seed. A module is not separate-compilation developer experience (`DX-06`/`B1`); `D-13-02b` remains its boundary. `internal/compiler/native/bindings.go` already confines paths and bounds snapshots, while `native_app.go` still reports dependency closure incomplete and `Cacheable:false`. C/Clang/SDK/runtime remain trusted components; input hashes alone do not make builds hermetic.

Portability and security obligations do not shrink with generated source. Candidate evidence must keep target facts and wire contracts independent of the current Apple arm64 host, and keep unsafe operations, FFI, secrets, and build authority explicit. Collaboration can instead center on auditable contracts, examples, bug reports, and conformance fixtures. Reject a community strategy that consists only of forecasting contributor behavior or accelerates packages without a concrete consumer and maintenance owner.

## Disposition and lessons for Schway

These are judgments from this review, not additional claims from the essay.

| Topic | Disposition | Benefit, cost, and reconsideration condition |
|---|---|---|
| Explicit obligations and layered assurance | Reinforce existing guarantee activation map | For each new property, identify construction rules, static derivation, runtime enforcement, or empirical evidence and their limits. Independent peers can share a specification mistake; retain independently authored answers and reached mutations. Revisit when a consumer changes an obligation. |
| Programmatic compiler queries | Dormant experiment | Could reduce navigation and repair work; costs include schema compatibility, freshness, bounded traversal, and another projection to validate. Start with one source unit and exact call occurrences. |
| Optional failure diagnosis | Dormant experiment | Could explain a real failure with less investigation; costs include collection overhead, attribution errors, privacy, and capture ingestion. Debugger integration remains a possible complementary interface. |
| Annotation and source design | Dormant comparison | Preserve human review and local reasoning while measuring whole-loop costs. Prefer declared boundary contracts without banning useful local inference. |
| Shared libraries and community | Existing roadmap home | Second-consumer reuse can concentrate fixes and conformance work; generated bespoke copies distribute maintenance and security responsibility. Reconsider with the byte/module consumer, not a package-count target. |
| New database, proof backend, runtime, or editor replacement | Defer; no current consumer justification | The essay supplies motivation to investigate, not evidence overturning standing technology choices. Reopen a choice only with a measured limitation and a smaller alternative considered. |

Runtime inspection must also retain the existing authority boundary: native app execution is process execution, not a sandbox. Source inspection of `internal/compiler/native/native_app.go` shows environment inheritance; observe permission must never silently imply arbitrary evaluation, mutation, or production control.

## Three dormant experiments

These are conditional seeds, not defects, requirements, or active M005 scope:

1. [SEED-261008-z1g: bounded symbol and caller query](../seeds/SEED-261008-z1g-bounded-symbol-and-caller-query.md) — consider an experiment when current query/search leaves a measured accuracy or cost gap in a named reviewer/agent task; future editor adapters may share the same facts.
2. [SEED-261008-trz: consumer-triggered failure observability](../seeds/SEED-261008-trz-consumer-triggered-failure-observability.md) — activate only when one named failure exceeds current capture, observer, and replay evidence.
3. [SEED-261008-z1s: compare contract and source forms](../seeds/SEED-261008-z1s-compare-contract-and-source-forms.md) — activate only when a concrete consumer contests a language design choice.

Library reuse, guarantee taxonomy, and locality stay with the existing roadmap mechanisms. The current top three remain, in order: Phase 27 EVD-12/DX-16; a bounded byte utility with reusable local module; named bounded JSON. Revisit relevant seeds only at existing milestone kickoff, pre-discussion/planning, and post-phase transitions; record a disposition in the active planning artifact. No background scheduler or repeated full-essay study is proposed.
