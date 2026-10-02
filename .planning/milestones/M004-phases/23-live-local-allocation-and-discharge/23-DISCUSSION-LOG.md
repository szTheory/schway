# Phase 23: Live Local Allocation and Discharge — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in `23-CONTEXT.md`; this log preserves alternatives.

**Date:** 2026-09-27  
**Phase:** 23-live-local-allocation-and-discharge  
**Areas discussed:** Application input and source boundary; file and output contract; acquisition and cleanup failure contract; independent evidence and CI

---

The configured GSD transition invoked discussion with `--auto`. The user's
standing instruction authorized automatic follow-through. Each recommended
option below was selected by the workflow; these are not represented as new
direct user answers.

## Application input and source boundary

| Option | Description | Selected |
|--------|-------------|----------|
| Recommended | Pass one bounded opaque caller-selected path token through the public app route to the source entry and explicit acquire operation; preserve U64 and add no general string API. | ✓ |
| General text | Add general string/path manipulation before the resource witness. | |
| Fixed or ambient input | Select a fixed repository fixture or read an ambient environment variable. | |

**Selection:** Auto-selected recommended default under the explicit `--auto` route.  
**Notes:** Maximum path token: 4096 bytes. Exact source spelling and CLI option are technical choices constrained by the narrow boundary.

---

## File and output contract

| Option | Description | Selected |
|--------|-------------|----------|
| Recommended | Accept exactly one raw byte; report empty and oversize as distinct typed errors; return its numeric value as decimal U64. | ✓ |
| Larger buffer | Read a larger prefix and introduce a byte-array result. | |
| Text output | Emit encoded text and add general output/string semantics. | |

**Selection:** Auto-selected recommended default under the explicit `--auto` route.  
**Notes:** Files containing `0x41` and `0x42` produce `65` and `66`; empty and greater-than-one-byte files fail acquisition.

---

## Acquisition and cleanup failure contract

| Option | Description | Selected |
|--------|-------------|----------|
| Recommended | Separate acquire, borrowed-use, and infallible consuming-release contracts; use a reached typed failure after acquisition. | ✓ |
| Hidden lifetime | Hide acquire/use/release in one C call or infer release pairing from an unrelated symbol. | |
| Synthetic-only failure | Use an unreachable cleanup mutation or process crash as the only error proof. | |

**Selection:** Auto-selected recommended default under the explicit `--auto` route.  
**Notes:** `0x43` is the documented unsupported-byte input for a real post-acquisition use error. The generated path must release before exposing the error; transfer and nonlocal exits remain later/refused work.

---

## Independent evidence and CI

| Option | Description | Selected |
|--------|-------------|----------|
| Recommended | Independently observe real allocation/use/free, reach cleanup mutations, and place high-value recurring checks in existing macOS/Linux CI lanes. | ✓ |
| Self-reported evidence | Rely on compiler events, engine equality, or manual README review. | |
| Duplicate suites | Add repeated full-suite/sanitizer jobs without a distinct evidence question. | |

**Selection:** Auto-selected recommended default under the explicit `--auto` route.  
**Notes:** Keep unavailable host evidence incomplete. Assign expensive lanes one owner per host. Objective tests replace subjective UAT.

---

## The agent's Discretion

- Exact source syntax, opaque path type, and public flag spelling within the
  selected boundary.
- Foreign ABI layout and independent native observer implementation.
- Placement of focused tests and ownership of expensive existing CI lanes.

## Deferred Ideas

- Transfer/error propagation through Lang calls (Phase 24).
- Shared/exclusive pointer helpers and the integrated utility (Phase 25).
- General strings, arrays, loops, arithmetic, owning aggregates, fallible
  destructors, unwind, cancellation, and broader FFI.
