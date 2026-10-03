# Schway Language Maturity — Current Evidence

**Re-assessed:** 2026-10-03, against current source and Phase 26's passed
5/5 goal verification. The checked scalar sum is now runnable through the
public app route. Current focused and full hosted receipts pass on Ubuntu and
macOS; details and remaining refusals are below. Earlier phase amendments are
preserved as dated evidence, not substituted for current implementation checks.
Future direction lives in [PRODUCT-ROADMAP.md](PRODUCT-ROADMAP.md).

### Phase 26 completion amendment — 2026-10-03

**Source inspection:** `examples/sum_to_n.schway` now uses local mutable U64
state, `<`, checked `+`, `if/else`, and pre-tested `while`; its immutable scalar
copy emits distinct dynamic occurrences. Checker, core validator, and origin
validator derive scalar admission independently with deterministic 65,536-work
budgets. Ownership, resources, loans, loan-derived provenance, and unsupported
cyclic events remain refused. The path oracle remains acyclic; it is not used
as a scalar-loop proof. No new backend or dependency was added.

**Newly executed hosted evidence:** focused run
[37118144404](https://github.com/szTheory/schway/actions/runs/37118144404) and full
run [37118315516](https://github.com/szTheory/schway/actions/runs/37118315516)
passed on Ubuntu and macOS at source revision
`46bee44ad87891d8a8547b2489f029d0fe237899`. Public inputs 0/10/1,000 produce
exactly `0\n`/`55\n`/`500500\n`; 1,001 fails without stdout. Direct overflow
fails with exit 65 and exact `schway: U64 addition overflow\n` stderr. Each
engine is checked against literal expected answers, including reached
wrong-result and skipped-iteration controls. Repeated-copy ordinals, forged
sequences, independent analysis exhaustion/refusals, and evidence capacity
controls pass. Full/race suites and current evidence aggregates pass too.
Independent verification passes 5/5 roadmap truths; review is clean and all
12 declared security mitigations are closed at ASVS L1.

**Limits and provenance:** no local tests or native runs were performed for
Phase 26. Prior M004 receipts remain historical. The 35-pair validation corpus
receipt is from hosted run 37098702559; current semantic proof comes from the
later implementation-bound runs above. Cold/warm latency distributions were
not measured. Capture remains separate from independent verification, and the
event peer checks structural attribution rather than proving overflow operands.
Equality, remainder, fixed text, shared output limits, and the complete M005
cross-program evidence/documentation contract remain Phase 27-owned.

### Phase 25 closeout amendment — 2026-10-01

**Source inspection:** the emitter admits only the integrated shared and
exclusive U64 read-copy helpers; the independent path and origin peers accept
the exact transfer chain. The emitted pointer manifest carries no unsupported
alias, `restrict`, `noalias`, capture, alignment, or ownership promise.
Mutation, forwarding, retention, callbacks, nonlocal exits, wider pointer
forms, and borrowed-value return bypasses remain refused. The hosted result
below closes EVD-10 without widening this implementation boundary.

**Newly executed checks:** `go test -count=1 ./...`, `go vet ./...`,
`go build ./...`, and `go test -race -count=1 ./...` passed locally on
macOS/arm64. The focused native contract tests and
`sh scripts/verify-phase25.sh` passed with Go 1.24.0 and Apple Clang 21.0.0 at
source revision `f00cdf8` with a modified working tree. Baseline `-O0`,
optimized `-O2`, and ASan+UBSan runs returned 65/66 and preserved the typed
0x43 failure before either helper. Reached wrong-result controls, distinct
conflict/escape controls, and pointer manifests passed. Across the three
lanes, cold Go-cache min/median/max was 2.44/2.72/2.73s and warm-cache
min/median/max was 0.27/0.43/0.53s. These are local checks, not hosted
receipts. The script reported every Linux family/lane row incomplete, so
EVD-10 remains open until hosted macOS and Linux runs report every family and
applicable lane.
**Historical receipts:** Phase 24 run 36856048690 and Phase 23 run 36707529870
remain scoped to their recorded source revisions and do not prove Phase 25
pointer behavior.

### Phase 25 security-audit amendment — 2026-10-02

**Source inspection:** at code revision `b19446f`, the command gate consults
`corevalidate`, then `originvalidate`, then `pathoracle`. The path oracle now
independently derives straight-line loan ancestry, last uses, access-family
conflicts, and terminal borrowed-result escapes from checked operations and
place/type facts, ignoring empty or forged `LoanEndpoints`. Its existing CFG
endpoint routine remains unchanged in scope. Origin problems identify the
function and return operation structurally; command diagnostics project those
facts to the actual result and borrow source tokens. The wire schema remains
`lang.diagnostic/0`, and the escape refusal has no repair suggestion.

**Newly executed checks:** the exact task commands passed locally with the Go
cache redirected to `/private/tmp/schway-gocache`. Named controls include
`TestPhase25PointerPathLiveOverlap`,
`TestPhase25PointerPathBorrowedResultEscape`,
`TestPhase25IndependentPeerMutations`,
`TestPhase25CheckCommandPeerGate`, `TestPhase25CheckCommandPeerObservation`,
`TestPhase25EscapeDiagnosticSourceAttribution`,
`TestPhase25EscapeDiagnosticFunctionIdentity`, and
`TestPhase25FamilyConflict`. They cover endpoint-claim mutations, separate
core/origin/path peer verdicts, command refusal, both escape access families,
exact primary/cause bytes, stable diagnostic identity, and no unsafe repairs.
Corevalidate's escape verdict is its non-callable signature fact;
originvalidate and pathoracle each return their own direct refusal. These are
Go unit controls, not native host evidence.

**Historical receipts:** the macOS/arm64 Phase 25 native lanes remain at
revision `f00cdf8`, and the earlier Phase 23/24 hosted runs remain scoped to
their own source revisions. This follow-up ran no native evidence script or
hosted workflow. Linux rows were incomplete; hosted macOS/Linux family-by-
lane receipts were still required, so EVD-10 stayed open at the time of that
follow-up. The three ranked next capabilities and their user witnesses,
blockers, slices, checker changes, evidence/debt, owners, and reprioritization
observations are unchanged.

### Phase 25 hosted-evidence closeout amendment — 2026-10-02

**Newly executed hosted checks:** GitHub Actions run `36971855722` tested
branch head `421b5b94eb867e940a207dfd64d7971dc8198172` at PR merge revision
`a90c27c5b432ef6fc59fbafaa68b50a1374ae138`. Its `checks` and `current evidence
aggregate` jobs passed on both native hosts. Checks included vet, build, full
Go tests, and race tests; the aggregate included the Phase 25 evidence script.
Linux ran Go `linux/amd64`, target `x86_64-pc-linux-gnu`, Ubuntu Clang 18.1.3;
macOS ran Go `darwin/arm64`, target `arm64-apple-darwin25.6.0`, Apple Clang 21.
Foreign, shared, and exclusive each passed baseline `-O0`, optimized `-O2`,
and ASan+UBSan on both hosts: all 18 host/family/lane rows passed, each with
expected/actual 65/66 and typed `0x43` use failure before helper calls. Cold
min/median/max was 9.350/9.540/9.980s (Linux) and 12.470/12.670/12.930s
(macOS); warm was 0.390/0.490/0.550s and 0.820/1.400/2.130s respectively.
Single-host script logs mark only the other host absent in that invocation;
the paired native aggregate supplies its independently run rows.

**Provenance boundary:** this hosted receipt is distinct from source
inspection, the local macOS Go/native checks at earlier revisions, and
historical Phase 23/24 receipts. It closes EVD-10 and establishes the current
Phase 25 matrix. It does not broaden the exact utility-chain admission or its
refusals: mutation, forwarding, retention, callbacks, nonlocal exits, wider
pointer shapes, and unsupported alias/alignment/capture/ownership attributes
remain unclaimed. The 2026-10-02 security assessment remains SECURED, ASVS L1,
15/15 mitigations closed, with zero open threats.

The next ranked capabilities are Phase 26 scalar computation (`sum_to_n`, then
FizzBuzz), one bounded checksum consumer with only the byte operations and
module imports it needs, and a bounded JSON configuration consumer only after a
real tool needs one. Each candidate requires its smallest complete slice,
checker/guarantee changes, evidence/debt, owner/next action, and a
reprioritization observation, as detailed below.

For each candidate, the smallest complete slice, checker changes, evidence,
owner/next action, and reprioritize observation determine whether its ranking
should change.

### M005 kickoff — 2026-10-02

The user accepted M005 Practical Computation: run bounded `sum_to_n`, then
FizzBuzz through ordinary Schway source. Kickoff source inspection confirms
`check.cfg_back_edge` still refuses cyclic CFGs in
`internal/compiler/check/check.go`; `examples/checksum.schway` remains a
provisional refused loop witness pinned by `TestPhase20ChecksumFrontier`
in `internal/compiler/session/session_phase20_test.go`. U64 support remains
limited to constants; arithmetic is absent. No project tests or native
evidence were run for this planning transition. Phase 26 now owns the smallest
runnable scalar computation slice. Preserve the resource/loan back-edge refusal
until fixed-point analysis and independent controls justify widening it.

The next recommendations remain: (1) complete M005 practical computation;
(2) after a named consumer, admit the minimum byte operations and local module
boundary for a bounded checksum; (3) admit a bounded JSON configuration reader
only when a real Schway tool needs it. Keep this order unless a named user
program shows that another smaller, safe runnable slice has greater value.

### Phase 26 discussion amendment — 2026-10-02

The user adopted the Phase 26 specialist recommendations: mutable local
`U64`/`Bool` scalar state, pre-tested `while`, `if/else` blocks, and only `<`
for the sum witness. U64 addition is checked; overflow maps to a program-level
application failure with bounded stderr and no stdout, rather than a
source-catchable typed error. Equality and remainder stay with Phase 27. These
are planning decisions, not implemented capability or new execution evidence.
Source inspection still finds `check.cfg_back_edge` in
`internal/compiler/check/check.go`, `TestBackEdgeRejected` in
`internal/compiler/check/check_test.go`, and the refused checksum frontier in
`TestPhase20ChecksumFrontier`. No project checks were run for this discussion;
M004 hosted receipts remain historical and do not prove M005 semantics.

### Phase 24 closeout amendment — 2026-10-01

Source inspection confirms that `examples/phase24/transfer.schway` and
`error.schway` admit the bounded file-byte owner through a helper return and a
later typed-error path. The acquisition-derived peers distinguish repeated
dynamic activations; the C emitter and native app tests exercise cleanup; an
independent physical observer checks actual allocation/use/free and five
reached destructor controls. Hosted CI 36856048690 passes the Phase 24 focused
aggregate and full/race suites on Linux/x86_64 and Darwin/arm64. This is newly
executed hosted evidence at `ed94ef7972b25deb85ab90fafbf3403dc31629f5`, not a
local test receipt. Model replay still does not claim physical cleanup or host
IO. Both shared and exclusive by-pointer bodies remain refused by native
emission.

### Documentation-gate amendment — 2026-09-27

Source inspection found the existing Phase 22 README covers its build, run,
manifest, evidence, replay, trust, and closure boundaries but lacked an explicit
same-run event-evidence command. The focused `TestPhase22READMEContract` and
named CLI checks passed on this macOS host after adding that example; no CI run
or Linux execution is claimed here. These are newly executed documentation and
CLI checks, distinct from the earlier Phase 22 verifier receipt. The regenerated
report at `2026-09-27T23:05:37Z` passes with 17/17 truths and the objective README
contract UAT. The contract test does not establish subjective readability.
This dated note records the earlier Phase 22 documentation gate; the current
thresholds below reflect the later Phase 23 closeout.

## What exists

| Capability | Observed boundary | Source / evidence anchor |
|---|---|---|
| Source → checked core → interpreter/native C17 | A working compiler exists; Go 1.24/Clang remain the development path | `cmd/schway/main.go`, `internal/compiler/session/session.go` |
| Retained native application build/run | `schway app build` creates a retained artifact from an admitted source and closed local-C manifest; `schway app run` accepts bounded U64 input, launches once, and preserves ordinary streams/outcomes | `cmd/schway/main.go`, `session.go`, `internal/compiler/native/native_app.go`; Phase 22 `TestPhase22IdentityApplicationBuildAndRunCLI`, `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes` |
| Separate application evidence | Capture is a distinct report with disabled/incomplete/complete/capacity states and never claims verification; explicit `app verify` replays isolated inputs against independent expected answers | `native_app.go`, `main.go`; `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, `TestPhase22AppVerifyIndependentIdentityCases`, `TestPhase22AppVerifyModelOnlyOutcomes` |
| Live local foreign allocation | A public file-byte path admits only exact acquire/use/release contracts; Schway holds the acquired byte until local release, and an independent physical observer plus negative controls prove malloc → use → matching free → exit | `examples/phase23/file_byte.schway`, `examples/phase23/adapter.c`, `internal/compiler/cgen/cgen_program.go`, `internal/compiler/native/phase23_observer_test.go`, `scripts/verify-phase23.sh`; the focused script passed in hosted Ubuntu/macOS evidence aggregates at `f991298b`; local macOS/Linux-container receipts at `c50430d` are historical |
| Bounded transfer and typed-error cleanup | The Phase 23 file-byte owner can move through the Phase 24 helper/return path; repeated calls retain distinct identities and the `0x43` typed-error path releases C,B,A exactly once | `examples/phase24/transfer.schway`, `examples/phase24/error.schway`, `internal/compiler/session/session_phase24_model_test.go`, `internal/compiler/native/phase24_observer_test.go`, `scripts/verify-phase24.sh`; hosted CI 36856048690 passes at `ed94ef79` on Linux/x86_64 and Darwin/arm64 |
| Bounded shared/exclusive pointer successors | The integrated utility calls separate shared and exclusive U64 read-copy helpers; independent straight-line validation rejects incompatible overlap and borrowed-result escape, and command escape diagnostics name the result and borrow origin; only plain pointer shapes are emitted | `examples/phase24/transfer.schway`, `internal/compiler/cgen/cgen_program.go`, `TestPhase25PointerPathLiveOverlap`, `TestPhase25PointerPathBorrowedResultEscape`, `TestPhase25IndependentPeerMutations`, and session source-attribution tests; local Go controls and hosted run 36971855722 (18/18 native rows) pass |
| Calls between Schway functions | Multi-function programs run; call-graph cycles remain refused | Phase 11 archive; `testdata/phase07/call_basic.schway`; `session.RunInterpreter` / native path |
| Returns independent of parameter type | Admitted through the production pipeline | Phase 17 archive and `session_phase17_test.go` |
| Computed `Result` matches and payload returns | Admitted; this is not unrestricted statement control flow | `session_phase18_payload_test.go`, Phase 18 archive |
| U64 constants and checked scalar computation | Constants, checked addition, `<`, Bool branches, mutable U64/Bool state, and pre-tested scalar loops run; equality/remainder are not yet admitted | `examples/sum_to_n.schway`, `examples/phase26/checked_add_overflow.schway`, `TestPhase26ExactSumMatrix`, `TestPhase26CheckedAddInterpreter`, `TestPhase26CheckedAddC17` |
| Bounded scalar CFG and repeated events | Three independent analyses admit scalar state with deterministic bounds; repeated copies have per-invocation/site occurrence identity | `check/scalar.go`, `corevalidate`, `originvalidate`, `TestPhase26AnalysisExhaustion`, `TestPhase26PeerOccurrenceOrder`, `TestPhase26SourceRepeatedCopyEvidence`; authority-bearing carries remain refused |
| Ownership/borrow checking | Affine ownership, interprocedural loans, independent core/origin validation, bounded path oracle | `check`, `corevalidate`, `originvalidate`, `pathoracle` |
| Evidence | Interpreter/native comparison, mutation controls, manifests, scoped optimizer/sanitizer lanes, structured diagnostics/repair | Phase 14–20 records; each claim retains its scope/grade |
| Phase 21 | Completed contract and emitter-retirement work, six plans, seven UAT cases | `milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/` |

Calls became executable in Phase 11. U64 constants landed in Phase 19. Older
claims that calls only check or that Byte is the only scalar are superseded.
The available compiler is not a general application runtime yet.

## What prevents ordinary programs

| Missing or constrained | Practical consequence | Current evidence |
|---|---|---|
| General caller input and IO | The app route currently accepts one bounded U64 token; it does not provide general file, byte-string, or network IO | `cmd/schway/main.go:runApplicationRun`, Phase 22 CLI tests |
| Legacy differential run route | `schway run --engine=native` remains a conformance harness with synthesized inputs and O0/O3 comparison; effectful application code must use `schway app run` | `session.go:runNative`, `cmd/schway/main.go` |
| Schway foreign operations and pointer parameters | Phases 23–24 admit only the exact file-byte acquire/use/release and bounded transfer/error shapes. Phase 25 admits shared/exclusive U64 read-copy helpers only in its exact utility chain; other foreign-call shapes and broader pointer bodies remain structurally refused | `internal/compiler/cgen/cgen_program.go`, `TestPhase23OperationContract`, `TestPhase25ExclusivePointerRefusal`, `TestPhase25PointerPathLiveOverlap`, `TestPhase25PointerPathBorrowedResultEscape`, `TestPhase25EscapeDiagnosticSourceAttribution`, and peer mutation controls; hosted run 36971855722 passes all 18 family/host/lane rows |
| Host build closure | Known source/header/compiler/target inputs are identity-bound, but the SDK/linker/runtime closure is incomplete; app artifacts are not cacheable | `native/bindings.go`, `examples/phase22/BINDINGS.md`, Phase 22 build receipts |
| Schway-owned physical cleanup | A bounded owner stays live through helper use and is released exactly once on normal and typed-error paths; general cleanup shapes, fallible destructors, and nonlocal exits remain unadmitted. The C adapter owns file descriptors and frees partial buffers on acquisition failures | `examples/phase24/error.schway`, `examples/phase24/README.md`, `internal/compiler/native/phase24_observer_test.go`, `scripts/verify-phase24.sh`; hosted CI 36856048690 |
| Remaining computation/output surface | Checked addition, `<`, Bool branches, and scalar loops run; equality, remainder, fixed text, and the shared M005 output contract still block FizzBuzz | Phase 26 source/tests and 26-VERIFICATION.md; Phase 27 owns U64-02/APP-08/APP-09/EVD-12/DX-16 |
| Resource/loan loop carries and cyclic path enumeration | Scalar-loop admission does not authorize ownership, resources, loans, or loan-derived provenance across back edges; the path oracle remains acyclic | Phase 26 category/peer mutation controls and `TestCompositionCycleGuardFailsClosed` |
| General strings, arrays, collections, usable library modules | Ordinary JSON/HTTP libraries are not yet writable | current syntax/checker frontier; design wiki is prospective |
| Recursion and broader resource control | Call cycles, nonlocal exits, cancellation, general fallible cleanup and escaped pointers require further contracts | callgraph and current foreign refusal boundaries |

Phase 23 adds the first public foreign-C file-byte path. Its three exact
operation contracts acquire a byte allocation from a file, keep the local
owner live through use, and discharge it before returning. An independent
observer witnesses physical allocation, use, the matching free, and no
outstanding pointer; reached negative controls reject omitted, premature,
duplicate, and wrong-resource destruction. Acquisition tests include partial
allocation and injected close failures. The refreshed verification report
passes 5/5 truths; hosted run 36707529870 supplied passing Ubuntu and macOS
focused aggregates at `f991298b`. Earlier focused macOS and Linux-container
receipts after the close-failure fix remain historical evidence. No local
suite was rerun for the refresh. Phase 24 extends the same resource through a
bounded helper and caller, distinct dynamic activations, and a real typed error
after successful acquisitions. Its independent physical observer proves
post-transfer use, C,B,A destruction, zero outstanding pointers, and five
reached physical controls; hosted run 36856048690 passed focused receipts in
25s on Linux/x86_64 and 22s on Darwin/arm64, plus both full/race suites and
evidence aggregates. Model-only replay continues to report
`actual_host_io=false` and `physical_cleanup=false`. These checks establish the
bounded Phase 24 contract; they do not admit general cleanup or broader pointer
mutation, forwarding, retention, or escape shapes.

`examples/checksum.schway` is a refused integration target with provisional syntax.
Its Phase 20 test pins the `loop` refusal; moving that diagnostic does not by
itself demonstrate a checksum, file IO, or output. `wiki/example-tour.md` remains
design exploration, not a supported language specification.

## Corpus and guard census

Corpus: **153 `.schway` programs, 4,964 lines total** (~32 lines average,
193-line maximum). These counts match the current machine-checked tree; the
corpus predominantly contains focused semantic fixtures, not applications.

A non-test AST census finds **19 `len(Functions) != 1` guards across 6 files in 1 packages** (52 including tests):

| Package | Guards | Notable sites |
|---|---|---|
| `session` | 19 | Bounded historical verification/reducer lanes; inspect each site's purpose rather than inferring that public multi-function run is refused |

`internal/compiler/session/self_describing_docs_test.go` independently derives
these numbers with Go's parser. The count describes a syntactic predicate; it
cannot determine public capability. The model-only foreign-outcome interpreter
now carries acquisition-derived owner identity across the helper return into
the caller; historical single-function evidence lanes and multi-function
production execution still coexist. The reducer also supports multi-function
seeds; its `<= 1` and `== 1` shortcuts are outside this census.

Re-verify (approximate only — Go AST evidence is authoritative):
`rg 'len\([^)]*Functions\) != 1' internal cmd`.

Machine check: `GOCACHE=/tmp/schway-verification-gocache go test ./internal/compiler/session -run '^(TestLanguageMaturityCountsAreCurrent|TestSelfDescribingDocsGuardIsNotInert)$' -count=1`.

## Next useful thresholds

1. **Complete Phase 27 with exact FizzBuzz.** User-visible program: ordinary
   `fizzbuzz.schway` prints the exact lines for inputs 0–1,000. Current blocker:
   equality, remainder with defined zero-divisor behavior, fixed text, and the
   shared output-limit contract are not admitted. Smallest complete slice:
   reuse Phase 26 scalar loops and decimal output, add only the operators and
   fixed writes needed by FizzBuzz, and pin both programs' 65,536-byte ceiling.
   Checker/guarantee changes: typed equality/remainder, static zero-divisor
   refusal, dynamic checked failure, and bounded output effects; keep resource
   and loan back-edge refusals. Evidence/debt: Phase 26 U64-01/FLOW-01/FLOW-02/
   APP-07 are verified; Phase 27 owns U64-02, APP-08, APP-09, EVD-12, and DX-16,
   including independent literal answers, reached mutations, and both-host
   evidence. Owner/next action: Phase 27 discussion and planning. Reprioritize
   only if a named consumer exposes a smaller safe slice with greater runnable
   value or a Phase 26 safety regression blocks reuse.
2. **Build a bounded byte utility and reusable local module boundary.**
   User-visible program: a `checksum` command reads a size-limited file and
   computes one specified checksum with exact success/error results. Current
   blocker: byte views, lengths/indexing, required checksum operations, and
   local modules are absent; scalar loops now exist, but resource/loan carries
   remain refused. Smallest complete slice: name one consumer and algorithm,
   specify empty/maximum/truncated behavior, then admit only its byte operations
   and an import used by a second consumer. Checker/guarantee changes: bounds,
   initialization, necessary scalar operators, borrow lifetime around iteration,
   cleanup on errors, and module visibility. Evidence/debt: Phase 24 RES-06
   cleanup, Phase 25 pointer controls, Phase 26 exact-loop and refusal tests,
   plus new byte-boundary and wrong-result controls; the provisional checksum
   fixture is still a refusal witness. Owner/next action: post-M005 consumer
   selection. Reprioritize if a real file tool needs a smaller byte API first.
3. **Enable one bounded JSON configuration consumer when a real need is named.**
   User-visible program: validate a size-limited config against one fixed schema,
   with exact success and typed-error results. Current blocker: byte indexing,
   aggregates, a bounded parser stack, and reusable modules remain unavailable;
   scalar loops alone do not supply them. Smallest complete slice: choose a
   consuming tool, define schema and byte/depth limits, and implement only its
   parser forms with failure cleanup. Checker/guarantee changes: bounds and
   initialization, parser-state ownership, borrow lifetime during iteration,
   and owned-buffer cleanup on every error. Evidence/debt: Phase 24 RES-06,
   Phase 26 scalar/refusal controls, and consumer-specific malformed, truncated,
   oversized, depth, and schema cases with independent answers. Assign new
   requirement/evidence IDs at kickoff; make no general JSON completeness claim.
   Owner/next action: post-M005 planning after a named Schway tool needs JSON.
   Reprioritize if another format or a smaller bounded representation serves
   the actual consumer.

M005 has delivered `sum_to_n`; reuse its checked scalar computation for
FizzBuzz. Equality/remainder, fixed text, shared output limits, and the complete
cross-program evidence/documentation contract remain Phase 27 work. Keep
resource/loan back-edge refusals until a named consumer and independent safety
evidence justify a wider contract.

Do not assign percentages to assurance or language completeness: neither has
a stable denominator. Report runnable witnesses, known refusals, observed
feedback cost, and next dependencies. D-12-43's wrong-slot mutation became
constructible in Phase 18; its historical unconstructibility was not a reason
to wait for arithmetic or aggregates.

## Refresh triggers

Refresh after a capability lands, a refusal changes, a milestone closes, or a
selected user program exposes an incorrect claim. At each planning transition,
review the current three recommendations in PRODUCT-ROADMAP and connect changed
syntax/core operations to the necessary checker/evidence obligations. Keep
historical receipts scoped to their revision, host, compiler, and exercised
paths. AGENTS.md specifies this agent-executed review; no background automation
is implied.
