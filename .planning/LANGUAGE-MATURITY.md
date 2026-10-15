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

### Phase 27 progress amendment — 2026-10-03

**Source inspection:** `examples/fizzbuzz.schway` composes the Phase 26 scalar
loop with U64 equality/remainder and literal-only text writes. The checker,
core validator, and origin validator derive the remainder and write facts
independently. Constant zero divisors are refused; dynamic zero divisors fail
in the interpreter and native app. Both `fizzbuzz.schway` and
`sum_to_n.schway` use the shared 65,536-byte whole-write output bound. Resource,
loan, loan-derived, and unsupported cyclic event carries remain refused.

**Newly executed local evidence:** focused Plan 27-07/27-08 selectors passed
on macOS/arm64 at implementation commit `21fb123`, covering exact answer
matrices, reached arithmetic/remainder/iteration mutations, overflow and zero
controls, repeated event identity and order, output/evidence capacity, and the
public app CLI. Interpreter and native builds were exercised. These focused
checks are not full-suite or hosted Phase 27 receipts.

**Still open:** Plan 27-09 owns the native host driver, same-revision receipt
gate, and CI matrix wiring; Plan 27-10 owns the clean-checkout guide and actual
macOS/Linux hosted receipts. EVD-12 and DX-16 remain open until those artifacts
and hosted runs are complete. Phase 25/26 receipts remain historical.

### Phase 25 closeout amendment — 2026-10-01

**Source inspection:** the emitter admits only the integrated shared and
exclusive U64 read-copy helpers; the independent path and origin peers accept
the exact transfer chain. The emitted pointer manifest carries no unsupported
alias, `restrict`, `noalias`, capture, alignment, or ownership promise.
Mutation, forwarding, retention, callbacks, nonlocal exits, wider pointer
forms, and borrowed-value return bypasses remain refused. The hosted result
below closes EVD-10 without widening this implementation boundary.

2026-10-10 (Phase 28): `core.entry_ambiguous` is a source refusal that exits 2 on check, build, both run engines and app verify; functions `main` never calls are pruned after validation so such programs run (examples/phase28/unused_function.schway); app verify source refusals exit 2 with the refusing code and still write the fail report, while capacity-exhausted cases stay unsupported with exit 3; core and origin refusals exit 2 on every command; one exit-code table in internal/compiler/exitcode, published in the README; app verify schema lang.app-verification/2 carries an outcome digest for capacity-exhausted cases; focused local tests (28-01..07), hosted CI not yet run.

2026-10-11 (Phase 29): source inspection and newly executed focused local checks (29-00..29-12); hosted CI run 38186068052 at 906d648 (macOS and Linux, including `go test -race` and the strict spike 008 step) passes; the 65,536-byte cap is kept, measured through the real fixture owner on both hosts in strict mode (spike 008 VALIDATED, plan 29-12; Stage A in 29-00-GATE-VERDICT.md was a gate-only proxy). A `ByteView {drop}` parameter is admitted on scalar loop functions that carry only U64 and Bool state; `byte_len` and `byte_at` have one meaning across check, corevalidate, originvalidate, pathoracle, the interpreter and C17 emission, with unsigned widening (`0x80` reads 128, `0xFF` reads 255). The bounds check is never elided: an out-of-bounds read exits 65 with stderr exactly `schway: byte index out of bounds` and empty stdout in the interpreter and native at -O0, -O3 and -O3 with LTO (examples/phase29/byte_index_oob.schway). `adler32(bytes: ByteView) -> U64` is pinned on both engines (empty 1, one 0x00 byte 65537, `a` 6422626, `Wikipedia` 300286872, 4,096 x 0x00 268435457, 5,552 x 0xFF 4052720524). Defined meaning (APP-10): Adler-32 follows RFC 1950 with modulus 65521 reduced on every byte (no deferred NMAX batching) and combined as b * 65536 + a via sixteen checked self-additions. Fixture bytes reach the program through a test-only fixture owner: the native gain is `schway build examples/phase29/adler32_fixture.schway --manifest examples/phase29/fixture_bytes.bindings.json --output ART` then `schway app run ART -- N`, the interpreter gain is `interp.RunWithFixtureBytes`; `schway run --engine=*` is unchanged. D-22 scope: releasing the owner before the last use of a view is refused by `check.release_while_view_live`, `core.release_while_view_live` and `pathoracle.release_while_view_live`; an owner move while a foreign-borrow view is live is refused only by generic or closed-shape matchers, not a loan law. Sanitizer evidence (examples/phase29/README.md) ran on darwin/arm64 with AddressSanitizer: adapter acquire/release, ten Adler-32 vectors and the out-of-bounds control each ran with no report; LeakSanitizer does not exist on darwin so no leak detection ran there (`detect_leaks=0`), and no Linux sanitizer lane exists yet (the hosted Linux run executes the plain and -race suites, not ASan); no lane is recorded PARTIAL for an unavailable AddressSanitizer. Known gaps: the formatter mis-formats scalar `var` statement bodies (pre-existing, not fixed), check's fixed point has no operand-type law for byte ops (source-level refusal exists), pathoracle refuses every cyclic CFG identically, and phase28 manifest rows for adler32_fixture.schway still show `native.build_failed` because that walker builds without a manifest.

