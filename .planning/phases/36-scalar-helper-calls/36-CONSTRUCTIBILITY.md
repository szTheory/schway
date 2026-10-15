# Phase 36 bounded storage decision — 2026-10-14

Status: planning constructibility receipt only. No compiler implementation or feature selector passed here.

Inspected revision: `c499cd69b92fc4a568246fbb70f2f146dc7bd182`. Local tools newly observed: Go 1.24.0 darwin/arm64; Apple Clang **21.0.0** (clang-2100.1.1.101), arm64-apple-darwin25.6.0. The research's 21.0.1 label is not this probe's tool version.

## New scratch observation

Only `/tmp/schway36-storage-probe.c` and its `/tmp` binaries were created. Source SHA-256: `113d88a7b6bbc0ba5911f6014cd7f1956e0acbd99dcfeb682f0d3bdcb23ab639`.

Commands actually run:

- `clang -std=c17 -Wall -Wextra -Werror -fsanitize=address,undefined -O0 /tmp/schway36-storage-probe.c -o /tmp/schway36-storage-probe-o0 && /tmp/schway36-storage-probe-o0` — exit 0.
- `clang -std=c17 -Wall -Wextra -Werror -O3 -flto /tmp/schway36-storage-probe.c -o /tmp/schway36-storage-probe-o3lto && /tmp/schway36-storage-probe-o3lto` — exit 0.
- `/tmp/schway36-storage-probe-o3lto conformance` — exit 69 with `schway: event capacity exhausted` on stderr.

Both successful runs printed four exact PASS lines: `stable-after-return nested-reset two-site`, `path-exact-plus-one`, `ordinal-before-wrap`, and `capture-disabled-work-over-4096-late-defect`. These assert stable saved pointers after nested function return, two parent activations with child ordinal zero, independent call sites, path byte exact-limit/plus-one, event exact-limit/plus-one, INT_MAX ordinal increment before overflow, and 5,001 numeric iterations followed by a simulated defect 65 after optional recording stopped. Sanitizers reported no error. The scratch uses short ASCII path components and a simulated numeric defect: canonical escaping, generated C, real runtime defects, wire size and production runners still require the named implementation tests.

## Chosen mechanics (D-11 through D-14)

Use a fixed, append-only C path byte arena for new scalar-value-call programs. Events retain pointers to arena storage until final serialization; never retain callee-stack pointers. Each emitted function activation owns a zero-initialized counter for each static call site. Build one child path once before entry and reuse it for the child frame and call/terminal attribution. Escape static components using the existing canonical mechanical spelling. Check all additions, remaining capacity, decimal formatting and ordinal representability before advancing or writing. A newly entered parent resets child-site counters. Existing tables remain the static preflight, with 4,096 static nodes; they are not a dynamic runtime count.

For new-call capture, use the existing byte ceiling as the bounded path arena (16 MiB conformance; 64 KiB app evidence). Bound retained event slots by `byteBudget / 128` and retain the existing final encoded-document/output bound; the slot budget is an internal evidence-memory bound, not a promise that every document below the byte ceiling fits, and never a semantic call/input limit. App evidence-disabled or exhausted paths stop all identity/counter/event work and continue real program execution. Conformance evidence exhaustion takes named exit 69; optional app capture reports the existing capacity-exhausted status and preserves the actual process result, streams and later defect. Final serialization retains its existing bounded writes.

Use a portable evidence ordinal ceiling of 2,147,483,647 (INT32_MAX) for newly recorded activations, below both supported Go int widths and C uint64_t. Guard the next value before narrowing/increment; after the maximum representable emitted ordinal, a subsequent capture attempt exhausts evidence. Tests use a scoped small limit seam and a near-maximum counter to reach the guard without billions of calls. This ceiling does not stop ordinary application computation. The wire grammar and existing nonnegative-int parser remain unchanged.

## Interpreter retention audit and decision

Source inspection of `interp.Run`, `RunApplication`, `runProgramFrameStack`, `runFrameStack`, their event appends and `assignEventOccurrences` found an in-memory event slice with no dynamic event/document cap at this driver. `RunApplication` separately bounds application output. `execution.CanonicalBytes` marshals after execution; it does not bound retention. This is a source finding, not a new execution result.

For the newly admitted scalar-call path, add an internal bounded event collector and bounded path accounting using the 16 MiB execution-document budget. Count retained paths/encoded event bytes before append; never retain a partial trace as complete. On exhaustion, disable evidence collection/counters/path construction, continue numeric/application execution, and return the computed Execution/ApplicationResult together with a named `interp.evidence_capacity_exhausted` error. No wire field/schema is added. Ensure RunApplication finishes automatic decimal output and preserves real terminal defect/output even when the returned error denotes incomplete evidence. Existing callers already refuse error-bearing evidence; no new public file route or semantic iteration limit follows. File/fixture entry points using the same driver retain the same fail-closed result/error contract. Existing old-call/no-new-call behavior stays on its current branch.

The interpreter and executionpeer derive counters independently; only canonical mechanical formatting may be shared. Executionpeer bounds processing by observed-document/event bytes and does not unfold dynamic activations under the static 4,096 ceiling. Depth 128, output limits, static cycle/depth rules and all resource restrictions remain distinct.

## Evidence still required

Plans 36-04/05 must implement real compiler tests for exact canonical paths, all call seams, integer overflow guards, path/event encoded boundaries and post-exhaustion real defects. Plan 36-06 must prove numeric values and actual defects independently at O0/O3/O3-LTO. The scratch result cannot stand in for these tests, peer acceptance, Linux, hosted binding, performance distributions or human readability.
