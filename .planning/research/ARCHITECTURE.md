# Architecture Research: M006 Bounded Byte Utility and Local Modules

**Domain:** Ownership-checked native language toolchain (Go 1.24 stdlib host, C17 via Clang)
**Researched:** 2026-10-10
**Mode:** Ecosystem/integration (subsequent milestone; existing architecture is the subject)
**Confidence:** HIGH for statements about current code (read directly, several behaviors re-run
on a scratch build of `cmd/schway`); MEDIUM for external-language analogies (Rust NLL, Zig,
Go, OCaml) which come from prior knowledge and were not re-fetched; LOW only where marked
"spike first".

Evidence convention: `[code]` = read in this repo on 2026-10-10; `[ran]` = executed against a
scratch build; `[analogy]` = external language lesson, not re-verified.

---

## 1. What exists today (the facts the design must respect)

### 1.1 Pipeline as built

```
 source bytes (ONE file)
   -> syntax.Parse            (lossless tree + ast.Program; one `module` per file)
   -> check.Program           (per-function dispatch; ONE program.Module; semanticID "s1:<module>:fn:<name>")
        scalar CFG path  : checkScalarFunction (scalar.go)  U64 param, U64|Unit return only
        linear/owner path: checkLinear / checkFallibleLinear / fixture-shaped Phase 23/24/25 matchers
   -> core.Program            (Functions[], ONE Module/ModuleID, Schema lang.core/1)
   -> corevalidate.Validate   (peer 1; own scalar fixed point, own classifier hasScalarCFG)
   -> originvalidate.ValidatePublished (peer 2; own originScalarFixedPoint, own originHasScalarCFG)
   -> pathoracle.ValidateLocalOwnerPaths (peer 3; ACYCLIC only; fixture-shaped owner replay)
   -> interp (oracle)  |  cgen.EmitApplication / Phase16ControlNativeC  -> Clang -> native runner
   session orchestrates; cmd/schway maps protocol.Status to exit 0/2/3/4/64
```

### 1.2 Load-bearing facts discovered while reading `[code]`/`[ran]`

| # | Fact | Consequence for M006 |
|---|------|----------------------|
| F1 | Scalar CFG functions are a *separate path*: `checkScalarFunction` requires `Parameter.Type == "U64"`, return `U64|Unit`; every place is U64/Bool; any `OpMove/OpBorrow*/OpForeignCall/OpRelease` or non-scalar `OpCopy` reached by the fixed point is refused as `check.cfg_back_edge` with a category (`scalarAuthorityCategory`). | A byte loop must enter this path; the refusal for owner/loan/resource carries stays intact for everything except one new, provably constant place class. |
| F2 | The scalar grammar has **no calls and no parentheses**. A scalar-body function cannot call anything (R1b), but a *linear* caller can call a scalar-CFG callee today `[ran]` (`let r = total(n) r` builds and `app run` prints 55). | Library function `adler32` is a scalar-CFG callee; its callers are linear. No scalar-to-scalar calls are needed. |
| F3 | Four independent classifiers decide "is this a scalar CFG function": `check` (statement presence), `corevalidate.hasScalarCFG`, `originvalidate.originHasScalarCFG`, `cgen.hasScalarOperations`. They are copied on purpose (D-27-07) and already diverge by design (D-27.2-04). | Every new scalar op kind must be added to all four, plus the dispatch control `TestAllOperationKindsHandledAtEverySite` fixtures. Drift = silent misclassification. |
| F4 | Scalar C emission is `emitProgramScalarFunction`: all places become `uint64_t`/`int` locals; `OpCopy` records a `value.copied` event **per execution**; `OpAddChecked/Less/Equal/Remainder/ScalarStore` record nothing. Event array capacity is static (`invocationEventCapacity` = operation count per activation); overflow exits with the named `native.event_capacity_exhausted` (exit 69). Shells differ: application shell prints a stderr line and `exit(65)`; conformance shell emits a `defect` terminal document and `exit(0)`. | Byte ops must record **no per-iteration events**, and the checksum loop must use `var` + assignment (OpScalarStore), never `let` copies, or a 64 KiB file exhausts capacity exactly as `sum_to_n` did at n>=15. Dynamic OOB mirrors checked remainder: defect in conformance, `exit(65)` + stderr in the app shell. |
| F5 | The **owner/file path is fixture-shaped across five packages**: check (`check.go` ~3690-4090), pathoracle (`ValidateLocalOwnerPaths` hardcodes `schway_file_byte_acquire/use/release`, mode sequence acquire,borrow,consume), cgen (`emitProgramLocalFileByteFunction`, transfer emitters), interp (`RunWithForeignOutcomes`, model-only, `function.Name == "main"`), native (adapter). Name-keyed matchers (`functionNames["acquire"]`, `function.Name == "main"`) exist. | A new file-bytes owner is a new parallel shape, not a parameter of the old one; and name-keyed matchers will break under module qualification. Build modules *before* the owner shape (see build order). |
| F6 | `session.Check`, `RunInterpreter`, `BuildApplication`, `VerifyApplicationCases` take `[]byte`. Only `*File` wrappers touch the filesystem (`readBoundedFile`). `publishedOriginProblemFile(path)` **re-reads** the file after `Check` already read it. | A resolver can live behind the `*File` wrappers with an in-memory `SourceSet` below them; the re-read must go (TOCTOU multiplies with N files). |
| F7 | `diagnostic.Span` is `{Start,End}` byte offsets with **no file identity**. `debugmap` and `evidence` carry spans as data. | Multi-file needs attribution that does not perturb existing bytes (see D3). |
| F8 | `callgraph.EntryFunction` requires exactly one in-degree-zero root. A single file with an unused function already fails `[ran]`: `check` passes, `run` reports `tool.run_failed` (exit 3) and `build` reports `native.build_failed ... callgraph.entry_ambiguous` (exit 3). | A reusable library necessarily has functions the entry does not call. A link step with explicit entry plus reachability pruning is mandatory, and `entry_ambiguous` is a *source-attributable* refusal that belongs in exit class 2. |
| F9 | `interface export/core/check`, `core.Interface`, `FunctionSignature` (Parameters, Return, Callable, Foreign, Merkle `ClosureDigest`) and `callSignatureTable` already implement a body-blind cross-function contract; `ParameterContract.Mode` already admits `"shared"|"exclusive"` but every producer emits `"owned"`. | The cross-module call admission should consume the signature table, never a callee body; `shared` becomes live for the ByteView parameter. This is the separate-compilation seam (D-13-02b) and it already exists. |
| F10 | Dynamic failures in the app shell are uniformly `exit(65)` plus one `schway ...` stderr line (U64 overflow, remainder by zero, defect, typed errors like `schway_file_byte_acquire: EmptyFile`). `app run` adds 65 for oversized/NUL input, 124 timeout, 125 launch, `128+N` signal. | New byte errors reuse 65 plus distinct stderr text; do not introduce sysexits 66 now. |
| F11 | `pathoracle` is acyclic by design (T-03-11); spike 002 showed the *liveness fixpoint* handles loop-carried loans (live through the back edge, end on the loop-exit edge) but the independent peer cannot corroborate that. | Admitting loop-carried loans would ship a producer with no independent peer. Do not. |
| F12 | `examples/checksum.schway` is pinned by `TestPhase20ChecksumFrontier` (digest + first diagnostic `syntax.expected_rbrace` at 521-527). `loop buffer { next(buffer) }` is provisional syntax. | Leave the file frozen as the historical refused witness; build the real program under a new path. Do not reuse the iterator-style syntax. |