2026-10-12 (Phase 32.1): `schway run` on an entry that cannot be bound from the command line (a parameterised `main`, engine interpreter or native) is a usage refusal `tool.run_input_required` (exit 64), not an operational failure; a `ByteView` `main` is refused at `check`; interpreter and native agree on non-bounds defects with a live fixture owner; the module-aware set is named once in `check.ModuleAwareRoutes()` (check, run, build, app verify, debug-map) and every other route reads a single file by design, so `explain` and `query` report the `module.imports_not_loaded` refusal as a passing lookup (exit 0) while the remaining single-file routes exit 2; the usage text lists every dispatched route including `stats` and `build --manifest`. Focused local tests (32.1-02..07); hosted CI not yet run for this phase.

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

2026-10-10 (Phase 27.1): schway build and schway app verify now apply the origin peer after corevalidate through the shared application admission helper, in the same order as schway check (CheckCommandFile: checker, corevalidate, origin peer); the interpreter application route is API-only; source inspection plus focused local tests (27.1-01); hosted run 38066884218 at e6ec151 failed on session meta-tests (fixed in 7936dac, no behavior change); hosted run 38068888687 at 7936dac passed on macOS and Linux with fresh receipts passing the pinned gate; see 27.1-HOST-EVIDENCE.md.

2026-10-10 (Phase 27.2): branch-free scalar bodies (let, var, mixed, pure copy, pure constant, Unit write-only) are admitted again end to end; producer-side fix in the checker scalar builder and cgen; validators unchanged; focused local tests (27.2-01, 27.2-02); hosted run 38080236918 at 112eaf7 green on macOS and Linux, see 27.2-HOST-EVIDENCE.md.

2026-10-10 (Phase 27.3): app verify of loop programs now names event-capacity exhaustion per case (native.event_capacity_exhausted, reported unsupported) instead of aborting the whole set; sum_to_n verifies n <= 14 through app verify; loop-free conformance C and all application C byte-identical; the Phase 27 guide writes outputs under ignored .schway/phase27; validators unchanged; focused local tests (27.3-01..03); hosted two-host run 38091738463 at d7ffbcd succeeded (workflow_dispatch), see 27.3-HOST-EVIDENCE.md.

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

M006 final hosted receipts: reopened 2026-10-13 (historical codedigest b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c)

The hosted receipts cover the M006 programs (checksum, adler_low, the adler32 library demo, the Phase 29 Adler-32 fixture and out-of-bounds control) plus sum_to_n and FizzBuzz (N2 closed). Receipt rows are default-build only; the optimization matrix is bound by the same run's go test lanes. Receipts were re-taken after Phase 32.1 at the 32.1 freeze; run 38315799626 is kept as historical; any later code change reopens them.

Amended 2026-10-13: Hosted run 38335175306 (headSha 9f8f878f) receipts were bound to codedigest b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c and are now historical; Phase 33 code changes reopened them. Phase 33 binds no hosted receipt. M006 receipts are historical; current hosted evidence resumes in Phase 38.

