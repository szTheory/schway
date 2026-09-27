# Feature Landscape: Native Emission Ownership and Resource Discharge

**Project:** Codename Lang  
**Researched:** 2026-09-27  
**Scope:** M004 feature research, application dependency ladder, and subsequent product direction  
**Evidence boundary:** Current capability claims below come from source inspection. Acceptance criteria are proposed work, not claims that implementation or verification has passed. No programs or tests were executed for this research.

## Recommendation

M004 should deliver a real native application boundary and one useful ownership slice: a bounded file read produces a malloc-backed buffer whose ownership passes into Lang, survives observable use and function transfer, and is discharged on every admitted ordinary exit. Introduce the standalone scalar application smoke before the full resource chain. Preserve the milestone's native resource priority and completed Phase 21 work.

Resolve the three NAT-09 family assignments through separate, explicit admission decisions. Admit ordinary foreign resources and narrowly bounded shared/exclusive read-copy pointer shapes through the single whole-program emitter, with no `restrict`. Retain named refusals for every broader shape whose prerequisites remain unproved. Restoring historical emitter bodies or making an old fixture pass is insufficient acceptance.

The following milestone should finish iterative `sum_to_n` and FizzBuzz as standalone programs with externally supplied inputs. Their additional dependencies are defined integer arithmetic, Bool/comparison, usable conditional continuation, bounded scalar loops, fixed text output, and U64 formatting. A general String runtime, modules, generic collections, a VM, and a full standard library are unnecessary prerequisites.

This sequencing spends the current milestone on the user's explicit low-level commitments while opening a visible application path immediately. It also limits resource work: general FFI, arbitrary pointer operations, and every historical adversarial fixture do not become current scope merely because they share a backend family name.

## Current Inspected Capability

The current compiler performs genuine checked interpretation and native computation for its admitted source forms. Its public execution boundary still serves a conformance harness.

| Capability | What the inspected tree actually provides | Planning consequence | Evidence |
|---|---|---|---|
| Native CLI execution | `run` takes an engine and source path; there is no executable build command or user-argument route. | Add a public application build/run contract. | `cmd/lang/main.go:89`, `:470` |
| Session input | The session synthesizes Byte `7`, Buffer `01020304`, or each data alternative. Its linear-input switch does not admit U64. | Lower-layer U64 support does not establish a public numeric application. | `internal/compiler/session/session.go:943` |
| Interpreter input | Internal execution accepts decimal U64; payload alternatives receive canonical buffer/byte payloads. | Arbitrary external payload decoding is new work. | `internal/compiler/interp/interp.go:327` |
| Native entry | Generated C accepts decimal U64 internally; other supported inputs are selected from fixed values or alternatives with canned payloads. | Reuse valid scalar parsing where appropriate, but define the public entry adapter explicitly. | `internal/compiler/cgen/cgen_program.go:863` |
| Application output | Generated `main` emits execution JSON. The emitter currently ignores its `executionJSON` parameter. | Ordinary output requires a real mode boundary and separate evidence channel. | `internal/compiler/cgen/cgen_program.go:558`, `:889` |
| Executable lifetime | The runner creates a temporary C file and executable and removes the directory after the run. | A retained, directly executable build artifact is a distinct deliverable. | `internal/compiler/native/native.go:457` |
| Process observation | The runner captures stdout as strict execution JSON and rejects successful-process stderr. | Program stdout, stderr, exit status, and evidence need independent contracts. | `internal/compiler/native/native.go:585`, `:635` |
| Native run semantics | The session interprets, executes `-O0` and `-O3`, compares them, and returns interpreter executions through the command result. | Real user effects must not inherit this repeated-execution behavior. | `internal/compiler/session/session.go:1057`, `:1086`, `:1235` |
| Foreign interpretation | A foreign operation copies its source value to the target and follows the success edge; an enabled nonlocal-exit policy triggers a synthetic second-call outcome. | The interpreter is not a real host-I/O oracle. A deterministic host model/replay contract must be named separately. | `internal/compiler/interp/interp.go:962` |
| Foreign source surface | Foreign arguments must be the containing function's original parameter; `Function` carries one foreign contract. | Usable acquire/use/transfer sequences require the relevant operand and per-call contract generalization. | `internal/compiler/check/check.go:3972`; `internal/compiler/core/core.go:148` |
| Foreign linkage | Four known fixture symbols resolve to repository C files. | A fresh application's explicit foreign build inputs must be supported without adding its symbol to a compiler switch. | `internal/compiler/native/foreign_nonlocal.go:27` |
| Resource tracer | `lang_res_open` allocates and frees inside the C wrapper, returning the request byte. | This tracer does not establish ownership of a live allocation in Lang. | `native/lang_foreign_resource.c:34` |
| Current native resource admission | Whole-program emission refuses foreign bodies/contracts and the cut pointer families. The live-resource result is empty for currently admitted shapes. | Phase 21's contract is a prerequisite, not proof that runtime ownership is implemented. | `internal/compiler/cgen/cgen_program.go:63`, `:586` |
| Scalar/control surface | U64 constants exist. The token and operation inventories contain no arithmetic or loop forms; the core function still has one parameter. | Do not describe M003's computation work as sufficient for FizzBuzz. | `internal/compiler/syntax/token.go:7`; `internal/compiler/core/core.go:127`, `:629` |

