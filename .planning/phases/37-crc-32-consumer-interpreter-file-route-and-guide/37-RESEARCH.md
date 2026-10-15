> **Resolved research amendment (2026-10-15):** The orchestrator amended CONTEXT D-10 to preserve the observed, existing-test-pinned empty-token tuple `(65, "", "")`. The conflict descriptions and original copied D-10 below are retained as the research discovery history; they are resolved, and they do not authorize changing generated C or the old empty-token test. Current CONTEXT wins. The bounded scratch CRC probe returned `3421780262` for `123456789`; no new repository test was run.

<user_constraints>
## User Constraints (from CONTEXT.md)

The locked decisions, discretion areas and deferred ideas below are copied verbatim from the phase context. [VERIFIED: .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-CONTEXT.md:14-57, 127-132]

### Locked Decisions

- **D-01:** Implement CRC-32/ISO-HDLC exactly: reflected polynomial `0xEDB88320`, initial state `0xFFFFFFFF`, XOR each unsigned byte into state, eight least-significant-bit-first right-shift/XOR rounds per byte, final XOR `0xFFFFFFFF`. Return a U64 whose value is in `0..4294967295`; print bare unsigned decimal plus one newline. Empty input returns zero. No table, POSIX CKSUM variant, CRC-32C, text normalization, hex output, cryptographic/authentication claim or host CRC call in the Schway implementation.
- **D-02:** Expose only `crc32(bytes: ByteView) -> U64` from `examples/crc32/lib/crc32.schway`, declaring `module lib.crc32`. Both entries import `lib.crc32 { fn crc32 }`. A top-level `import crc32` would resolve to the required entry file and self-import. Use a private arity-one `crc_byte(mixed: U64) -> U64` helper containing the eight bit rounds; the outer byte loop calls it once per byte with `crc ^ byte_at(bytes, i)`. This reduces maximum activations from 524,288 to 65,536 while retaining a bit-at-a-time algorithm and a real helper consumer. Keep the 32-bit state invariant explicit: the seed, byte XOR, logical right shift and polynomial XOR preserve it. Arbitrary external U64 state is not part of the API.
- **D-03:** Add `examples/crc32/crc_low.schway` as the second consumer. It calls the identical imported `crc32` and prints the low 16 bits, following `adler_low`; `123456789` must give `14630`. Keep all CRC byte/bit loops in the one library. Define the single-copy scan over the new library plus its two entry closures, accounting explicitly for the historical Phase 36 bit-step example instead of falsely demanding one CRC-related loop repository-wide. An armed library mutation must change both real application outputs, including the low bits.
- **D-04:** Reuse the Phase 31 FileBytes foreign contract and copy-local adapter/header/manifest triple without semantic changes. Repeat entry-local foreign declarations and `FileError` data where required because imported foreign/data declarations remain unsupported. Check adapter-copy equality/drift against the established contract. No dependency, generic IO package, new binding dialect, arity-N, aggregate, subtraction, writable byte buffer, dynamic shift, owner-loop widening or optimizer assumption is justified. Keep the existing formatter limitation explicit; do not claim these examples are format-canonical unless separately demonstrated.

### Public file-run grammar and authority

- **D-05:** Add separator-marked ordinary application forms: `schway run --engine=interpreter SOURCE -- PATH` and `schway run --engine=native SOURCE --manifest MANIFEST -- PATH`. The native form builds into a private temporary location and runs once, preserving application streams without a build-success banner. Existing retained `build --manifest ... --output ...` plus `app run ARTIFACT -- PATH` remains the guide's reusable native workflow. The existing no-tail `run --engine=E SOURCE` remains the synthetic-input conformance route; a PathToken/U64 entry with no supplied input retains its usage refusal. No new installed `crc32` subcommand or default input is added.
- **D-06:** Dispatch the unambiguous new `run` separator form before the generic JSON/depth/kind/cursor/expand extractors. Parse only the pre-separator options; exactly one token after `--` is opaque file-path data, including `--json`, `--depth=2`, `--`, leading dashes, spaces, Unicode and shell metacharacters. Reject malformed/duplicate/unknown options and wrong operand counts with usage 64. File application mode emits ordinary application streams, and does not accept JSON/report/evidence/query flags. Preserve the old no-tail global-option behavior. Never form a shell command from a file path.
- **D-07:** The interpreter file mode has no manifest and rejects one if supplied. Its source authority is the independently checked exact `schway_file_bytes_{acquire,view,release}` ABI, FileBytes/FileError types and PathToken-to-U64 owner path, implemented by one fixed production Go semantic adapter. It performs real host file reads, not arbitrary foreign C execution. No Clang invocation, ignored/ceremonial manifest, C digest allowlist, test-support import or hidden fallback to native. Native alone requires and uses its explicit manifest via the existing closed resolver and byte snapshots. Parity is claimed for the pinned canonical adapter/program cases, not arbitrary replacement C implementations.
- **D-08:** Apply the full existing command admission sequence before entry pruning: module-aware check, independent corevalidate, originvalidate and the applicable file-owner path oracle. Then resolve the entry and enforce the exact admitted file application shape and known foreign contracts. Add neighboring-source/raw-core controls rather than trusting symbol names or a successful main result. No silent read authority for a differently shaped foreign call. Keep modules confined under the entry root with the existing import path/symlink rules.
- **D-09:** Preserve the existing exit table: source refusal 2, engine/tool/manifest failure 3, old conformance disagreement 4, CLI grammar 64; file typed failures and checked application defects retain 65, and existing conformance/capacity/timeout/launch classes retain their separate 69/124/125 meanings where applicable. Record the new route in the Phase 28-owned table and current usage/route inventories. Pin real `(exit, stdout, stderr)` values, not prose categories alone. File read errors have empty stdout and exactly `FileError.FileTooLarge\n`, `FileError.FileUnavailable\n` or the existing supported `FileError.AllocFailed\n`, with no private input path. Do not collapse an actual defect into a successful zero checksum.
- **D-10:** Match the existing PathToken byte contract: at most `4,096` bytes, NUL and `4,097+` bytes refused at the input boundary with 65. Exactly `4,096` bytes is syntactically admitted even if the OS subsequently rejects the pathname. Empty string currently reaches the C adapter and gives FileUnavailable; pin that actual parity and correct the old checksum guide's contrary prelaunch-empty claim as a scoped documentation correction. Source paths in source diagnostics and private application input paths are different concerns; the fixed file-error messages must not echo the latter.

### Bounded host input and outcome/capture separation

