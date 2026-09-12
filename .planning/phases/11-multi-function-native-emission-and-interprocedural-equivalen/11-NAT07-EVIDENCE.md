---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 02
requirement: NAT-07
captured: 2026-09-12
status: measured
---

# NAT-07 Evidence Record: Composition-Only LTO Divergence Control

**Purpose:** pin the composition-only negative control (D-11-23) to a
recorded toolchain identity, and state in writing what this control does
and does not prove, per D-11-22/D-11-25/D-11-26.

## 1. Toolchain identity

Captured on the execution host via `clang --version`, verbatim:

```
Apple clang version 21.0.0 (clang-2100.1.1.101)
Target: arm64-apple-darwin25.6.0
Thread model: posix
InstalledDir: /Applications/Xcode.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/bin
```

**Date of capture:** 2026-09-12

This is a fresh capture on this execution host, not copied from
`11-CONTEXT.md`'s prior-researcher report — the value happens to match
(Apple clang 21.0.0, arm64-apple-darwin25.6.0), which is itself useful
corroboration, but it was independently re-run per D-11-22's instruction to
never assume it.

## 2. The 4x3 matrix

Produced by `TestCompositionOnlyLTODivergence`
(`internal/compiler/native/native_lto_test.go`), driving hand-written C
through `native.Runner` with `Options.LTO` toggled — never a second,
hand-rolled clang invocation path. Cell values are the decoded execution
document's `outcome.value` (a decimal byte sum; see the test's topology
comment for the full three-translation-unit design).

| Injection config | -O0 | -O1 | -O3 | -O3 -flto |
|---|---|---|---|---|
| none (no restrict, aliased) | 104 | 104 | 104 | 104 |
| restrict-only (restrict, not aliased) | 10 | 10 | 10 | 10 |
| restrict+write (restrict, aliased) | 104 | 104 | 104 | **10 (RED)** |

**The single red cell is `restrict+write` at `-O3 -flto`.** Every
single-injection row (`none`, `restrict-only`) is green at every tier,
and `restrict+write` is green at `-O0`/`-O1`/`-O3` — the divergence is
observable ONLY once cross-translation-unit inlining is enabled, which on
this host required `-flto` specifically.

## 3. Observed divergence signature

The observed signature on this host is:

```
interp == -O0 == -O1 == -O3 != (-O3 -flto)
```

This **confirms D-11-20's claim** as re-measured on this specific
execution: the interpreter and `-O0` agree (unmutated code is honestly
executed everywhere), and the `-flto` tier — and only the `-flto` tier —
is the disagreeing one. (D-11-20 stated `interp == -O0 == -O3 != (-O3
-flto)`; this measurement additionally confirms `-O1` agrees too, which
is consistent with, and a strict superset of, D-11-20's claim — no
`ESCALATION` applies.)

## 4. Toolchain-pinned re-measurement clause (D-11-22)

Per D-11-22: the control's red cell is re-measured against the recorded
`clang --version` above; a toolchain change that extinguishes the
divergence is an escalation, not a pass. `TestCompositionOnlyLTODivergence`
is committed and runs on every invocation of this test suite — it is not a
one-time measurement archived only in this file. If a future host's clang
version produces an all-green matrix, the test itself fails via `t.Fatal`
(D-11-21's own lane-failure rule), surfacing the toolchain drift as a test
failure rather than a silently stale evidence record.

## 5. D-11-25: the production corpus's own LTO tier is inert by construction

Because all Lang functions land in one translation unit (D-11-24 — this
control is the ONE deliberate exception, and even it does not widen
`native.Runner`'s single `program.c` path; the extra translation units are
supplied entirely through the existing `ForeignSources` mechanism),
**`-flto` is INERT BY CONSTRUCTION for Lang-to-Lang code in criterion 2's
corpus.** The existing LTO lane's non-inertness (`TestLTOTierIsNotInert`,
this same file) is borrowed entirely from the *foreign* TU boundary
(`testdata/phase5/inline_across_foreign.lang`) — a foreign C function
compiled as its own translation unit, not from any Lang-to-Lang call.

This control (`TestCompositionOnlyLTODivergence`) proves the `-O3`/`-flto`
interprocedural optimizer TIER can be exploited at all, on this toolchain.
It does **not** make the production corpus's LTO tier meaningful for
Lang-to-Lang calls — that tier stays inert by construction until Lang
gains a genuine multi-translation-unit emission target, which is
explicitly out of scope for Phase 11 (D-11-24).

## 6. D-11-26: honest scope statement

The control survives as **tier evidence plus a mutation-killed validator
test** (`TestCompositionOnlyLTODivergence` fails closed: an all-green
matrix is asserted as a lane failure via `t.Fatal`, never a skip or a
silent pass — see assertion 4 in the test, and the "none"/"restrict-only"
rows are the mutation-kill demonstrating neither injection alone is
sufficient).

It does **NOT** survive as evidence that a shipped Lang `restrict` is
sound. Phase 11 emits **zero** call-boundary `restrict`/noalias-shaped
attributes (D-11-09) — this control's false `restrict` is confined
entirely to hand-written test C, compiled only through
`native.Runner.Run` inside `native_lto_test.go`, and never enters `cgen`'s
emission path or any shipped artifact.

**Disclosed residual, named alongside `EscapeCoordinatedSourceToCoreFalseClaim`:**
the `interpreterInput` stipulation inherited from M001 applies here too —
the injected demonstration (a false `restrict` promise on a hand-written C
function, deliberately violated by an aliased probe pointer) has **no
Lang-level meaning**. No Lang program can express this shape; it exists
purely to demonstrate that the compiler's optimizer, when given a false
promise, will act on it. This is disclosed, not laundered, per D-11-26's
instruction.