The historical maturity document contains stale expressiveness claims, including calls that supposedly cannot run and an inventory without current U64 support. Use this inspected baseline when updating it. Phase 21 is complete and archived; `.planning/STATE.md:106` is the current handoff, while several prospective passages in `.planning/PROJECT.md` still describe earlier milestone assumptions.

## Table Stakes for M004

These features make the selected application and ownership claim complete. Labels in this research table are planning names, not newly ratified requirement IDs.

| Feature | Why expected | Complexity | Proposed acceptance boundary |
|---|---|---|---|
| Retained native executable | A program author needs a result usable outside the compiler harness. | Medium | A public build command produces a named executable, which runs after compilation finishes and from a different working directory. Its runtime dependencies are declared. |
| Explicit bounded application input | Generated programs must respond to user data. | Medium | The first scalar smoke accepts at least two distinct valid external values and rejects malformed/out-of-range input predictably. The resource application receives an explicit bounded path or selected input capability, with decoding and size limits documented. |
| Compile/build/run once | Real external effects cannot be repeated automatically for differential comparison. | Medium | Building performs no application I/O; running executes the application once. Any replay or optimizer-tier comparison is an explicit verification operation over controlled inputs. |
| Ordinary process outputs | CLI programs must compose with shell tools and other processes. | Medium | Application stdout and stderr are byte streams; exit behavior is specified. Compiler diagnostics, application failure, process signals, and evidence capture remain distinguishable. |
| Separate optional evidence transport | JSON protocol output must not corrupt the application's own output. | Medium | Application bytes that resemble JSON remain application bytes. Evidence goes to an explicit separate destination. Disabled or exhausted evidence collection has declared behavior and cannot silently change an application result. |
| Lang-owned file-read buffer | Ownership must cover a live, useful native resource. | High | An audited bounded file-read adapter returns a malloc-backed allocation and length into Lang. Generated cleanup frees it after the last owner finishes; it remains alive while Lang uses it. |
| Ownership transfer across functions | Local-only cleanup cannot establish a usable ownership model. | High | A caller transfers a buffer to a callee or receives an owned buffer back; exactly one owner remains responsible. A moved-from use is refused. A return transfer does not free the allocation before the receiver uses it. |
| Borrowed read-copy access | The owner must support useful observation without duplicate ownership. | High | Shared and exclusive read-copy cases are separately admitted and evidenced. Borrowed access cannot release or outlive the owner, and copying observed bytes does not copy the release obligation. No `restrict` is emitted. |
| Defined ordinary failure cleanup | Failed reads and later failures are everyday outcomes. | High | Failure before acquisition has no phantom release; partial adapter acquisition is cleaned at its boundary; failure after completed acquisition releases each still-owned allocation once. Transfer changes the responsible owner explicitly. |
| Fresh foreign build inputs | Programs must not depend on compiler-recognized fixture names. | Medium | Declared audited support code can be compiled and linked through the public build path. Input identities participate in relevant build invalidation. Runtime success does not require the compiler's source checkout. |
| Structured diagnostics and repairs | The language's author is expected to work through tools. | Medium | Invalid transfer, expired borrow, unsupported pointer extension, and malformed application input produce the appropriate stable structured response. Compiler repairs remain scoped to source defects. |
| Exact host evidence | Initial host priorities are macOS and Linux. | Medium | Each reopened pointer family meets its own cross-host prerequisite. A missing host result remains explicitly unproved rather than being inferred from the other host. |