- **D-11:** Implement the independent production Go file reader for macOS/Linux using open-once descriptor semantics, nonblocking open, descriptor Stat regular-file check and an actual limit+1 read capped before allocation. Do not trust Stat.Size or pre-stat then reopen. Follow input symlinks to regular files like the existing C adapter; this differs deliberately from module-import symlink refusal. Replacement after open stays bound to the open descriptor. Preserve close-on-exec/no-controlling-terminal posture where the Go/POSIX opening route supports it. Map open/stat/read/normal-close failure to FileUnavailable; preserve the primary failure on cleanup. Cover EINTR/short-read/close behavior at the appropriate seam rather than promising identical syscall schedules.
- **D-12:** Keep the `65,536`-byte cap specific to this built-in interpreter adapter and the canonical example C adapter. FileBytes generally remains binding/configuration-dependent; native can select a different admitted C binding. Update `schway.surface/1`/witnesses only where needed to expose truthful route-specific limits and input forms, preserving the independent finite census and ten-field wire shape. No false universal language cap or removal of existing restrictions.
- **D-13:** Add a file-application interpreter boundary with event/identity collection explicitly disabled for ordinary execution. Merely calling `RunWithFileBytes` and ignoring its error/events afterward is insufficient: scalar calls currently auto-create evidence and exhaust the 16 MiB document/event budget. Disabled execution must avoid retaining events, dynamic invocation identities and growing counter tables while still executing all computation, output and cleanup. Reuse the current execution engine and distinct application output sink/terminal outcome; no second evaluator or raised evidence ceiling. Keep the existing evidence-enabled seams and their callers compatible.
- **D-14:** Pin maximum-size CRC completion with capture disabled and small controlled executions with capture disabled/exhausted. Actual value, prior stdout, typed error, checked defect and process exit remain independent from evidence status. Optional evidence failure cannot authorize application success, mask a later defect, or imply complete verification. Normal return/typed-error cleanup must discharge the owner exactly once; checked abort-like defects retain the existing explicitly defined cleanup policy rather than inventing destructor guarantees. Numeric truth belongs to independent engine/oracle assertions, source intent to the checker, causal structure to the causal peer. Complete large-file traces and source-intent/numeric proofs from causal acceptance are outside the claim.
- **D-15:** Do not add a generic interpreter fuel/runtime system here. Public case execution and test processes must have explicit existing bounded timeouts/output capture. Record that arbitrary nontermination/timeout parity is unproved: native has a process deadline while the interpreter currently lacks context/fuel cancellation. APP-14 is the stated pinned file-program case contract, not equality for arbitrary divergent programs. Observe actual maximum-case cost before choosing focused test timeouts; no invented performance or cold/warm loop claim.

### Independent evidence, reuse and guide

- **D-16:** Store a finite independent CRC oracle/case table with literal values, byte recipes/digests, exact Python generating commands and Python/zlib compile/runtime versions. New calculation here used Python 3.14.4 and zlib 1.2.12. Pins: empty `0`; `00` `3523407757`; `123456789` `3421780262`; `65,536` x FF `3735780942`; `80` `1069182125`; `80 FF` `1465410873`; all byte values 00..FF `688229491`. Low16 projections are independently pinned too. Go standard-library `hash/crc32.ChecksumIEEE` can cross-check recurring tests; do not derive expected values from Schway or require Python/zlib as a compiler/runtime dependency. Oversize input has no application CRC answer.
- **D-17:** One shared matrix must run the real public interpreter and native file application forms and compare each separately to the expected tuple, then to each other. Cover both new entries and existing checksum/Adler consumers at their pinned rows: empty, zero, check text, binary high bytes, exact cap, cap+1, missing, directory, FIFO, device/nonregular, unreadable, opaque flag-shaped/space/Unicode paths and empty/long input. Native arguments include its manifest; interpreter arguments omit it. Keep unreadable-mode testing in a named root-sensitive subtest with an honest environment skip; do not skip the matrix or count a skip as a pass. Unsupported host capabilities are separately named, not silently green.
- **D-18:** Reach all five APP-13 controls against independent outputs: wrong polynomial, a non-reflected variant, missing final XOR, off-by-one bit/byte bound and shared-library mutation affecting both importers. A mutant must check/admit/build and reach the discriminating assertion; syntax/compiler/tool failure, timeout, zero runs or a skip is not the expected RED. Restore source/overlays and check raw bytes. Reuse finite mutation tables/helpers instead of a generic mutation framework. A single-copy text scan alone does not prove actual reuse; require both output changes and module-digest identity changes, including a library-only edit.
- **D-19:** Tie reader/adapter changes to actual bound and cleanup controls: cap and cap+1, forged file size, nonregular/FIFO refusal without blocking, read/stat/close faults and once-only descriptor/resource discharge. Reuse reached native physical observer/allocator controls where unchanged; model release events are not proof of native free. Go allocation failure is not claimed to be recoverable simply because the source has AllocFailed; use explicit modeled/injected alternatives where appropriate and distinguish host behavior. Add a focused sanitizer lane only for the actual changed/native consumer boundary with named host availability; no generic sanitizer matrix or leak claim on Darwin.
- **D-20:** Write `examples/crc32/README.md` as a short executable path from repository root: prerequisites, tracked-source build, ignored `.schway/crc32/` outputs, exact fixture creation, native retained build/run, both public source engines, second consumer, result table and fixed errors. State the cap, CRC-32/ISO-HDLC versus POSIX cksum/CRC-32/CKSUM, binary byte treatment, fixed interpreter adapter and optional-capture limits. Mark executable fences and run the actual commands in a temp copy of tracked files only. The old Phase 32 copy helper includes untracked nonignored files; reusing it unchanged cannot establish a tracked-only claim. Check exact command exit/stdout/stderr, fixture JSON/table agreement and an armed guide drift. Do not generalize this into a docs framework.
- **D-21:** Integrate high-value focused tests with existing recurring Go CI; no new workflow or recurring dependency installation unless a concrete gap requires it. Locally use exact named selectors/narrow CLI commands with `GOCACHE=/tmp/ai-lang-verification-gocache`. No project-wide suite, whole-package sweep, broad native matrix, push, hosted dispatch or receipt binding in Phase 37. Preserve prior C corpora/goldens and exact old route tests; any inventory update must enumerate actual new consumers and carry drift/negative controls rather than weakening the census.
- **D-22:** Target about eight coherent plans, starting with a runnable vertical tracer and a fixed pre-change baseline for affected routes. Organize expansion around actual consumer, route-authority, outcome/bounds, independent mutations and executable guide obligations rather than one plan per role. Additional plans need a named runnable witness or safety obligation. SUMMARY coverage must cite real registered tests with the existing kind vocabulary; no invented pass counts, completeness percentages or subjective review claims.
- **D-23:** Handle prohibition evidence during planning, not as a surprise at closeout. Keep recalled safety prohibitions in canonical plan metadata, preserve unresolved/classifier history, and arrange real supported bad/clean/default enforcement proof where required. Existing Go assertion tests and prose alone do not satisfy a producer that requires a machine-readable descriptor. Prefer small phase-local dependency-free proof wiring over a GSD framework change; never silently dismiss an unverified prohibition or fabricate human approval. Record exact source/fixture binding and distinguish independent numeric evidence from structural evidence.
- **D-24:** Retain z1g/trz/z1s dormant with the scoped transition dispositions below. Current search, checked source, bounded capture, observers and replay explain the known boundaries; no measured navigation gap, unexplained failure or contested source-form decision was found. No database, runtime telemetry, proof backend or background scheduler follows from the generic fan-out prompt. Preserve archived decisions and add dated living-document amendments rather than rewriting history.

### Planner's Discretion

Exact local API/file/test names and small helper boundaries; CLI parser implementation within the fixed grammar; whether small route-specific facts need new manifest rows or just accurate existing restriction text; concrete reached fault injection and sanitizer selectors; finite case allocation across focused owners. Keep production host IO independent from test oracle code. Revisit a recommendation only with a concrete source/behavior counterexample and record the scoped change.

### Deferred Ideas

