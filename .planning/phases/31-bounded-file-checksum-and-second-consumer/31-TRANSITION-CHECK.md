# Phase 31 Transition Check

Executed 2026-10-12 by plan 31-09 under the AGENTS.md "Living language roadmap and proactive planning" section. Hosted CI has not run for Phase 31 (see section 5); nothing below claims a hosted receipt.

## 1. Claims checked

| Claim | Where | Evidence class | Witness |
|-------|-------|----------------|---------|
| `checksum` prints the pinned answers (1, 65537, 300286872, 2006388466) through `build --manifest` plus `app run -- PATH`, and refuses 65,537 bytes | ROADMAP SC1, examples/checksum/README.md, LANGUAGE-MATURITY | newly executed focused check (local, native lane not skipped) | TestPhase31ChecksumPublicRoute, TestPhase31ChecksumInterpreter |
| Every error row exits 65 with empty stdout, a fixed path-free stderr line and the same result on both engines | ROADMAP SC2, README exit table | newly executed focused check | TestPhase31CaseMatrix, TestPhase31ExpectedMatrixJSON |
| The buffer is released exactly once on success, oversize, AllocFailed and read error; the size check is a limit+1 read that never trusts the reported size | ROADMAP SC2 | newly executed focused check (ASan lane ran) | TestPhase31ReleaseExactlyOnce, TestPhase31BoundedReadMutants |
| `adler_low` imports the one library and prints 920 for Wikipedia; a library edit flips both applications, a one-sided edit flips one | ROADMAP SC3, MOD-06 | newly executed focused check | TestPhase31AdlerLowPublicRoute, TestPhase31SingleCopyAndDualFlip |
| Build identity, receipt and evidence carry sorted `modules` rows when a program imports; no-import bytes are unchanged | MOD-06, D-11, D-19 | newly executed focused check | TestPhase31ModuleDigests, TestPhase31MovedLibraryIdentity |
| The third owner shape is accepted and refused independently by check, corevalidate, originvalidate, pathoracle and cgen | D-06, D-16 | newly executed focused check | TestPhase31FileBytesSeededFaults, TestPhase31FileBytesNeighborRefusals, TestPhase31FileErrorStderrTable |
| Each pinned value is cross-checked against Go `hash/adler32` beside its Python command; each EVD-14 control is observed to fail | EVD-14, D-12, D-13 | newly executed focused check | TestPhase31OracleRows, TestPhase31Controls, TestPhase31ControlInventory, TestPhase31LoopNeighborRefusals |
| The exit-class row for 65 names the checksum outcomes with no constant change | D-08 | newly executed focused check | TestExitCodeDocsTable, TestNoDuplicatedExitLiterals |
| Corpus registers match the tree (163 `.schway` programs) | LANGUAGE-MATURITY corpus sentence | newly executed focused check | TestLanguageMaturityCountsAreCurrent, TestPhase28EntryExitClassManifest |
| Modules are whole-program composition, not separate compilation, not B1/DX-06 | PRODUCT-ROADMAP, LANGUAGE-MATURITY, examples/checksum/README.md | source inspection | the three documents state it only as a non-claim |
| Hosted macOS and Linux CI for Phase 31; M006 receipts; clean-checkout guide | none | not yet run | runs at the phase gate; run IDs go in 31-VERIFICATION.md; receipts and guide belong to Phase 32 |

Known gaps carried honestly:
- The host clang on this darwin machine does not reproduce executables, so import-order and dual-flip tests compare receipt `modules` rows and the ExecutableDigest-free input ID rather than equal executables (31-05, 31-07). A reproducible toolchain applies the strict check.
- The interpreter file route needed the same byte-index-defect exemption as the fixture route (31-08).
- The off-by-one mutant is `n = n + 1` because the language has no `<=` operator (31-08).
- The pre-existing formatter defect on scalar `var` statement bodies is unchanged, so the library and the entries are committed verbatim.
- Imported data types, foreign declarations and type imports are still refused, so the adapter block is repeated in each entry.

## 2. Seed table

Seeds scanned: `.planning/seeds/`. Source appraisal consulted: `.planning/sources/DASHBIT-AI-ERA.md`.

