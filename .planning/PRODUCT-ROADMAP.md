# Schway — Living Product Roadmap

Updated 2026-10-01. This document carries current direction; ROADMAP.md owns
committed milestone phases, REQUIREMENTS.md owns acceptance, and
LANGUAGE-MATURITY.md records demonstrated capability. Historical milestone
forecasts and evidence remain in their archives.

## Intent and present position

Build a usable general-purpose language for AI-authored, human-audited software:
fast structured feedback, readable canonical source, explicit costs, low-level
ownership, deterministic cleanup, and reproducible evidence. A native compiler
already exists (Go 1.24 → readable C17 → Clang), alongside an interpreter. A new
VM or LLVM backend is not a prerequisite for useful programs.

M001 established source-to-native behavior; M002 added executable Schway calls;
M003 added independent returns, computed matches/payloads, and U64 constants.
Phase 21 completed contract/retirement prework. Phase 22 delivered its
application/build/evidence implementation and passed objective README contract
UAT; its refreshed 2026-09-30 verifier report passes 5/5 truths. No subjective
readability claim or human UAT remains. The separate `schway app build` /
`schway app run` route retains a native app, accepts a bounded U64 input, starts
it once, and preserves ordinary streams and process outcomes. Evidence capture
and explicit differential replay remain separate. The older `schway run` path
is still the synthetic-input O0/O3 conformance harness.

