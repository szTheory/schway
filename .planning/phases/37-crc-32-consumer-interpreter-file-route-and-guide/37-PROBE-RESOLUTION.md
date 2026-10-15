# Phase 37 probe resolution — planning only

Fallback is ON (key absent). Both canonical SPEC sections were absent, and the deterministic raw report is preserved unchanged in 37-EDGE-COVERAGE.json: 9 applicable, 0 resolved, 9 unresolved, 2 unclassified. 37-EDGE-RESOLUTION.json is a separate derived planning disposition: 7 explicit criteria, 2 unresolved/unclassified, zero backstops/dismissals. This is acceptance design, not executed evidence.

## Edge lift and no-silent-drop

| Requirement/category | Explicit acceptance / assumption | Plan tasks |
|---|---|---|
| APP-13/boundary | Cap-1/cap/cap+1 actual bytes; exact numeric and refusal tuples, no answer for oversize. | 37-04-01;37-05-01;37-05-02 |
| APP-13/precision | Exact unsigned U64 state in 0..4294967295, high-bit vectors, exact low16; literal independent IEEE answers. | 37-01-01;37-05-01;37-05-02 |
| APP-13/concurrency | Two bounded concurrent public applications with isolated inputs/cwd/artifacts preserve independent literal answers; arbitrary interpreter cancellation remains unproved. | 37-05-02 |
| MOD-07/unclassified | FLAGGED unresolved classifier assumption; no auto-resolution/backstop. Actual concrete reuse/guide requirements are nevertheless fully planned. | 37-06-02 |
| APP-14/adjacency | First separator owns prefix; suffix token -- and option-shaped strings remain one opaque distinct operand. | 37-03-01 |
| APP-14/empty | Empty file yields numeric zero for CRC; empty path is65/empty/empty; zero/multiple suffix operands are64. | 37-03-01;37-05-02 |
| APP-14/ordering | Program output chunks keep order before later typed failure/checked defect independently of capture, with exact literal streams. | 37-02-02;37-03-02;37-05-02 |
| APP-14/concurrency | Parallel CLI cases are isolated and bounded; native deadline/launch outcomes retain their codes; no claim of generic interpreter interruption parity. | 37-03-01;37-05-02 |
| DX-24/unclassified | FLAGGED unresolved classifier assumption; no auto-resolution/backstop. Actual concrete reuse/guide requirements are nevertheless fully planned. | 37-08-01/02 |

Each of the seven resolved criteria is also represented by plain must_haves.truths in its owning plan. MOD-07 and DX-24 retain flagged must_haves.assumptions. The equality is 9 surfaced = 7 authored criteria + 2 flagged assumptions. Native deadline and bounded case execution do not prove arbitrary interpreter cancellation.

## Prohibition recall → precision

Each row lists the deliberately broad recall candidates, then precision disposition. Routine correctness goes to edges/code review; canon authority/security goes to the security plan. Only bespoke evidence/claim integrity remains.

| Requirement | Stage 1 recall (ten candidates) | Stage 2 precision |
|---|---|---|
| APP-13 | wrong polynomial; wrong reflection; missing final XOR; signed output; wrong empty value; off-by-one loop; host CRC masquerading as source; crypto claim; causal acceptance sold as numeric truth; oversized read | Numeric/loop/empty/byte-bound concerns are ordinary contract checks; source/crypto constraints are already explicit D-01, not newly minted fallback prohibitions. KEEP causal-versus-numeric transparency as P-02. |
| MOD-07 | copied loops; self-import; stale digest; second consumer unused; unchanged low bits; module escape; import symlink; copied manifest drift; wrong low16; unstable identity | Sharing/digest/low16 are normal requirement correctness, explicitly planned in37-06; module escape/symlink are canon security. No additional bespoke prohibition. |
| APP-14 | shell interpolation; flag consumption; arbitrary foreign C; manifest ignored; test oracle imported; FIFO hang; unbounded read; capture exhaustion sold as success; late defect hidden; private path echoed | Canon security/authority issues go to source/reader/CLI STRIDE controls. KEEP the two outcome/capture candidates as one P-01 safety/transparency constraint. Other cases are ordinary contract correctness. |
| DX-24 | untracked helper; ignored build reused; prose commands not executed; wrong table; missing fence; stale receipt; missing stderr; guessed pass counts; omitted prerequisite; false clean-checkout statement | These are explicit D-20/D-21 documentation/evidence correctness and guide drift obligations, not a new bespoke values policy. No extra fallback prohibition; all are planned in37-08. |

Canon breadcrumb: injection, traversal, read authority and private-path disclosure are owned by $gsd-secure-phase and the Go/CLI exact authority/input controls (no JavaScript application or eslint surface exists). No duplicate canon prohibition is minted.

## Canonical projection and early enforcement

37-PROHIBITION-PROJECTION.json is the exact output of installed probe-core.cjs projectProhibitions for the two recalled unresolved items. The same scalar projections appear as sibling must_haves.prohibitions in37-01-PLAN.md. No check_* descriptor is authored at planning time. Both currently dispose flagged/unverified, and neither is dismissed.

- P-01: optional capture exhaustion must not become success or hide a later checked defect. Plan37-01-03 creates real supported native bad/clean/default proof early;37-02-02 extends it to the new interpreter result path.
- P-02: causal structural acceptance must not be sold as independent CRC numeric truth. Plan37-01-03 uses the existing causal-limit assertion and exact overlay control;37-08-03 reruns final source-bound canonical producer evidence.

The raw/history count is 2 applicable, 0 resolved, 2 unresolved. Successful execution may produce supported descriptors via projectProhibitions and real canonical proof receipts, while preserving this initial history. No plan metadata asserts a nonexistent check is already wired. Missing sources, nonunique anchors, compile/tool failure, timeout, missing/skip/zero roots and inert mutants remain fail-closed. No global GSD change or human approval fabrication is authorized.

## Remaining flagged assumptions

MOD-07 and DX-24 classifier rows remain unresolved, as required. Linux syscall runtime behavior is not demonstrated locally. Maximum-case latency is unmeasured until37-02-01 runs. Arbitrary interpreter cancellation/timeout parity, subjective readability, cold/warm whole-loop latency, complete large-file causal traces, host Go allocation-failure recovery and Darwin leak detection remain outside the claim. Phase38 owns final audited hosted binding.

