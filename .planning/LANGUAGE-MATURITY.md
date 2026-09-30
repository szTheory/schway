# Schway Language Maturity — Current Evidence

**Re-assessed:** 2026-09-30, against the current source and Phases 22–23
artifacts. Phase 22's refreshed report passes 5/5 truths and its completed
objective README UAT remains preserved. Hosted run 36707529870 passed Ubuntu
and macOS check suites plus both current evidence aggregates, including the
Phase 23 script, at source SHA `f991298b29779838a2b1a5c3cd5ac90aafcb84fc`.
Phase 23's refreshed 2026-09-30 verifier passes 5/5 truths, binds that hosted
receipt, and marks the phase complete. Its seven plans are complete; the phase
context requires no human UAT. No local project suites were run during the
refresh. The Phase 23 security review covers 18/18 threats. Objective checks
cover acceptance; subjective readability is not claimed. Future direction lives
in [PRODUCT-ROADMAP.md](PRODUCT-ROADMAP.md).

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
| Schway foreign operations and pointer parameters | Phase 23 admits only its exact acquire/use/release operation contracts; other foreign-call shapes and both shared/exclusive by-pointer families remain structurally refused | `cgen_program.go:emitProgram`, `TestPhase23OperationContract`, Phase 23 source/emitter refusal tests |
| Host build closure | Known source/header/compiler/target inputs are identity-bound, but the SDK/linker/runtime closure is incomplete; app artifacts are not cacheable | `native/bindings.go`, `examples/phase22/BINDINGS.md`, Phase 22 build receipts |
| Schway-owned physical cleanup | A local owner now stays live through use and is physically released before function exit; ownership transfer across calls, typed errors, and general cleanup remain unadmitted. The C adapter owns file descriptors and frees partial buffers on acquisition failures | `examples/phase23/adapter.c`, `internal/compiler/native/native_app_test.go`, `native/phase23_observer_test.go`; Phase 24 owns transfer and error cleanup |
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
suite was rerun for the refresh. Transfer through calls, typed errors, and
shared/exclusive pointer families remain separate successors.

`examples/checksum.schway` is a refused integration target with provisional syntax.
Its Phase 20 test pins the `loop` refusal; moving that diagnostic does not by
itself demonstrate a checksum, file IO, or output. `wiki/example-tour.md` remains
design exploration, not a supported language specification.

## Corpus and guard census

Corpus: **146 `.schway` programs, 4,809 lines total** (~33 lines average,
193-line maximum). These counts match the current machine-checked tree; the
corpus predominantly contains focused semantic fixtures, not applications.

A non-test AST census finds **19 `len(Functions) != 1` guards across 6 files in 1 packages** (50 including tests):

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

1. **Phase 24 — transfer and typed errors.** User-visible program: the
   file-byte reader passes its live owner to a helper, consumes its result, and
   releases exactly once after normal return and a real post-acquisition typed
   error. Current evidence: Phase 23 is complete at 5/5 with passing hosted
   Ubuntu/macOS aggregates, an independent physical observer, and reached
   cleanup controls. Blocker: resource identity is still local to a
   straight-line activation; transfer and typed-error cleanup across frames
   remain refused. Smallest complete slice: one caller/callee transfer, a
   distinct dynamic activation, reverse-order generated cleanup, and one real
   typed error after acquisition with independent observer and refusal
   controls. Checker/guarantee changes: independently derive transferred
   owner facts at admission, peer validation, interpretation, and emission;
   reject double/wrong release and unsupported exits. Evidence/debt:
   RES-05/06, OWN-10/11/12, EVD-09, extending EVD-11. Owner/next action: Phase
   24; its context is absent, so run `$gsd-discuss-phase 24` before planning.
   Reprioritize if the Phase 23 contract cannot witness the call boundary or a
   concrete consumer proves pointer access is a prerequisite.
2. **Phase 25 — bounded pointer families and utility.** User-visible program:
   a documented byte utility with distinct shared and exclusive read-copy
   helpers. Blocker: both families remain refused by emission and lack
   family-specific pointer contracts. Smallest complete slice: one bounded
   helper per family, independent borrow/access checks, positive and
   conflict/escape controls, and a reproducible utility command. Checker and
   guarantee changes: family-specific admission and peer checks plus
   conservative interpreter/C behavior, without unsupported alias attributes.
   Evidence/debt: NAT-11/12/13, EVD-10 and DX-14/15. Owner: Phase 25 after
   Phase 24. Reprioritize if a Phase 24 witness requires one pointer family to
   prove ownership transfer or cleanup.
3. **Next milestone — ordinary scalar computation.** User-visible program:
   `sum_to_n`, then FizzBuzz from ordinary Schway source with exact output and
   boundary behavior. Blocker: defined U64 arithmetic/remainder and overflow,
   comparison/Bool, scalar-loop fixed points, and bounded text/decimal output
   do not exist. Smallest complete slice: one bounded-input `sum_to_n` command
   with exact expected output, arithmetic boundaries, and a loop-carried-state
   mutation control, then use that core for FizzBuzz. Checker/guarantee
   changes: typed operators with defined C behavior, overflow/error semantics,
   CFG fixed-point state/loan analysis, and bounded output effects.
   Evidence/debt: NAT-09 and current arithmetic/CFG refusal witnesses; assign
   the milestone's remaining requirement IDs at kickoff. Owner: post-M004
   milestone. Reprioritize only if a selected consumer demonstrates that
   computation is more valuable than completing the committed resource scope.

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
