# Requirements: Schway M007

**Milestone:** Bitwise Checksum and Composable Scalar Helpers
**Defined:** 2026-10-13
**Core Value:** Give an AI agent and a human reviewer the shortest reliable path from intent to sound, reproducible evidence without wasting iteration time or hiding runtime costs.

Research: [M007 fan-out synthesis](research/M007/SUMMARY.md). Five lenses
(product, compiler/PL, security, devops/evidence, DX), each with an
adversarial pass, were synthesized and auto-followed on 2026-10-13. M006's
acceptance record is kept at
[milestones/M006-REQUIREMENTS.md](milestones/M006-REQUIREMENTS.md).

## M007 Requirements

### Carried debt and hardening (first code-moving phase)

- [ ] **QLT-13**: The native app reads the evidence capture file through a reader bounded at limit+1 bytes, before it allocates anything. An oversize capture yields the named capacity outcome without allocating the full file. A seeded mutant that restores the unbounded read is caught.
- [ ] **QLT-14**: The 32.1 review warnings are closed:
  - WR-01: the freeze gate gives the same verdict from any working directory, with a subdirectory test.
  - WR-02: `isAbortLikeDefect` accepts only the abort-like defect reasons it names, with the live owner checked, and a mutant that widens it is caught.
  - IN-01..03 and the deferred coverage-`kind` pin are each fixed or dispositioned in writing.
- [ ] **QLT-15**: Phase 29 WR-02 and WR-03 are triaged, and each is either fixed with a test or recorded with a reason. Phase 28 gains VALIDATION.md and SECURITY.md built from existing evidence.
- [ ] **EVD-16**: One milestone-agnostic freeze and receipt gate replaces the phase-specific `check-phase32-*` scripts. It takes the milestone's program list and codedigest as input, keeps all existing refusal behaviour, and needs no per-phase copy in later milestones.

### Diagnostics for agent authoring

- [ ] **DX-20**: Text diagnostics show `file:line:col` and a one-line source snippet with a caret. JSON diagnostics add 1-based `line` and `col` to each span, next to the existing byte offsets. Codes are unchanged.
- [ ] **DX-21**: A single syntax error yields one diagnostic for its recovery point, not a cascade of follow-on `expected }` / `expected declaration` errors. A pinned corpus of agent-style drafts shows the reduction.
- [ ] **DX-22**: Constructs that remain absent are refused with a dedicated `surface.unsupported_*` code and a one-line hint. They include `&&`/`||`/`!`, more than one parameter, if without else, Bool literal tails, subtraction and dynamic shift amounts. The previous misleading parse errors no longer appear.
- [ ] **DX-23**: `schway surface --json` prints a versioned, deterministic manifest of the admitted surface: types, operators, statements, builtins, call rules and limits. A meta-test fails if an admitted construct or `surface.unsupported_*` code is missing from the manifest, or if the manifest names one that is not admitted.

### Bitwise operators and literals

- [ ] **BIT-01**: U64 `^`, `&` and `|` are admitted in scalar code. Interpreter and native agree with literal expected answers at -O0, -O3 and -O3 with LTO. Each peer (check, corevalidate, originvalidate, executionpeer) derives the ops independently, and a wrong-result mutant per op is caught.
- [ ] **BIT-02**: U64 `<<` and `>>` are admitted only when the shift amount is an integer literal from 0 to 63.
  - Amounts of 0, 31, 32 and 63 give pinned answers on both engines and at every optimization level.
  - An amount of 64 or more, or a non-literal amount, is refused at check, core and origin with a named code. Generated C never contains an unguarded shift.
  - `<<` drops high bits by definition, which the manifest and guide document.
- [ ] **BIT-03**: Hex integer literals (`0x…`, case-insensitive digits) are admitted wherever decimal U64 literals are. Overflowing or malformed hex is refused with a named code, and the formatter keeps the source spelling.
- [ ] **BIT-04**: `examples/checksum/adler32.schway` uses `b << 16` in place of 16 doublings. All of its pinned answers, and the checksum and `adler_low` outputs, stay byte-identical on both engines.