Final hosted binding Phase 38 after audit; M008 arity-N/scalar records/exported types for wc/longest_line, then bounded key=value/JSON. Generic foreign execution in interpreter, a general IO/evidence/docs framework, runtime fuel/cancellation, full large-file causal traces, separate compilation, module-aware query/evidence/interface, numeric spelling unification and whole-loop performance work retain their existing owners/future consumer triggers. No new dormant seed is promoted.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| APP-13 | examples/crc32/crc32.schway reads a file of at most 65,536 bytes and prints its CRC-32/ISO-HDLC as bare decimal. It is built from an ordinary-source crc32 module that uses a bit-at-a-time loop with no table. | Scratch ordinary-source witness passed public check and native build/run for 123456789; the independent oracle and mutation controls are locked. [VERIFIED: 37-CONTEXT.md:19-22, 43-46; executed scratch probe 2026-10-15] |
| MOD-07 | A second consumer imports crc32 from the module without copying it and prints a pinned answer on both engines. | Module resolution and the existing single-copy consumer evidence pattern are inspected; the real dual-output and digest-mutation check is required. [VERIFIED: 37-CONTEXT.md:20-22; internal/compiler/modload/resolve.go and session_phase31_single_copy_test.go inspected] |
| APP-14 | schway run --engine=interpreter runs file-bytes programs such as crc32 and checksum from the CLI, and gives the same stdout, stderr and exit code as native for every pinned case. Interpreter/native parity for file input is no longer Go-level only. | Existing RunWithFileBytes is a caller-reader Execution API; the CLI lacks a suffix route and scalar calls auto-create evidence. A separate file application boundary and public tuple matrix are needed. [VERIFIED: internal/compiler/interp/interp.go:601-657, 1695-1713; cmd/schway/main.go:36-66, 120-130] |
| DX-24 | examples/crc32/README.md is an executable clean-checkout guide, run by a test the way TestPhase32GuideCleanCheckout runs its guide. | Existing selector and guide copy limitation are inspected; the new guide needs a tracked-only runner plus an armed drift control. [VERIFIED: cmd/schway/main_phase32_guide_test.go:347; 37-CONTEXT.md:47] |

Discrete requirement values copied from the source of truth include “at most 65,536 bytes,” “schway run --engine=interpreter,” and “TestPhase32GuideCleanCheckout.” [VERIFIED: .planning/REQUIREMENTS.md:68-80]
</phase_requirements>

# Phase 37: CRC-32 Consumer, Interpreter File Route, and Guide - Research

**Researched:** 2026-10-15
**Domain:** Checked ordinary-source CRC consumer, bounded FileBytes application execution, CLI routing, documentation/evidence
**Confidence:** MEDIUM

## Summary

Phase 37 joins three proof obligations around one user-visible operation: compute CRC-32/ISO-HDLC over an opaque file path with the checked Schway implementation, execute the same admitted file program through interpreter and native entry points, and make the workflow reproducible from a tracked-source checkout. Keep the CRC implementation in one shared module; keep host file access behind the fixed interpreter adapter and existing native manifest/adapter boundary. [CITED: .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-CONTEXT.md:8-10, 19-39, 41-50]

The existing execution architecture supplies the needed checked pieces: module-aware loading, source/core admission, owner-path validation, a checked PathToken-to-U64 shape, scalar helper calls, ByteView and a native FileBytes contract. The new work is boundary integration: a public opaque-argument route, fixed Go reader, capture-free application result path, and focused evidence for outcomes, bounds, consumer reuse and guide drift. [VERIFIED: internal/compiler/session/session_modules.go:24-49; internal/compiler/session/session_admit.go:11-55; internal/compiler/pathoracle/pathoracle.go:57-75; internal/compiler/interp/interp.go:601-657; internal/compiler/cgen/cgen_program.go:3215-3250]

**Primary recommendation:** Start with one runnable tracer proving the shared CRC source checks, builds and returns the independent 123456789 answer on the retained native path. Then add interpreter file execution with pre-prune admission, exact outcome mapping and explicit evidence-disabled semantics. [LOCKED: .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-CONTEXT.md:21, 43, 49; VERIFIED: scratch probe 2026-10-15]

## Project Constraints (from AGENTS.md)

- Stage 0 uses the Go 1.24 standard library and the development native path emits readable C17 to installed Clang. Prefer standard-library-only dependencies. Exact project text: “Go 1.24 standard library hosts Stage 0; readable C17 emitted to installed Clang is the development native path” and “Prefer standard-library-only and shallow audited boundaries; every dependency must earn more than copy-local implementation.” [VERIFIED: AGENTS.md:35-38]
- Safe-code and FFI guarantees must derive from checked facts; source/runtime authority stays explicit; macOS and Linux are initial host priorities. [VERIFIED: AGENTS.md:29-32, 44-47]
- Evidence lanes answer distinct questions and have explicit cost boundaries; structural receipts cannot substitute for numerical correctness or native cleanup proof. [VERIFIED: AGENTS.md:41-43; .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-CONTEXT.md:38, 45-48]
- The phase init operation ran before this research artifact write. Project workflow instruction: “Before using Edit, Write, or other file-changing tools, start work through a GSD command” and “Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.” [VERIFIED: AGENTS.md:82-90; GSD init.phase-op 37 executed 2026-10-15]
- CLAUDE.md routes native FFI and differential-harness work to spike-findings-ai-lang. I read its differential-verification and native-FFI references: keep checker/oracle implementations independent, arm concrete mutations, bind FFI cleanup facts to evidence, and keep oracle, native, optimized and sanitizer claims distinct. Apply only the relevant patterns here; current Phase 37 limits still rule out broad matrices and general harness work. [VERIFIED: CLAUDE.md:6-11; .claude/skills/spike-findings-ai-lang/SKILL.md:21-23; references/differential-verification-harness.md:18-41, 48-62; references/native-lowering-ffi-contract.md:11-16, 28-45; LOCKED: 37-CONTEXT.md:48, 50]

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|---|---|---|---|
| CLI file-run grammar and opaque tail | CLI/process boundary | — | Must dispatch before generic extractors so one post-separator token stays opaque; the command owns usage and exit classification. [VERIFIED: cmd/schway/main.go:36-66, 120-130; LOCKED: 37-CONTEXT.md:26-30] |
| Module admission and entry authority | Compiler session | Checker, core/origin validators, file-owner path oracle | Validate the full merged module program before entry pruning, then enforce the exact checked file-owner shape. [VERIFIED: internal/compiler/session/session_admit.go:11-55; session_modules.go:24-49; LOCKED: 37-CONTEXT.md:28-30] |
| Interpreter host file access | Interpreter adapter boundary | OS descriptor layer | Implement fixed FileBytes semantics with Go host reads; do not evaluate arbitrary foreign C or treat names as authority. [LOCKED: 37-CONTEXT.md:28, 35-37] |
| File application output and terminal result | Interpreter application boundary | Existing frame-stack engine and output sink | The file app accepts PathToken, unlike the current U64 app API. Keep output, typed failure, checked defect, cleanup and capture status distinct. [VERIFIED: internal/compiler/interp/interp.go:1084-1163; LOCKED: 37-CONTEXT.md:37-39] |
| Native application execution | Native build/runner boundary | Generated C and closed manifest resolver | Native keeps explicit manifest resolution, temporary build lifetime, direct argument passing and child outcome reporting. [VERIFIED: internal/compiler/session/session.go:1438-1486; internal/compiler/native/native_app.go:21-26, 431-452; CITED: pkg.go.dev/os/exec] |
| Executable guide | Example/documentation test boundary | CLI and fixture owners | Run guide fences in a temporary copy of tracked sources only; existing Phase 32 copy includes untracked nonignored files. [LOCKED: 37-CONTEXT.md:47; VERIFIED: cmd/schway/main_phase32_guide_test.go:347] |