The file descriptor used internally by the file-read adapter can remain adapter-owned and closed before the call returns. M004's Lang-owned resource is the returned allocation. State this distinction explicitly; the selected slice does not establish a general Lang file-handle API.

## Proposed User Stories and Acceptance Cases

### 1. Build and run a scalar program before resource integration

As an author, I can build a checked Lang source file into an executable, pass a bounded number to it, and observe its result through ordinary process output.

Proposed evidence:

- Two runtime inputs produce their corresponding results without editing or rebuilding the source.
- Running the retained artifact does not invoke the Lang compiler or consult the compiler checkout.
- Invalid input produces the documented application-boundary failure; it is not a native execution-document parse failure.
- Application stdout/stderr and an optional evidence destination can be observed independently.
- The scalar smoke is described as an application-boundary demonstration. Identity or constant return does not establish arithmetic capability.

### 2. Read real file bytes into a Lang-owned buffer

As an author, I can select a bounded file, obtain its bytes as an owned buffer, pass or borrow that buffer through Lang functions, observe the bytes, and finish without leaking or freeing them twice.

Proposed evidence:

- Distinct file contents produce distinct observed bytes. Include empty input, embedded zero bytes, a size boundary, and a refused oversized input.
- An allocation remains physically live after the foreign call returns and throughout the admitted borrow/use interval.
- A Lang-directed transfer changes the owner while preserving the allocation and contents. Cleanup frees the allocation at its final admitted release point.
- A read error and a later modeled operation failure exercise different cleanup paths. An output failure cannot bypass cleanup of an owned buffer.
- The adapter's own partial failure cleanup is distinguished from Lang's cleanup after a successful acquisition.
- An omitted release, duplicate release, premature release, or wrong-resource release cannot be hidden by a matching synthetic event stream.

Fix the maximum input size and empty-buffer representation in the phase contract. An empty result may have no allocation or a valid owned allocation, but ownership, dereference, and release behavior must agree across the checker, runtime boundary, interpreter model, and native implementation.

This demonstrates useful native orchestration and resource control. The adapter may implement bounded file acquisition and byte output; it must not hide the complete acquire/use/release lifetime that the Lang demonstration claims to prove.

### 3. Use shared and exclusive read-copy views within their own contracts

As a reviewer, I can tell which aliasing shapes are admitted and see why the generated C has only the guarantees justified by those shapes.

| NAT-09 family | M004 consumer | Required admission result | Continuing refusal |
|---|---|---|---|
| D-16-11, former `emitLinearForeign` | File-read allocation, Lang ownership transfer, ordinary success/error cleanup | Real emitted acquisition/use/discharge through `emitProgram`, with independent physical observations and checked ownership facts | Opaque nonlocal transfer, unsupported unwinding/cancellation, and foreign shapes outside the selected contract |
| D-16-12, former `emitLinearBorrowedByPointer` | Exclusive borrowed view used only for the admitted read-copy operation | Exact-shape ownership/discharge evidence on macOS and Linux; no `restrict`; no inference from a historical restrict probe | General pointer mutation, retention, escaping views, and broader alias guarantees |
| D-16-13, former `emitLinearBorrowedByPointerPlain` | Shared borrowed view used only for the admitted read-copy operation | Its own shared-alias and lifetime evidence on macOS and Linux; no substitution of exclusive-family evidence | Shared access combined with unsupported mutation/retention or owner destruction |

