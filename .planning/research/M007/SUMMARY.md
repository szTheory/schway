# Project Research Summary — M007 Bitwise Checksum and Composable Scalar Helpers

**Researched:** 2026-10-13 via five-lens fan-out (product/consumer, compiler/PL,
security/adversarial, devops/evidence/debt, DX/agent-ergonomics), synthesized
and auto-followed per the standing fan-out preference.

## Key Findings

### Consumer choice
- CRC-32 was the pick of the compiler, security and DX lenses, and the
  product lens's runner-up (0.25 vs `wc` at 0.65). It reuses M006's file cap,
  ByteView, module and receipt pipeline unchanged and has a universal
  independent oracle (CRC-32/ISO-HDLC check value 0xCBF43926 for "123456789";
  Python `zlib.crc32`).
- Name exactly one variant: POSIX `cksum` is CRC-32/CKSUM (non-reflected,
  length appended, check 0x765E7680) — a different algorithm. M007 uses
  ISO-HDLC and must not claim cksum compatibility.
- Bit-at-a-time loop avoids a 256-entry table (no aggregates). Output stays
  bare decimal through existing `write_u64` (no hex output surface).
- Product lens's `wc` + `longest_line` (K&R ch.1 sequence) forces arity-N,
  scalar-only records and exported types — the correct NEXT step toward JSON,
  staged as M008, then a bounded key=value config parser, then JSON.

### Compiler cost (file evidence)
- Remainder (Phase 27) touched ~12 production files; byte_at (Phase 29) ~11.
  Most cost is the defect path (cgen evidence reservation
  cgen_program.go:229-243, guard emission :2735, interp defect events
  interp.go:455/:1725, executionpeer :386/:532). Non-defecting ops join the
  event-free lists only; pathoracle/originvalidate walk terminators only.
- Arity is exactly 1 (`core.Function.Parameter`, core.go:146/:466; 65 check
  sites; `check.call_arity_unsupported` check.go:5072-5094). Changing it churns
  the core IR JSON schema → L. Aggregates XL; writable buffers XL and highest
  risk to the evidence model.
- Scalar functions cannot call other functions (scalar.go:625-680). Operators
  are only `+ < == %` (parser.go:700-710). adler32 computes b<<16 by 16
  doublings.
- Bitwise xor/and/or: S. Shifts: S/M — literal amount < 64 only (precedent:
  static-zero refusal scalar.go:638) keeps them event-free and avoids C17 UB
  (§6.5.7). Scalar→scalar calls: M. Hex literals: lexer-only.
- Free differential witness: re-express adler32 with `<< 16`; pinned answers
  must stay byte-identical.

### Security
- HTTP worst (framing differentials/smuggling, out of scope); JSON high
  (Seriot "Parsing JSON is a Minefield", JSONTestSuite, json-c CVE-2020-12762);
  base64 medium (curl CVE-2016-8617 size math); CRC-32 low.
- Pre-feature fix: `os.ReadFile(capturePath)` precedes the capacity check in
  internal/compiler/native/native_app.go (~:364) — child-written file → host
  unbounded allocation. Use io.LimitReader(limit+1) + seeded oversize mutant.
- Required CRC controls: empty (0), single 0x00, "123456789", 64 KiB of 0xFF,
  65,536 accepted / 65,537 refused; shift counts 0/31/32/63 accepted, 64 and
  non-literal refused by check, core and origin peers; mutant removing each
  refusal is caught; interp/-O0/-O3/-O3 -flto agree.

### Debt / CI
- Next phase number: 33 (continues after 32.1).
- Debt phase ≤5 plans, no receipt binding: WR-01 (freeze script cwd-relative
  pathspecs), WR-02 (`isAbortLikeDefect` interp.go:432-461 too broad — touches
  byte-op parity), IN-01..03, coverage-`kind` pin, 29 WR-02/WR-03 triage,
  Phase 28 VALIDATION/SECURITY backfill, generalize check-phase32-* into a
  milestone-agnostic freeze/receipt gate.
- Fold interpreter file-bytes CLI route (`RunWithFileBytes` is test-only) into
  the first file-reading feature phase. R1b only if a phase needs it.
- Drop N4 (obsolete after EVD-13); defer D-27.1-10, test-seam refactor,
  module-aware interface/query.
- Hosted runs ~15–22 min, 5/15 recent failures. M006 lessons: bind receipts
  once at the very end after `/gsd-audit-milestone`; cap phases ≈8 plans
  (table-driven per-peer seeded faults in one plan); require one_liner and
  requirements-completed in SUMMARY frontmatter.

### DX
- Scratch-authoring trial: CRC-32 and word-counter drafts produced ~3
  diagnostics per real cause (cascading `expected }`), misleading messages
  (arity-2 → `expected )`; Bool tail → `syntax.expected_equal`), byte-offset-only
  spans, no surface manifest, `query SRC fnname` returns not_captured.
- Floor to bundle: line:col + snippet (text) and line/col (JSON); one diagnostic
  per recovery point; `surface.unsupported_*` codes with one-line hints for
  absent constructs (&&/||/!, arity>1, if-without-else, Bool tail literal…);
  `schway surface --json` capability manifest pinned by meta-tests so hints
  cannot drift.
- Foreign-adapter boilerplate (~40 lines per byte program) is a hidden tax;
  candidate for M008 unless cheap.

## Implications for Roadmap

Suggested order (roadmapper decides final shape):
1. Phase 33 — debt + security fix + generic freeze/receipt gate (≤5 plans).
2. Diagnostics floor + surface manifest (no semantics change).
3. Bitwise ops, literal-amount shifts, hex literals; adler32 `<< 16` witness.
4. Scalar→scalar calls (arity 1).
5. `crc32` module + program + second consumer + interpreter file-bytes CLI
   route + clean-checkout guide.
6. Milestone receipts bound once after audit (may be a final plan of 5 rather
   than its own phase).

## Sources
RFC 8259; RFC 4648 §10; Seriot 2016 + github.com/nst/JSONTestSuite; curl
CVE-2016-8617; json-c CVE-2020-12762; ISO C17 §6.5.7; CRC RevEng catalogue;
K&R *The C Programming Language* ch.1; CWE-787/125/190/1335; repo files cited
inline.
