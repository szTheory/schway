# Phase 37: CRC-32 Consumer, Interpreter File Route, and Guide — Discussion Log

> **Audit trail only.** Do not use as input to planning, research or execution agents. Decisions are captured in CONTEXT.md.

**Date:** 2026-10-15
**Areas:** CRC API/reuse; file CLI/host authority; independent evidence/guide; adversarial synthesis.

The user requested all relevant decision points, breadth/depth through relevant roles, primary-source research, an adversarial pass and coherent recommendations to auto-follow. This explicitly delegates the option choices. No question was asked or represented as a separate human answer. Config enables auto-advance; no existing context/plans/SPEC/checkpoint or pending todo needed a user choice.

Three typed `gsd-advisor-researcher` agents ran as background specialists with fresh resolved `gpt-6-luna`/`xhigh` routing: `/root/discuss37_crc`, `/root/discuss37_cli`, `/root/discuss37_evidence`. All returned terminal results before synthesis. Source files and primary references were read; no broad compiler suite ran. Parent independently reviewed the current dispatch/extractors, file adapter, interpreter APIs/evidence setup, admission and manifest/guide facts and calculated the Python oracle rows. Agent answers distinguish code inspection from newly executed oracle calculations.

## CRC API and second importer

| Option | Pros | Cons / footguns | Impact / risk | Selected |
|---|---|---|---|---|
| Public crc32(ByteView), private one-argument eight-round crc_byte | Existing arity/types, byte authority stays outer, one activation/byte | Dynamic events still exceed large-run capture; state invariant must remain internal | Library/two entries, existing adapters; bounded calls and file ownership | Yes |
| One-bit helper called eight times/byte | Small direct recurrence; resembles Phase 36 witness | 8x activations/capture/dispatch cost with no consumer gain | Same surfaces, larger runtime/evidence demand | No |

**Delegated selection:** per-byte helper; nested lib.crc32 avoids the mandatory crc32.schway entry's self-import collision. The second consumer uses low16, a known simple projection with its own pin and a mutation that changes both outputs. Private state makes the 32-bit invariant audit-local. Existing parser/checker shapes appear sufficient; this is source inspection, not a successful new Schway program run.

## File application CLI and authority

| Option | Pros | Cons / footguns | Impact / risk | Selected |
|---|---|---|---|---|
| `run --engine=E SOURCE [native manifest] -- PATH` | Clear supplied-input discriminator; old no-tail conformance preserved | Two run modes require precise usage; early parsing is essential | CLI/session/interpreter, route/exit tables; flag-shaped path risk | Yes |
| Add `--file-bytes` mode flag | More explicit mode in argv | Redundant grammar/duplicate-option surface | Same surfaces plus mode flag | No |
| Retained native build/app run only as native matrix route | Existing proven boundary | Does not give the symmetrical source/native file form | Existing route, less new code | Guide retained workflow, not sole new-route evidence |

**Delegated selection:** suffix-marked input, early dispatch, opaque post-separator path, native explicit manifest, interpreter no manifest and fixed checked semantic Go adapter. A ceremonial interpreter manifest or exact C digest allowlist adds authority confusion without proving C execution. Interpreter does actual host reads but makes no arbitrary-C semantic claim. The root-sensitive unreadability case, symlink distinction and empty-path documentation discrepancy are explicit. Native process timeout is not falsely generalized to arbitrary interpreter divergence.

## Independent evidence and executable guide

| Option | Pros | Cons / footguns | Impact / risk | Selected |
|---|---|---|---|---|
| Finite literal oracle + shared public CLI matrix + tracked-only guide | Each engine has independent expected outcomes; checks parsing/streams/exits/reuse | Must connect fixtures, guide and actual commands; native requires Clang | Focused CLI/test-support/example files, no dependency | Yes |
| Go-level file seams plus CLI smoke | Smaller/faster low-level checks | Cannot satisfy per-case public CLI parity; agreement can be wrong | Misses APP-14's explicit requirement | Supporting evidence only |
| Generic test/docs/mutation framework | Potential future reuse | More abstractions/drift with no named current need | Broad framework churn | No |

**Delegated selection:** pinned Python/zlib answers and independent Go cross-checks, exact public tuples, five reached algorithm/library controls, new module identity and dual-output changes, existing physical bound/cleanup controls, marked executable guide in a tracked-only temp checkout. Phase 32's current copy helper includes untracked files and cannot support that stronger claim unchanged. Actual Python 3.14.4/zlib 1.2.12 pins include maxFF `3735780942`; no Schway output generated the expectations. Python remains oracle tooling, not a runtime dependency.

## Adversarial cross-lens pass

Compiler/PL and independent-checker lenses: preserve exact arity/types, full-program admission before prune, fixed owner shape, module identity and 32-bit state; no new source/runtime semantics merely to make examples pass. Systems/security lenses: preallocation/read cap+1, open-descriptor regular check, opaque bytes/argv, bounded options, path-free errors, independent C/Go readers, actual cleanup controls; do not conflate file symlinks with import confinement. Product/DX/docs lenses: one copy of algorithm, useful low projection, bare decimal streams, concise reproducible commands, explicit cksum variant, exact errors. Ops/evidence/portability lenses: bounded focused local checks, real reached mutants, actual capture disabled, honest root/sanitizer availability, existing CI, hosted receipts only38. Human readability and whole-loop latency stay unmeasured.

The pass rejected: using conformance JSON as app stdout; parsing a path with global flag extractors; ignoring a manifest while implying native equivalence; importing testsupport into production; trusting Stat.Size; buffering unlimited events then discarding them; treating evidence failure as app success; a scan that flags the historical one-bit fixture as duplicate full CRC; a mutant that only fails compilation; a tracked-checkout test that copies untracked files; and labeling source inspection as newly passing execution.

Further fan-out would repeat the same boundaries without a new consumer/counterexample. The context captures 24 decisions, precise source witnesses, verification owners, all three relevant dormant-seed dispositions and the next three capabilities. No generic framework or source-language expansion is promoted.

## Agent's Discretion

Concrete source/API/test layout, focused fault injections, exact finite owner selectors, small production helper boundaries, scoped manifest wording/witness deltas and test timeout choices after observation. Any source-grounded correction is recorded, not silently substituted.

## Deferred Ideas

Phase 38 audit/final hosted binding; M008 named arity-N/record/wc consumer; arbitrary C interpretation; general cancellation/fuel, IO/evidence/docs frameworks, full large-file traces, separate compilation, numeric spelling unification and untriggered query/observation/source-form seeds. See CONTEXT for owners and triggers.

## Research amendment — 2026-10-15

The bounded native source probe produced `3421780262` for `123456789`. It also disproved the initial source-inspection premise in D-10: generated C rejects an empty PathToken before the C adapter and `cmd/schway/main_test.go` already pins `(65, "", "")`. Under the delegated evidence-backed correction rule, D-10 now preserves that existing tuple. No native preflight or old empty-token test is to be changed merely to satisfy the draft premise. The checksum guide statement remains correct. The context now also names the existing project spike-findings skill discovered by research. These corrections do not expand scope or claim phase verification.