| Capability | Observed boundary | Source / evidence anchor |
|---|---|---|
| Source → checked core → interpreter/native C17 | A working compiler exists; Go 1.24/Clang remain the development path | `cmd/schway/main.go`, `internal/compiler/session/session.go` |
| Retained native application build/run | `schway app build` creates a retained artifact from an admitted source and closed local-C manifest; `schway app run` accepts bounded U64 input, launches once, and preserves ordinary streams/outcomes | `cmd/schway/main.go`, `session.go`, `internal/compiler/native/native_app.go`; Phase 22 `TestPhase22IdentityApplicationBuildAndRunCLI`, `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes` |
| Separate application evidence | Capture is a distinct report with disabled/incomplete/complete/capacity states and never claims verification; explicit `app verify` replays isolated inputs against independent expected answers | `native_app.go`, `main.go`; `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, `TestPhase22AppVerifyIndependentIdentityCases`, `TestPhase22AppVerifyModelOnlyOutcomes` |
| Live local foreign allocation | A public file-byte path admits only exact acquire/use/release contracts; Schway holds the acquired byte until local release, and an independent physical observer plus negative controls prove malloc → use → matching free → exit | `examples/phase23/file_byte.schway`, `examples/phase23/adapter.c`, `internal/compiler/cgen/cgen_program.go`, `internal/compiler/native/phase23_observer_test.go`, `scripts/verify-phase23.sh`; the focused script passed in hosted Ubuntu/macOS evidence aggregates at `f991298b`; local macOS/Linux-container receipts at `c50430d` are historical |
| Bounded transfer and typed-error cleanup | The Phase 23 file-byte owner can move through the Phase 24 helper/return path; repeated calls retain distinct identities and the `0x43` typed-error path releases C,B,A exactly once | `examples/phase24/transfer.schway`, `examples/phase24/error.schway`, `internal/compiler/session/session_phase24_model_test.go`, `internal/compiler/native/phase24_observer_test.go`, `scripts/verify-phase24.sh`; hosted CI 36856048690 passes at `ed94ef79` on Linux/x86_64 and Darwin/arm64 |
| Bounded shared/exclusive pointer successors | The integrated utility calls separate shared and exclusive U64 read-copy helpers; independent straight-line validation rejects incompatible overlap and borrowed-result escape, and command escape diagnostics name the result and borrow origin; only plain pointer shapes are emitted | `examples/phase24/transfer.schway`, `internal/compiler/cgen/cgen_program.go`, `TestPhase25PointerPathLiveOverlap`, `TestPhase25PointerPathBorrowedResultEscape`, `TestPhase25IndependentPeerMutations`, and session source-attribution tests; local Go controls and hosted run 36971855722 (18/18 native rows) pass |
| Local modules (whole-program composition) | An entry imports named functions from other `.schway` files under its own directory; the loader resolves each file through one closed route (no search path; `..`, absolute paths and symlinks refused), merges the loaded modules into one program, and the merged program is checked, interpreted and emitted together. Imported functions carry their module in identity, in `unit=` diagnostics and in native C names. *Added 2026-10-12 (Phase 30).* Not separate compilation; not B1/DX-06. *Amended 2026-10-12 (Phase 31, MOD-06):* build identity, the receipt and the evidence report carry a sorted `modules` list (path and sha256 of every loaded module file) whenever the program imports, and a program with no import is byte-identical to before | `examples/checksum/adler32.schway`, `examples/checksum/adler32_cli.schway`, `TestPhase30AdlerModuleAppCLI`, `TestPhase30AdlerModuleInterpreter`, `TestPhase30LibraryIsSingleCopy`, `TestPhase30CheckMergedProgram`, `TestPhase30ImportCycles`, `TestPhase30SameBareNameTwoModules`, `examples/checksum/adler_low.schway`, `TestPhase31ModuleDigests`, `TestPhase31SingleCopyAndDualFlip`; hosted CI not yet run for Phase 30 or Phase 31 |
| Bounded file checksum and second consumer | `checksum` reads one file of at most 65,536 bytes through a copy-local file adapter and prints its Adler-32 (Wikipedia gives 300286872); `adler_low` imports the same single library and prints the low 16 bits (920). Files over the limit, missing, directory, FIFO, `/dev/null` and unreadable paths are refused with `FileError.FileTooLarge`, `FileError.FileUnavailable` or `FileError.AllocFailed`, exit 65 (`AppInputRefused`), empty stdout and no path in stderr. The shape is a third owner shape (FileBytes) accepted independently by check, corevalidate, originvalidate, pathoracle and cgen. *Added 2026-10-12 (Phase 31).* Whole-program composition only; hosted CI not yet run for Phase 31 | `examples/checksum/checksum.schway`, `examples/checksum/adler_low.schway`, `TestPhase31ChecksumPublicRoute`, `TestPhase31AdlerLowPublicRoute`, `TestPhase31CaseMatrix`, `TestPhase31ReleaseExactlyOnce`, `TestPhase31BoundedReadMutants`, `TestPhase31Controls`, `TestPhase31ControlInventory`, `TestPhase31OracleRows`; hosted CI not yet run for Phase 31 |
| Calls between Schway functions | Multi-function programs run; call-graph cycles remain refused | Phase 11 archive; `testdata/phase07/call_basic.schway`; `session.RunInterpreter` / native path |
| Returns independent of parameter type | Admitted through the production pipeline | Phase 17 archive and `session_phase17_test.go` |
| Computed `Result` matches and payload returns | Admitted; this is not unrestricted statement control flow | `session_phase18_payload_test.go`, Phase 18 archive |
| U64 arithmetic and scalar computation | Constants, checked addition, `<`, `==`, `%`, Bool branches, mutable U64/Bool state, and pre-tested scalar loops run; constant zero remainder is refused and dynamic zero fails with a checked app outcome | `examples/sum_to_n.schway`, `examples/fizzbuzz.schway`, `check/scalar.go`, `TestPhase26ExactSumMatrix`, `TestPhase27EqualityRemainder`, `TestPhase27StaticZero`, `TestPhase27DynamicZero` |
| Fixed text and bounded application output | Literal-only text writes combine with U64 decimal/LF writes; a 65,536-byte shared whole-write ceiling fails closed while evidence capacity stays separate | `internal/compiler/execution/execution.go`, `interp_phase27_output_test.go`, `cgen_phase27_output_test.go`, `TestPhase27OutputEvidenceSeparation`, `TestPhase27OutputEvidenceExactLimit` |
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
| Bounded M005 applications | `sum_to_n` and exact FizzBuzz run from ordinary source with checked equality/remainder, literal writes, and bounded output; general strings and byte/file IO remain unavailable | `examples/sum_to_n.schway`, `examples/fizzbuzz.schway`, Phase 26/27 focused controls; Phase 27's actual hosted macOS/Linux receipts and clean-checkout guide remain open |
| Resource/loan loop carries and cyclic path enumeration | Scalar-loop admission does not authorize ownership, resources, loans, or loan-derived provenance across back edges; the path oracle remains acyclic | Phase 26 category/peer mutation controls and `TestCompositionCycleGuardFailsClosed` |
| General strings, arrays, collections, reusable library packages | Ordinary JSON/HTTP libraries are not yet writable. Phase 30 local modules (whole-program composition) let one entry import functions from sibling files; there is no package or version story, no imported data or foreign declarations, no separate compilation; since Phase 31 loaded modules are hashed in build identity and evidence | current syntax/checker frontier; design wiki is prospective; `examples/checksum/README.md` |
| Recursion and broader resource control | Call cycles, nonlocal exits, cancellation, general fallible cleanup and escaped pointers require further contracts | callgraph and current foreign refusal boundaries |
| Entry selection | Entry is `main` first, then a unique in-degree-zero root, then a strictly larger closure; a genuinely ambiguous entry is refused with `core.entry_ambiguous` (exit 2) on check, build, run and app verify, and uncalled functions no longer block a run (`examples/phase28/unused_function.schway`) | `TestUnreachableFunctionProgramRuns`, `TestExitClassesUniformAcrossSurfaces`, `TestAppVerifySourceRefusalExitsInvalid` |

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

`examples/checksum.schway` was retired on 2026-10-12; see the PROJECT.md Phase 32
amendment. `wiki/example-tour.md` remains design exploration, not a supported
language specification.

## Corpus and guard census

Corpus: **162 `.schway` programs, 5,352 lines total** (~33 lines average,
193-line maximum). These counts match the current machine-checked tree; the
corpus predominantly contains focused semantic fixtures, not applications.

A non-test AST census finds **19 `len(Functions) != 1` guards across 6 files in 1 packages** (55 including tests):

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

1. **Finish Phase 27's exact FizzBuzz and cross-host evidence.** User-visible
   programs: ordinary `fizzbuzz.schway` and `sum_to_n.schway` produce exact
   bounded output. Current blocker: actual macOS and Linux native receipts at
   the same final revision, plus the linked clean-checkout guide. Smallest
   complete slice: implement the 27-09 host driver, receipt gate, and existing
   CI matrix wiring; then complete 27-10's guide and collect the actual hosted
   artifacts. Checker/guarantee changes: require two distinct hosts, matching
   source/fixture hashes, declared inputs, and successful jobs; unavailable or
   stale evidence stays open. The language checks already enforce static-zero
   refusal, dynamic checked failure, the shared output ceiling, and resource/
   loan loop refusals. Evidence/debt: Phase 26 U64-01/FLOW-01/FLOW-02/APP-07
   and Phase 27's focused local answer, mutation, output, and CLI checks are
   current; EVD-12 and DX-16 remain open. Owner/next action: execute Plans
   27-09 and 27-10; historical Phase 25/26 receipts do not count. Reprioritize
   by reacquiring both hosts if code, compiler, fixture, or CI inputs change
   after the claimed SHA, or leave EVD-12 open if either host is unavailable.
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

## Phase 30 transition amendment

**2026-10-12.** Phase 30 added local modules as whole-program composition: the
loader merges an entry and its imports into one program that is checked and
emitted as a unit. The refusal frontier is the `module.*` code family (for
example `module.not_found`, `module.path_mismatch`, `module.import_cycle`,
`module.not_exported`, `module.escape`, `module.symlink` and the cap codes),
and a refusal inside an imported file reports `unit=<module>` with a span in
that file. The runnable gain is the Phase 29 Adler-32 program with the library
in `examples/checksum/adler32.schway`, imported rather than copied.

Explicit non-claims: modules are not separate compilation, add no per-module
interface artifacts, do not close B1/DX-06, and are not hashed in build
identity or evidence until Phase 31 (MOD-06). Hosted CI for Phase 30 has not
run. Full check: [30-TRANSITION-CHECK.md](phases/30-local-modules/30-TRANSITION-CHECK.md).

## Phase 31 transition amendment

**2026-10-12.** Phase 31 added a bounded file checksum (`checksum`, limit 65,536
bytes, outcomes `FileTooLarge`, `FileUnavailable`, `AllocFailed` all exiting 65),
a second consumer (`adler_low`) importing the one `adler32` library, and module
hashing: build identity, the receipt and the evidence report carry sorted
`modules` rows when a program imports. The runnable gain is in
`examples/checksum/README.md`. Explicit non-claims: modules remain whole-program
composition, not separate compilation, and do not close B1/DX-06; the
`examples/checksum.schway` frontier was retired on 2026-10-12 (see the PROJECT.md
Phase 32 amendment); the M006 receipts and the guide belong to Phase 32; hosted
CI for Phase 31 has not run. Full check:
[31-TRANSITION-CHECK.md](phases/31-bounded-file-checksum-and-second-consumer/31-TRANSITION-CHECK.md).

## Refresh triggers

**2026-10-08 source-review amendment.** The [Dashbit AI-era source appraisal](sources/DASHBIT-AI-ERA.md) and relevant
[dormant seeds](seeds/) are conditional research inputs, not capability
receipts. At existing planning transitions, compare a triggered seed with a
named consumer, present source witness, and refusal boundary; record a scoped
disposition in the active planning artifact. No recurring article study or
background scheduler is implied.

Refresh after a capability lands, a refusal changes, a milestone closes, or a
selected user program exposes an incorrect claim. At each planning transition,
review the current three recommendations in PRODUCT-ROADMAP and connect changed
syntax/core operations to the necessary checker/evidence obligations. Keep
historical receipts scoped to their revision, host, compiler, and exercised
paths. AGENTS.md specifies this agent-executed review; no background automation
is implied.

## Phase 33 transition amendment — 2026-10-13

Phase 33 implements bounded capture ingestion, owner/reason-scoped interpreter defects, generic freeze/receipt gates, and grounded carried-debt dispositions. Focused current checks and reached mutants are recorded in its five summaries; the review follow-up refuses malformed control lists. Phase 33 binds no hosted receipt. M006 receipts are historical; current hosted evidence resumes in Phase 38.

The current three capabilities are: agent diagnostics plus `schway surface --json` (Phase 34), bitwise/literal-shift/hex computation for CRC-32 (Phase 35), and scalar helper calls used by the CRC-32 module (Phase 36; integrated consumer Phase 37). Their named programs, blockers, smallest slices, checker obligations, evidence owners and reprioritization observations are in [33-TRANSITION-CHECK.md](phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-TRANSITION-CHECK.md). Earlier recommendation sections remain historical; this dated amendment owns the current order.

All three conditional seeds remain dormant. The capture-allocation concern in the Dashbit source appraisal is resolved by `readApplicationCapture`, `TestPhase33CaptureReadBounded`, and the reached unbounded-read mutant. Current source search, capture, observers and replay supplied the diagnosis; no broader query, observability or source-form experiment is promoted.

## Phase 34 transition amendment — 2026-10-13

This evidence update distinguishes current source inspection and focused local execution from historical M006 receipts. It records the present runnable diagnostic and source-free discovery behavior, then retains Phase 35 and Phase 36/37 as planned capability ownership. Captured source may locate a diagnostic for presentation; build identity stays byte-based and repair remains authorized only by its exact byte edit. Human readability and end-to-end latency have not been measured.

| Program | Status | Blocker | Smallest complete slice | Checker/guarantee | Evidence/debt | Owner/next action | Reprioritization observation |
|---|---|---|---|---|---|---|---|
| Broken CRC/word-counter check and schway surface --json | implemented | None remaining for this bounded diagnostic and discovery slice | Keep the runnable broken-program check and source-free public surface discovery usable from one checkout; preserve refusal behavior when source is unavailable | Captured source supplies diagnostic locations; byte identity remains unchanged and repair authority remains the exact byte span and replacement, with cross-unit edits declined | DX-20; `internal/compiler/session/session_phase34_routes_test.go#TestPhase34SourceRoutePresentation`; `internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestWitnesses`; `internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCProgramGolden`; `internal/compiler/session/session.go#BuildApplicationCommandFile`; `cmd/schway/main.go#runSurface` | Phase 34 | Reconsider if an admitted source route loses its snapshot or an independent cause changes the diagnostic contract |
| CRC-32 operator expression x ^ (y << 3) & 0xFF and Adler b << 16 | planned | bitwise operators and shifts are absent | Add literal shifts and independent vectors for the named bitwise forms | Prove literal operands, range bounds, and safe shift width before emission; retain refusal for dynamic shifts | BIT-01; EVD-17; CRC-32 independent vectors | Phase 35 | Reprioritize if a named consumer needs dynamic shift or different bounds first |
| Imported crc_step scalar helper inside a loop | planned | scalar helper calls are absent | Add an arity-one imported helper and one file consumer with an independent answer | Require arity-1, acyclic call structure, and result-defect behavior across checker and execution peers | CALL-01; APP-13; MOD-07; independent imported-helper result and defect cases | Phase 36; file integration in Phase 37 | Reprioritize if a second consumer or arity-N call is required |

