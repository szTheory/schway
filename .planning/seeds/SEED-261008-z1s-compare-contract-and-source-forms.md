---
id: SEED-261008-z1s
status: dormant
planted: 2026-10-08
planted_during: M005/Phase 27 quick task
trigger_when: A named consumer or reviewer contests a concrete language design choice involving explicit contracts, inference, or canonical source form.
scope: Compare two forms of one real program under the same semantic obligations and bounded evidence plan.
audit_acknowledged:
  milestone: M005
  at: 2026-10-10
  status: dormant
---

# SEED-261008-z1s: Compare contract and source forms

## Why This Matters

Agent tolerance for annotations does not prove that duplicated facts help checking, and reduced typing effort does not make canonical readability irrelevant. Explicit boundary contracts coexist with local scalar inference; any change must improve guarantees or human review without hiding checker or maintenance cost.

## When to Surface

**Trigger:** A selected program or actual reviewer identifies a contested choice in explicit contracts, local inference, or canonical source. Do not trigger from general forecasts about agent syntax preferences.

## Smallest Consumer and Obligations

Take the same program and semantic obligations in two proposed forms. Include malformed edits and seeded defects. Specify which checker/guarantee changes each form needs and retain independent semantic answers. Measure actual human reviewers’ defect detection and time; if no human review is performed, label that dimension unmeasured rather than proxying it with formatter output or an AI. Count tokens alongside full generate/verify/run/observe/repair cold and warm p50/p95, plus checker complexity and maintenance burden.

## Independent Answers and Negative Controls

Use literal expected program behavior and independently derived guarantee results for both forms. Seed syntax and semantic defects, including malformed edits, and verify each is caught without weakening existing refusal controls. Reject added syntax/formality if no measurable review or guarantee gain appears, if checker trust grows beyond the named need, or if end-to-end cost worsens without a compensating user-observed benefit. Bounded state models and independent traces may serve a selected resource/loan loop witness; no general SMT/proof backend follows from this seed.

## Cost, Owner, and Next Action

**Cost:** A paired source and reviewer experiment plus only the checker work required by the selected guarantee.

**Owner/next action:** The language/checker owner for the selected consumer, with actual reviewers for human claims, should at a relevant planning transition, identify the program, real review task, and expected guarantee before promoting through roadmap/requirements ownership.

## Supersession and Breadcrumbs

Supersede when the dispute resolves without a code or review consequence. Reject if no named consumer, reviewer, or independently checkable obligation exists. Existing homes: `.planning/PRODUCT-ROADMAP.md` and `.planning/LANGUAGE-MATURITY.md` own guarantee taxonomy and locality. Source anchors include `internal/compiler/ability/ability.go`, `internal/compiler/check/scalar.go`, and `internal/compiler/session/session_phase27_remainder_test.go` (`TestPhase27StaticZero`, `TestPhase27ForgedRemainder`, `TestPhase27DynamicZero`). See the [Dashbit assessment](../sources/DASHBIT-AI-ERA.md).
