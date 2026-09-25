# Phase 20 PRC-02 Debt Baseline

## Counted population

The denominator is the exact ten-ID carry-forward table in
`.planning/milestones/M002-MILESTONE-AUDIT.md` §4, combined with the seven
distinct unowned rows newly present in the live M003 `PHASE-14-DEBT.md`.
Current dispositions are read from each source register's `Landing phase` cell;
only `UNOWNED(...)` rows count. The qualified starting population is 13 unique
open, unowned IDs:

| Source | IDs | Count |
|---|---|---:|
| M002 audit carry-forward (open/unowned at phase start) | D-10-C04, D-11-02, D-11-27, D-12-36, D-12-43, D-13-34 | 6 |
| Live M003 additions in Phase 14 register (open/unowned at phase start) | D-14-45, D-14-46, D-14-47, D-14-50, D-14-51, D-14-52, D-14-54 | 7 |
| **Qualified starting total** | **13 unique IDs** | **13** |

Starting open/unowned IDs (before the Phase 20 dispositions): `D-10-C04`, `D-11-02`, `D-11-27`, `D-12-36`, `D-12-43`, `D-13-34`, `D-14-45`, `D-14-46`, `D-14-47`, `D-14-50`, `D-14-51`, `D-14-52`, `D-14-54`

The ten historical audit-cohort IDs are D-10-C04, D-11-02, D-11-27,
D-11-51, D-12-21, D-12-36, D-12-43, D-13-02b, D-13-10a, and D-13-34.
Four already have a current `CLOSED(...)` or `P<NN>` disposition and therefore
do not count as open/unowned. `TestPhase20UnownedDebtPopulation` pins both the
ten-ID historical cohort and the exact 13-ID open/unowned starting set, while
deriving each current disposition from the source registers.

## Scope distinctions

`.planning/UNREACHABLE-CLAIMS.md` is generated from probe-backed Grade/Witness
rows. Its eight `UNOWNED(...)` entries are a qualified subset, not the PRC-02
denominator. The raw archival count of 40 `UNOWNED(...)` cells includes older
M001/M002 liabilities outside the M002 audit's explicit current carry-forward;
counting those would substitute archive history for the current milestone
boundary. The gate rejects duplicate qualified source rows, missing source
rows, unknown landing syntax, and drift in the audit's ten IDs.

## Dispositions in this plan

At the Phase 20 execution start, 13 qualified IDs were open and unowned. After
the evidence-backed Phase 16 ownership reconciliation and Phase 21's explicit
ownership of D-14-45, the current source-derived set is nine:
D-10-C04, D-12-43, D-13-34, D-14-46, D-14-47, D-14-50, D-14-51, D-14-52,
and D-14-54. Four stale validation citations (D-14-50/51/52/54) are the
remaining QLT-10 repair targets in Plan 07; that route reaches five without
using D-13-34's human decision. D-10-C04 remains open and D-13-34 remains
pending Plan 06.
