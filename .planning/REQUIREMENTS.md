# Requirements: Schway M006

**Milestone:** Bounded Byte Utility and Local Modules
**Defined:** 2026-10-10
**Core Value:** Give an AI agent and a human reviewer the shortest reliable path from intent to sound, reproducible evidence without wasting iteration time or hiding runtime costs.

Research: [M006 synthesis](research/SUMMARY.md). Its "Resolved Decisions" table
records the auto-followed design choices and the alternatives that were
rejected. M005's acceptance record is preserved at
[milestones/M005-REQUIREMENTS.md](milestones/M005-REQUIREMENTS.md).

## M006 Requirements

### Exit classes and verification reports (M005 carry-over)

- [x] **DX-17**: Every source refusal exits 2 with the refusing layer's code: checker, core validator, origin validator, and entry resolution. Tool, IO, and timeout failures exit 3, and engine disagreement exits 4. `build`, `check`, and `run` report the same class and code for the same source, and so do `run --engine=interpreter` and `run --engine=native` (closes W2 and W3). Exit codes come from one named table instead of duplicated literals (closes N3).
- [x] **DX-18**: A program that contains functions unreachable from `main` runs instead of failing with `callgraph.entry_ambiguous` as an operational error. Unreachable functions are pruned before validation and emission. A genuinely ambiguous entry is a source refusal (exit 2). Programs whose functions are all reachable keep byte-identical generated C.
- [x] **EVD-13**: `app verify` keeps the named `native.event_capacity_exhausted` reason for every exhausted case, including `sum_to_n` at n = 173 and n = 1,000 and a set of 16 exhausted cases. The report stays within its size limit and is never counted as success (closes N1).

### Byte views

- [x] **BYTE-01**: Schway source can read the length of a read-only byte view and read the byte at an index, as an unsigned U64 value from 0 to 255. Checking, independent validation, interpretation, and C17 emission give these operations one defined meaning.
- [x] **BYTE-02**: Every byte read is bounds-checked at run time on every engine and optimization level, and no check is elided. An out-of-range index is a defined defect: exit 65, stderr exactly `schway: byte index out of bounds`, and empty stdout.
- [x] **BYTE-03**: A scalar loop function can take a read-only byte view parameter and iterate over it while carrying only U64 and Bool state. No loan, owner, or resource crosses a loop back edge. Each unsound neighbor is refused with its existing stable code: borrowing inside the loop, copying, storing or returning the view, releasing or moving an owner in the loop, and owner or loan parameters on a loop function.

### Checksum applications

- [x] **APP-10**: An `adler32` function written in ordinary Schway computes RFC 1950 Adler-32 over a byte view using only checked addition, comparison, and remainder. Interpreter and native results equal pinned literals for empty input, one `0x00` byte, `a`, `Wikipedia`, 4,096 `0x00` bytes, and 5,552 `0xFF` bytes.
- [x] **APP-11**: A caller can run the `checksum` application on one file path through the public application route and get the Adler-32 value as bare decimal plus a newline. Pinned cases: an empty file gives `1`, `Wikipedia` gives `300286872`, and exactly 65,536 `0xFF` bytes gives `2006388466`. The 65,536-byte case passes on every engine.
- [x] **APP-12**: `checksum` refuses a file of 65,537 bytes or more with its own named oversize outcome. Missing, unreadable, and non-regular files (directory, FIFO, device) share one named outcome. Every error leaves stdout empty, uses fixed stderr text that never echoes the path, and releases the owned buffer exactly once on every path. The size check uses a bounded limit+1 read and never trusts the reported file size.

### Local modules

- [x] **MOD-03**: A Schway program can import another local source file as a module through an explicit import declaration. Each file is one module, and the declared `module` name must match its path under the entry directory. Resolution never leaves the entry directory: no search path, `..`, absolute path, or symlink escape. *Amended 2026-10-12 (Phase 30 CONTEXT D-06):* the header/path check applies to imported modules. The entry module's own header is exempt, so existing programs without imports keep their headers and their generated C unchanged.
- [x] **MOD-04**: Only names listed in a module's `export` list are usable from importing modules. Using a private name is refused with a stable code, and an independent validator re-derives visibility. Cross-module calls are admitted from signatures alone, without reading the callee's body.
- [x] **MOD-05**: Each of these module-graph errors is refused with its own named, source-attributed diagnostic: import cycles (self, length 2, and length 3), missing modules, duplicate modules, header/path mismatch, and name collisions. A diamond import loads and emits its shared module once. A program without imports produces byte-identical C, and diagnostics in imported files name that file.
- [x] **MOD-06**: A second application imports the same `adler32` library and prints a different pinned result, without a copy of the library source. Mutating the library flips both applications' pinned outputs, and both build identities and evidence records hash every module in sorted order.