Phase 23 admits a bounded file-byte allocation that remains live through
Schway-directed use and generated local cleanup. Its independent physical
observer and close-failure controls have historical focused receipts on macOS
and a local Linux ARM64 container. Hosted run
[36707529870](https://github.com/szTheory/schway/actions/runs/36707529870)
passed the Ubuntu/macOS check suites and both evidence aggregates, including
`scripts/verify-phase23.sh`, at source SHA
`f991298b29779838a2b1a5c3cd5ac90aafcb84fc`. The refreshed 2026-09-30 verifier
passes 5/5 truths and Phase 23 is complete. No local project suites or UAT were
rerun for that refresh; the hosted receipt supplies current behavioral
evidence, and the phase context requires no human UAT. Phase 24 now admits the
bounded file-byte owner transfer through Schway calls and returns, repeated
helper activations, and exactly-once cleanup on normal and typed-error paths.
Its independent physical observer proves use and destruction, while model
replay remains explicit about not proving host IO or physical cleanup. The
2026-10-01 verifier passes 5/5 truths; hosted run
[36856048690](https://github.com/szTheory/schway/actions/runs/36856048690)
passed the focused aggregate and full/race suites on Ubuntu Linux/x86_64 and
macOS Darwin/arm64 at source SHA
`ed94ef7972b25deb85ab90fafbf3403dc31629f5`. No local project tests or scripts
were run for closeout. Remaining M004 work is the bounded shared/exclusive
read-copy pointer families and the integrated utility in Phase 25. Phase 25's
exact shared/exclusive U64 read-copy helpers and integrated utility are now
present in source; focused local native receipts are recorded below. Hosted
macOS/Linux Phase 25 family-by-lane receipts remain pending, so EVD-10 and
phase acceptance stay open until both native aggregates report them. App build
receipts remain incomplete/non-cacheable when runtime closure is unknown.

The user explicitly requested another deep fan-out and automatic adoption of
recommendations. The 2026-09-27 decision integrates product/DX, compiler and
resource architecture, FFI/security, verification, portability, and delivery
reviews. Research and dissent: [M004 research](research/M004/SUMMARY.md).

## Horizons

| Horizon | User-visible result | Dependencies and scope | Exit observation |
|---|---|---|---|
| **Now: M004** | Complete bounded shared/exclusive read-copy pointer families and a reproducible utility | Phases 22–24 are complete. Phase 25 source admits exact U64 read-copy helpers and the utility; focused local native witnesses exist. Hosted Phase 25 receipts are still required | Existing dual-host CI records each family on macOS/Linux at baseline, optimized, and sanitizer lanes; the documented utility returns 65/66 and preserves typed 0x43 failure ordering |
| **Next milestone: practical computation** | `sum_to_n`, then FizzBuzz from ordinary source | Defined U64 arithmetic/remainder and overflow, comparison/Bool, continuation control flow, scalar loops, fixed text/byte literals, bounded writes and decimal formatting | Public command produces exact expected FizzBuzz output; boundary/error cases and changed-assumption checker controls pass |
| **Mid term: reusable libraries** | Small byte/file utilities and checksum; a reusable bounded JSON parser/serializer | Arity-N and small aggregates as consumers require, explicit byte views/lengths/indexing, fallible APIs, local modules, explicit resource transfer | Second consumer imports a library without copying it; malformed/truncated/oversized input has specified behavior |
| **Mid term: network branch** | A bounded HTTP client or server for a selected use case | Explicit sockets/timeouts/body/framing/error ownership; audited C/OS adapter or dependency; cleanup on every admitted outcome | One documented real endpoint flow plus adversarial protocol cases; TLS policy specified when needed |
| **Long term: production ecosystem** | Larger independently built libraries and real services/tools | Separate compilation/contracts, package/version reproducibility, richer types/generics when justified, effects/concurrency only with resource semantics | Independent consumers and measured limitations justify each expansion |

These are dependency horizons, not dated delivery promises. JSON and HTTP are
independent branches: an HTTP client can fetch bytes before JSON, and a JSON
tool can consume stdin before networking. Prefer a bounded explicit parser
stack over making language recursion a prerequisite. A local module system
need not introduce separately verified binary contracts.

The former M003-era forecast (M004 loops, M005 aggregates, M006 modules) is
superseded prospectively. The next milestone now targets FizzBuzz; further
milestone numbers are assigned at kickoff. `examples/checksum.schway` remains
a provisional integration target; no unsupported syntax is treated as settled.

## Decisions and alternatives

| Decision | Recommendation and benefit | Cost / rejected alternative | Reopen when |
|---|---|---|---|
| Native backend | Keep Go/standard library, C17/Clang, interpreter oracle | C ABI and defined-semantics lowering require care; a VM/LLVM rewrite adds another backend before user value | A measured consumer/latency/target limit cannot be resolved within this path |
| M004 versus arithmetic first | Honor resource commitments, establish real application IO and lifetime first; deliver an early scalar application witness | Delays arithmetic by this bounded milestone; computation-first retains synthetic execution and leaves NAT-09 unresolved | Resource scope grows beyond the specified witness; reduce shapes before adding phases |
| First owned resource | Audited C adapter returns a live bounded allocation; Schway uses/transfers/releases it | Adapter still owns file open/close; this does not demonstrate Schway-owned file handles | A consumer needs direct file/socket ownership with specified close-error semantics |
| Pointer families | Separate bounded shared/exclusive read-copy support, no added `restrict`/`noalias` | More work than ordinary foreign alone; no generic mutation/escape/retention | A concrete consumer requires a broader access contract and independent evidence |
| Cleanup | Successful acquisition creates an obligation; release consumes it, transfer moves it, borrow preserves it | Requires real dynamic identities and per-operation contracts; release-op-seeded tracking can miss deleted cleanup | Aggregates, loops, cancellation or fallible destructors change the model |
| Iteration | Start with scalar loop-carried state; retain resource/loan-across-back-edge refusal initially if needed | Needs CFG fixed-point analysis, event occurrence identity and bounded oracle policy | FizzBuzz/sum witness exposes a necessary broader construct |
| Text and libraries | Fixed text + decimal output first; byte-oriented reusable APIs next | Full String/Unicode/collection abstractions remain separate work | A selected consumer needs encoding/ownership behavior beyond byte literals |
| Planning automation | AGENTS procedure + two living documents | Requires agents to perform the review; no scheduled/background service | Repeated missed triggers demonstrate a need for a small executable check |

## Checker and guarantee activation map

The six existing operation consumers are `check`, `corevalidate`,
`originvalidate`, `pathoracle`, `interp`, and `cgen`. Update affected consumers
within the feature slice. Preserve independent derivation at trust crossings;
do not make all peers import the producer's conclusion. A registry entry or
engine agreement alone cannot prove correctness.

| Capability trigger | Analysis/contract that changes | Decisive evidence and refusal boundary | Owner |
|---|---|---|---|
| Real application IO | Entry inputs, ordinary IO/exit versus evidence; side effects | Phase 22's automated tests establish bounded U64 input, one native launch, ordinary streams, separate evidence capture, and explicit independent replay; `TestPhase22READMEContract` pins public forms and limits; general file IO remains unimplemented | Phase 22 closed; extend only for a concrete consumer |
| Live owned allocation | Acquisition-based obligation conservation; per-operation ABI/destructor; noncopyability | Omitted physical destructor with unchanged events must fail; discard must not erase obligation | M004 acquisition/discharge phase |
| Transfer/calls/errors | Owner and resource identity across frames; reverse completion order; failed acquisition | Phase 24's ordinary success and typed-error apps, distinct activations, acquisition-derived peers, and physical observer passed on both hosted hosts; general cleanup shapes remain refused | Phase 24 complete; extend only for a selected broader cleanup consumer |
| Shared/exclusive pointer access | Borrow endpoints, independent straight-line overlap/escape refusal, source-attributed escape diagnostics, actual C ABI and emitted attributes | `TestPhase25PointerPathLiveOverlap`, `TestPhase25PointerPathBorrowedResultEscape`, `TestPhase25IndependentPeerMutations`, and command/diagnostic attribution checks pass locally; hosted macOS/Linux receipts remain required. No unsupported alias/alignment/capture promises | Phase 25 evidence aggregate; keep NAT-11/12/13 and EVD-10 open until hosted receipts are bound |
| Arithmetic and Bool | Typed operators; overflow/divide/remainder rules and C definedness | Boundary expected values, wrong-result controls, interpreter/native agreement; no reliance on C undefined behavior | Following practical-computation milestone |
| Scalar CFG back edges | Fixed-point state/loans, dynamic occurrence identity, finite oracle exploration | Loop spike before planning; separate application semantics from evidence-budget exhaustion; reject unproved resource/loan carries | Following practical-computation milestone |
| Byte views/indexing/aggregates | Bounds, initialization, subobject layout/ownership and error paths | Empty/one/max/truncated cases; out-of-bounds controls; only admitted moves/copies | Byte-library capability |
| JSON/HTTP input | Untrusted bytes, limits, framing/encoding, allocation/error cleanup | Malformed input, depth/size budget, split/truncated messages; adversarial cases from selected protocol | Respective library consumer |
| Separate compilation | User-declared facts not verifiable inside producer; contract compatibility | Real B1 boundary witness before DX-06 closure; modules alone do not establish it | Separate-compilation capability (D-13-02b) |
| FFI expansion/async | Unwind, partial acquisition, cleanup failure, suspension/cancellation | Define consumed/retained/unknown resource states and error precedence first | Future selected consumer; refused in M004 |

Security and low-level control remain explicit: local C inputs are build
authority, foreign declarations are trusted assertions, external inputs have
limits, and ordinary pointers do not authorize unchecked Schway aliasing. This
roadmap does not promise to prove arbitrary foreign C correct.

## Current three recommendations

1. **Deliver ordinary scalar computation in Phase 26.** User-visible program:
   `sum_to_n`, followed by FizzBuzz from ordinary Schway source with exact
   output and boundary behavior. Current blocker: defined U64 arithmetic and
   remainder/overflow, comparison/Bool, scalar-loop fixed points, and bounded
   text/decimal output are absent. Smallest complete slice: one bounded-input
   `sum_to_n` command with pinned output, arithmetic boundaries, and a
   loop-carried-state mutation control; reuse that core for FizzBuzz.
   Checker/guarantee changes: typed operators with defined C behavior,
   explicit overflow/error semantics, CFG fixed-point state/loan analysis,
   and bounded output effects. Evidence/debt: NAT-09 and current arithmetic,
   CFG, and output refusal witnesses; add independent interpreter/native
   answers and changed-assumption controls. Owner/next action: Phase 26
   kickoff after M004 acceptance. Reprioritize if a named consumer demonstrates
   a smaller safe computation slice with equal runnable value.
2. **Build a bounded byte utility and reusable local module boundary.**
   User-visible program: a `checksum` command that reads a size-limited file,
   computes a specified checksum, and returns exact success and error results.
   Current blocker: byte views/indexing, bounded iteration, arithmetic, and
   reusable local modules are not admitted. Smallest complete slice: choose one
   checksum consumer, specify maximum input and empty/truncated behavior, then
   admit only the byte operations and module imports it needs. Checker/guarantee
   changes: prove bounds and initialization, analyze loop-carried state and
   resource loans, preserve cleanup on errors, and define module visibility.
   Evidence/debt: build on Phase 24 RES-06 and Phase 26 arithmetic/loop tests;
   use malformed, empty, maximum-size, and wrong-result controls. Owner/next
   action: post-Phase-26 planning after a checksum consumer is named.
   Reprioritize if a concrete file tool needs a smaller non-checksum byte API.
3. **Enable one bounded JSON configuration consumer after a real need is named.**
   User-visible program: read and validate a size-limited JSON config file with
   independently specified success and typed-error results for malformed,
   truncated, and oversized input. Current blocker: general byte views/indexing,
   arrays/strings, bounded loops, and reusable modules are not admitted; Phase
   25 pointer support alone does not provide them. Smallest complete slice:
   choose one consuming tool, define its schema and byte/depth limits, and
   implement only the parser forms required by that config with bounds checks
   and failure cleanup. Checker/guarantee changes: prove bounds and
   initialization, analyze loop-carried state and resource loans, preserve
   owned-buffer cleanup on parse errors, and independently mutate truncation,
   length, and schema cases across interpreter/native runs. Evidence/debt:
   build on Phase 24 RES-06 typed-error cleanup and define consumer-specific
   requirement/evidence IDs at kickoff; make no general JSON completeness
   claim. Owner/next action: post-Phase-26 planning, only after a named Schway
   tool needs this format. Reprioritize if a real consumer needs another input
   format or its config fits a smaller bounded representation.

For each candidate, the smallest complete slice, checker changes, evidence,
owner/next action, and reprioritize observation determine whether its ranking
should change.

The next milestone can then deliver `sum_to_n` and FizzBuzz. Its current
blockers are defined U64 arithmetic/remainder, comparison/Bool, scalar loop
fixed points, and minimal text/decimal output. Do not pull that work ahead of
M004 unless a selected consumer or measured safety constraint demonstrates a
more valuable complete program.

## Maintaining pace and truth

- At kickoff, before phase discussion/planning, and after phase completion,
  inspect source witnesses and refusal boundaries, then refresh this document
  and LANGUAGE-MATURITY only where facts changed. Surface three ranked next
  capabilities without waiting for a prompt (procedure in AGENTS.md).
- Every feature phase names a source/input/output witness and expected result
  before plans expand. A refused frontier is a starting point; acceptance
  requires the intended behavior, not merely a moved first diagnostic.
- Allow at most one consecutive enabling phase without a runnable gain. A
  second triggers scope review. When a phase exceeds its initial plan estimate,
  identify the specific witness/safety obligation before adding more plans.
- Record plans and effort per runnable gain, remaining blockers, and measured
  cold/warm feedback distributions. Avoid percentage estimates of language
  completeness. M003 delivered 85 plans against research forecasts of 40–57;
  that variance justifies tighter slices, not removing correctness evidence.
- Use focused evidence during edits and full required lanes at integration
  boundaries. Assign each expensive CI lane one owner. The known duplicate
  full/race/vet work is a measured-maintenance candidate, not kickoff scope.
- Keep source inspection, executed checks, and historical receipts distinct.
  macOS-only evidence cannot close a Linux claim; missing evidence keeps its
  family incomplete. Do not rerun completed Phase 21 UAT to rewrite history.
- At milestone close, review this arc and promote only the next useful scope.
  The charter may change; record why, successor ownership, and affected claims.

## Provenance

**2026-10-01 Phase 24 closeout amendment.** Phase 24 completes bounded
PathToken/file-byte owner transfer through a helper and return, distinct
repeated activations, and reverse cleanup on success and a real typed error.
The verifier passes 5/5 roadmap truths. Hosted run 36856048690 passed the full
and focused suites on Linux/x86_64 and Darwin/arm64 at source SHA
`ed94ef7972b25deb85ab90fafbf3403dc31629f5`; focused Phase 24 receipts were
25s and 22s. No local project tests or scripts were run. The source still
refuses by-pointer function bodies; Phase 25 owns separate shared/exclusive
pointer witnesses and the integrated utility.

**2026-10-01 Phase 25 Plan 06 amendment.** Source inspection confirms the
production emitter admits only the bounded shared and exclusive U64 read-copy
helpers used by `examples/phase24/transfer.schway`; their manifests make no
unsupported alias, capture, alignment, or ownership promise. Mutation,
forwarding, retention, callbacks, nonlocal exits, and wider pointer forms
remain refused before C serialization. Newly executed local Plan06 checks—the
focused native contract tests and `sh scripts/verify-phase25.sh`—passed on this
macOS/arm64 host with Go 1.24.0 and Apple Clang 21.0.0 at source revision
`f00cdf8` with a modified working tree. Baseline `-O0`, optimized `-O2`, and
ASan+UBSan native utility runs returned 65/66 and preserved the typed 0x43
failure before either helper. Reached wrong-result controls, separate
conflict/escape controls, and pointer manifests also passed. Across the three
lanes, cold Go-cache min/median/max was 2.44/2.72/2.73s and warm-cache
min/median/max was 0.27/0.43/0.53s. The full local test suite, race suite,
`go vet ./...`, and `go build ./...` also passed on macOS/arm64. The Phase 25
script marks every Linux family/lane row incomplete; these local checks are
not hosted receipts. Hosted macOS and Linux Phase 25 family-by-lane receipts
remain pending, so no hosted pass is claimed and EVD-10 remains open.
Historical receipts are the Phase 24 run 36856048690 and earlier Phase 23
runs, scoped to their archived source revisions and behaviors.

**2026-10-02 Phase 25 security-audit follow-up.** Source inspection at
`b19446f` confirms that the command gate invokes core, origin, and path
validation in precedence order. The independent path oracle derives
straight-line borrow ancestry, last use, conflict families, and terminal
escape without reading endpoint claims; the origin peer supplies structured
function/return identity for source projection. Newly executed local tests
`TestPhase25PointerPathLiveOverlap`,
`TestPhase25PointerPathBorrowedResultEscape`,
`TestPhase25IndependentPeerMutations`,
`TestPhase25CheckCommandPeerGate`, `TestPhase25CheckCommandPeerObservation`,
`TestPhase25EscapeDiagnosticSourceAttribution`, and
`TestPhase25EscapeDiagnosticFunctionIdentity` pass, along with the retained
conflict/schema/boundary controls. Historical Phase 25 native results remain
scoped to source revision `f00cdf8`; this follow-up ran no native or hosted
matrix. EVD-10 and phase acceptance remain open pending hosted family-by-lane
receipts.

**2026-09-30 Phase 23 closeout amendment.** Phase 22's verifier passes 5/5
roadmap truths and its completed objective README UAT is preserved. Phase 23's
refreshed verifier also passes 5/5, closes the hosted Ubuntu receipt gap, and
records Phase 23 complete. Hosted run 36707529870 passed the Ubuntu/macOS
check suites and both evidence aggregates, including the Phase 23 script, at
source SHA `f991298b29779838a2b1a5c3cd5ac90aafcb84fc`. No local project suites
or human UAT were rerun/needed for Phase 23 closeout. The earlier same-date
pre-closeout snapshot is historical; 2026-09-27 and 2026-09-28 observations
below remain true at their recorded revisions.

Repository research baseline: `d9bde05`; research date 2026-09-27. Current
observations are refreshed after Phase 22 from source inspection, its named
tests, and the full Go suite. The refreshed verifier at 2026-09-27T23:05:37Z
reports 17/17 truths and `passed` after the objective README contract UAT. No
subjective readability claim is made. These are macOS
observations; Linux remains a separate required host lane. Roadmap rows are
proposed future acceptance, not evidence of implementation. Detailed source
paths, official ecosystem references, and adversarial findings are in
[research/M004](research/M004/SUMMARY.md).
