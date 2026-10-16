# Phase 37.1 counted reason storage repair

**Date:** 2026-10-16

**Scope:** T-37.1-08 bounded compiler storage correction

The security review found that the counted structured-reason and application
stderr helpers formed automatic compound-literal byte arrays before their
existing write bounds were applied. Both helpers now emit a small block-scoped
`static const unsigned char` array and pass its `sizeof` length to the existing
counted writer. The application path uses the same static storage for its
prefix, reason bytes, and newline. The ordinary printable-ASCII path is
unchanged. Capture handling, output limits, exits, and exact byte semantics
remain in place.

`TestPhase371CountedDefectReasonBytes/static_readonly_counted_storage` checks
generated native and application C from the small U+0001+A fixture. It requires
the named static readonly buffers and their counted uses, and rejects the
former reason-specific automatic compound-literal forms. The old helper was
run through the adapter's exact original-source overlay and failed this
specific invariant; the overlay also retained the independently required
parser/core semantic REDs. The overlay manifest records its baseline revisions,
source hashes, and successful source-restoration check in the machine evidence.

The post-fix cgen command passed all three requested roots. The counted-reason
root preserved the literal native record and application stderr bytes for its
small ASCII, control, NUL, non-ASCII, replacement-character, and backslash
fixtures across no-call, dynamic-entry, and dynamic-helper witnesses. The
existing native O0/O3/O3-LTO lanes and application tuples passed. The focused
ASan/UBSan lane was available on this host and exited cleanly. The separate
six-root frozen-baseline command passed; the immutable census and amendment
files were not changed.

Fresh `--prove all` produced green, unflagged canonical test-tier evidence for
all four existing prohibitions. Bad, clean, and default controls passed their
expected semantic lanes without skips. A subsequent independent `--sample all`
observed the same 19-root union: 254 passing subtests, zero skips, with 93
executionpeer, 70 session, and 91 cgen subtests. The counted-reason root now
has 76 subtests, one more than its preceding 75-subtest snapshot. The added
case is `static_readonly_counted_storage`; the other 75 are the existing
reason, witness, and optimization cases.

The live lifecycle remains `resolved/test`. The original projection remains
`unresolved` with null verification, and the original pre-resolution proof
snapshot remains hash-bound and unchanged. The resolution record appends three
verified source-evolution entries: the first captures the emitter correction,
the second captures the combined native/application invariant refinement, and
the third preserves the adapter's two-state lifecycle support. Each entry
preserves its previous resolved proofs and root-union receipt before replacing
current proof records.
The final source bindings and canonical producer outputs are recorded in
[`371-STORAGE-EVIDENCE.json`](371-STORAGE-EVIDENCE.json), the current
[`371-PROHIBITION-PROJECTION.json`](371-PROHIBITION-PROJECTION.json), and
[`371-PROHIBITION-RESOLUTION.json`](371-PROHIBITION-RESOLUTION.json).

No memory-pressure experiment, large-input construction, stack-exhaustion
reproduction, or native stress probe was run. The earlier diagnostic native
signal's cause remains unestablished. This repair and its local proofs do not
constitute the parent-owned security rerun, independent review, or phase
completion.