### Independent evidence and usability

- [x] **EVD-14**: Checksum and byte-view results are compared with independently pinned answers from Python `zlib.adler32`, with the generating command recorded. Every reached control is seen to fail:
  - a wrong modulus
  - wrong initial state
  - swapped halves
  - an off-by-one loop bound
  - signed-char widening (a `-fsigned-char` mutant)
  - a deleted bounds check (observed by ASan)
  - each refused loop neighbor
  - a library mutation
- [x] **EVD-15**: Hosted macOS and Linux receipts are taken at the final M006 revision and cover the M006 programs and the carried-forward M005 programs (closes N2). A later code change reopens the receipts.
- [x] **DX-19**: From a clean checkout, a developer can follow documented commands to build and run `checksum`, the second module consumer, and the `adler32` library. The guide states the expected results, the 65,536-byte limit, and the error outcomes. The provisional `examples/checksum.schway` frontier is retired with a dated amendment.

## Future Requirements

These capabilities are acknowledged but not part of M006 acceptance.

- Bitwise operators (xor, shift, and) and CRC-32 or POSIX `cksum` compatibility.
- Multiplication, subtraction, and general division beyond the operators already admitted.
- Byte mutation, writable views, slicing, and collections.
- Loans, owners, or resources carried across loop back edges.
- Module aliasing, re-exports, exported types and data across modules, and package versioning.
- Separate compilation and per-module interface artifacts (MOD-01, MOD-02, B1/DX-06).
- Stdin input, multiple file arguments, and multi-token application ABI.
- Loop-aware event capacity that lets `app verify` cover `sum_to_n` across its whole input domain (D-27.3-03), and app verify byte cases.
- M005 residuals N4 (guide n ≤ 14 note) and R1b (scalar first let then call).

## Out of Scope

| Capability | Reason |
|---|---|
| New third-party dependency, VM, or backend | Go stdlib, libc, and C17/Clang cover the scope; copy-local code beats a new dependency. |
| Iterator syntax such as `loop buffer { next(buffer) }` | It needs mutable cursor state across the back edge; an index loop over a read-only view does not. |
| Hex output or a filename suffix in `checksum` output | Bare decimal matches `zlib.adler32` and the existing `write_u64` output. |
| Search paths, environment roots, remote imports, package manager | Import resolution stays closed under the entry directory. |
| Editing the Phase 23 adapter, its receipts, or its observer tests | They pin historical evidence; M006 uses a separate copy-local adapter. |
| Describing modules as separate compilation or as closing B1/DX-06 | Whole-program composition does not establish a separately checked contract boundary. |

## Traceability

Each M006 requirement has one phase owner (see [ROADMAP.md](ROADMAP.md)).

| Requirement | Phase | Status |
|---|---|---|
| DX-17 | Phase 28 | Complete |
| DX-18 | Phase 28 | Complete |
| EVD-13 | Phase 28 | Complete |
| BYTE-01 | Phase 29 | Complete |
| BYTE-02 | Phase 29 | Complete |
| BYTE-03 | Phase 29 | Complete |
| APP-10 | Phase 29 | Complete |
| APP-11 | Phase 31 | Complete |
| APP-12 | Phase 31 | Complete |
| MOD-03 | Phase 30 | Complete |
| MOD-04 | Phase 30 | Complete |
| MOD-05 | Phase 30 | Complete |
| MOD-06 | Phase 31 | Complete |
| EVD-14 | Phase 31 | Complete |
| EVD-15 | Phase 32 | Complete |
| DX-19 | Phase 32 | Complete |

**Coverage:** 16 requirements; 16 mapped to Phases 28–32; 0 unmapped.

---
*Requirements defined: 2026-10-10*
*Traceability filled by the M006 roadmap: 2026-10-10*