| Seed | Disposition | Current source witness | Named consumer | Trigger/reason | Next action |
|---|---|---|---|---|---|
| z1g | retain dormant | `internal/compiler/session/session_phase6_query.go#QueryCommandFile`; `internal/compiler/session/session_phase6_explain.go#ExplainCommandFile` | crc_step declaration and all direct call occurrences | measured named query gap | Let this seed remain dormant until a named reviewer task shows a measured gap in current query plus ordinary search; then scope one source-unit query |
| trz | retain dormant | `internal/compiler/native/native_app.go#readApplicationCapture`; `internal/compiler/native/native_phase33_capture_test.go#TestPhase33CaptureReadBounded`; `internal/compiler/session/session_app_verify.go#VerifyApplicationCasesFile` | one named checksum/file utility failure | failure unexplained by current capture, observers, and replay | The old allocation concern is historical and was resolved in Phase 33 by the bounded read and reached unbounded-read mutant; remain dormant until a named failure remains unexplained |
| z1s | retain dormant | `internal/compiler/syntax/format.go#Format` | CRC/word-counter canonical source review | named canonical-source or contract dispute | Let this seed remain dormant until a named consumer or reviewer identifies a concrete canonical-source or contract dispute |

| Kind | Status | Evidence/claim |
|---|---|---|
| Source inspection | inspected | Read `internal/compiler/session/session.go`, `cmd/schway-repair/repair.go`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, the three named seeds, and the Dashbit source appraisal; source inspection does not establish execution behavior |
| Focused execution | passed | Newly executed focused selectors: `TestPhase34SourceRoutePresentation`, `TestPhase34DiagnosticRouteInventory`, `TestPhase34RepairJSONCompatibility`, `TestPhase34CompatibilityDeltaLedger`, `TestPhase34PreChangeCProgramGolden`, `TestPhase34PreChangeCGoldenIsNotInert`, `TestPhase34BoolNameBaseline`, `TestPhase34IdentityBaseline`, `TestPhase34SurfaceManifestWitnesses`, `TestPhase16PublicEmitterConsumerInventory`, `TestPhase321RunInputRequiredCLI`, `TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity`, `TestPhase28PrunedEmitMatchesGolden`, and `TestEmittedCExitLiteralsMatchTable`; Phase 34 document controls are checked against the committed amendment |
| Historical M006 receipts | historical | M006 receipt records remain historical evidence scoped to their original revision and hosts |
| Phase 38 hosted binding | pending | Phase 38; not yet bound |
| Human readability | unmeasured | No human review; readability remains unmeasured |
| Whole-loop latency | unmeasured | No cold/warm whole-loop distributions; end-to-end latency remains unmeasured |