---

## 2. Recommended architecture

### 2.1 One-paragraph summary

**Do not let any loan cross a back edge. Make the loop function's buffer a read-only
`ByteView` *parameter* that is constant by construction, so the loan lives in the caller,
straight-line, and ends at the call.** Add two scalar read operations (`len`, bounds-checked
index) at all six consumers with dynamic-only failure. Add local modules as a producer-side
loader that merges all modules into one checked `core.Program` (module-qualified IDs,
explicit entry, reachability pruning) so every peer still sees a single whole program, while
cross-module call admission goes through the existing body-blind signature table so
separate compilation is not foreclosed. Land the file-reading owner path last, as a new
parallel shape on top of the module-aware pipeline.

### 2.2 Target data flow

```
 root.schway ──┐
 (dir of entry) │   modload (session, IO)         check.Link (pure)             peers (unchanged shape)
 lib.schway  ───┼─> SourceSet{ModuleName->bytes,──> merged ast units ──> core.Program ──> corevalidate
 other.schway ──┘   digest}   cap: 16 modules       per-decl Module       Functions[] have          originvalidate
                    cycle/escape/case checks        export table          Module, Exported;         pathoracle (caller shapes)
                    single read, no re-read         signature-only        Program.Modules[] digests      │
                                                    imports                                              v
                                          callgraph.Prune(entry) ──> linked core.Program ──> interp | cgen -> clang
                                          (explicit entry = root `main`)                       |
                                                                       evidence/cache/build-id bind SourceSet digest
```

### 2.3 Component boundaries (new vs modified)

**NEW**

| Component | Location | Responsibility | Talks to |
|-----------|----------|----------------|----------|
| `OpBytesLen`, `OpBytesAt` | `core/core.go` (+ `AllOperationKinds`) | Read-only scalar ops on a ByteView place | all six consumers |
| `ByteView` type shape | `ability` (+ `check`, `corevalidate`, `originvalidate` tables) | Non-owning, non-copyable, non-escaping, droppable view; no obligations | scalar fixed points |
| `ImportDecl` + `len(x)` / `x[i]` expr kinds | `ast`, `syntax` (lexer `[` `]`, `TokenImport`, parser, `format.go`) | Surface for modules and byte reads | check |
| `modload` (resolver) | `session/modload.go` (IO) + small pure graph helper | Resolve import -> file, cap, cycle, case, symlink checks; produce `SourceSet` | session only |
| `check.Link` | `check/` | Check N units as one program; per-declaration module; export/import tables; signature-only cross-module contracts | modload, check internals |
| `callgraph.Prune(program, entryID)` | `callgraph/` | Reachability closure from explicit entry | session (pre-peer for build/run) |
| `core.Function.Module`, `Exported`; `core.Program.Modules[]` | `core/core.go` (additive omitempty) | Independent re-derivation anchors for visibility and set identity | corevalidate, originvalidate, evidence |
| `Diagnostic.Unit` (omitempty) | `diagnostic/` | Which module a diagnostic's file-local span belongs to | protocol, debugmap |
| `FileBytes` owner shape + `bytes_adapter.c` + binding manifest | `check`, `pathoracle`, `cgen`, `interp`, `examples/checksum/` | Phase-31 owner path: acquire -> borrow ByteView -> call -> release | native bindings |
| `verify-phase3x.sh` + CI lane | `scripts/`, `.github/workflows/ci.yml` | Dual-host receipt at final revision | CI |