## Standard Stack

### Core

| Library / component | Version | Purpose | Why standard |
|---|---|---|---|
| Schway source module and current checker/runtime | Current repository | CRC bit loop, ByteView processing, checked helper call and file app | Locked substrate; the scratch witness passed public check, native build and one public app vector without packages. [LOCKED: 37-CONTEXT.md:19-22; VERIFIED: scratch probe 2026-10-15] |
| Go standard library | Project Go 1.24 baseline; local probe go1.24.0 darwin/arm64 | CLI, fixed reader, filesystem descriptors, bounded read and process lifecycle | Use portable os/io APIs for ordinary file operations and isolate the platform seam needed for nonblocking open. [VERIFIED: AGENTS.md:35-38; environment probe 2026-10-15; CITED: https://pkg.go.dev/os, https://pkg.go.dev/io#LimitReader, https://pkg.go.dev/syscall] |
| Existing C17 FileBytes adapter contract | Copy-local adapter; cap scoped to canonical sample and interpreter adapter | Native read, status mapping and bounded FileBytes owner | Existing adapter is the canonical open-once/nonblocking/stat/read/cleanup pattern; the copy triple is reused unchanged. [VERIFIED: examples/checksum/adapter.c:36-50, 68-112; LOCKED: 37-CONTEXT.md:22, 35-36] |

### Supporting

| Component | Version | Purpose | When to use |
|---|---|---|---|
| Python zlib | Oracle-only; locked generator environment Python 3.14.4 / zlib 1.2.12 | Generate fixed expected CRC answers and byte recipes | Fixture generation/checking only; not a compiler or runtime dependency. [LOCKED: 37-CONTEXT.md:43] |
| Go hash/crc32 | Existing standard library | Independently cross-check recurring cases | Cross-check selected answers, never implement Schway CRC or serve as sole oracle. [LOCKED: 37-CONTEXT.md:43] |

**Installation:** None. No dependency is added. [LOCKED: 37-CONTEXT.md:22]

### Alternatives Considered

| Instead of | Could use | Tradeoff |
|---|---|---|
| One shared ordinary-source CRC module | Copy the bit loop into each entry | Violates the single implementation and shared-consumer requirement; require both output and digest changes under library mutation. [LOCKED: 37-CONTEXT.md:20-22, 45] |
| Fixed production Go FileBytes adapter | Execute arbitrary foreign C in interpreter | Expands execution authority; only exact checked ABI is admitted. [LOCKED: 37-CONTEXT.md:28-30] |
| Direct argv path passing | Construct shell command | Shell metacharacters and flag-shaped paths are opaque data. Official os/exec docs say it does not invoke a shell or expand globs. [LOCKED: 37-CONTEXT.md:27; CITED: pkg.go.dev/os/exec] |

## Package Legitimacy Audit

Not applicable: no external package is installed. [LOCKED: 37-CONTEXT.md:22]

## Architecture Patterns

### System Architecture Diagram

~~~mermaid
flowchart LR
  Args["schway run ... -- PATH"] --> Route["Early CLI route and grammar"]
  Route --> Load["Module-aware load and full admission"]
  Load --> Entry["Resolve exact PathToken-to-U64 FileBytes entry"]
  Entry --> Engine{"Selected engine"}
  Engine -->|interpreter| Reader["Fixed Go reader: open once, stat descriptor, bounded read"]
  Reader --> Frames["Existing interpreter frame stack, evidence disabled"]
  Engine -->|native| Manifest["Explicit closed manifest resolver"]
  Manifest --> CBuild["Temporary C17 build and generated adapter"]
  CBuild --> Child["Direct argv child execution"]
  Frames --> Result["Application streams and terminal outcome"]
  Child --> Result
~~~

The route owns the opaque suffix before generic flag extraction; both engine paths share admission and entry shape, only native uses the explicit binding manifest. [VERIFIED: cmd/schway/main.go:36-66, 120-130; internal/compiler/session/session.go:1438-1486; LOCKED: 37-CONTEXT.md:26-30]

### Recommended Project Structure

~~~text
examples/crc32/
├── README.md
├── crc32.schway
├── crc_low.schway
├── lib/
│   └── crc32.schway
├── adapter.c
├── adapter.h
└── file_bytes.bindings.json
~~~

The example paths and shared-module shape are locked; reuse the checksum adapter/header/manifest triple and compare copied bytes for drift. [LOCKED: 37-CONTEXT.md:20-22, 47; VERIFIED: examples/checksum/adapter.c, adapter.h and file_bytes.bindings.json inspected]

### Pattern 1: Admit the full source set before entry pruning

**What:** Use module-aware session loading and full command admission before resolving/pruning the entry. Existing admission runs core validation, origin validation and the local-owner path oracle over checked source. [VERIFIED: internal/compiler/session/session_modules.go:24-49; session_admit.go:11-55]

**When to use:** Every public source file route, before interpreter reads or native manifest resolution. [LOCKED: 37-CONTEXT.md:28-30]

**Example:** A bounded scratch ordinary-source CRC module passed public check, built through the public native command with the copied FileBytes manifest, then returned stdout `3421780262\n`, stderr empty, exit 0 for `123456789`. This checks one source/ABI shape and vector only; it is not a committed test or complete phase result. [VERIFIED: scratch probe commands on macOS arm64, 2026-10-15]

### Pattern 2: Separate application result from evidence retention

**What:** Add a file-application boundary that reuses the current frame stack and application output sink, with explicit event/identity/counter collection mode. RunWithFileBytes returns evidence-oriented Execution; runProgramFrameStack creates evidence for scalar-call programs when none is provided. [VERIFIED: internal/compiler/interp/interp.go:601-657, 1091-1163, 1695-1713]

**When to use:** Public no-report interpreter file application, especially the maximum-size CRC case; keep evidence-enabled APIs compatible. [LOCKED: 37-CONTEXT.md:37-39]

**Example shape:** Return ordinary output and the terminal outcome while a capture-disabled mode prevents evidence identity construction and event/occurrence growth. Calling current API and discarding its evidence-capacity error afterward does not meet the locked behavior. [LOCKED: 37-CONTEXT.md:37-38; VERIFIED: internal/compiler/interp/interp.go:146-180, 1362-1378, 1695-1713, 1978-2002]

### Pattern 3: Keep CLI path data opaque after separator

**What:** Recognize the new separator route before generic JSON/depth/kind/cursor/expand extractors. Parse only options before separator, require one path token after it, and pass path as data. Existing app run parser is a boundary precedent; existing no-tail run is post-extractor. [VERIFIED: cmd/schway/main.go:36-66, 120-130, 211-264]

**When to use:** Both public file-run forms. Interpreter rejects manifests; native requires/resolves one. [LOCKED: 37-CONTEXT.md:26-30]

~~~text
schway run --engine=interpreter SOURCE -- PATH
schway run --engine=native SOURCE --manifest MANIFEST -- PATH
~~~

These forms and option ownership are copied from locked context. [LOCKED: 37-CONTEXT.md:26-28]

### Anti-Patterns to Avoid

- **Prune before whole-source admission:** imported declarations could escape validation; admit the merged program first. [LOCKED: 37-CONTEXT.md:28-30]
- **Use RunWithFileBytes and ignore capture exhaustion:** scalar calls auto-create evidence and can make results capacity-dependent. [VERIFIED: internal/compiler/interp/interp.go:601-657, 1695-1713; LOCKED: 37-CONTEXT.md:37-38]
- **Treat names as authority:** enforce checked foreign contracts and owner path; test neighboring-source/raw-core controls. [LOCKED: 37-CONTEXT.md:28-30]
- **Route through generic flags or shell:** flag-shaped input could be consumed or interpreted; keep the path suffix opaque and direct. [LOCKED: 37-CONTEXT.md:27; CITED: pkg.go.dev/os/exec]
- **Treat a copy scan as proof of sharing:** mutate the library and require both output and module-digest changes. [LOCKED: 37-CONTEXT.md:21, 45]
- **Assume empty path reaches the adapter:** current generated C and an existing test prove the opposite; resolve this scoped conflict. [VERIFIED: internal/compiler/cgen/cgen_program.go:1099-1104; cmd/schway/main_test.go:328-345; scratch probe 2026-10-15]

## Don't Hand-Roll

| Problem | Don't build | Use instead | Why |
|---|---|---|---|
| FileBytes ownership and native ABI | New binding dialect, loose symbol-name matching or duplicated ABI | Exact existing contract, copied adapter/header/manifest and cgen/session validators | Owner path checks exact acquire/view/release contracts and ordered failures; a second dialect creates authority/layout drift. [VERIFIED: internal/compiler/cgen/cgen_program.go:3215-3250, 3316-3320; LOCKED: 37-CONTEXT.md:22, 28-30] |
| Independent CRC answers | Expected values computed by Schway or production CRC helper | Literal receipt generated with Python zlib; cross-check selected rows with Go hash/crc32 | Independent answers are a distinct evidence lane from checker/cause evidence. [LOCKED: 37-CONTEXT.md:43, 45] |
| General interpreter fuel, IO, evidence or documentation framework | Reusable framework without a current consumer | Small phase-local reader/result mode and tracked-only guide runner | Broad systems are explicitly deferred and lack a named consumer in this phase. [LOCKED: 37-CONTEXT.md:50-51, 130] |

**Key insight:** One checked source program and one fixed native ABI define the authority boundary. Interpreter and native need independent host file-read implementations; expected numbers and route outcomes need independent case ownership. [LOCKED: 37-CONTEXT.md:28, 35-38, 43-45]

## Common Pitfalls

### Pitfall 1: D-10 empty-path premise conflicts with live code and a current test

**What goes wrong:** D-10 says an empty path reaches the C adapter and produces FileUnavailable. Generated C returns 65 before calling the adapter, and the existing command test expects exit 65 with empty stderr. [CITED: 37-CONTEXT.md:31; VERIFIED: internal/compiler/cgen/cgen_program.go:1099-1104; cmd/schway/main_test.go:328-345]

**Evidence:** The generated guard is:
~~~go
out.WriteString("  if (argv[1][0] == '\\0' || strlen(argv[1]) > 4096u) return 65;\n")
~~~
~~~c
if (argv[1][0] == '\0' || strlen(argv[1]) > 4096u) return 65;
~~~
The adapter's exact empty-path check is “if (path == NULL || path[0] == '\0') return acquire_failure(SCHWAY_FILE_BYTES_UNAVAILABLE);”. It separately maps that case to unavailable; the existing empty-token test pins code 65 with no stderr. A newly built scratch public app also returned `(65, "", "")` for an empty path. [VERIFIED: examples/checksum/adapter.c:76-80; internal/compiler/cgen/cgen_program.go:1101-1104; cmd/schway/main_test.go:328-345; scratch probe 2026-10-15]

**How to plan:** Resolve this scoped context/source conflict before pinning the cross-engine tuple matrix. To satisfy D-10's desired result, plan a narrow C generator change that lets the empty token reach the adapter while retaining the length check; update the existing empty-token test and pin both engines to `FileError.FileUnavailable\n`. Alternatively, date-amend D-10 to the currently tested tuple. Do not silently carry the false current-behavior premise into implementation. [INFERRED from the cited conflict; 37-CONTEXT.md:31, 55 permits revisiting only with a concrete counterexample]

### Pitfall 2: Evidence changes ordinary interpreter completion

RunWithFileBytes calls runProgramFrameStack; scalar-call programs auto-create an evidence recorder. The recorder holds occurrence and event data, and each frame owns invocation identity/call-ordinal state. Disabled mode must prevent this data construction/growth, not erase events afterward. [VERIFIED: internal/compiler/interp/interp.go:146-180, 1362-1378, 1695-1713, 1978-2002; LOCKED: 37-CONTEXT.md:37-38]

### Pitfall 3: Portable Go file I/O alone does not prove FIFO safety

The C adapter opens nonblocking, then checks the opened descriptor for regular-file type and reads through that descriptor. Go docs recommend portable os APIs where possible and classify syscall as system-specific; nonblocking behavior needs a narrow target seam and macOS/Linux checks. [VERIFIED: examples/checksum/adapter.c:38-50, 68-82; CITED: pkg.go.dev/os, pkg.go.dev/syscall; LOCKED: 37-CONTEXT.md:35]

### Pitfall 4: Stat-size checks can miss file growth or race with replacement

The C adapter reads up to cap+1 from an already-open descriptor and distinguishes oversize from that read. D-11 forbids trusting Stat.Size or pre-stat/reopen; inject forged-size/read/close faults at the reader seam. [VERIFIED: examples/checksum/adapter.c:77-108; LOCKED: 37-CONTEXT.md:35, 46]

### Pitfall 5: Guide test can validate untracked inputs

The Phase 32 copier includes untracked nonignored files, so it cannot prove a tracked-only clean-checkout claim. Add a tracked-source copy path and make executable fence drift fail the focused guide test. [LOCKED: 37-CONTEXT.md:47; VERIFIED: cmd/schway/main_phase32_guide_test.go:347 and helper inspected]

## Prohibition Enforcement Constructibility

Phase 36 provides the minimal dependency-free precedent for D-23: a phase-local Node adapter routes closed fixtures through the existing projectProhibitions contract and the canonical check prohibition-enforcement producer. Its accepted proof binds exact source anchors and test roots, and checks one armed violation, clean, and default behavior; the producer reports the exact values “status=green”, “flagged=false”, and “failFirstProof=violation-fixture”. [VERIFIED: .planning/phases/36-scalar-helper-calls/36-PROHIBITION-EVIDENCE.md:8-16, 26-32]

Plan an early Phase 37 task to bind each applicable prohibition to a closed source/fixture selector, then run the canonical producer over bad, clean and default cases. Preserve unresolved or classifier history and fail closed on skipped/missing roots; do not count Go assertions or prose alone as machine-gate evidence. Keep the adapter and fixture set local to Phase 37 and dependency-free rather than changing generic GSD infrastructure. [LOCKED: 37-CONTEXT.md:50; VERIFIED precedent: .planning/phases/36-scalar-helper-calls/36-PROHIBITION-EVIDENCE.md:8, 28-32]

## Code Examples

### CRC consumer and independent vector

The locked source shape is one crc_byte helper called once per input byte, with eight LSB-first rounds and U64 state constrained to 32 bits. A temporary ordinary-source implementation using the current admitted operators and explicit if/else branches passed public check and native public execution for 123456789; output was `3421780262\n`. [LOCKED: 37-CONTEXT.md:19-21; VERIFIED: scratch source/check/build/run probe 2026-10-15]

~~~text
module lib.crc32

fn crc32(bytes: ByteView) -> U64 {
  // Seed, byte loop, one crc_byte call for each byte, final XOR.
}

fn crc_byte(mixed: U64) -> U64 {
  // Eight least-significant-bit-first right-shift/XOR rounds.
}
~~~

This is a shape sketch, not a formatter-canonical sample; exact API and algorithm values are in the verbatim locked decisions above. [LOCKED: 37-CONTEXT.md:19-22]

### Existing FileError order

The generated-C validator requires this exact ordered list:
~~~go
[]string{"FileTooLarge", "FileUnavailable", "AllocFailed"}
~~~
Repeat those values in source FileError and output tests. [VERIFIED: internal/compiler/cgen/cgen_program.go:3316-3320]

### Existing exit-code values

Opened source quotes:
~~~go
Success = 0
Unrepaired = 1
Invalid = 2
Operational = 3
Mismatch = 4
Usage = 64
AppInputRefused = 65
EventCapacity = 69
AppTimeout = 124
AppLaunchError = 125
SignalBase = 128
~~~
Preserve meanings and pin output streams beside status. [VERIFIED: internal/compiler/exitcode/exitcode.go:6-36]

## State of the Art

| Existing boundary | Phase 37 boundary | Impact |
|---|---|---|
| RunWithFileBytes accepts a caller-supplied reader and returns evidence-oriented Execution. | Public file mode adds one production reader and an ordinary application result with capture explicitly disabled. | Makes file execution public while keeping evidence-enabled callers compatible. [VERIFIED: internal/compiler/interp/interp.go:601-657; LOCKED: 37-CONTEXT.md:28, 37-39] |
| schway run --engine=... SOURCE has no input suffix and is conformance-oriented. | Add separate -- PATH grammar and preserve no-tail conformance form. | Makes path authority explicit and avoids consuming flag-shaped paths. [VERIFIED: cmd/schway/main.go:120-130; LOCKED: 37-CONTEXT.md:26-28] |
| Phase 32 guide copier accepts tracked and untracked nonignored files. | New guide verifier copies tracked files only. | Verifies commands from the claimed checkout surface. [LOCKED: 37-CONTEXT.md:47] |

**Deprecated/outdated:** No external library/API is replaced; no package is added. [LOCKED: 37-CONTEXT.md:22]

## Transition Review: Three useful next capabilities

| User-visible program | Current blocker | Smallest complete slice | Checker/guarantee changes | Evidence/debt | Owner / next action | Observation that changes priority |
|---|---|---|---|---|---|---|
| Bounded file CRC-32 and crc_low | Only reflected-bit helper witnesses exist; complete CRC module and importer pair absent. | One ByteView CRC module/private byte helper, two file entries and existing adapter. | Reuse checked scalar/call and file-owner laws; prove 32-bit invariant, numeric answers and shared use. | APP-13/MOD-07, Phase 36 call evidence, zlib pins and five reached controls. | Phase 37 planner/executors. | Discriminating vector or admitted-source witness demonstrates a missing semantic operation. [VERIFIED: 37-CONTEXT.md:113-116] |
| Public interpreter/native file run and CRC guide | Interpreter file read is Go-level only; current run grammar has no opaque path suffix; guide absent. | Fixed Go reader, suffix grammar, application output/outcome separation and tracked guide. | Full module admission; route-specific read cap; independent host implementation; no capture-dependent result. | APP-14/DX-24; current FileBytes/C adapter, Phase 32 guide and route tests. | Phase 37 alongside consumer. | A pinned tuple differs, a flag-shaped path is consumed, or maximum case cannot complete in a measured bounded lane. [VERIFIED: 37-CONTEXT.md:113-117] |
| Audited CRC/checksum on Ubuntu/macOS | No final M007 hosted binding; audit has not run after Phase 37. | Complete Phase 37, audit and close gaps, then bind both hosted lanes to one head/fixture set. | Same-source independent host results and non-skipped roots; local receipts do not substitute. | EVD-17; prior M006/Phase 36 receipts remain scoped. | Phase 38 after milestone audit. | Host drift, changed source or new audit defect invalidates freeze. [VERIFIED: 37-CONTEXT.md:117; REQUIREMENTS.md:84] |

### Dormant seed dispositions

Keep z1g dormant until a measured reviewer lookup fails; trz until an actual unexplained consumer failure; z1s until concrete paired source forms and independent reviewer evidence exist. Do not promote a query, observability or source-form framework without those triggers. [VERIFIED: 37-CONTEXT.md:119-123; .planning/sources/DASHBIT-AI-ERA.md inspected]

## Assumptions Log

| # | Claim | Section | Risk if wrong |
|---|---|---|---|
| A1 | Go environment probes on Apple arm64 do not establish correct syscall behavior on Linux. | Standard Stack / Environment | The reader may be target-specific; check both host lanes before claiming support. [ASSUMED] |
| A2 | A focused maximum-file test can finish within an acceptable local timeout. | Validation Architecture | If measured cost is too high, use a bounded focused lane or address the issue; do not invent a speed target. [ASSUMED] |

## Open Questions — RESOLVED planning decisions (2026-10-15)

All three questions below are resolved for plan acceptance. This resolves implementation strategy, not execution evidence; the original discovery questions remain visible. No local Linux, maximum-case latency, FIFO or descriptor behavior pass is claimed.

1. **D-10 empty-path behavior — RESOLVED:** The dated CONTEXT amendment preserves `(65, "", "")` on both public routes, without generated-C or old-test changes. Original discovery: source/test contradicted the copied locked sentence; letting empty input reach FileUnavailable was an alternative, now superseded. [VERIFIED: internal/compiler/cgen/cgen_program.go:1101-1104; cmd/schway/main_test.go:328-345; examples/checksum/adapter.c:76-80; scratch probe 2026-10-15]
2. **Go open flags on macOS/Linux — RESOLVED strategy:** Plans 37-02/04 use open-once nonblocking descriptor semantics, a narrow descriptor/read/close seam and isolated target-specific opening code only where needed. Real FIFO-without-writer, regular/symlink/nonregular and descriptor-replacement tests run under bounded subprocess/test deadlines; seam tests separately cover EINTR/short-read/fault precedence. Existing recurring macOS/Linux Go CI discovers these exact owners without a new workflow. Local macOS evidence and any unavailable host subcases stay labeled; Linux was unavailable locally, and the cross-host claim requires actual Linux lane evidence, with final hosted binding owned by Phase38. [LOCKED: D-11/D-19/D-21; planning decision, no new host execution]
3. **Maximum-case test budget — RESOLVED rule:** Plan37-02-01 first runs the actual 65,536-byte FF CRC with capture disabled under the existing 600s kill bound and records observed duration, root result and host/toolchain. Choose any tighter per-child budget only after that observation; Plan37-05-02 consumes it. Exceeding the existing bound is a failed acceptance check to investigate, not permission to drop the maximum case, invent a latency target or add generic fuel. The nine-byte scratch probe gives no maximum completion or cold/warm guarantee, and arbitrary timeout parity remains unproved per D-15. [LOCKED: D-14/D-15/D-21; planning decision, maximum case not yet executed]

## Additive source-inventory amendment — 2026-10-15

Source inspection during plan review found an additional D-21 safety obligation:
`session_phase34_prechange_test.go#phase34AdditionProblems` pins exactly three
Phase36 helper ledger rows, while
`session_phase36_prechange_test.go#phase36InventoryProblems` closes the entire
current tracked Schway inventory in both directions. Appending CRC rows to the
Phase36 ledger alone breaks its Phase34 consumer; leaving the new sources
unregistered breaks the Phase36 baseline once they are tracked.

Plan37-01 therefore owns a separate exact-three-path Phase37 ledger with actual
final source hashes/reasons, explicitly composed into both historical consumers.
Preserve the historical ledger/schema/hashes and both frozen corpora unchanged.
Create all three complete new sources before registering/staging them, retain
strict finite path checks and named old controls, and add reached
omission/hash/duplicate/unsafe/false-exemption controls in both owners. Capture
the existing four roots before additions and the six-root final owner set after
explicit owned-source staging. This is inspected source plus a resolved planning
mechanism; no new ledger or implementation test has run during this revision.

## Environment Availability

| Dependency | Required by | Available | Version | Fallback |
|---|---|---|---|---|
| Go | CLI, interpreter and focused probes | Yes | go1.24.0 darwin/arm64 | None; matches project bootstrap. [VERIFIED: environment probe; AGENTS.md:35-36] |
| Clang | Native C17 path and native consumer tests | Yes | Apple clang 21.0.0, Xcode clang generation 2100.1.1.22 | Interpreter-only work can proceed, but parity acceptance still needs native. [VERIFIED: environment probe 2026-10-15] |
| Python / zlib | Independent fixture generation only | Yes | Python 3.14.4; zlib compile/runtime 1.2.12 | Checked-in literals plus Go hash/crc32 cross-check; no runtime dependency. [VERIFIED: environment probe; LOCKED: 37-CONTEXT.md:43] |
| Linux host | Cross-host reader behavior | No | — | Use named Linux recurring-CI lane; do not claim local Linux execution. [VERIFIED: environment probe darwin/arm64; AGENTS.md:44-45] |

**Missing dependencies with no fallback:** None for local source/interpreter/native work. A Linux host remains necessary for the cross-host reader claim. [VERIFIED: environment probe; AGENTS.md:44-45]

## Validation Architecture

Nyquist validation is enabled in .planning/config.json; focused validation architecture is required. [VERIFIED: .planning/config.json:22-28]

### Test Framework

| Property | Value |
|---|---|
| Framework | Go testing package with existing repository selectors. [VERIFIED: cmd/schway/main_phase32_guide_test.go:347; internal/compiler/native/native_phase31_matrix_test.go:40] |
| Config file | go.mod; existing code uses Go toolchain. [VERIFIED: go.mod inspected] |
| Quick route-inventory command | GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase32PublicRouteAudit|TestPhase34SourceRoutePresentation|TestPhase34DiagnosticRouteInventory)$' -count=1 -timeout=600s |
| Focused interpreter baseline | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase31ChecksumInterpreter$' -count=1 -timeout=600s |
| Focused native baseline | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase31CaseMatrix$' -count=1 -timeout=600s |

