# Phase 25: Separate Pointer Successors and Integrated Utility - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-10-01
**Phase:** 25-separate-pointer-successors-and-integrated-utility
**Areas discussed:** Integrated utility, pointer helper contract, diagnostics,
evidence and onboarding

---

The user supplied a general fan-out preference: consider relevant technical and
product roles, tradeoffs, patterns and footguns, make an adversarial pass, and
synthesize recommendations in line with Schway's goals. Three read-only
specialist reviews covered pointer semantics, utility composition, and
diagnostic/evidence UX. Current source was inspected; no project tests, builds,
or CI were run during discussion. The user selected all four recommendations.

## Integrated utility

| Option | Description | Selected |
|--------|-------------|----------|
| Compose both families in one utility | Transfer and use the existing file-byte owner, then pass the returned U64 through shared and exclusive read/copy helpers in sequence. | ✓ |
| Separate family-specific applications | Keep positive pointer-family examples as separate applications beside the Phase 24 utility. | |
| C wrapper performs composition | Let a new C wrapper call both helpers and hide the sequence from Schway source. | |

**User's choice:** Adopt the recommended single utility composition. Each
helper's result must contribute to the result. It borrows the copied U64, not
the owner allocation. Keep 0x41→65, 0x42→66, and the inherited 0x43 typed use
failure; that failure occurs before the infallible pointer helpers.
**Notes:** The Phase 24 caller checker currently recognizes a fixed two-binding
shape, so bounded admission changes must reach that caller and its independent
peers alongside emitter admission. This is source-inspection evidence only.

---

## Pointer helper contract

| Option | Description | Selected |
|--------|-------------|----------|
| Separate ordinary Schway helpers | Use existing shared borrow and exclusive borrow mut forms in two named, read/copy-only helpers; lower each checked source shape to actual pointer-parameter C. | ✓ |
| New access-mode parameter or annotation | Add a generalized mode-bearing interface to express both families. | |
| New foreign-C pointer contracts | Introduce a second foreign ABI surface for the two helper families. | |

**User's choice:** Adopt the separate helper recommendation. Keep shared and
exclusive witnesses, peer checks, and controls distinct. Exclusive access
prevents overlapping access; it does not authorize mutation here. Add no
unsupported optimizer attributes.
**Notes:** Existing function parameters are ordinary by-value source
parameters; the pointer C shape comes from the checked native lowering. The
historical restrict golden is not the Phase 25 contract. A shared loan may end
before the exclusive one starts; an overlapping pair is rejected.

---

## Diagnostics

| Option | Description | Selected |
|--------|-------------|----------|
| Extend existing structured records | Use stable codes, a primary offending span, borrow/conflict causes, and actionable boundary text. | ✓ |
| Prose-only refusal documentation | Explain refused cases only in the example documentation. | |
| Add a diagnostic schema or repair framework | Introduce a new protocol or generalized remediation surface. | |

**User's choice:** Adopt the existing structured diagnostic route. Distinguish
unsafe borrow conflicts from unsupported but otherwise safe shapes; avoid
repairs unless an edit is demonstrably safe.
**Notes:** Preserve existing diagnostic wire identifiers and compatibility.

---

## Evidence and onboarding

| Option | Description | Selected |
|--------|-------------|----------|
| One utility quick start and family-indexed evidence | Share setup, then show each family's source, expected result, host, and applicable lane separately. | ✓ |
| Independent documentation and workflow per family | Duplicate setup and evidence instructions for each family with no shared index. | |
| One combined pass label | Report one Phase 25 result without family-level status detail. | |

**User's choice:** Adopt one quick start with a family/host/lane matrix and
incomplete states when receipts are absent. Reuse current CI lanes when they
answer the same evidence question.
**Notes:** Each helper needs a reached wrong-result control plus its own
conflict/escape controls and actual emitted pointer C. Phase 24's physical
observer proves owner allocation/use/free; it does not prove by-pointer scalar
behavior. Clang's sanitizer evidence is a distinct native evidence lane, not a
replacement for semantic checks.

---

## the agent's Discretion

- Choose names, minimal grammar/admission extensions, diagnostic wording,
  focused fixture placement, public command layout, and CI lane ownership while
  preserving the accepted bounded composition.
- Follow the user's standing preference for Go standard-library/copy-local
  implementations; a dependency needs a concrete justification.

## Deferred Ideas

- Direct pointers into the owned buffer, mutation, forwarding, retention,
  callbacks, nonlocal exits, and wider pointer shapes remain outside Phase 25.
- Arithmetic, loops, FizzBuzz, general bytes/arrays/strings, JSON, and reusable
  libraries remain future consumer-led capabilities.