These are semantic family dispositions. The retired emitter implementations stay retired; the production path remains `emitProgram`. Each narrowed admission must have a reachable unsupported-neighbor fixture so widening cannot accidentally erase its refusal boundary. Phase 21's archived resource-discharge contract remains the source for exit classifications and prerequisites; widening requires an explicit successor contract and executed evidence.

### 4. Author and repair through the shipped interface

As an AI author, I can learn the admitted syntax from a compact current guide, build the selected application, and repair a seeded ownership mistake through structured feedback. As a human, I can review the resulting source and understand the resource lifetime and observable behavior.

Run a bounded fresh-agent exercise once a complete application slice exists. Give it the shipped CLI, current syntax documentation, examples, and a task description. Record successful task completion, repair rounds, diagnostics consulted, tool calls, and cold/warm timings. Include at least one held-out resource-flow variation. Treat the result as evidence about the sampled tasks and agent, not proof of universal agent productivity.

## Differentiators

| Feature | Value proposition | Complexity | Proposed observable evidence |
|---|---|---|---|
| Human-auditable lifetime in small canonical source | A reviewer can locate acquisition, transfer/borrowing, error propagation, and discharge responsibility. | Medium | A short file-read example and a source-linked ownership explanation agree with the emitted behavior. |
| Checked resource meaning across native and model boundaries | Generated C does not invent stronger ownership or alias guarantees. | High | Physical resource controls and independent ownership checks catch different failures; each claim names which observation establishes it. |
| Useful structured failure | Agents can repair source without reverse-engineering prose or C failures. | Medium | A held-out source defect is diagnosed and repaired through the supported protocol, then rerun through the application path. |
| Explicit evidence cost | Ordinary iteration stays practical while expensive checks remain available. | Medium | Report cold/warm end-to-end build/run and repair distributions with host, toolchain, workload, and lane scope; distinguish check, build, run, and verification costs. |
| Precise refusal boundaries | A growing language remains predictable to agents and reviewers. | Medium | Every deferred shape has a stable explanation, owner, and unblocking capability, rather than an accidental backend failure. |

These differentiators implement the intent-to-evidence loop in `wiki/vision.md:16` and the cost doctrine in `wiki/compute-efficiency-constitution.md:23`. They should be exercised on user-facing application tasks rather than counted only as internal mechanisms.

## Feature Dependencies and Delivery Order

```text
Checked core + current C17 backend
  -> application entry/input/output/exit contract
  -> retained build artifact + compile/build/run-once path
  -> scalar standalone smoke

Phase 21 discharge contract + explicit foreign build inputs
  -> real file-read allocation represented as Lang-owned value
  -> per-call resource operands/contracts + owner transfer
  -> shared/exclusive read-copy views under separate narrow contracts
  -> acquire/use/release + failure/transfer application evidence

Standalone application boundary
  + U64 arithmetic + Bool/comparison + conditional continuation
  + bounded scalar loop + checker/event changes
  -> iterative sum_to_n
  + remainder + fixed text/byte output + U64 formatting
  -> FizzBuzz

Resource buffer slice + scalar iteration
  + byte length/index/bounds + EOF/error behavior + useful call arity
  -> checksum and byte-oriented tools
  -> reusable local libraries and typed JSON work

Resource/byte I/O + explicit network authority + bounded transport adapter
  -> blocking HTTP client
  -> optional concurrent service/runtime work
```

The scalar smoke should not wait for the ownership chain. Resource implementation and application-boundary implementation can use separate internal seams, but milestone completion requires their composed public application. The numerical milestone should follow immediately; broader pointer work must not consume its place by default.

## Capability Ladder After M004

