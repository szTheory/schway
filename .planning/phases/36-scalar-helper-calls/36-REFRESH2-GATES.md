# Phase 36 execution refresh gates — 2026-10-15

Invocation: `$gsd-execute-phase 36`, no flags. Canonical init: eight complete
plans, 21 tasks, seven waves, zero incomplete/halted/blocked/checkpoint plans,
zero duplicate threat IDs. No executor/worktree or completed-UAT replay was
needed. Runtime identity is @opengsd/gsd-core1.16.0. The manual invocation
cleared only the ephemeral chain flag; persistent user config is preserved.
Explicit canonical model/effort resolutions reach typed review/verifier agents
despite the baked-config notice. MVP/TDD false; execute:pre has no hooks;
only the regression-gate section applies.

Gate order: validate-phase, secure-phase, UI applicability, independent code
review/disposition, bounded prior-phase regression, independent goal verifier.
The initial mapped sample passed84roots, then review discovered CR-01.
Native eligibility omitted existing structured unused-Byte/ByteView Unit
callers. Original-source overlays reproduce both exact emission failures.
The bounded repair in d69380ed and two new tests are in36-REFRESH2-FIX.md.
Post-repair validation passes84mapped+2repairroots (86total/13packages),
zero missing/fail/skip;36-REFRESH2-RUN.json binds85inputs/14logs. The final
plan08 owner lists were parsed; historical capture producers were not replayed.
Read-only assertions preserve acceptance revision and read its single allowed
historical Phase35 report at that revision, since its current report is later.
No historical receipt, plan, summary, frozen expected fixture or UAT was altered.

Both existing prohibition descriptors freshly proved actual-source bad RED,
clean GREEN and default GREEN after repair. Their canonical results and source
bindings are in36-REFRESH2-PROHIBITIONS.json; no waiver was inferred.

Security StateA retains24distinct authored closed threats, zero open, configured
ASVS L1/high. Source inspection and fresh controls support the authored L1
shortcut; no new accepted risk or L2/L3 claim. Fresh UI gate has frontend,
hasUiFiles,hasUiSpec,block all false; existing N/A review reused under the
automatic lifecycle branch. No numerical graphical or readability score.

Independent standard review:54initial explicit paths, then2repair code paths
(one new),55cumulative unique paths; a final one-row inventory supplement makes56. Closure reviewer independently ran both
new roots. CR-01 and historical WR-01 are fixed; WR-02 remains deferred to
the core-validator maintainer's next semantic edit. Current review has one
warning and is not clean. Canonical disposition preserves all three IDs,
zero open, one deferred. No fallow/external reviewer lane or extra execute:post
gate applies. Phase33CR-03/04 remain separate deferred BLOCKER propagation
findings for mandatory M007 audit before Phase38 freeze; verifier must assess
must-have overlap independently. No source/digest verdict is supplied by parent.

Regression results, independent verification and completion routing are added
after those gates return. A focused pass does not refresh earlier reports;
read canonical progress after completion. No whole-package/project local suite,
new dependency/backend/wire field, push, hosted dispatch or host receipt occurs.


## Regression-discovered consumer bookkeeping

The first combined union passed all reached behavior except the Phase16 public
emitter inventory: the new direct EmitNative test call lacked one registry row
(source116 vsregistry115). Commit0e56a3be adds its exact AST call/witness record;
independent reviewer confirmed the row and both inventory/control roots pass.
The original failed union log remains at/tmp/schway36-r2-regression-13.log.
No failing behavior is credited as passed. A fresh final union follows.


## Final bounded regression

The fresh final union passed239distinct names /251root executions across16
packages at0e56a3be, zero missing/fail/skip in136.534seconds. It includes the
84mapped owners, two repair roots, current Phase35 (including Phase33/34)
selectors and the historical Phase36 prior-phase names. Historical lists
supply names only; all final logs are newly executed. Commands are canonical
one-shot normalized, count1, JSON captured, Go/process bounded600seconds.
36-REFRESH2-REGRESSION.json binds179current inputs and16raw logs; it also
preserves the initial failed attempt and its correction. Exact native
optimization/sanitizer/mutation controls ran; no hosted lane earns credit.