## Phase 34 completion review — 2026-10-13

The source-review follow-up corrected manifest route applicability and aggregate outer-body binding scope. The newly executed TestPhase34SurfaceLimitConsumers, TestPhase34InterfaceLimitWitnesses and TestPhase34LinearBindingLimitScope provide finite consumer, executable interface-threshold and direct AST-guard evidence; see the [review-fix record](phases/34-agent-diagnostics-floor-and-surface-manifest/34-REVIEW-FIX.md). Parser/checker admission and enforcement are unchanged. The next three runnable capabilities and retained dormant-seed dispositions are now recorded in [34-TRANSITION-CHECK.md](phases/34-agent-diagnostics-floor-and-surface-manifest/34-TRANSITION-CHECK.md): Phase 35 bitwise/CRC expressions, Phase 36 imported scalar helpers, and Phase 37 file CRC, second-consumer reuse and CLI interpreter parity. Final hosted binding remains Phase 38; human readability and whole-loop latency remain unmeasured.

## Phase 35 discussion review — 2026-10-13

At the pre-feature source inspection, hex is partially implemented rather than absent: generic number tokens retain spelling, lexer accepts `0x`/`0X`, scalar conversion accepts both, and legacy linear conversion recognizes only lowercase `0x`. Unprefixed `010` also has different current values by checker route; Phase 35 preserves that behavior as named debt instead of normalizing it. Bitwise operators, literal shifts and scalar grouping are still absent at this discussion baseline. No feature behavior tests were run here.