| Horizon / capability | Demonstrator and user-visible gate | New dependencies and checker work | Capabilities not required |
|---|---|---|---|
| M004 application boundary | Build once, run retained scalar artifact with different external inputs, observe ordinary output | Entry ABI, bounded input decoding, process outputs, distinct evidence transport | Arithmetic, general String, VM |
| M004 owned native buffer | Read external file bytes, use/transfer/borrow in Lang, release on admitted ordinary exits | Physical allocation identity, length/initialization, owner transfer, borrow lifetime, per-call FFI facts, checked discharge | Arbitrary pointer arithmetic, `restrict`, general Lang file handles, tracing heap |
| Following milestone: iterative `sum_to_n` | External `n` produces the declared sum; Lang performs the iteration | Arithmetic/overflow contract, Bool/comparison, continuation/join semantics, loop-carried scalar values, supported back-edge checker | General aggregates, arity-N, resource-bearing loops, full stdlib |
| Same following milestone: FizzBuzz | External limit drives ordinary line output with correct number/Fizz/Buzz/FizzBuzz decisions | Remainder, branching, fixed byte/text literals, U64 decimal formatting, bounded writes | Owning String runtime, Unicode manipulation, modules, collections |
| Byte-oriented tools | A declared checksum of real external bytes, plus a second byte-processing consumer | Byte views, safe indexing, conversions, streaming/EOF/error contract, resource-bearing iteration where needed; arity-N as demanded by real APIs | JSON, HTTP, registry, general effect rows |
| Reusable local libraries | Two local consumers share byte utilities without copying source; an edit invalidates relevant facts | Module/name boundaries, explicit dependency/build inputs, reusable contracts; separate compilation/evidence reuse receives its own gate | Package registry, remote cache, ecosystem-scale version management |
| Typed JSON | External JSON is parsed/validated, changed by Lang code, and emitted with bounded resource use | Records/variants, text validation, escapes, numeric/duplicate-key policy, bounded nesting, allocation or bounded storage, precise parse errors | HTTP, scheduler; language recursion if an explicit parser stack suffices |
| Bounded networking | A blocking client fetches a bounded response using an audited transport boundary | Byte I/O, timeout/failure contract, resource cleanup, explicit network authority; TLS/provider work scoped to the selected client | JSON, actors, async scheduler, VM |
| Concurrent services and ecosystem | A selected real application justifies each added facility | Task/cancellation ownership, deadlines, bounded queues, operational observation; then packaging, tooling and broader libraries | No automatic commitment to a global tracing collector or a new execution engine |

The `sum_to_n` gate must require iteration; a closed-form expression would not exercise the back-edge work. The checksum gate must declare its algorithm and overflow behavior. The existing `examples/checksum.lang` is a refused frontier artifact; `next`, `add`, and the loop in that artifact do not settle EOF, conversion, or accumulator semantics.