These exact existing names are suggested pre-change baselines; no repository test selector was run during this research. [VERIFIED: internal/compiler/session/session_phase31_interp_test.go:62; internal/compiler/native/native_phase31_matrix_test.go:40; cmd/schway/main_phase32_routes_test.go:686; cmd/schway/main_phase34_routes_test.go:21,168]

### Phase Requirements → Test Map

| Requirement | Behavior | Test type | Automated command / selector | File exists? |
|---|---|---|---|---|
| APP-13 | Independent literal vectors including maximum, exact failure tuple and reached mutations. | Unit + public integration + mutation. | Add a shared phase-local consumer matrix and mutation controls; selector name belongs in plan. | No Phase 37 test yet. |
| MOD-07 | Both consumers import one library; library-only edit changes both outputs and module digest. | Integration + mutation. | Extend/add the existing single-copy pattern; select exact owner in plan. | Phase 31 pattern exists; consumers absent. |
| APP-14 | Opaque route, early dispatch, full admission, fixed reader, native manifest route and tuple parity. | CLI integration + reader fault/unit. | Add exact selectors for grammar, reader faults and public tuple rows. | Suffix route absent. |
| DX-24 | Every guide fence runs from tracked-only temp checkout; drift is detected. | Documentation integration + negative control. | Add Phase 37 guide runner and clean-checkout selector. | Phase 32 runner exists; new guide absent. |