**MODIFIED**

| Component | Change |
|-----------|--------|
| `check/scalar.go` | accept `ByteView` parameter; lower `len`/index; `checkerScalarFact.view`; refusal categories for any other use of the view |
| `corevalidate` | own view-parameter law, op structural law, `hasScalarCFG` set, call-arg provenance law, module/export law |
| `originvalidate` | own scalar transfer for the two ops; `BuildInterface` emits `Mode:"shared"`, `Drops:false` for ByteView params; module visibility vs publishable |
| `pathoracle` | extend straight-line replay to the caller shape; keep acyclic refusal; refuse owner/view ops in any cyclic function |
| `interp` | ByteView value, two ops, hex-string entry input, defect on OOB, model-only `FileBytes` outcome |
| `cgen` | `SCHWAY_BYTES` struct, scalar emission with view param, guard, reserved identifiers, module-qualified C names (multi-module only), file-app shell |
| `session` | `admitPublishedApplication` core-refusal diagnostic (W2), command-layer admission before execution (W3), single-load pipeline, app-verify report shrink (N1), SourceSet plumbing |
| `native` | `BuildIdentityInputs` gains module inventory (additive); bindings for new adapter |
| `cmd/schway` | no new flags: module root = directory of the entry file; exit mapping follows session |
| `evidence`, `cache`, `debugmap` | SourceSet digest, per-unit attribution (single-module bytes unchanged) |

---

## 3. Decision 1: iterating a borrowed byte buffer inside a scalar loop

### 3.1 Recommendation (ONE)

**Constant ByteView parameter; loan confined to the caller; loop function stays scalar.**

```
fn adler32(data: ByteView) -> U64 {          # library, scalar CFG, ByteView param is read-only
  var a = 1
  var b = 0
  var t = 0
  var i = 0
  let n = len(data)                          # n: U64 scalar
  while i < n {
    t = a + data[i]                          # OpBytesAt: bounds-checked, no event
    a = t % 65521
    t = b + a
    b = t % 65521
    i = i + 1
  }
  ...16 doublings of b, then + a...          # B*65536 + A with existing operators only
  t
}
```

Caller (Phase 31), straight-line and fallible, as in Phase 23/24:
`acquire(path) -> OpBorrowShared(view <- owner) -> OpCall adler32(view) -> OpRelease(owner) -> OpReturn`.

Why this and not the alternatives:

| Option | Verdict |
|--------|---------|
| **A. Constant ByteView parameter, loan in caller (chosen)** | No loan, owner, or resource exists inside the loop function, so FLOW-02's back-edge refusal is *unchanged*. The only thing live across the back edge is a parameter place that cannot be assigned (`assign` needs `Mutable`; parameters are not), moved, copied, borrowed, stored, or returned. Escape is impossible by construction: scalar functions return `U64|Unit` and ByteView grants no abilities. `pathoracle` (acyclic) is never asked to reason about a cycle. The library is pure, so the interpreter/native differential can run it on literal bytes without any file. |
| B. Loop-invariant owner + per-access transient borrow | Carries a *resource place* across the back edge (needs "invariant authority" analysis in three cyclic analyses) and couples the algorithm to IO. More invasive, weaker library story. |
| C. Generalize loan liveness over back edges | Spike 002 shows the producer fixpoint can; no independent cyclic oracle exists (F11). Ships a producer without a peer. Reject for M006. |
| D. Iterator protocol (`loop buffer { next(buffer) }`, the provisional fixture) | Needs a mutable cursor = exclusive-borrow state carried across the back edge. Strictly harder than C. Reject; keep the file as the frozen refused witness. |

Lessons applied `[analogy]`: Rust NLL keeps a shared loan live around a loop whenever a
reference is used in a later iteration; our design avoids that analysis entirely by never
creating the reference in the loop. Zig passes slices as `(ptr,len)` parameters with
safety-checked indexing, which is exactly this shape, minus lifetime tracking; Schway keeps
the lifetime check in the caller. Go's `for i := 0; i < len(b); i++` index loop is the
simplest sound form.

### 3.2 Soundness argument (what each consumer must independently establish)

1. **View place is constant.** Parameter place of type `ByteView`, initialized at entry, never
   a target of any op, never `Mutable`. The only legal operand uses are `SourceID` of
   `OpBytesLen` and `OpBytesAt`.
2. **No escape.** Function return type is `U64|Unit` (already enforced by `checkScalarFunction`);
   ByteView has no `copy`/`escape` ability; `let v = data`, `var v = data`, `return data`,
   passing it to anything but the two ops are refused.
3. **Index is U64; failure is dynamic.** Static refusal is only possible for *type/shape*
   errors. A length-based static refusal is not sound because length is a runtime value; do
   **not** add range analysis or guard-elision in M006 (guard is always emitted).
4. **Caller loan ends at the call.** Owner not released/moved/exclusively borrowed while the
   view place is live; escape of the view out of the caller refused; both reuse the straight-line
   loan law (verify release-while-loan-live is refused by all three layers: spike item).

### 3.3 Per-consumer change and independent re-derivation

