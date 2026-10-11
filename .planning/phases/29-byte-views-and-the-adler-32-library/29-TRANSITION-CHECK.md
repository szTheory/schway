# Phase 29 Transition Check

Executed 2026-10-11 by plan 29-11 under the AGENTS.md "Living language roadmap and proactive planning" section. Hosted CI has not run for Phase 29, and the Stage-B cap record (plan 29-12) has not run; nothing below claims the 65,536-byte cap is authoritatively measured.

## 1. Claims checked

| Claim | Where | Evidence class | Witness |
|-------|-------|----------------|---------|
| Byte views are admitted on scalar loop functions only, as `ByteView {drop}` | LANGUAGE-MATURITY, ROADMAP SC4 | source inspection | testdata/phase29/adler32_library.schway, examples/phase29/adler32_fixture.schway; TestPhase29NeighborBaseline |
| adler32 returns the pinned literals on both engines | ROADMAP SC1, LANGUAGE-MATURITY | newly executed focused check (29-00, 29-01, 29-10, local) | TestPhase29GateInterpreterVectors, TestPhase29GateNativeProxyVectors, TestPhase29NativeFixtureVectors |
| Out-of-bounds read exits 65 with exact stderr at -O0, -O3, -O3 LTO | ROADMAP SC3 | newly executed focused check (29-10, local) | TestPhase29OOBApplicationLanes |
| Release before the last view use is refused by one code per peer | ROADMAP SC4 (D-22) | newly executed focused check (29-09, local) | TestPhase29ReleaseBeforeViewUseRefused, TestPhase29OwnerLoopBlockRefused |
| Loop-block source rows are parse refusals, not ownership proofs | ROADMAP SC4 note | source inspection and focused check | TestPhase29LoopNeighborParseRefusals |
| Native runnable gain is build with a bindings manifest plus `app run` | ROADMAP runnable gain (D-16) | newly executed focused check | TestPhase29FixtureAppCLI |
| Sanitizer evidence: AddressSanitizer ran on darwin/arm64, no leak detection on darwin | examples/phase29/README.md, LANGUAGE-MATURITY | newly executed focused check (29-10) | README "Sanitizer evidence" table; no Linux lane yet |
| Stage-A cost is a gate-only proxy, far under budget | 29-00-GATE-VERDICT.md | historical receipt of plan 29-01 | not the cap record (D-14) |
| Corpus registers match the tree (159 `.schway` files) | LANGUAGE-MATURITY corpus sentence | newly executed focused check | TestLanguageMaturityCountsAreCurrent, TestPhase272AcceptedCorpusByteIdentity, TestPhase273ConformanceCorpusByteIdentity |
| Hosted macOS and Linux CI receipts for Phase 29 | none | not yet run | plan 29-12 / hosted dispatch |

Known gaps carried honestly: the pre-existing formatter defect on scalar `var` statement bodies (29-02, not fixed); no operand-type law for byte ops in check's fixed point (29-03); pathoracle refuses every cyclic CFG identically (29-03); a narrow `fixtureViewCall` exemption in corevalidate's replay law (29-06); phase28 manifest rows for adler32_fixture.schway show `native.build_failed` because that walker builds without a manifest (29-07); interpreter `terminalLiveResources` reports the whole stack only on the fixture-generator route (29-08); owner move while a foreign-borrow view is live is refused only by generic matchers (29-09).

## 2. Seed table

Seeds scanned: `.planning/seeds/`. Source appraisal consulted: `.planning/sources/DASHBIT-AI-ERA.md` (sections on contracts and inference, program queries, runtime observation).

| Seed | Trigger | Current witness and named consumer | Disposition | Reason |
|------|---------|------------------------------------|-------------|--------|
| SEED-261008-trz | A named program failure cannot be diagnosed from current capture, observers, debug map and source replay | Consumer: the adler32 fixture app and out-of-bounds control. Every Phase 29 failure (refusal codes, exit 65 stderr, sanitizer lanes) was diagnosed from source, refusal codes and the existing tests. | retain dormant | No failure was left unexplained by current evidence. |
| SEED-261008-z1g | A named reviewer or agent task shows an accuracy or material cost gap with bounded query and ordinary search | Consumer: finding adler32 callers and ByteView parameters. Ordinary search and the callgraph tests sufficed. | retain dormant | No measured reviewer gap; re-check at Phase 30 when module imports create cross-file callers. |
| SEED-261008-z1s | A named consumer or reviewer contests a concrete choice in contracts, inference or canonical source form | Consumer: byte-read surface syntax (`byte_len`, `byte_at`, parameter named `bytes`) settled by D-01 and D-18 in discussion; no reviewer contested a concrete form. | retain dormant | The Phase 28 note that byte-read syntax was a Phase 29/30 discussion item is now settled; nothing contested, so it is not run. Re-open only if a reviewer contests the form. |

## 3. Next capabilities

1. **Phase 30 local modules importing adler32.** User-visible program: an entry program imports `adler32` from a separate module file and prints the same pinned answers. Current blocker: no import or module-qualified ID exists; the library is copied into the fixture (examples/phase29/adler32_fixture.schway). Smallest complete slice: module header matching path, import of one function, refusal of `..`, absolute and symlink escapes. Checker and guarantee changes: module-qualified function IDs, import resolution in check, and the single-exit-class table for new refusals. Evidence: Phase 29 pinned literals reused as the independent answers, plus escape-refusal controls. Owner and next action: ROADMAP Phase 30 (ownership of the module requirements in REQUIREMENTS.md); discuss then plan. Would change priority: the Stage-B cap record failing or lowering the cap, which would first need a library change.
2. **Phase 31 `checksum FILE` with the cap from spike 008.** User-visible program: `checksum FILE` prints Adler-32 for files up to the cap. Current blocker: the file adapter and the authoritative Stage-B cap record (plan 29-12) do not exist; the fixture owner is test-only. Smallest complete slice: bounded file read into the owner shape with typed truncation and oversize errors. Checker and guarantee changes: promote the fixture owner shape to a real adapter without widening release or move laws beyond D-22. Evidence: Stage-B cap measurements, the existing sanitizer lanes plus a Linux lane (LeakSanitizer applies there), and negative controls for missing, empty and oversized files. Owner and next action: plan 29-12 first, then Phase 31. Would change priority: the cap measurement exceeding budget, or a named consumer needing a larger file size.
3. **Alternative: bitwise U64 operators with independent evidence (CRC-32 or `cksum` compatibility).** User-visible program: CRC-32 output over the same byte view. Current blocker: no xor, shift or and operator; byte views now exist, so the consumer side is ready. Smallest complete slice: one operator set with checked semantics on both engines and a pinned CRC-32 vector set. Checker and guarantee changes: new scalar operators in all peers, with a type law for byte ops that check's fixed point currently lacks (29-03 gap). Evidence: literal CRC-32 vectors, mutants per operator, both-engine differential. Owner and next action: unowned; promote only through requirements. Would change priority: a named consumer needing interop that Adler-32 cannot provide, or the formatter and byte-op type-law gaps being fixed first.

## 4. Runnable-gain status

Phase 29 is a feature phase with a runnable gain: yes (D-16). `schway build examples/phase29/adler32_fixture.schway --manifest examples/phase29/fixture_bytes.bindings.json --output ART` plus `schway app run ART -- N` prints pinned checksums natively, the interpreter gain is the session API `interp.RunWithFixtureBytes`, and the out-of-bounds control is `examples/phase29/byte_index_oob.schway`. `schway run --engine=*` is unchanged. No scope review is triggered.