Requirement wording comes from the source-of-truth table; new selector names are left to planning because they do not exist. [VERIFIED: .planning/REQUIREMENTS.md:66-80; 37-CONTEXT.md:47-50]

### Sampling Rate

- **Per task commit:** Run exact changed selectors using the specified GOCACHE; keep each command narrow. [LOCKED: 37-CONTEXT.md:48]
- **Per wave merge:** Run focused consumer, route and guide selectors; each engine must first compare independently to oracle rows, then to the peer engine. [LOCKED: 37-CONTEXT.md:44, 48]
- **Phase gate:** Focused tests/reached controls and named host limitations; hosted receipt binding remains Phase 38. [LOCKED: 37-CONTEXT.md:48; REQUIREMENTS.md:84]

### Wave 0 Gaps

- [ ] Add an independent literal CRC fixture with byte recipes, exact Python generation command and Python/zlib compile/runtime versions. [LOCKED: 37-CONTEXT.md:43]
- [ ] Add public shared interpreter/native tuple rows and named boundary/fault rows, including root-sensitive unreadable-file skip and nonblocking FIFO. [LOCKED: 37-CONTEXT.md:44, 46]
- [ ] Add reader seam tests for limit+1, nonregular refusal, forged file size, read/stat/close faults and once-only descriptor discharge. [LOCKED: 37-CONTEXT.md:35, 46]
- [ ] Add a capture-disabled scalar-call test proving no event, invocation identity or occurrence-counter growth while output, terminal result and cleanup continue; retain capture-enabled tests. [LOCKED: 37-CONTEXT.md:37-38; VERIFIED: interp.go:146-180, 1362-1378, 1695-1713]
- [ ] Add reached wrong-polynomial, non-reflected, missing-final-XOR, off-by-one and shared-library mutation controls. [LOCKED: 37-CONTEXT.md:45]
- [ ] Add tracked-files-only guide runner, exact command-stream checks, table/fixture consistency and armed drift control. [LOCKED: 37-CONTEXT.md:47]
- [ ] Add route/usage/surface inventory rows with finite census and negative controls; resolve empty-path conflict before tuple expectations. [LOCKED: 37-CONTEXT.md:30, 36, 48; VERIFIED: cmd/schway/main.go:720-721 and source/test conflict]

