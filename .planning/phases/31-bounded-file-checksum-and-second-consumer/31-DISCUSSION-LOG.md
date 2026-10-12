# Phase 31: Bounded File Checksum and Second Consumer - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-10-12
**Phase:** 31-bounded-file-checksum-and-second-consumer
**Areas discussed:**
- library home and second consumer;
- file-error outcomes and exit codes;
- path input and adapter surface;
- evidence, identity and controls.

**Mode:** advisor fan-out, with one `gsd-advisor-researcher` per area at the
minimal_decisive tier, plus a codebase scout. Every recommendation was
auto-followed under the user's standing fan-out boilerplate, which was
re-pasted this session. No AskUserQuestion was issued.

---

## Library home and second consumer

| Option | Description | Selected |
|--------|-------------|----------|
| Move the library and CLI into `examples/checksum/` | All consumers sit beside one library (needed by the D-05 entry-dir root) | ✓ |
| Freeze the phase30 copy and add a new library | Two live copies, which contradicts MOD-06 | |
| `adler_low` = `adler32 % 65536` | A half; `Wikipedia` → 920 | ✓ |
| `adler32 % 65521` | Mixes both halves, but has no user meaning | |

**Choice:** Move, plus `adler_low`. A dual-flip control and a reverse
(B-only) control are added.

## File-error outcomes and exit codes

| Option | Description | Selected |
|--------|-------------|----------|
| Every outcome exits 65, distinguished by a named alternative and fixed stderr | Phase 23 precedent, no table change | ✓ |
| sysexits split (66/71/74) | Adds rows to the one-way table and causes exit-pin ripple | |

**Choice:** 65 for all outcomes, with `FileError =
FileTooLarge | FileUnavailable | AllocFailed`.

**Notes:** The researchers disagreed on the stderr text: new `checksum: …` lines
against the existing generated `Type.Alternative` line. The synthesis chose the
existing convention, so both engines share one mechanism.

## Path input and adapter surface

| Option | Description | Selected |
|--------|-------------|----------|
| The entry type `PathToken` selects the input kind | Already built in cgen and native | ✓ |
| A `--path` flag | Duplicates a fact and allows drift | |
| A copy-local foreign acquire/view/release triple | Same shape as Phase 30 | ✓ |
| A `read_file` builtin | New surface that collapses the FFI boundary | |
| An independent interpreter read (O_NONBLOCK, IsRegular, limit+1) | A real parity read | ✓ |
| The interpreter models outcomes only | Not independent | |
| Fold the `run --engine` input todo | Re-pins the Phase 28 manifest | |

**Choice:** PathToken entry, the foreign triple, and the independent
interpreter read. The todo stays pending.

## Evidence, identity and controls

| Option | Description | Selected |
|--------|-------------|----------|
| A trailing omitempty `Modules[]` list, sorted by name | No-import bytes unchanged | ✓ |
| Replace `SourceDigest` with one closure digest | Breaks every golden | |
| Extend the Phase 29 mutant harness | Reuse | ✓ |
| An external mutation tool | Adds a dependency | |
| Literal rows with `cmd` and `python` fields plus a Go `hash/adler32` cross-check | Provenance is enforced | ✓ |
| Run the cap on every CI run | 376 ms worst case | ✓ |
| Put the cap behind an env opt-in | Leaves the cap unproven in CI | |

**Adversarial findings:**
- A literal `-fsigned-char` flag is an equivalent mutant. The retype mutant
  stays the must-fail control, and a flag lane that must pass is added.
- The ASan deleted-bounds-check lane needs a non-empty out-of-bounds input.

## Claude's Discretion

- Adapter, owner and symbol names.
- Plan count and wave split.
- Helper layout of the case matrix.

## Deferred Ideas

- `run --engine` input.
- sysexits split.
- Imported foreign declarations.
- The formatter todo (`/gsd-quick` or Phase 32).
