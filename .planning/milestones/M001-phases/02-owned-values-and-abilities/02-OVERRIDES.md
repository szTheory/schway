---
phase: 02-owned-values-and-abilities
recorded: 2026-09-04
code_head: 3399ddc
overrides: 1
---

# Phase 02: Accepted Plan Overrides

## OV-02-01 — `Buffer` grants the `share` ability

**Decided by:** the developer, 2026-09-04, during the Phase 02 close-out review.
**Implemented in:** `2549311`.

### What the plans said

- `02-01-PLAN.md`: "`Buffer` is compiler-defined with `drop`, `send`, and `escape`,
  but without `copy` or `share`."
- `02-02-PLAN.md`: `Box<Buffer>` and `Pair<Byte, Buffer>` derive as **not**
  copy/share.
- `02-RESEARCH.md:143` states the same, marked `[PROPOSED]`.

### Why it changed

The plans contradicted themselves. `02-RESEARCH.md:158` and the shipped control
fixture `testdata/phase2/move_while_borrowed.lang` both do `let view = borrow buffer`
— borrowing a `Buffer` — while the ability table denied `Buffer` the `share`
ability that `corevalidate` requires for `OpBorrowShared`.

The implementation only avoided the contradiction by accident of ordering: the
checker had no ability gate on `borrow` (CR-01), so `move_while_borrowed` fired
before the missing-share condition was ever evaluated. Closing CR-01 forced the
contradiction into the open — a share gate in the checker would have rejected the
phase's own control fixture.

Two coherent resolutions existed: give `Buffer` the `share` ability, or keep it
unshareable and rewrite the control fixture onto a shareable type. The developer
chose the former, on the grounds that a shared loan of a noncopyable value is the
canonical purpose of `borrow`, and the whole Phase 02 narrative ("move during an
active loan") presupposes that a `Buffer` can be loaned.

### What it changed

- `internal/compiler/ability/ability.go`: `Buffer` = `{drop, share, send, escape}`;
  the `AbilityShare` negative witness removed. `copy` stays denied with its witness
  intact — **`Buffer` remains noncopyable**.
- `internal/compiler/corevalidate/corevalidate.go`: `deriveAbility`'s Buffer arm
  denies only `copy`. Mandatory, not optional — the validator re-derives every
  `TypeFact` and compares with `reflect.DeepEqual`, so leaving it would have made
  every Buffer program fail `core.ability_mismatch`.
- `testdata/phase2/evidence.golden.json`: `core_digest` and manifest `id` only.
  Verified key-by-key as parsed JSON; `source_digest` and `c_digest` unchanged.
- Phase 1 goldens and `testdata/phase2/owned_transfer.golden.c`: **byte-identical**.

### Consequences accepted with the decision

1. **The `ownership.borrow_requires_share` gate is unreachable from source.** With
   `Buffer` shareable, `Byte` and `Buffer` both grant `share` and `Box`/`Pair` only
   conjoin their children, so by structural induction every source-reachable type
   grants `share`. The gate is defence in depth against `core.ability.share_denied`,
   not a source-level control. Its law is covered by an exhaustive
   production-vs-oracle differential over a synthetic non-shareable fact that grants
   copy (so the copy gate cannot mask a missing share gate), and
   `TestShareIsUniversallyGrantedAfterBufferShare` enumerates >100 shapes to depth 4
   and fails the moment any type withholds `share` — making the unreachability claim
   self-invalidating rather than assumed.

2. **The Buffer ability row is not cross-checked by the differential.** Both
   admission layers were edited in one commit, so production and validator cannot
   independently corroborate *that specific row*. Every other row remains
   cross-checked, and the derivation *mechanism* is untouched.

3. **`AbilityShare` has exactly two consumers**, both borrow gates. Enumerated at
   `3399ddc`; no authority-escalation path is created. `control:ability.forged_copy`
   still fires.

### Verification at `3399ddc`

`move_while_borrowed.lang` is byte-unchanged and still produces
`control:ownership.move_while_borrowed`. All nine Phase 2 controls fire. Phase 1
work 18, Phase 2 work 52. `go test`, `-race`, `vet`, and `verify-phase2.sh` all
exit 0.

---

_Recorded: 2026-09-04 at `3399ddc`_
