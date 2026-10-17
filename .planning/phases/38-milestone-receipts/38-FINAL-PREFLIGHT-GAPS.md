# Final pre-freeze source findings

No freeze or remote action is authorized. Earlier independent reports retain their exact source and scope in history. Two source-grounded gaps were found before the new audit.

- PF-01: scripts/check-milestone-binding.sh freeze mode hardcodes the current uncommitted workspace config SHA. TestPhase38FreezeWorktreeControls/config-hash-and-staging copies the caller checkout config and expects acceptance. The committed config SHA is 4cd5087d87f94aa1607a5b9337dee159b8afdc0c9e365472987f7d5569b8d461, while the protected user workspace SHA is 148ccc4c56db8516788c264216ae277d9860d43e00b92de299d8ccff2fd59521. A clean hosted checkout therefore cannot pass that control. Capture the workspace baseline as explicit freeze data and exercise both committed and dirty-config baselines in isolated fixtures; preserve the actual user file unchanged/unstaged.
- PF-02: final mode only invokes validate_run and prints its hosted head_sha. It does not inspect actual HEAD, enforce ingest_base_sha ancestry/planning-only changes or reject final source/worktree drift. Its comment proposes extending it in Plan03, which is prohibited because Plan03 is planning-only. Implement all final gates now with reached positive planning-descendant and wrong-source/nonplanning/history/worktree controls, printing actual candidate HEAD.

Freeze must include reconciled validation documentation in its committed tree: the hosted full suite reads that document, so selecting a source commit that still contains status:draft would fail despite the code digest ignoring Markdown. Choose the actual freeze only after the metadata/audit commit; then rebind the source-equivalent preparation records explicitly to that full SHA before a planning-only freeze-record commit.

These are preparation correctness/portability fixes, not new language, dependency, backend or general evidence scope. Existing EVD-17 ownership and explicit actual-commit approval remain unchanged.

Independent incremental review at efff39f98bf9f7ac9aba6dc5f514d1b49acfc5a6 found two additional blockers (review history preserved):

- PF-03 / CR-01: Preserve the workspace config in history as well as in the current worktree/index. Reject every post-freeze commit that touches .planning/config.json, including commit-and-restore and merge-only changes.
- PF-04 / CR-02: Enumerate every parent diff for committed path history. Default git log --name-only omits merge diffs; a merge-only non-planning create/remove must refuse even with a clean endpoint tree. Confirm the narrow freeze history guard also rejects any digest-input merge-only changes.

The prior SECURED verdict at efff39f9 is historical and is superseded for these two source-grounded findings; no active clearance is inferred from it.

Final disposition: all listed preparation defects and the freeze-commit config edge are fixed at 06e72dda0addd2744106e22ef35fec539337999b; code review clear and security SECURED7/7. Independent ordinary/merge freeze-commit controls pass. Actual freeze, hosted authorization/results and final retained-byte binding remain pending.