**Research execution evidence:** A temporary ordinary-source CRC witness passed public check and native build/run using the copied Phase 31 binding; its 123456789 run returned exit 0, stdout `3421780262\n`, empty stderr. The empty-path probe returned exit 65 with both streams empty. These were scratch probes only: no repository test, max-size case, full suite, broad native matrix, Linux lane, hosted receipt or sanitizer ran. [VERIFIED: commands executed 2026-10-15 on macOS arm64; LOCKED scope: 37-CONTEXT.md:11, 48]

## Security Domain

Security enforcement is enabled at ASVS Level 1 and blocks high findings. Exact config values: security_enforcement: true, security_asvs_level: 1, security_block_on: "high". [VERIFIED: .planning/config.json:49-51]

### Applicable ASVS Categories

| ASVS Category | Applies | Standard control |
|---|---|---|
| V1 Encoding and Sanitization | Yes | Keep the path out of shell interpretation and ensure OS calls receive the intended opaque argument. [CITED: OWASP ASVS 5.0.0 CSV V1.2.5; LOCKED: 37-CONTEXT.md:27] |
| V2 Validation and Business Logic | Yes | Enforce the grammar, NUL/length limits, exact owner shape, descriptor type and bounded read at trusted boundaries. [CITED: OWASP ASVS 5.0.0 CSV V2.2.1; LOCKED: 37-CONTEXT.md:27-30, 35-36] |
| V5 File Handling | Yes | Bind file access to the intended path and actual opened regular file; reject special files and bound bytes read. [CITED: OWASP ASVS 5.0.0 CSV V5.3.2; LOCKED: 37-CONTEXT.md:28-30, 35-36] |
| V6 Authentication | No | No user authentication surface is added. [INFERRED from 37-CONTEXT.md:8-10; ASVS chapter name CITED: OWASP ASVS 5.0.0 CSV] |
| V7 Session Management | No | No session or credential lifecycle is added. [INFERRED from 37-CONTEXT.md:8-10; ASVS chapter name CITED: OWASP ASVS 5.0.0 CSV] |
| V8 Authorization | Yes, at the source-capability boundary | The exact checked FileBytes owner path governs read authority; no arbitrary C execution or authority from a function name. [CITED: OWASP ASVS 5.0.0 CSV V8.2.1; LOCKED: 37-CONTEXT.md:28-30] |
| V11 Cryptography | No | CRC is an accidental-corruption checksum only; no cryptographic integrity or authentication claim. [CITED: OWASP ASVS 5.0.0 CSV V11.4.3; LOCKED: 37-CONTEXT.md:19] |