JSON and HTTP are separate branches once byte I/O is usable. JSON can run over local files/stdin, while an HTTP client can return bytes before a JSON library exists. Local modules improve reuse; a single-file JSON parser is possible before full separate compilation. The official JSON specification defines a data format with strings, numbers, arrays, objects, booleans and null, and discusses numeric and Unicode interoperability; it does not prescribe a Lang module system or networking runtime. [RFC 8259](https://www.rfc-editor.org/rfc/rfc8259.html)

## Checker Work to Surface Before Numerical Iteration

The next milestone must budget these changes explicitly rather than discovering them during loop syntax implementation:

1. **Loop state and joins:** initialization, mutation or value flow, and invariants for values crossing the header and exit. Start with copyable scalar state; refuse unsupported ownership-bearing loop state by name.
2. **Independent back-edge admission:** `pathoracle` currently refuses cycles (`internal/compiler/pathoracle/pathoracle.go:130`, `:237`). A bounded loop subset needs a sound independent checking strategy; bounded execution alone cannot prove safety for all admitted executions.
3. **Loan/resource extension trigger:** once borrows or owned buffers cross/recur across the edge, revisit their liveness, transfer, and cleanup rules. The first scalar loop milestone must not claim this broader result.
4. **Dynamic event occurrences:** repeated operations and repeated calls within a loop require an identity model that distinguishes occurrences and can be checked independently.
5. **Evidence capacity:** native invocation/event capacity currently derives from static acyclic structure (`internal/compiler/cgen/cgen_program.go:16`, `:149`). An ordinary long-running loop must not silently inherit a fixed fixture-sized trace assumption.
6. **Budget outcome:** distinguish application semantics from exhausted verification steps, trace capacity, or incomplete replay. Make incompleteness observable.
7. **Defined numeric lowering:** choose overflow, remainder/division-by-zero, comparison, and conversion behavior before emitting C or optimizer guarantees.

## Alternatives Considered

| Alternative | Strongest argument | Opportunity cost / risk | Decision |
|---|---|---|---|
| Computation first | Fastest direct path to useful algorithms and an AI-authoring task with clear expected output. | Defers the user's low-level resource commitment again and postpones the actual host-effect boundary. | Keep as the explicit counterargument; adopted scope honors resources now and makes numerical applications the immediate next gate. |
| Restore all pointer/foreign families | Retires old backend debt in one campaign. | Historical fixtures can dominate the roadmap; broad aliasing, retention, unwinding and platform work expand without a selected application. | Reject blanket restoration. Resolve each NAT-09 family through the selected narrow consumer and its own evidence. |
| Resource contract only | Lowest implementation risk and easy structural validation. | Phase 21 already established this layer. Another contract-only milestone would leave the promised native behavior unexercised. | Reject as M004 completion. |
| Application output through execution JSON | Reuses the current strict harness and comparator. | Prevents ordinary stream composition and conflates verification with user effects. | Retain for explicit verification; add ordinary process behavior. |
| Full String/stdlib before FizzBuzz | Produces reusable text APIs early. | Delays a cheap, meaningful numerical application behind unrelated ownership and library design. | Use fixed byte/text literals and U64 formatting first. |
| VM or bytecode now | Could eventually improve interactive execution or portability. | Adds another engine while input/output, effects, and surface gaps remain. | Continue C17 native compilation and the reference interpreter; reconsider only against measured requirements. |

Native applications need appropriate startup, I/O and allocation support. They do not inherently require a VM. Lang's own runtime direction already identifies the interpreter as a conformance/development engine (`wiki/runtime-profiles-and-dogfooding.md:25–54`). As external precedent, Rust separates its platform-independent core from hosted startup and OS facilities. This supports separating runtime facilities from the choice of execution engine; it does not prove Lang's current support is complete. [The Embedded Rust Book: no_std](https://doc.rust-lang.org/stable/embedded-book/intro/no-std.html)

Clang's documented toolchain includes compilation, object generation, linkage, and target-dependent runtime support. A retained native artifact fits the existing backend direction; the missing Lang work is the application contract and build product around it. [Clang: Assembling a Complete Toolchain](https://clang.llvm.org/docs/Toolchain.html)

## Anti-Features for This Milestone

| Anti-feature | Why avoid now | Do instead |
|---|---|---|
| Automatic real-I/O replay across optimizers | Can repeat writes or other external effects and cannot use synthetic foreign results as an oracle. | Run user applications once; use deterministic controlled worlds or explicit replay for verification. |
| A release event as the sole resource proof | Serialization can agree while the wrong allocation is freed or none is freed. | Couple checked obligations to independent physical acquire/use/release observations. |
| A C helper owning the entire claimed Lang lifetime | A wrapper can allocate, consume, and free internally while Lang only sees a byte. | Return a live allocation with an actual Lang ownership obligation. |
| Compiler knowledge of each application's symbol names | Turns new programs into compiler modifications. | Explicit bounded foreign build inputs and per-call contracts. |
| `restrict` or blanket no-alias promises | Narrow borrow admission does not automatically justify every C optimization contract. | Emit conservative pointers for the selected read-copy families. |
| General pointer mutation, escape, retention or arithmetic | Broadens the semantic and host evidence burden beyond the consumer. | Named refusals and explicit later unblocking cases. |
| General unwind/cancellation semantics | Changes the exit and resource model substantially. | Preserve Phase 21 classifications and guarantee cleanup only for admitted exits. |
| Full String runtime, generics, package ecosystem or VM | None is needed for the selected resource consumer or next numerical demos. | Add the smallest coherent reusable capability when its consumer arrives. |
| Closure through documentation alone | Names and completion counts can outrun execution capability. | Attach each product capability to a source program, public command, observable output, and current evidence status. |

## MVP Recommendation and Living Roadmap Contract

Prioritize in this order:

1. Build/run-once application contract and retained scalar smoke artifact.
2. Genuine malloc-backed file-read buffer, explicit foreign linkage, and Lang ownership transfer.
3. Separately justified shared/exclusive read-copy access and ordinary failure cleanup.
4. Composed public application evidence, narrowly scoped cross-host gates, and fresh-agent authoring/repair observation.

Carry the following numerical milestone as a concrete acceptance promise: standalone iterative `sum_to_n` and FizzBuzz. Carry later byte libraries, JSON, networking and ecosystem work as capability-led directions, revised when their consumers expose real constraints. Avoid assigning new long-range milestone numbers in this research.

Each roadmap capability should carry: current status; example source; public command; expected observable behavior; first refusal/blocker; owning phase or future capability; and evidence scope. Rebind the old loop/aggregate/module debt destinations prospectively and preserve their historical records. A module requirement for separate-compilation blame is different from the basic local-reuse feature and should retain its specific trigger.

Reassess after each application gate: can an external input change the result, can the artifact run independently, does the Lang code perform the claimed work, can a fresh author recover from a relevant defect, and what does the full feedback loop cost? These questions should change the roadmap when their answers contradict the assumed dependency ladder.

## Sources and Remaining Uncertainty

**Local implementation evidence:** `cmd/lang/main.go`; `internal/compiler/session/session.go`; `internal/compiler/interp/interp.go`; `internal/compiler/cgen/cgen_program.go`; `internal/compiler/native/native.go`; `internal/compiler/native/foreign_nonlocal.go`; `internal/compiler/check/check.go`; `internal/compiler/core/core.go`; `internal/compiler/syntax/token.go`; `internal/compiler/pathoracle/pathoracle.go`; `native/lang_foreign_resource.c`; `examples/checksum.lang`. Specific anchors accompany the findings above. Line numbers refer to the inspected pre-M004 implementation and may move.

**Project commitments and design intent:** `.planning/STATE.md:106`; `.planning/PROJECT.md:194`; `.planning/LANGUAGE-MATURITY.md`; `.planning/research/M003/ROADMAP-STRATEGY.md`; `.planning/milestones/M003-REQUIREMENTS.md:110`; `.planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-CONTEXT.md:9`; the adjacent `21-RESOURCE-DISCHARGE-CONTRACT.json`; `wiki/vision.md:16`; `wiki/compute-efficiency-constitution.md:23`; `wiki/runtime-profiles-and-dogfooding.md:25`; `wiki/syntax-and-cognitive-ergonomics.md:16`. Historical strategy and candidate wiki designs are context, not implementation evidence.

**Primary external sources consulted:** [Clang toolchain documentation](https://clang.llvm.org/docs/Toolchain.html), [Rust's no_std explanation](https://doc.rust-lang.org/stable/embedded-book/intro/no-std.html), and [RFC 8259](https://www.rfc-editor.org/rfc/rfc8259.html). The documentation pages were read on 2026-09-27; RFC 8259 is the December 2017 standard. These support the narrow precedent/format statements attached to them, not Lang completion claims. Direct official-source lookup was used under the orchestrator's scoped research instructions.

The inspected gaps have direct code evidence. Delivery sizing, the exact public input syntax, the owned-buffer ABI, empty-buffer policy, instrumentation transport, and the minimum loop checker remain phase-design questions. No evidence in this research establishes that the proposed applications already build, that the pointer families pass either host lane, or that a fresh agent can yet author them successfully.
