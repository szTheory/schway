# Language Maturity — Where Codename Lang Actually Stands

**Re-assessed:** 2026-09-27, after Phase 22 implementation, against the current
source and Phase 22 artifacts. Source inspection, named tests, the full Go test
suite, and the regenerated goal-verification report were checked. It confirms
17/17 truths with `passed` status after the objective README contract UAT.
Subjective readability is not claimed. Evidence is macOS-only; this does not
claim a Linux run. Future direction lives
in [PRODUCT-ROADMAP.md](PRODUCT-ROADMAP.md).

### Documentation-gate amendment — 2026-09-27

Source inspection found the existing Phase 22 README covers its build, run,
manifest, evidence, replay, trust, and closure boundaries but lacked an explicit
same-run event-evidence command. The focused `TestPhase22READMEContract` and
named CLI checks passed on this macOS host after adding that example; no CI run
or Linux execution is claimed here. These are newly executed documentation and
CLI checks, distinct from the earlier Phase 22 verifier receipt. The regenerated
report at `2026-09-27T23:05:37Z` passes with 17/17 truths and the objective README
contract UAT. The contract test does not establish subjective readability. The three capability
recommendations and Phase 23/24/25 ordering below remain unchanged.

## What exists

| Capability | Observed boundary | Source / evidence anchor |
|---|---|---|
| Source → checked core → interpreter/native C17 | A working compiler exists; Go 1.24/Clang remain the development path | `cmd/lang/main.go`, `internal/compiler/session/session.go` |
| Retained native application build/run | `lang app build` creates a retained artifact from an admitted source and closed local-C manifest; `lang app run` accepts bounded U64 input, launches once, and preserves ordinary streams/outcomes | `cmd/lang/main.go`, `session.go`, `internal/compiler/native/native_app.go`; Phase 22 `TestPhase22IdentityApplicationBuildAndRunCLI`, `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes` |
| Separate application evidence | Capture is a distinct report with disabled/incomplete/complete/capacity states and never claims verification; explicit `app verify` replays isolated inputs against independent expected answers | `native_app.go`, `main.go`; `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, `TestPhase22AppVerifyIndependentIdentityCases`, `TestPhase22AppVerifyModelOnlyOutcomes` |
| Lang-to-Lang calls | Multi-function programs run; call-graph cycles remain refused | Phase 11 archive; `testdata/phase07/call_basic.lang`; `session.RunInterpreter` / native path |
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
| General caller input and IO | The app route currently accepts one bounded U64 token; it does not provide general file, byte-string, or network IO | `cmd/lang/main.go:runApplicationRun`, Phase 22 CLI tests |
| Legacy differential run route | `lang run --engine=native` remains a conformance harness with synthesized inputs and O0/O3 comparison; effectful application code must use `lang app run` | `session.go:runNative`, `cmd/lang/main.go` |
| Lang foreign operations and pointer parameters | Local C can be declared and linked, but Lang foreign-call and both by-pointer families remain structurally refused by the sole emitter | `cgen_program.go:emitProgram`, `TestPhase22AppVerifyRejectsOrdinaryForeignScriptsAndLocalC`, Phase 22 verification |
| Host build closure | Known source/header/compiler/target inputs are identity-bound, but the SDK/linker/runtime closure is incomplete; app artifacts are not cacheable | `native/bindings.go`, `examples/phase22/BINDINGS.md`, Phase 22 build receipts |
| Lang-owned physical cleanup | The old C resource shim allocates and frees before returning a byte; current tracking cannot establish a live resource crossing Lang calls | `native/lang_foreign_resource.c`, M004 architecture research |
| Arithmetic/comparison and scalar iteration | Cannot add, compute remainder, loop, or write FizzBuzz | syntax/core operation inventory; `pathoracle` rejects CFG cycles |
| General strings, arrays, collections, usable library modules | Ordinary JSON/HTTP libraries are not yet writable | current syntax/checker frontier; design wiki is prospective |
| Recursion and broader resource control | Call cycles, nonlocal exits, cancellation, general fallible cleanup and escaped pointers require further contracts | callgraph and current foreign refusal boundaries |

There is no currently admitted public foreign-C hello-world path. Phase 22 can
compile/link declared local C, but the identity program does not call it; a
named foreign declaration or historical standalone C probe is not native
source admission. Phase 22 completed application execution/evidence separation.
Phase 23 must now prove that an allocation returned live from C is owned, used,
and physically discharged by Lang.

`examples/checksum.lang` is a refused integration target with provisional syntax.
Its Phase 20 test pins the `loop` refusal; moving that diagnostic does not by
itself demonstrate a checksum, file IO, or output. `wiki/example-tour.md` remains
design exploration, not a supported language specification.

## Corpus and guard census

Corpus: **143 `.lang` programs, 4,670 lines total** (~33 lines average,
193-line maximum). These counts are retained from the machine-checked snapshot;
the corpus predominantly contains focused semantic fixtures, not applications.

A non-test AST census finds **19 `len(Functions) != 1` guards across 6 files in 1 packages** (49 including tests):

| Package | Guards | Notable sites |
|---|---|---|
| `session` | 19 | Bounded historical verification/reducer lanes; inspect each site's purpose rather than inferring that public multi-function run is refused |

`internal/compiler/session/self_describing_docs_test.go` independently derives
these numbers with Go's parser. The count describes a syntactic predicate; it
cannot determine public capability. Multi-function production execution and
historical single-function evidence lanes coexist. The reducer also supports
multi-function seeds; its `<= 1` and `== 1` shortcuts are outside this census.

Re-verify (approximate only — Go AST evidence is authoritative):
`rg 'len\([^)]*Functions\) != 1' internal cmd`.

Machine check: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestLanguageMaturityCountsAreCurrent|TestSelfDescribingDocsGuardIsNotInert)$' -count=1`.

## Next useful thresholds

1. **Phase 23 — live file-byte allocation:** the `0x41`/`0x42` file consumer
   must hold a bounded malloc-backed value after acquisition, use it through
   Lang, and prove generated destruction with an observer independent of
   compiler events. Blockers are the refused foreign call, missing acquisition-
   derived obligation, and absent physical destructor witness. Trigger changes
   in `check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, and
   `cgen`; cover FFI-03, RES-04/07/08/09 and the Phase 23 EVD-09 slice. Owner:
   Phase 23; its 2026-09-27 research resolution fixes source spellings, the C17 ABI, and the negative-control design for execution, without claiming the capability is implemented.
2. **Phase 24 — transfer and typed errors:** pass that live owner through a
   helper, then prove exactly-once cleanup on normal and actual post-acquisition
   error paths. Blockers are activation-specific resource identity and
   cross-frame cleanup. Trigger the same independent checker peers and emitted
   cleanup; evidence covers RES-05/06, OWN-10/11/12 and EVD-09. Owner: Phase 24
   after Phase 23 establishes the local owner contract.
3. **Phase 25 — bounded pointer families and utility:** admit distinct shared
   and exclusive read-copy helpers, each with positive, conflict/escape
   controls and macOS/Linux evidence, then document the complete byte utility.
   The emitter currently refuses these shapes. Trigger family-specific borrow
   checks and conservative C lowering with no unproved alias attributes;
   evidence covers NAT-11/12/13, EVD-10 and DX-14/15. Owner: Phase 25.

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
