# Phase33 CR-02 repair — 2026-10-15

Independent standard-depth review confirmed terminal false accepts in the
execution peer. A temporary overlay probe accepted returned events labeled
as defect/typed_failure, a forged return source/type, and a user-defect reason
unrelated to core. This was reproduced locally, not inferred from review.

## Bounded repair

The named safety obligation is structural terminal attribution in
executionpeer.Validate. OpReturn, OpFail and OpDefect now bind source/type and
an empty target to core; explicit defects bind their reason. Root terminal
kind must agree with the document outcome; defects require an empty outcome
value. Legacy and dynamic activations close and reject later events. Child
returns remain independent of the root's eventual outcome. Numeric returned
and failed values remain outside this peer's claim and 999 stays accepted
when structurally plausible. No language admission, wire field, new dependency
or execution backend changed.

TestPhase33TerminalAttribution adds clean and adversarial return/failure/
defect cases. The actual pre-fix peer source overlay fails30 test nodes;
clean repaired controls and existing dynamic-call diagnostic controls pass.
Two positive synthetic fixtures now supply the outcome or attribution fields
already present in real producer documents. The independent closure inspected
exactly four repair files and closed CR-02. The initial25-file scope and
four-file closure have28 cumulative distinct paths, with only peer.go in both.

## Remaining review findings

CR-03 (legacy nested defect stack unwinding) and CR-04 (modeled typed failure
through OpReturn) remain confirmed source-inspection findings, classified
BLOCKER by the reviewer. They predate this repair; they are false rejections
in older non-scalar propagation paths, rather than the CR-02 false accepts.
They are deferred to the execution-peer maintainer through the mandatory
M007 audit before Phase38 freeze. That owner must choose scoped gap work and
independent producer/core propagation controls. In particular CR-04 cannot
be repaired by broadly admitting failed return events without validating the
failure contract and caller error edge. These are not fixed, accepted risks,
or a clean review. The Phase33 verifier must independently assess whether
they invalidate any current goal claim; review disposition alone is advisory.

## Evidence

33-REFRESH2-RUN.json binds all18 mapped rows plus current inherited selectors,
all18 execution-peer roots and metadata controls:184 names /196 executions
in16 packages, zero missing/fail/skip. It binds current working-tree inputs
and raw logs, including actual-source RED and clean GREEN. The harness
corrected a nonexistent metadata selector and a receipt schema assumption;
passed logs were retained within this invocation and checked against exact
root sets. No historical passing result supplied execution credit. The final
harness resume duration is not a full-run latency measurement.

Original plans/summaries/UAT and source-bound receipts are preserved. This
is an execution deviation for a named safety obligation, not a new feature
phase or a scope expansion into general error propagation.