| Seed | Trigger | Current witness and named consumer | Disposition | Reason |
|------|---------|------------------------------------|-------------|--------|
| SEED-261008-trz | A named program failure cannot be diagnosed from current capture, observers, debug map and source replay | Consumer: the Phase 31 file checksum. Every failure (the shape faults, the clang non-reproducibility, the interpreter defect route) was diagnosed from refusal codes, ASan output, the counting observer and focused tests. Weighed against DASHBIT-AI-ERA.md, "Runtime observation, security, and authority" and the "Optional failure diagnosis" row. | retain dormant | No failure was left unexplained by current evidence. |
| SEED-261008-z1g | A named reviewer or agent task shows an accuracy or material cost gap with bounded `schway query` and ordinary search | Consumer: `adler_low` importing the same library, the re-check the Phase 30 disposition requested. Finding the callers of `adler32` across `checksum.schway`, `adler_low.schway` and the Phase 29 fixtures was done by ordinary search, and TestPhase31SingleCopyAndDualFlip enumerates the importers and the single loop copy independently. Weighed against DASHBIT-AI-ERA.md, "Program queries, source locality, and cost" and the "Programmatic compiler queries" row. | retain dormant | The re-check is done: ordinary search plus the single-copy test found every caller, with no measured accuracy or cost gap. Next consumer to re-check against: a third importer or a library with a private helper shared across modules. |
| SEED-261008-z1s | A named consumer or reviewer contests a concrete choice in contracts, inference or canonical source form | Consumer: the checksum entries' source form (trailing `low16` call, `try` acquire). Nobody contested a form. Weighed against DASHBIT-AI-ERA.md, "Contracts, inference, and human review" and the "Annotation and source design" row. | retain dormant | Nothing contested. |

## 3. Next capabilities

1. **Phase 32: retire the provisional checksum frontier, record the M006 hosted receipts and write the guide.** User-visible program: a clean-checkout guide that builds and runs `checksum FILE` on macOS and Linux. Current blocker: no hosted receipts yet, and `examples/checksum.schway` still carries its provisional frontier. Smallest complete slice: the receipt gate for the Phase 31 programs plus the guide, with `examples/checksum.schway` retired or reconciled. Checker and guarantee changes: none expected; the receipt gate must require both hosts at the same revision. Evidence: hosted run IDs, the pinned answers and the case matrix on both hosts. Owner and next action: ROADMAP Phase 32 (EVD-15); discuss then plan. Would change priority: a hosted failure on either OS, or a toolchain that cannot reproduce builds.
2. **A bounded JSON configuration consumer, once a tool is named.** User-visible program: validate a size-limited config against one fixed schema with exact success and typed-error results. Current blocker: no aggregates and no bounded parser stack; the file read, bounded-size policy and module reuse now exist. Smallest complete slice: pick the consuming tool, fix the schema and byte and depth limits, and admit only its parser forms with failure cleanup. Checker and guarantee changes: aggregate and parser-state ownership, borrow lifetime during iteration, owned-buffer cleanup on every error. Evidence: truncated, oversized and malformed-input controls with independent expected answers. Owner and next action: unowned; promote only through requirements once a tool is named. Would change priority: a named consumer, or aggregates landing first.
3. **Alternative: bitwise U64 operators with independent evidence (CRC-32 or `cksum` compatibility).** User-visible program: CRC-32 over the same file bytes, in its own module beside the Adler-32 library. Current blocker: no xor, shift or and operator, and no operand-type law for byte ops. Smallest complete slice: one operator set with checked semantics on both engines and a pinned CRC-32 vector set. Checker and guarantee changes: new scalar operators in every peer. Evidence: literal CRC-32 vectors, a mutant per operator, a both-engine differential. Owner and next action: unowned; promote only through requirements. Would change priority: a named consumer needing interop Adler-32 cannot provide, or a need for files above the 65,536-byte limit.

## 4. Runnable-gain status

Phase 31 is a feature phase with a runnable gain: yes. `schway build examples/checksum/checksum.schway --manifest examples/checksum/file_bytes.bindings.json --output ART` plus `schway app run ART -- PATH` prints the Adler-32 of a file (TestPhase31ChecksumPublicRoute, run locally with clang present), and the same commands with `adler_low.schway` print the low 16 bits (TestPhase31AdlerLowPublicRoute). The interpreter route gives the same answers (TestPhase31ChecksumInterpreter). The commands are in `examples/checksum/README.md`. No scope review is triggered.

## 5. Hosted CI at the phase gate (recorded 2026-10-12)

Hosted full and race CI (`go test ./...` and `go test -race`) on ubuntu-latest and macos-latest runs at the phase gate, before `/gsd-verify-work`. Its run IDs are recorded in 31-VERIFICATION.md, not here. Status: not yet run. If the hosted run is unavailable at the gate, the verifier records an external-blocker note in 31-VERIFICATION.md and opens no gap plan. M006 receipts and the guide belong to Phase 32 (EVD-15); this phase does not claim them.
