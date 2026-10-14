# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals - Discussion Log

> **Audit trail only.** Do not use as input to planning, research or execution agents. Decisions are captured in CONTEXT.md; this records the alternatives and recommendation corrections.

**Date:** 2026-10-13
**Areas discussed:** expression grammar/grouping; shift trust contract; numeric spelling/diagnostics; evidence/consumer migration.
**Selection authority:** User explicitly asked to consider all relevant areas with specialist breadth/depth, primary research and adversarial passes, then auto-follow the coherent recommendations. No AskUserQuestion response is fabricated. Three `gsd-advisor-researcher` agents ran initial and adversarial turns with fresh SDK-resolved `gpt-6-luna`/`xhigh` routing. All returned terminal results before synthesis. No repo edits or behavior tests were delegated to them.

## Expression grammar and bounded grouping

| Option | Pros | Cons/footguns | Selection |
|---|---|---|---|
| Rust-family bitwise tiers, retain old comparisons | `crc & 1 == 0` has useful grouping; separates arithmetic/shift/bitwise | Written final `&` masks the shifted operand, not the XOR | Selected; literal vector 264 versus final-mask 8 |
| Go-family compact tiers | Fewer tiers, familiar to bootstrap implementer | Arithmetic and bitwise interleave unexpectedly | Rejected for current CRC/DX goals |
| C-family comparisons above bitwise | Familiar C spelling | `crc & 1 == 0` groups into a Bool operand | Rejected |
| Transparent bounded scalar parentheses | Keeps exact runnable roadmap source; no IR operation | Must bound nesting and preserve generic closers | Selected, expression-depth 64 and contextual tokens |
| Staged let-only witness | Avoids grouping change | Alters published runnable expression | Rejected for this witness |

**Agent recommendation under delegation:** D-03..07. Adversarial pass found absent parentheses, recursive unbounded atoms and `>>`/nested-type collision. Parent retained token-strategy discretion; new resource-refusal delta must be recorded rather than silently changing old successful goldens. Source/token caps alone do not prove recursion safety. Human readability was not measured.

## Literal shift trust contract

| Option | Pros | Cons/footguns | Selection |
|---|---|---|---|
| Immediate op with existing ConstU64 | No dynamic RightID, canonical range, old omitempty wire stable | Extend kind-exclusive laws at every consumer | Selected |
| Dedicated ShiftAmount field | Explicit immediate | New field unnecessary given existing canonical numeric field | Superseded on adversarial pass |
| RightID pointing at OpConst | Uniform binary form | Constant definition does not prove source literal; requires def-use/provenance machinery | Rejected |
| Mask or guard dynamic counts | Can serve a broader language | Changes settled semantics/defect path and consumer scope | Deferred |

**Agent recommendation under delegation:** D-08..12. Initial adviser preferred a dedicated immediate field; second pass found existing ConstU64 can provide the same structural safety with less wire churn. Source checker proves number syntax, independent core/origin prove IR shape/range; no AST provenance is invented. C unsigned-long-long cast avoids promoted signed shift, with exact-width cast retaining low 64 bits. Parent inspected the actual field/trigger/consumer contracts; malformed or overflowing numeric tokens retain stage-appropriate named failures.

## Hex spelling, recovery and compatibility

| Option | Pros | Cons/footguns | Selection |
|---|---|---|---|
| Existing generic token + narrow legacy 0X patch | Raw spelling/lossless formatting, small footprint | Existing numeric policy differences persist | Selected |
| Distinct hex token/value AST | Explicit radix | More consumer churn than required | Rejected |
| Uniform decimal/base parser | Uniform semantics | Changes scalar 010 from 8 to 10; violates old no-feature C contract | Rejected after adversarial challenge |
| Shared helper with explicit route policy | Reuses mechanics while preserving values | Adds abstractions without demonstrated need | Planner discretion only if justified |

**Agent recommendation under delegation:** D-13..15. Source showed hex/binary/separators are already partly admitted, correcting the kickoff's lexer-only estimate. First-pass leading-zero normalization was rejected; preserve legacy decimal versus scalar base-0 interpretation and binary prefix differences as named debt. Lexer accepts both hex prefixes and existing between-digit underscores. Malformed whole candidate gets exact-span ownership; no broad unknown-token suppression. Parent found scalar overflow uses `type.invalid_u64`, while legacy uses `check.literal_out_of_range`; context preserves stage-specific codes instead of claiming nonexistent uniform behavior.

## Independent evidence, consumer and cost lanes

| Option | Pros | Cons/footguns | Selection |
|---|---|---|---|
| Independent check/core/origin transfers plus engine literal vectors | Separates static safety from actual answers | Needs reached transfer/result controls and honest finite scope | Selected |
| Pure-op runtime events or causal-peer scalar replay | Could let executionpeer prove values | New runtime/input API, capacity/semantics churn; exceeds event-free slice | Deferred |
| Per-op event-free and fabricated-event controls at executionpeer | Matches its actual causal contract | Does not validate numeric results | Selected with dated BIT-01/criterion clarification before plans |
| Broader local/hosted CI matrix now | More host coverage | Repeats final binding and exceeds standing local authority | Existing recurring CI; final hosted owner Phase 38 |

**Agent recommendation under delegation:** D-16..23. Table-driven per-peer controls remain required where each peer has the data to prove the property; no causal numeric claim is fabricated. The live acceptance text must be dated/amended from its impossible causal-only wrong-result demand before planning. Preserve original wording/provenance in that amendment. Existing Adler/second-consumer answer tests remain unchanged, while source replaces only the high-word construction. Immutable pre-change corpus first, focused named lanes, no dependency, ~8 plans maximum absent a named safety/witness justification.

## Primary-source research and stopping rule

Consulted Rust precedence/tokens, Go spec/scanner, official WG14 C17 draft and Clang UBSan. Links and relevant claims are in CONTEXT.md. Parent fetched Rust/Go/Clang pages; web parser could not decode WG14 PDF, explicitly disclosed. Ecosystem grammars are contrasts, not Schway specification. No star counts, general framework comparison, HTTP/GUI/rendering study or invented maturity scores were useful for this U64 compiler slice.

Two passes resolved the meaningful conflicts: parentheses, mixed precedence, raw-core shift trust, event versus value proof, and collateral numeric normalization. Remaining choices are local implementation details delegated to planner/researcher. No self-feeding pass over the newly written context was used.

## Deferred Ideas

Numeric policy consistency and causal-peer value replay need separate owners/compatibility inputs. Scalar calls Phase 36, file CRC/second reuse/CLI interpreter parity Phase 37, hosted binding Phase 38, word counter future M008. z1g/trz/z1s remain dormant with named consumers/triggers in CONTEXT.md. No pending todos matched Phase 35. No codebase maps, phase SPEC, interrupted checkpoint or advisor profile was found at initialization.
