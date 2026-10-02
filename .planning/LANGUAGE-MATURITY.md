# Schway Language Maturity — Current Evidence

**Re-assessed:** 2026-10-01, against current source and Phases 22–25
artifacts. Phase 22's passed 5/5 report and completed objective README UAT are
preserved. Phase 23's refreshed 2026-09-30 report passes 5/5 truths and its
seven plans are complete. Phase 24's verifier passes 5/5 truths with no
behavior-unverified items. Hosted run 36856048690 passed full vet/build/test/
race suites and current evidence aggregates on Ubuntu Linux/x86_64 and macOS
Darwin/arm64 at source SHA `ed94ef7972b25deb85ab90fafbf3403dc31629f5`; the
focused Phase 24 receipts took 25s and 22s. No local project suites or scripts
were run for Phase 24 closeout. Phase 23's security review covers 18/18 threats.
Objective checks cover acceptance; subjective readability is not claimed.
Future direction lives in [PRODUCT-ROADMAP.md](PRODUCT-ROADMAP.md).

### Phase 25 closeout amendment — 2026-10-01

**Source inspection:** the emitter admits only the integrated shared and
exclusive U64 read-copy helpers; the independent path and origin peers accept
the exact transfer chain. The emitted pointer manifest carries no unsupported
alias, `restrict`, `noalias`, capture, alignment, or ownership promise.
Mutation, forwarding, retention, callbacks, nonlocal exits, wider pointer
forms, and borrowed-value return bypasses remain refused.

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

The next ranked capabilities are Phase 26 scalar computation (`sum_to_n`, then
FizzBuzz), one bounded checksum consumer with only the byte operations and
module imports it needs, and a bounded JSON configuration consumer only after a
real tool needs one. Each candidate requires its smallest complete slice,
checker/guarantee changes, evidence/debt, owner/next action, and a
reprioritization observation, as detailed below.

For each candidate, the smallest complete slice, checker changes, evidence,
owner/next action, and reprioritize observation determine whether its ranking
should change.

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
| Bounded shared/exclusive pointer successors | The integrated utility calls separate shared and exclusive U64 read-copy helpers; only plain pointer shapes are emitted, with no `restrict`, `noalias`, capture, alignment, or ownership promise | `examples/phase24/transfer.schway`, `internal/compiler/cgen/cgen_program.go`, path/origin peer checks, and `internal/compiler/native/phase25_utility_test.go`; focused local receipts are run on this checkout, while hosted macOS/Linux receipts remain required |
| Calls between Schway functions | Multi-function programs run; call-graph cycles remain refused | Phase 11 archive; `testdata/phase07/call_basic.schway`; `session.RunInterpreter` / native path |
| Returns independent of parameter type | Admitted through the production pipeline | Phase 17 archive and `session_phase17_test.go` |
| Computed `Result` matches and payload returns | Admitted; this is not unrestricted statement control flow | `session_phase18_payload_test.go`, Phase 18 archive |
| U64 constants | Literals and `OpConst` run; arithmetic does not exist | `session_phase19_test.go`, `testdata/phase19/` |
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
| Schway foreign operations and pointer parameters | Phases 23–24 admit only the exact file-byte acquire/use/release and bounded transfer/error shapes. Phase 25 admits shared/exclusive U64 read-copy helpers only in its exact utility chain; other foreign-call shapes and broader pointer bodies remain structurally refused | `internal/compiler/cgen/cgen_program.go`, `TestPhase23OperationContract`, `TestPhase25ExclusivePointerRefusal`, and Phase 25 family/core/origin/path controls; hosted family receipts remain pending |
| Host build closure | Known source/header/compiler/target inputs are identity-bound, but the SDK/linker/runtime closure is incomplete; app artifacts are not cacheable | `native/bindings.go`, `examples/phase22/BINDINGS.md`, Phase 22 build receipts |
| Schway-owned physical cleanup | A bounded owner stays live through helper use and is released exactly once on normal and typed-error paths; general cleanup shapes, fallible destructors, and nonlocal exits remain unadmitted. The C adapter owns file descriptors and frees partial buffers on acquisition failures | `examples/phase24/error.schway`, `examples/phase24/README.md`, `internal/compiler/native/phase24_observer_test.go`, `scripts/verify-phase24.sh`; hosted CI 36856048690 |
| Arithmetic/comparison and scalar iteration | Cannot add, compute remainder, loop, or write FizzBuzz | syntax/core operation inventory; `pathoracle` rejects CFG cycles |
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

Corpus: **151 `.schway` programs, 4,929 lines total** (~33 lines average,
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

1. **Phase 26 practical computation.** User-visible program: `sum_to_n`, then
   FizzBuzz from ordinary Schway source with exact output and boundary
   behavior. Current blocker: defined U64 arithmetic/remainder and overflow,
   comparison/Bool, scalar-loop fixed points, and bounded text/decimal output
   do not exist. Smallest complete slice: one bounded-input `sum_to_n` command
   with exact expected output, arithmetic boundaries, and a loop-carried-state
   mutation control, then reuse that core for FizzBuzz. Checker/guarantee
   changes: typed operators with defined C behavior, overflow/error semantics,
   CFG fixed-point state/loan analysis, and bounded output effects.
   Evidence/debt: NAT-09 and current arithmetic/CFG refusal witnesses;
   independent interpreter/native answers and changed-assumption controls.
   Owner/next action: Phase 26 kickoff after M004 acceptance. Reprioritize if a
   named consumer demonstrates a smaller safe computation slice with equal
   runnable value.
2. **A bounded byte utility with a reusable local module boundary.**
   User-visible program: a `checksum` command that reads a size-limited file,
   computes a specified checksum, and returns exact success/error results.
   Current blocker: byte views/indexing, bounded iteration, arithmetic, and
   local modules are absent. Smallest complete slice: choose one checksum
   consumer, define input limits and empty/truncated behavior, then admit only
   the byte operations and imports it needs. Checker/guarantee changes: prove
   bounds and initialization, analyze loop-carried state and resource loans,
   preserve cleanup on errors, and define module visibility. Evidence/debt:
   build on Phase 24 RES-06 and Phase 26 arithmetic/loop tests; cover malformed,
   empty, maximum-size, and wrong-result controls. Owner/next action:
   post-Phase-26 planning after a checksum consumer is named. Reprioritize if a
   concrete file tool needs a smaller non-checksum byte API.
3. **A bounded JSON configuration consumer if a real need is named.**
   User-visible program: validate a size-limited JSON config file with pinned
   success and typed-error results for malformed, truncated, and oversized
   input. Current blocker: general byte views/indexing, arrays/strings, bounded
   loops, and reusable modules are not admitted; Phase 25 pointers alone do not
   provide those capabilities. Smallest complete slice: choose one consumer,
   fix its schema and byte/depth limits, and implement only the parser forms it
   needs with bounds checks and cleanup on error. Checker/guarantee changes:
   prove bounds and initialization, analyze loop-carried state and resource
   loans, preserve owned-buffer cleanup, and independently mutate truncation,
   length, and schema cases across interpreter/native runs. Evidence/debt:
   build on Phase 24 RES-06 and define consumer-specific IDs at kickoff; make no
   general JSON completeness claim. Owner/next action: post-Phase-26 planning
   after a named Schway tool needs JSON configuration. Reprioritize if a real
   consumer needs another input format or a smaller bounded representation.

After M004, practical computation can target `sum_to_n` and FizzBuzz. Its
blockers remain defined arithmetic/remainder, comparison/Bool, scalar-loop
fixed points, and bounded text/decimal output.

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
