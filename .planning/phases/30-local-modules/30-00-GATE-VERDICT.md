# Phase 30 Wave-0 Gate Verdict

Written 2026-10-12 by plan 30-00. The gate is this one plan: the no-import golden first, then flags (a) to (f), each with an armed control.

## Flags

| Flag | Pass | Evidence |
|------|------|----------|
| (a) Closed resolution: escapes, symlinks, FIFO | yes | TestPhase30GateRootEscapesAndSymlinks (modload) refuses an escaping file link, an escaping directory link, an in-root file link and an in-root directory link as module.symlink, and a dotdot segment as module.escape; the tracer row reads real.schway and parses with zero diagnostics. Its control engages skipComponentLstatForTest and observes the in-root directory link read through. TestPhase30GateFIFONonBlocking (modload, unix) returns module.not_regular inside 2 s without opening the FIFO; its control engages openBeforeLstatForTest and observes the call block past the deadline. |
| (b) Case-exact resolution on every OS | yes | TestPhase30GateCaseExact (modload) refuses lib with only Lib.schway, and sub.n with only Sub/, as module.case_mismatch. Its control engages caseFoldAcceptForTest and observes a result other than case_mismatch. |
| (c) Unit omitempty invariance and consumer inventory | yes | TestPhase30GateUnitOmitemptyInvariance (diagnostic) keeps both identity structs byte-identical and the pinned ID diagnostic:d34396fcc8ea6530a1f04dc3 with a trailing omitempty Unit field; a Unit field without omitempty is observed to change the bytes. TestPhase30GateUnitConsumerInventory (session) pins the sorted non-test consumer list with a negative control. TestDiagnosticZeroBytesUnchanged and TestRepairFieldEnumerationIsExhaustive stay green. |
| (d) Injective module C names | yes | TestPhase30GateCNamesInjective (cgen): 4,672 pairs, pairwise distinct except 584 pure function-case collapses that allocate separates; no double underscore; every name matches the SCHWAY_ namespace; allocation is independent of input order; an entry named like a mangled name keeps the bare name. The plain-join seam is observed to collide a_b.c with a.b_c. TestGeneratedIdentifierNamespacesStayConfined stays green. |
| (e) Body-blind contract shape | yes | TestPhase30GateBodyBlindContractShape (check): calleeContract has fields ID, ParameterType, ReturnType and ReturnsBorrowOfParam only; a synthetic struct with a Body field is flagged; TestCallAdmissionBodyBlindControl exists and stays green. |
| (f) Merged-program size gap | yes, gap recorded | TestPhase30GateMergedProgramLimit (session): 1,024 functions parse, 1,025 is refused syntax.input_limit, and a programmatic merge of 600 and 425 functions (1,025) reaches check.Program with no limit diagnostic. |

## D-15 golden

- Revision: 4ea05fd7e2f8dec830fb6392331e44c86bda0e8a (PLAN_BASE).
- Entry count: 160 (159 walked .schway files plus the generated row).
- First commit subject: `test(30-00): pin no-import corpus bytes before modules (D-15)`; it holds only the golden test and the golden.
- The generated row is `generated:phase30/no-import-over-work`; it is built in a TempDir and never committed.
over-work N: 400

Five walked sources panic inside debugmap.Build at PLAN_BASE (testdata/phase1/comments.schway, testdata/phase1/toggle.schway, testdata/phase17/return_type_tracer.schway, testdata/phase6/derivation_match_defect.schway, testdata/phase6/heldout_match_defect.schway). The golden pins the recovered panic text as their debug_map field; plan 30-08 must keep it unless a later plan owns the fix. This is a pre-existing defect, not caused by this phase.

## Merged-program cost

Logged reading from TestPhase30GateMergedProgramLimit: corevalidate.Validate on the 1,025-function merged program took about 220 ms on one host, not pinned. D-21 `module.program_limit` at 1,024 functions and 4,096 declarations is the cap, owned by plan 30-04.

## D-16 split conditions

- Verdict SPLIT? No. Every flag passed and no control was weakened.
- Plan count: 10 including 30-00, at the ceiling and not over.
- Does Unit force a non-omitempty or schema change? No. Flag (c) shows omitempty keeps every empty-unit ID byte-identical.
- Does naming change any no-import C bytes? No. The golden stayed green after the cgen.go addition, which adds no caller.

## D-27 carve-outs

Authority: D-27 amendment (2026-10-12), 30-CONTEXT.md.

Two committed files outside `examples/phase30/` are accepted. The planner and executor add no carve-out of their own; any further file outside `examples/phase30/` stops the plan.

- `testdata/phase30/prechange-corpus.json`: required by D-15; it is evidence, not a fixture, and its generated over-work row is never committed as source.
- `testdata/phase16/public-emitter-consumers.json`: a line re-pin forced by the session.go edits in plan 30-08. It moves `:<digits>` suffixes only, with no corpus content change and no new call site.

## Decision

Flags (a) to (f) passed and every D-16 split condition is false. If a later plan hits a split condition, the proposed boundary is Phase 30 for plans 00-04 (syntax, loader, caps, graph refusals, Unit) and Phase 30.1 for plans 05-09.

verdict: PROCEED