| Consumer | Change | How it re-derives *independently* (no import of another consumer's conclusion) |
|----------|--------|-----------------------------------------------------------------------------------|
| `check` | `checkScalarFunction` admits `ByteView` param; lowers `len(x)` -> `OpBytesLen{Source:view,Target:U64}`, `x[i]` -> `OpBytesAt{Source:view,Right:i,Target:U64}`; fixed-point fact gets `view`; `OpBytesAt` requires `source.view && right.initialized`; all other touches of a view place -> `check.cfg_back_edge` category `loan_derived_provenance` (add `ByteView` to `scalarAuthorityCategory`, which already anticipates `View`). Parser: `scalarBodyStart` must accept ByteView params and a leading `let n = len(data)` (current `scalarLetStart` rejects `ident (` and requires U64 param). | Source-driven (AST) derivation. |
| `corevalidate` | `hasScalarCFG` += two kinds; `validateScalarFixedPoint` gets its own `scalarAbstractValue.view`; structural law per op (types: ByteView/U64/U64, no stray fields, target not parameter); call-argument law: an `OpCall` whose callee parameter type is ByteView must take either the caller's own ByteView parameter or a place produced by `OpBorrowShared` in the same function. | Reads core only; its own transfer function and its own refusal codes (`core.byte_view_*`), distinct from check's. |
| `originvalidate` | `originScalarFixedPoint` transfer for the two ops (own copy); `BuildInterface` derives `ParameterContract{Mode:"shared", Drops:false}` from the *type constructor* in its own table; return stays `owned`, `Fresh:false` (U64). | Recomputes from core types; ClosureDigest chain picks the new Mode up automatically. |
| `pathoracle` | Not asked to iterate the loop. Extend the caller-shape replay: ordered modes `acquire, borrow(view), call, consume`; the view must not outlive the call; any function whose CFG is cyclic **and** contains owner/loan/view/foreign ops is refused with the existing `pathoracle.cfg_back_edge`. | Operation-sequence replay only; ignores serialized `LoanEndpoints`. |
| `interp` | Value kind `bytes []byte`; `OpBytesLen/At`; OOB -> event `function.defected` with output `byte index out of bounds`, outcome `defect` (same construction as `U64 remainder by zero`). Entry param `ByteView` takes a hex string (extending the existing `Buffer` -> `"01020304"` convention). | Go semantics, no shared code with cgen. |
| `cgen` | `typedef struct SCHWAY_BYTES { const unsigned char *data; uint64_t length; }`; scalar function takes it by value; `if (i >= v.length)` guard before every read: app shell `fputs("schway: byte index out of bounds\n", stderr); exit(65);`, conformance shell `emitProgramDefectTerminal`; no `restrict`; new fixed identifiers added to the closed reserved set (`TestGeneratedIdentifierNamespacesStayConfined`). | Emission from core; evidence is the engine differential plus an omitted-guard mutation that ASan or a wrong result must kill. |

Retained refusals that must stay pinned as negative controls (all must still fail with their
current causes): `borrow` inside a loop body, `let v = data` copy, storing a view in `var`,
returning a view, moving/releasing an owner in a loop, owner or loan parameters on a loop
function.

### 3.4 Evidence rules (so size never changes the evidence shape)

* Byte ops are **not evented**. A 65,536-byte input yields the same O(1) event document as a
  1-byte input. Checksum body uses `var`+assign only.
* Independent oracle: Go `hash/adler32` in tests (stdlib; independent of the compiler). `[ran]`
  confirms `Checksum("Wikipedia") = 300286872` (0x11E60398), `Checksum(nil) = 1`, `Checksum({0}) = 65537`.
* Algorithm choice: **Adler-32** (RFC 1950 s8.2). It needs only add, `%`, `<` (all admitted).
  `B*65536` is composed with 16 self-additions, avoiding a seventh arithmetic operator across
  six consumers. Result printed as decimal U64 (existing `app run` convention); hex output is
  deferred. Empty input is a **success** (`1`), unlike Phase 23's adapter where empty is an error.

---

## 4. Decision 2: byte indexing and length

| Question | Recommendation |
|----------|----------------|
| Where do bounds checks live? | **Dynamic only**, emitted in every engine at every `OpBytesAt`, never elided. Static refusal covers type/shape only. |
| Mirroring M005? | Checked remainder: static `type.scalar_remainder_zero` for a *literal* zero, dynamic defect otherwise. Bytes have no literal-length analog, so the static half is intentionally empty and documented as such (do not invent range analysis). |
| Element type | `data[i]` is `U64` in `[0,255]`. No new scalar `Byte` type, no widening op: avoids a new type through the six consumers. |
| Dynamic OOB, application shell | stderr `schway: byte index out of bounds\n`, `exit(65)`, stdout empty. |
| Dynamic OOB, conformance/`run` | `defect` terminal outcome, reason `byte index out of bounds`; interpreter and native documents compared byte-for-byte after schema-2 projection. |
| Owner live at defect | OpDefect semantics are abort-only with no release; the execution document lists the owner under `live_resources`; process exit lets the OS reclaim. Do not build a landing pad (the multi-frame pad is open debt D-10-34). The shipped library cannot trigger this; only the deliberate control can. |
| Evidence events | No per-byte events. Existing events cover acquire, borrow, call/return, release, defect. |
| Required controls | `data[len(data)]`, `data[0]` on empty, `i <= n` loop mutation (off-by-one), omitted-guard cgen mutation (must die under ASan/wrong result), index of non-U64, index of a non-view place (forged core, corevalidate refuses). |

---

## 5. Decision 3: local modules

### 5.1 Recommendation (ONE per sub-decision)

| Sub-decision | Choice | Rationale |
|--------------|--------|-----------|
| Resolution site | **Session loader (`modload`), producer side, IO there only.** `syntax` stays single-file; `check`/`core`/peers do no IO. | Mirrors `readBoundedFile` placement; peers must not trust the loader (see re-derivation below). |
| Naming | Module `a.b.c` lives at `<root-dir>/a/b/c.schway`; the `module` header must equal the path-derived name (`module.path_mismatch`). Segments `[a-z0-9_]`. Root dir = directory of the entry file. No search path, no env var, no `..`, no absolute paths, no symlinks (Lstat), exact-case directory-listing match (macOS case-insensitive FS). | Closed, deterministic, explicit build authority. Go-style path=identity without GOPATH. |
| Syntax (provisional) | `import a.b.c { fn name, fn other }` — explicit item list, no glob, no alias, no re-export, no transitive visibility. Callee used bare by name; collision with a local or another import is an error. | Smallest surface that makes visibility explicit. |
| Visibility | The existing `export { fn ... }` block *is* the public set. Non-exported helpers are private; importing one is `module.not_exported`. Enforced by check **and** independently by corevalidate (`Function.Exported`, cross-module `OpCall` must target an exported callee). | Two independent derivations; negative control = import a private helper. |
| Cycles / caps | Cycle -> `module.import_cycle` (three-color DFS, copied not shared). Caps: 16 modules, depth 8, 4 MiB aggregate, 1 MiB per file (existing). Diamond imports dedup by canonical module name; digest mismatch for the same name is refused. | Fail closed, bounded. |
| Flow through core/peers | **Whole program after resolution**, one `core.Program`, one checker run, peers unchanged in shape. Function IDs stay `s1:<module>:fn:<name>`, now with the real defining module. Additive fields `Function.Module`, `Function.Exported`, `Program.Modules[{name,digest,imports}]` (omitempty; single-module bytes unchanged, no schema bump). | Every peer, interpreter and emitter already works on one Program; per-module analysis would need new composition machinery for no M006 gain. |
| Entry and unused functions | Entry = exported `main` of the root module, explicit. After check, `callgraph.Prune(program, entryID)` removes unreachable functions before `corevalidate`/`cgen` for build/run. `check <lib>` of a library with no `main` validates the whole unpruned program. Also reclassify `callgraph.entry_ambiguous` as a source refusal (exit 2). | F8: required for any library with an unused export. |
| Cross-module call admission | Via `callSignatureTable`/`calleeContracts` built from the imported unit's *signatures only*. Add a guard test in the SEM-05 style: the importer path never reads an imported body. | Preserves D-13-02b: swapping "source of imported signatures" for a `lang.interface/1` file later changes the input, not the admission. |
| Staleness | `ClosureDigest` Merkle chain already makes an importer's digest change when a callee in another module changes (cross-module calls are plain `OpCall`s in the merged program). | Free correctness. |
| C symbols | For multi-module programs only: allocate with preferred name `cName(module + "__" + name)` and key the ordinal by function ID, not slice index; single-module programs keep today's `cName(function.Name)` so accepted-corpus byte-identity tests (`TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity`) stay green. Function order in the merged program: topological by import graph, then source order; IDs are name-based so editing module B never renumbers module A. | Unique and stable across modules without moving pinned bytes; ordinal-indexed names would change when an unrelated function is inserted (hostile to future separate compilation). |
| Diagnostics/debugmap attribution | Keep `Span` file-local; add `Diagnostic.Unit` (omitempty) set at the single append site per function/decl in `check.Program`'s loop and by the parser result per file; `debugmap.Entry` gets the same field. Human projection prints `module:start-end`. | Zero change to existing JSON for single-file; no virtual-concatenation offsets (which would break every `source[Start:End]` slice and the formatter). |
| Evidence / build identity / cache | `SourceSet` digest = digest of the sorted list `(module, canonical-file-digest)`; for a one-module set it equals today's `SourceDigest`. `native.BuildIdentityInputs` gains an additive `modules` inventory (name, digest) beside `Inventory`; `Cacheable:false`, `DependencyClosure:"incomplete"` stay (host C closure still incomplete). Loader reads each file once; everything downstream uses the in-memory set (removes the `publishedOriginProblemFile` re-read). | Content-bound, closed set, no TOCTOU. |
| Formatter | `import` blocks formatted canonically, idempotence + lossless properties extended (existing generated-program and malformed-edit lanes). | Source authority constraint. |
| What modules do NOT claim | They are source-level whole-program linking. They do **not** establish the B1/DX-06 contract boundary (PRODUCT-ROADMAP: "modules alone do not establish it"). | Honest scope. |

### 5.2 Why not per-module analysis now

Per-module core would need per-module peers and an interface-consumption path for every
consumer simultaneously. The seams exist (F9) but exercising them end to end is the
separate-compilation capability. M006 keeps the *inputs* shaped for it (signature-only
imports, module-stable IDs, per-module digests) and defers the *plumbing*. `[analogy]` OCaml's
`.cmi` consistency errors show the failure to avoid: dependents silently built against a
different interface. The `ClosureDigest` chain plus `Program.Modules[].digest` is the same
guard, enforced at link time.

### 5.3 Adversarial pass (modules)

| Attack | Defense |
|--------|---------|
| `import ../../etc/x`, absolute, symlinked dir | name charset + Lstat-no-symlink + prefix check under root |
| `Lib.schway` vs `lib.schway` on case-insensitive FS | exact-case match against `ReadDir`; duplicate-after-fold refused |
| Header says `module a`, file is `b.schway` | `module.path_mismatch` |
| Same module reached two ways with different bytes | digest conflict refusal |
| Private helper reached by forged core | corevalidate `Exported` law |
| Unused export breaks build | Prune (F8) |
| Same function name in two modules | IDs qualified; C allocator closure; imports forbid ambiguous bare names |
| File changes between check and build | single read into `SourceSet`, no re-read |
| Cache hit after a dependency edit | module digests are cache/identity inputs; ClosureDigest chain |
| 1,000-file fan-out | caps |

---

## 6. Decision 4: carry-over fixes (concrete designs)

### 6.1 N1: app verify loses the named reason at n >= 173

Cause `[code]`: in `VerifyApplicationCases`, exhausted cases set
`Engines = {"interpreter": <full execution>}` (about 371 bytes per unit of n);
`writeApplicationVerificationReport` fails closed at `MaxReplayCasesBytes` (64 KiB), so the
whole report (and the named `native.event_capacity_exhausted` case) disappears.

**Design (one):**
1. For exhausted cases do not embed the execution. Add additive omitempty fields to
   `ReplayCaseReport`: `InterpreterOutcome *ReplayExpected` and `InterpreterEvidence *ReplayEvidenceSummary{Digest, EventCount}` (digest = sha256 of `execution.CanonicalBytes`). Keep `Engines` nil for those cases. Schema string stays `lang.app-verification/1` (additive).
2. Add a size governor in `writeApplicationVerificationReport`: if the encoded report exceeds the limit, degrade deterministically (strip `Engines` from non-pass cases, then from all cases keeping outcome+digest summaries); only if still too large return the existing error naming the case count. The named unsupported status is therefore never lost to size.
3. Keep the 64 KiB ceiling (fail-closed is intentional) but stop sharing one constant between input and report: introduce `MaxVerificationReportBytes = MaxReplayCasesBytes` as a distinct name so the two can diverge later (also reduces N3-style literal coupling).
4. Tests: `sum_to_n` at n = 15, 172, 173, 1000, and 16 simultaneous exhausted cases all exit 3 with `native.event_capacity_exhausted` text and a report <= 64 KiB; wrong-expected stays `fail`; one-level-only exhaustion still `fail` (WR-01).

Runnable observation: `schway app verify examples/sum_to_n.schway --cases <n=1000> --report r.json` now writes a report.

### 6.2 W2: build maps origin refusal to exit 2 but corevalidate refusal to exit 3

Cause `[code]`: `admitPublishedApplication` returns a `diagnostic` only for the origin layer;
`BuildApplication` returns `fmt.Errorf("core validation failed: %s", code)` for the core layer,
which `runBuild` maps to `native.build_failed` / `StatusOperational`.

**Design (one):** `admitPublishedApplication` sets `applicationAdmission.diagnostic =
peerRefusalDiagnostic(validated, checked.Program.ModuleID)` for `applicationRefusedCore` (the
helper `CheckCommandFile` already uses). `BuildApplication` returns
`[]diagnostic.Diagnostic{adm.diagnostic}, nil` for both layers; `runBuild` needs no change.
`RunInterpreterApplication` and `VerifyApplicationCases` use the same diagnostic (verify maps
a source refusal to exit 2 through a typed error that `runApplicationVerify` recognizes, as a
follow-on; checker diagnostics there currently also exit 3). Policy to state in the phase:
*checker, corevalidate, originvalidate, and entry-resolution refusals are class 2 (invalid
source); toolchain, IO, timeout, and launch failures are class 3 (operational); engine
disagreement is class 4.* Control: a source the checker accepts and corevalidate refuses
(use the documented divergence in `session_admission_divergence_test.go`) must exit 2 with the
peer's code from `build`, `check`, and `run`.

### 6.3 W3: `run --engine=native` foreign_origin_omitted exits 3 native.tool_failure

Cause `[ran]`: interpreter route reaches `publishedOriginProblemFile` after a clean run, so the
origin diagnostic wins (exit 2); native route errors inside `runNative` (cgen refuses) and
`RunNativeCommandFile` only consults the origin check when `err == nil`.

**Design (one):** add one command-layer helper `commandAdmission(sourceSet) (CheckResult,
[]diagnostic.Diagnostic)` = `Check` -> `corevalidate` -> `originvalidate.ValidatePublished`,
used by `RunInterpreterCommandFile` and `RunNativeCommandFile` *before* any engine runs, and
pass the already-validated program into the engine call (no second file read; this also
retires the deferred "re-read in publishedOriginProblemFile" item). Do **not** widen the
internal lanes (`RunInterpreter`, `RunNative`, verify corpora) that intentionally run
undeclared-borrow fixtures (D-27.1-02). Control: a table test over
`{foreign_origin_omitted, the core-divergence fixture, an unused-function program}` asserting
identical class, code and exit for `--engine=interpreter` and `--engine=native`.

### 6.4 N2: hosted two-host receipts

Not a code change. Receipts are pinned to an exact SHA (ancestor acceptance was rejected,
D-27.1-06), so N2 must be the **last plan of the last phase**, after the final code and
fixture change; add `scripts/verify-phase3x.sh` and a CI evidence-aggregate step beside
`verify-phase27.sh`, reusing `scripts/check-phase27-receipts.sh` mechanics. Any later code
change reopens it.

---

## 7. Decision 5: build order (phase decomposition, next integer 28)

| Phase | Name | Runnable gain (user-visible) | Depends on | Consumers touched |
|-------|------|------------------------------|-----------|-------------------|
| 28 | Source-refusal exit classes and verify report (W2, W3, N1, single-load, `entry_ambiguous` class) | `build`/`run` agree on exit 2 for source refusals; `app verify` at n=1000 reports the named unsupported case | none | session, cmd |
| 29 | Byte views and bounds-checked scalar loop | `schway run --engine=native` on `adler32` library fixture (synthetic hex input) matches interpreter; Go `hash/adler32` vectors; OOB control fails identically on both | 28 (new refusals must surface as exit 2) | all six + parser/formatter/ability |
| 30 | Local modules and second consumer | two programs (`checksum`-sample and a second consumer) import one library without copying; private-helper import refused; same entry/C naming guarantees; multi-file diagnostics/evidence/build identity | 29 (library to import) | syntax, check, core, corevalidate, originvalidate, callgraph, cgen naming, session, native identity, evidence, debugmap |
| 31 | Bounded file checksum application | `schway build ... --manifest ...` then `app run ./checksum -- file` prints the Adler-32 for empty/1-byte/65,536/oversized/missing/unreadable with exact stderr; cleanup observer | 30 (module-qualified IDs; name-keyed matchers), 29 | check, pathoracle, cgen, interp (model), native, examples |
| 32 | Closure: dual-host receipts and guide (N2) | clean-checkout guide; hosted macOS+Ubuntu receipts at the final SHA | all | scripts, CI, docs |

Ordering rationale:
* 28 first: tiny, independent, and every later refusal (ByteView misuse, import errors,
  entry ambiguity) reaches users through the paths it fixes.
* 29 before 30: modules need a library worth importing; the loop machinery is the highest
  *soundness* risk and gets the full six-consumer slice while the program is single-file.
* 30 before 31: highest *blast radius* (IDs, C names, diagnostics, identity) lands before the
  fixture-shaped owner path (F5), so the new owner shape is built module-aware once instead of
  retrofitted. Phase 30's consumers run through `run --engine=*` (ByteView entry, hex input);
  no file IO is needed to prove the module boundary.
* 31 is the user headline but the biggest per-slice cost; it composes finished parts.
* No two consecutive enabling phases (AGENTS.md rule): 28 fixes user-visible behavior, 29-31
  each run something new, 32 is closure.

Fixture-first gate for every phase: check in the target `.schway` as a refused frontier
fixture with its first diagnostic pinned, and gate on the pinned diagnostic moving. Phase 29
starts with `examples/checksum/adler.schway` refused at `len(`; leave
`examples/checksum.schway` and `TestPhase20ChecksumFrontier` untouched until Phase 32, then
retire them with a dated amendment.

### Dependency graph

```
28 (exit classes, N1) --> 29 (ByteView + loop) --> 30 (modules) --> 31 (file app) --> 32 (hosted closure)
                                    \------------------------------/
                                     library fn reused by 30 and 31
```

---

## 8. Patterns to follow

1. **Confine the loan, not the analysis.** Prefer shapes where the dangerous authority never
   enters the cyclic region over extending cyclic analyses that have no independent peer.
2. **Copy, do not share, peer tables.** New classifier sets, borrow-projection tables
   (`FileBytes -> ByteView`) and visibility rules are written separately in each consumer
   (D-27-07); add a registry test that every new op kind appears in all four classifiers and
   in `exhaustiveDispatchFixtures`.
3. **Additive, omitempty, no schema bump** for every new core/diagnostic/receipt field, with
   a byte-identity test over the existing accepted corpus.
4. **Closed bounds with named outcomes** (`MaxFileBytes = 65,536`, 16 modules, depth 8): every
   cap has an exact-N and N+1 control and a stable code.
5. **Audited C adapter with fault-injection hooks** (as `examples/phase23/adapter.c`): read
   until EOF with a cap of limit+1 and never trust `st_size`; `O_NONBLOCK` open, `S_ISREG`
   check; free and close on every failure path; macro hooks for open/read/close/malloc so a
   physical observer test proves cleanup on error paths.
6. **Verify-script plus CI lane per phase** (`verify-phaseNN.sh`, dual host receipt).

## 9. Anti-patterns to avoid

| Anti-pattern | Why bad | Instead |
|--------------|---------|---------|
| Iterator syntax (`loop buffer { next(buffer) }`) | Mutable cursor = carried exclusive state | Index loop over a constant view |
| `let` copies inside the byte loop | Evented `OpCopy` per iteration -> `native.event_capacity_exhausted` | `var` + assignment |
| New per-byte events for "evidence" | Document size grows with input; capacity wall | Outcome-level evidence + engine differential |
| Eliding the bounds check when `i < len` dominates | Unproven optimization in the trust path | Always emit; optimize only with a later proof |
| Virtual-concatenated source offsets | Breaks `source[Start:End]` users and the formatter | File-local spans + `Diagnostic.Unit` |
| Index-ordinal C names for multi-module | Unstable when an unrelated function is added | Name/ID-derived preferred names |
| Reusing the Phase 23 owner matchers by editing them | Moves pinned bytes and frozen evidence | New parallel shape; leave FileByteOwner frozen |
| Treating `callgraph.entry_ambiguous` as operational | Source-attributable | Prune + class 2 |
| Re-reading source files after check | TOCTOU grows with N files | One `SourceSet` |
| Hex-encoding a 64 KiB input into argv | ARG_MAX / `MaxApplicationArgumentBytes` | Differential on <= 4 KiB hex; max-size case via real file (Phase 31) and Go-level `[]byte` into the interpreter |

## 10. Risks and open items per phase (research flags)

| Phase | Flag | Item |
|-------|------|------|
| 28 | Standard | Mechanical; one open choice (verify's source-refusal exit class) is recommended above |
| 29 | **Spike first** | (a) Can an `OpCall` take a ByteView argument through `check`'s caller path and `originvalidate`'s call-site derivation without new `Mode` plumbing? (b) Parser lookahead: `scalarBodyStart` / `scalarLetStart` with `len(`, `[`, `]` (currently no postfix and `ident (` excluded); (c) confirm release-while-loan-live is refused by check, corevalidate, and pathoracle; (d) interpreter cost for 65,536 iterations (maps per op) |
| 30 | **Deeper research** | Entry/prune interaction with `corevalidate`/`EntryFunction`; name-keyed owner matchers; cgen allocator behavior with qualified preferred names; `Diagnostic.Unit` plumbing sites; `qlt03_shape_register`/`witness_registry` updates; formatter idempotence for `import` |
| 31 | **Moderate** | Reusing `Foreign *ForeignOperationContract` and `validateLocalOperationBindings` (already symbol-generic) vs new pathoracle replay table; typed-borrow projection `FileBytes -> ByteView` at five sites; macOS vs Linux `read`/`O_NONBLOCK` semantics for unreadable files; defect-with-live-owner documentation |
| 32 | Standard | Process only; must be last |

Session meta-tests (inventory, groundedness, validation-lifecycle, witness registry,
`self_describing_docs`) catch planning-doc and call-site drift; run them locally in every
phase before dispatching hosted CI (user memory).

## 11. Scalability and limits

| Concern | Limit chosen | Behavior at limit |
|---------|--------------|-------------------|
| File size | 65,536 bytes (matches the existing output ceiling) | `limit` succeeds; `limit+1` -> adapter error `FileTooLarge`, exit 65 |
| Interpreter work | ~12 ops x 65,536 iterations | Acceptable for tests; add step-budget awareness if a later size grows |
| Event capacity | independent of input size (no per-byte events) | n/a |
| Modules | 16 / depth 8 / 4 MiB total | named refusal codes |
| Report size | 64 KiB, degrade ladder | named case kept (N1) |

## 12. Exact result matrix to specify (for requirements)

| Input | Result |
|-------|--------|
| empty file | stdout `1`, exit 0 (Adler-32 of empty) |
| one byte `0x00` | `65537` |
| `Wikipedia` | `300286872` |
| exactly 65,536 bytes | Go `hash/adler32` value |
| 65,537 bytes | stderr `schway_bytes_read: FileTooLarge`, exit 65, no stdout, buffer freed |
| missing / not regular / unreadable | stderr `schway_bytes_read: OpenFailed` / `NotRegular` / `ReadFailed`, exit 65 |
| allocation failure | `AllocFailed`, exit 65 (fault-injected in the observer test) |
| deliberate OOB control | stderr `schway: byte index out of bounds`, exit 65 |
| path > 4096 bytes or NUL | existing `app run` exit 65 |

## Sources

* `[code]` `internal/compiler/check/scalar.go`, `check/check.go` (dispatch ~259-420; owner shapes ~3690-4090), `core/core.go` (operations, `Interface`, `FunctionSignature`), `corevalidate/corevalidate.go` (`hasScalarCFG`, `validateScalarFixedPoint`), `originvalidate/originvalidate.go` (`originScalarFixedPoint`, `BuildInterface`), `pathoracle/pathoracle.go` (`ValidateLocalOwnerPaths`, `backEdgeError`), `interp/interp.go` (`OpRemainderU64`, `RunWithForeignOutcomes`), `cgen/cgen_program.go` (`emitProgramScalarFunction`, `invocationEventCapacity`, `hasScalarOperations`), `callgraph/callgraph.go` (`EntryFunction`), `session/session.go` (`admitPublishedApplication`, `BuildApplication`, `RunNativeCommandFile`, `publishedOriginProblemFile`), `session/session_app_verify.go`, `cmd/schway/main.go` (`runBuild`, exit mapping), `native/native_app.go` and `bindings.go` (`BuildIdentityInputs`), `diagnostic.Span`, `examples/phase23|24`, `examples/checksum.schway`, `session_phase20_test.go`.
* `[ran]` scratch build of `./cmd/schway`: `foreign_origin_omitted` exit 2 (interpreter) vs 3 (native); unused-function program passes `check`, fails `run`/`build` with exit 3 `entry_ambiguous`; scalar callee called from a linear caller builds and prints 55; `examples/checksum.schway` first diagnostic `syntax.expected_rbrace`; Go `hash/adler32` vectors.
* `.planning/PROJECT.md`, `AGENTS.md`, `.planning/PRODUCT-ROADMAP.md` (Checker and guarantee activation map), `.planning/milestones/M005-MILESTONE-AUDIT.md` (N1-N4, W2, W3), `.planning/spikes/002-cfg-edge-last-use/README.md`, `007-loan-across-branch/README.md`.
* `[analogy]` (MEDIUM, prior knowledge, not re-fetched): Rust NLL region liveness around loops; Zig slice parameters with safety-checked indexing; Go package = directory, no import cycles, content-keyed build cache; OCaml `.cmi` interface consistency checks.