The [Phase 35 context](phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-CONTEXT.md) separates source literal admission, independently checked IR immediate semantics, static value transfers, engine literal results and causal event validation. `executionpeer.Validate` currently checks observed causal structure, not numeric replay; no broader guarantee is claimed. The three runnable successors and z1g/trz/z1s dormant dispositions remain scoped there. Phase 38 owns hosted final evidence; human readability and whole-loop latency are unmeasured.

## Phase 35 maturity transition amendment — 2026-10-14

Source inspection at `83c01058` confirms the real Adler consumer now uses the
bounded `b << 16` form with its final addition intact. Fresh local runs passed
the checksum and `adler_low` answer tables on both engines and reached a copied
wrong-shift control. The source-free operator guide is tied to current manifest
contracts; the public conformance observation `outcome.value == 264` remains
distinct from the U64 application output `264\n`, and the alternate grouping
returns 8 on both engines. These receipts establish only the named local paths.
They do not close phase-wide verification, security/code review, Phase 38 hosted
binding, human readability, or whole-loop cold/warm performance.

| User-visible program | Current blocker/status | Smallest complete slice | Checker/guarantee changes | Evidence/debt | Owner or next action | Observation that changes priority |
|---|---|---|---|---|---|---|
| Grouped bitwise example and imported Adler checksum consumers | Implemented and locally witnessed; phase closeout and independent review remain pending | `x ^ (y << 3) & 0xFF` with alternate-grouping control, plus one-line Adler shift through `checksum` and `adler_low` | U64 operators, literal shift bounds, low-64-bit/logical shift behavior, expression-depth guard, and unchanged Adler answers/cleanup | BIT-01..04, D-07/D-18/D-20, EVD-17; fresh `TestPhase35ExistingAdlerAnswers`, `TestPhase35AdlerWrongShiftControl`, `TestPhase35OperatorDocumentation`, `TestPhase35PublicDocumentValues`, public 264/8 and application 264 newline selectors | Finish plan evidence; parent verifier, security audit, and code review own phase acceptance; Phase 38 owns hosted binding | A named consumer needs dynamic shifts or a different bound |
| Imported `crc_step` helper in a checksum loop | Planned Phase 36; scalar-expression helper calls and independent import/call defect evidence are absent | One arity-one acyclic imported U64/Bool helper and a call from the loop consumer | Check import identity, call arity, argument/result types, acyclicity, and independently mutated call/result behavior | CALL-01/02 and Phase 36 helper call/defect evidence; integrated consumer belongs to Phase 37 | Phase 36 owner; then Phase 37 consumer | A named consumer requires arity-N or another call shape |
| File CRC-32, second importer, and interpreter CLI use | Planned Phase 37; helper and file-input CLI route are absent | Bounded ISO-HDLC CRC-32 file program, second importer, public interpreter CLI parity, and guide | Preserve bounded file handling and module composition; check independent vectors, failures, mutation, and second-consumer behavior | APP-13/14, MOD-07, DX-24; CRC vectors and wrong-result/file failure controls | Phase 37 owner; Phase 38 binds hosted evidence | A real file consumer needs a different bounded input or failure contract |

