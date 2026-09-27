# Language Maturity — Where Codename Lang Actually Stands

**Re-assessed:** 2026-09-27, M004 kickoff, against source at `d9bde05`.
This is a current capability snapshot. Inspected test code and archived
verification identify the evidence boundary; this review does not claim to
have rerun every implementation suite. Future direction lives in
[PRODUCT-ROADMAP.md](PRODUCT-ROADMAP.md).

## What exists

| Capability | Observed boundary | Source / evidence anchor |
|---|---|---|
| Source → checked core → interpreter/native C17 | A working compiler exists; Go 1.24/Clang remain the development path | `cmd/lang/main.go`, `internal/compiler/session/session.go` |
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
| Caller-selected public inputs / application mode | Public `run` synthesizes Byte/Buffer inputs; lower-layer U64 input support is not a complete CLI feature | `session.go:interpreterInputs`, `cmd/lang/main.go` |
| Single application execution | Native `run` interprets, then runs O0 and O3; real side effects would be duplicated | `session.go:runNative` |
| Ordinary stdout/stderr and retained executable | Native runner expects execution JSON, rejects successful stderr, and deletes temporary executable | `native/native.go`, `cgen/cgen_program.go` |
| Real native foreign resources | All three foreign/by-pointer families remain refused by `emitProgram` | `cgen_program.go:emitProgram` and Phase 16 public refusal tests |
| Lang-owned physical cleanup | The old C resource shim allocates and frees before returning a byte; current tracking cannot establish a live resource crossing Lang calls | `native/lang_foreign_resource.c`, M004 architecture research |
| Arithmetic/comparison and scalar iteration | Cannot add, compute remainder, loop, or write FizzBuzz | syntax/core operation inventory; `pathoracle` rejects CFG cycles |
| General strings, arrays, collections, usable library modules | Ordinary JSON/HTTP libraries are not yet writable | current syntax/checker frontier; design wiki is prospective |
| Recursion and broader resource control | Call cycles, nonlocal exits, cancellation, general fallible cleanup and escaped pointers require further contracts | callgraph and current foreign refusal boundaries |

There is no currently admitted public foreign-C hello-world path. A named
foreign declaration or historical standalone C probe is not native source
admission. M004 starts by separating application execution from evidence replay,
then proving an allocation returned live from C is owned and discharged by Lang.

`examples/checksum.lang` is a refused integration target with provisional syntax.
Its Phase 20 test pins the `loop` refusal; moving that diagnostic does not by
itself demonstrate a checksum, file IO, or output. `wiki/example-tour.md` remains
design exploration, not a supported language specification.

## Corpus and guard census

Corpus: **142 `.lang` programs, 4,661 lines total** (~33 lines average,
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

1. **M004:** a public build/run command takes actual input, executes once, and
   proves Lang's ownership of a live foreign allocation across use, calls,
   errors, and cleanup. Shared/exclusive read-copy pointers have separate
   admission/evidence boundaries without additional alias promises.
2. **Following milestone:** defined arithmetic and comparisons, scalar loops,
   fixed text and decimal output make `sum_to_n` and FizzBuzz possible. A loop
   design spike must address fixed points, dynamic event identity, and bounded
   evidence before implementation; a full String runtime is unnecessary.
3. **Reusable libraries:** byte views/indexing and small data/API/module features
   support file utilities and bounded JSON. HTTP is an independent branch with
   explicit protocol, timeout and resource requirements.

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