### Scalar helper calls

- [ ] **CALL-01**: A scalar function can call another scalar function with one U64 or Bool argument and use the U64 or Bool result in an expression or assignment. This includes calls to imported module functions. Recursion stays refused by the call graph. Callee defects (overflow, remainder, bounds) surface with the callee's reason and exit 65 on both engines.
- [ ] **CALL-02**: Calls are derived independently by check, corevalidate, originvalidate and the execution peer. Mutants that drop the argument, the result or a callee defect are caught. Programs without scalar calls keep byte-identical generated C.

### CRC-32 consumer

- [ ] **APP-13**: `examples/crc32/crc32.schway` reads a file of at most 65,536 bytes and prints its CRC-32/ISO-HDLC as bare decimal. It is built from an ordinary-source `crc32` module that uses a bit-at-a-time loop with no table.
  - Pinned answers:
    - empty input gives 0;
    - one `0x00` byte gives 3523407757;
    - `123456789` gives 3421780262 (0xCBF43926);
    - 65,536 bytes of `0xFF`;
    - 65,537 bytes is refused with the existing oversize outcome;
    - a missing or unreadable file gets the shared outcome.
  - Answers are checked against Python `zlib.crc32` with armed controls.
  - The guide states that this is not POSIX `cksum` (CRC-32/CKSUM).
- [ ] **MOD-07**: A second consumer imports `crc32` from the module without copying it and prints a pinned answer on both engines.
- [ ] **APP-14**: `schway run --engine=interpreter` runs file-bytes programs such as `crc32` and `checksum` from the CLI, and gives the same stdout, stderr and exit code as native for every pinned case. Interpreter/native parity for file input is no longer Go-level only.
- [ ] **DX-24**: `examples/crc32/README.md` is an executable clean-checkout guide, run by a test the way `TestPhase32GuideCleanCheckout` runs its guide.

### Evidence

- [ ] **EVD-17**: Hosted ubuntu and macOS receipts for every M007 and earlier pinned program are bound once to the final M007 codedigest, after the milestone audit and in its last plan. No per-phase hosted receipt binding is required.

## Future Requirements (named landing)

- **M008:** arity-N functions, records with only scalar fields, exported module types, and a `wc` + `longest_line` consumer pair. These are the first two prerequisites for JSON.
- **After M008:** a bounded key=value config parser (a record holding a view, errors carrying a line number), then a bounded JSON config validator with the JSONTestSuite y_/n_ subset as gates.
- Reducing the foreign-adapter boilerplate needed to run byte programs (about 40 lines per program).
- Checked U64 subtraction, once a consumer needs it.
- Module-aware `evidence`, `interface` and `query`.

## Out of Scope

| Feature | Reason |
|---------|--------|
| HTTP client/server | Framing differentials and request smuggling; network I/O cannot be digest-bound or deterministically double-engined; already out of scope in PROJECT.md |
| JSON in M007 | Needs records, arity-N and a bounded parser stack; staged after M008 |
| Dynamic shift amounts | Would add a defect path and C17 UB guard; literal-only covers CRC-32 |
| CRC lookup table / POSIX `cksum` variant | Table needs aggregates; CKSUM is a different algorithm, and naming one variant avoids a differential |
| Hex output / writable byte buffers | First writable buffer (CWE-787) deserves its own consumer; decimal output suffices |
| Test-seam refactor (29 WR-04, 30 IN-01), D-27.1-10 set | Churn with no feature need; revisit when those packages are rewritten |
| M005 N4 note | Obsolete after EVD-13 removed the n bound from the reason |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|

**Coverage:**
- M007 requirements: 20 total
- Mapped to phases: 0 (roadmap pending)

---
*Requirements defined: 2026-10-13*
*Last updated: 2026-10-13 after M007 kickoff*