The conditional seeds z1g, trz, and z1s remain dormant. z1g's current query
witness is `internal/compiler/session/session_phase6_query.go#QueryCommandFile`;
there is no measured navigation gap for the `crc_step` consumer. trz's witnesses
are bounded application capture (`readApplicationCapture` and
`TestPhase33CaptureReadBounded`) plus application verification; no unexplained
checksum failure exists, and the older allocation concern was resolved in
Phase 33. z1s's witness is `internal/compiler/syntax/format.go#Format`; no
canonical-source dispute is named. Revisit each only on its named trigger.
The Dashbit source appraisal remains contextual and supplies no evidence for a
generic database, telemetry, or source-query expansion.

Current execution receipts were produced locally on macOS/arm64 with Go
1.24.0 and Apple Clang 21.0.1 from checkout revision `26380801`. Historical
Phase 35-07 evidence remains scoped to revision `64dfec6f`: 33 fixture rows ran
at O0/O3/O3-LTO, and all 16 shift rows ran UBSan. It is not a fresh current
receipt. Phase 38 hosted evidence is pending; readability and complete
generate/verify/run/observe/repair cold/warm distributions remain unmeasured.

## Phase 35 closeout transition review — 2026-10-14

The implementation and independent repair work now support the bitwise/Adler slice. Phase acceptance is owned by [35-VERIFICATION.md](phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VERIFICATION.md), rather than the earlier implementation-only pending notes. The [transition check](phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-TRANSITION-CHECK.md) records current source witnesses, the 94-name regression refresh, the target-freshness repair, 26 admission mutation controls and all evidence boundaries.

The next three scoped capabilities are: Phase 36 imported `crc_step` calls in a loop; Phase 37 bounded file CRC-32 with a second importer and interpreter CLI parity; Phase 38 reproducible two-host checksum/CRC receipts. The linked transition table names each program, blocker, smallest slice, checker obligations, evidence/debt, owner and priority-changing observation. All three conditional seeds remain dormant on their named source witnesses. Historical M006 receipts remain historical; final hosted binding, human readability and whole-loop cold/warm distributions remain pending or unmeasured.

Phase 35 final acceptance update (2026-10-14): canonical verification now passes 38/38 unique truths; all 33 vectors and 16 UBSan shift rows were freshly rerun locally, together with 15 arithmetic and 26 admission mutation controls. The earlier implementation-only pending notes remain historical. Security/review are closed; final two-host receipts remain Phase 38-owned.

## Phase 36 discussion transition — 2026-10-14

Source inspection for the [Phase 36 context](phases/36-scalar-helper-calls/36-CONTEXT.md)
distinguishes existing general direct-call machinery from missing scalar call
expressions and Bool signatures. U64/Bool scalar helper results will be
initialized but value-unknown to independent static analyses. Existing
InvocationSegment.Ordinal supports unique repeated activations without a new
wire field, but current producers/peer hard-code zero and repeated-call/event
capacity support still needs implementation. These are scoped obligations,
not demonstrated new capability.

An exclusively U64/Bool callee cannot read bytes. CALL-01's bounds row therefore
retains the existing owning-wrapper/ByteView callee regression and separately
checks a bounds fault while computing a new helper argument. The causal peer
receives no scalar values; numerical argument/result/defect claims require
independent engine answers. See the dated requirements clarification and
36-CONTEXT D-18. No Go/native behavior check ran during discussion.