The sample V2–V6 labels in the generic research template use older ASVS chapter numbering. OWASP lists ASVS 5.0.0 as its latest stable version, where validation is V2, file handling V5, authentication V6, session management V7, authorization V8, and cryptography V11; use versioned category references in the security plan. [CITED: https://owasp.org/projects/asvs; https://github.com/OWASP/ASVS/raw/v5.0.0/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv]

### Known Threat Patterns

| Pattern | STRIDE | Standard mitigation |
|---|---|---|
| Flag/shell injection through path | Tampering / Elevation of Privilege | Dispatch before extractors; pass argv/path directly; prove flag-shaped and metacharacter paths stay opaque. [LOCKED: 37-CONTEXT.md:27; CITED: pkg.go.dev/os/exec] |
| FIFO or special-file blocking | Denial of Service | Nonblocking open, then descriptor fstat/type check before reading; test FIFO without blocking. [VERIFIED: examples/checksum/adapter.c:38-50, 77-82; LOCKED: 37-CONTEXT.md:35, 44] |
| Oversized/changing file | Denial of Service | Bound actual reads to cap+1 before allocation; no stat-size trust or reopen; keep cap route-specific. [LOCKED: 37-CONTEXT.md:35-36, 46] |
| Read authority through crafted source/core | Elevation of Privilege | Full module admission, exact ABI and owner-path oracle before pruning; negative controls for neighboring/raw-core shapes. [LOCKED: 37-CONTEXT.md:28-30] |
| Resource leak or false-success result | Denial of Service / Tampering | Preserve owner discharge and checked defects independently of evidence status; pin streams, exit and resource controls. [LOCKED: 37-CONTEXT.md:38, 45-46] |
| Private path disclosure | Information Disclosure | Fixed FileError stderr; do not echo application input path. [LOCKED: 37-CONTEXT.md:30] |

## Sources

### Primary
- AGENTS.md — bootstrap, dependency, portability, security and workflow constraints. [VERIFIED: AGENTS.md:29-47, 82-90]
- .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-CONTEXT.md — locked contracts, evidence requirements, transition priorities and seed dispositions. [VERIFIED: read in this session]
- .planning/REQUIREMENTS.md — APP-13, MOD-07, APP-14 and DX-24. [VERIFIED: .planning/REQUIREMENTS.md:66-80]
- internal/compiler/interp/interp.go, cgen/cgen_program.go, session/*, pathoracle/pathoracle.go, cmd/schway/main.go, examples/checksum/adapter.c — live source boundaries and current tests. [VERIFIED: cited ranges above]

### Secondary
- [Go os package documentation](https://pkg.go.dev/os) — portable open/stat/file descriptor APIs. [CITED: pkg.go.dev/os]
- [Go io.LimitReader documentation](https://pkg.go.dev/io#LimitReader) — bounded reader utility. [CITED: pkg.go.dev/io#LimitReader]
- [Go syscall package documentation](https://pkg.go.dev/syscall) — system-specific low-level API; prefer os where possible. [CITED: pkg.go.dev/syscall]
- [Go os/exec documentation](https://pkg.go.dev/os/exec) — exec does not invoke a shell or expand globs. [CITED: pkg.go.dev/os/exec]
- [OWASP ASVS 5.0.0 CSV](https://github.com/OWASP/ASVS/raw/v5.0.0/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv) — current validation, file-handling, authorization and cryptography chapter names and applicable L1 controls. [CITED: official OWASP ASVS v5.0.0]
- [RFC 1952 §8](https://www.rfc-editor.org/rfc/rfc1952#section-8), [PNG CRC algorithm](https://w3c.github.io/png/#crc-algorithm), [zlib checksum manual](https://www.zlib.net/manual.html#Checksum), [Python zlib.crc32](https://docs.python.org/3/library/zlib.html#zlib.crc32), [Go hash/crc32 source](https://go.dev/src/hash/crc32/crc32.go) — primary algorithm and answer references named by phase context. [CITED: 37-CONTEXT.md:91-93]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — project stack and local Go/Clang/Python versions checked; no package may be added. [VERIFIED: AGENTS.md:35-38; environment probes; 37-CONTEXT.md:22]
- Architecture: HIGH — CLI, interpreter, admission, native and generated-C seams inspected directly. [VERIFIED: cited source ranges above]
- Pitfalls: HIGH for empty-path contradiction and evidence auto-capture; MEDIUM for target-specific Go open flags pending both host lanes. [VERIFIED: source/test probes; CITED: pkg.go.dev/syscall]

**Research date:** 2026-10-15
**Valid until:** 2026-11-14 for stable source architecture; re-check official Go APIs/target behavior if toolchain baseline changes.

