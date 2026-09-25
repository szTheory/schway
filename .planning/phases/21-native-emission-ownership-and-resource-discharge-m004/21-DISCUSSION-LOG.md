# Phase 21: Native Emission Ownership and Resource Discharge (M004) - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-25
**Phase:** 21-native-emission-ownership-and-resource-discharge-m004
**Areas discussed:** Machine-checkable resource-discharge contract, foreign-boundary exit and cleanup classification, D-14-45 one-TU / -flto evidence scope

---

## Machine-checkable resource-discharge contract

| Option | Description | Selected |
|--------|-------------|----------|
| Machine-checkable contract | Machine-checkable acceptance cases/refusal rules, with executable native emitter witnesses reserved for later admission work. | ✓ |
| Design documentation only | Normative contract prose without a machine-checkable acceptance surface. | |
| Add executable native emitter witnesses now | Exercise emitter behavior within Phase 21. | |

**User's choice:** Machine-checkable contract.
**Notes:** Structural contract checks establish completeness and consistency; they do not prove runtime cleanup behavior or authorize admission.

---

## Foreign-boundary exit and cleanup classification

| Option | Description | Selected |
|--------|-------------|----------|
| Classify modeled exits; refuse unproved paths | Explicitly classify normal/error returns, supported unwind, nonlocal transfer, defects/cancellation, and process termination; admit only paths with checked discharge rules. | ✓ |
| Support normal and error returns only; forbid unwind | Keep the contract to explicit return paths and forbid unwinding. | |
| Specify broader unwind/cancellation semantics now | Expand Lang's semantic scope to define more unwind/cancellation behavior in this phase. | |

**User's choice:** Classify modeled exits; refuse unproved paths.
**Notes:** This decision classifies existing/modelable behavior; it does not create new Lang unwind or cancellation semantics.

---

## D-14-45 one-TU / -flto evidence scope

| Option | Description | Selected |
|--------|-------------|----------|
| Scoped measurement plus structural CI guard | Preserve the one-TU/no-restrict/refusal boundary, measure one emitted-fixture compiler comparison, and run the structural guard in CI; defer broader recurring matrices until justified. | ✓ |
| Preserve current evidence and boundary only | Carry forward existing evidence without a new emitted-fixture comparison. | |
| Build a recurring multi-host/toolchain LTO matrix | Add a broader recurring measurement matrix across hosts and toolchains. | |

**User's choice:** Scoped measurement plus structural CI guard.
**Notes:** Keep behavioral equivalence distinct from optimization/performance claims and bound conclusions to measured fixture, toolchain, and host lanes.

---

## the agent's Discretion

Contract encoding, structural assertions, fixture details, and CI lane placement, provided they preserve the user decisions and Phase 21's no-emitter-admission boundary.

## Deferred Ideas

None.