The context's transition table records the next three capabilities: imported
helper in a loop (36), file CRC/second consumer/public interpreter CLI and guide
(37), final two-host revision binding after audit (38). Each has a named user
program, blocker, complete slice, guarantee/evidence owner and reprioritization
observation. All three conditional seeds remain dormant on named witnesses;
Phase 33 resolved the historical capture-allocation concern. Hosted receipts
remain historical/pending as previously labelled. Human readability and whole
feedback-loop cold/warm distributions remain unmeasured.

## Phase 36 completion — 2026-10-15

Scalar helper composition now has a runnable local/importer witness and bounded fresh evidence for source lowering, independent raw call contracts, engine values, causal attribution, ByteView bounds, capture outcomes, and the source-free surface manifest. The [Phase 36 acceptance receipt](.planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md) records the exact owner commands and logs. This is implementation and focused-check evidence; CALL-01/CALL-02 remain unclassified/unresolved, the descriptorless prohibitions remain unverified, and independent phase/security/review gates remain open. Human readability and full-loop cold/warm performance remain unmeasured.

Source inspection: The independent call transfers and refusal boundaries are in `internal/compiler/check/scalar.go`, `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/originvalidate/originvalidate.go`, and `internal/compiler/pathoracle/pathoracle.go`; repeated values and call ancestry live in `internal/compiler/interp/interp.go` and `internal/compiler/executionpeer/executionpeer.go`; the runnable local/importer programs are `examples/helper_calls/crc_step.schway` and `examples/helper_calls/importer.schway`.

New focused checks: The fixed final capture records `internal/compiler/session/session_phase36_calls_test.go#TestPhase36ByteViewCallerArgumentBoundary`, `internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeByteViewCallerArgumentBoundary`, and `internal/compiler/session/session_phase36_surface_test.go#TestPhase36SurfaceManifestCallWitnesses` as fresh owner results, with the remaining claim mappings in the linked receipt.

Historical receipts: Historical only; not fresh Phase 36 evidence. [Phase 35 verification](.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VERIFICATION.md) passed on 2026-10-14 for its recorded source revision and host.

| Phase | Program | Blocker | Smallest slice | Guarantee/evidence | Owner/next action | Priority changes if |
|---|---|---|---|---|---|---|
| Phase 36 | Repeated `crc_step` helper calls through a local module, an importer, and a public application route | The implementation slice is present; phase verifier, security and code review are pending, and requirement edges have no SPEC classification | Finish independent acceptance of the bounded arity-one U64/Bool composition already implemented | Source lowering, independent raw/origin call contracts, interpreter/native values and defects, causal-only attribution, both bounds obligations, capture outcome, manifest census, and frozen C have separate receipt rows | Phase 36 verifier/security/review owners close their gates; keep CALL requirements pending until then | A named program requires another argument authority or call shape, or the exact independent answer/mutant evidence disagrees |
| Phase 37 | Bounded file CRC-32, second importer, interpreter CLI, and guide | No integrated file consumer, second importer, or interpreter file route is available yet | Use the current helper in one complete bounded file program and a second module consumer, with CLI parity and a runnable guide | Add literal independent CRC vectors, file-boundary/failure controls, module-identity checks, and reached wrong-result mutations across the applicable engines | Phase 37 owner plans and executes the complete consumer after Phase 36 audit handoff | A named file task needs a different size limit, input route, or error contract |
| Phase 38 | Final checksum/CRC evidence on Ubuntu and macOS | No final M007 hosted receipt is bound to the audited source revision | Run the agreed bounded acceptance on both hosts after preceding audits and bind exact logs to the same source and fixture revision | Require fresh non-skipped roots, consistent independent answers, and preserved app outcome on each host; do not transfer Phase 35 historical results | Phase 38 owner records hosted evidence after verifier/security/review handoff | Any host mismatch, revision drift, or platform-specific outcome changes the guarantee |

| Seed | Disposition | Source witness | Trigger observation | Next action |
|---|---|---|---|---|
| z1g | retain dormant | `internal/compiler/session/session_phase6_query.go#QueryCommandFile` and ordinary source search | The helper name and callsites are findable; no named task has measured an accuracy or material-cost gap | Keep dormant until a reviewer records a concrete lookup gap, then test one source-unit query |
| trz | retain dormant | `internal/compiler/native/native_app.go#readApplicationCapture`; `internal/compiler/session/session_app_verify.go#VerifyApplicationCasesFile` | The current bounded capture and replay explain the observed cases; Phase 33 resolved the old allocation concern, and no new unexplained failure is named | Keep dormant until one named program failure remains unexplained by current capture, observers, debug map, and replay |
| z1s | retain dormant | `internal/compiler/syntax/format.go#Format` | No named consumer or reviewer disputes a concrete source form or contract | Keep dormant until a real dispute exists, then compare paired forms against the same independent obligations |
