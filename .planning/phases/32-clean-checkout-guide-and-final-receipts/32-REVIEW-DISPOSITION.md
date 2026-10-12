---
phase: 32
review: 32-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: deferred
    title: "Freeze gate pathspec includes `.md`, contradicting the digest definition it claims to follow"
  - id: WR-02
    severity: warning
    disposition: deferred
    title: "`codedigest.sh` is cwd-dependent and cannot detect an empty listing"
  - id: IN-01
    severity: info
    disposition: deferred
    title: "Env override of the expected digest in the production gate"
  - id: IN-02
    severity: info
    disposition: deferred
    title: "Living-claim marker only matches at column 0"
  - id: IN-03
    severity: info
    disposition: deferred
    title: "Formatter guard has no negative-context tests"
  - id: IN-04
    severity: info
    disposition: deferred
    title: "Retirement citation check is satisfied by any `| superseded |` line mentioning the name"
  - id: IN-05
    severity: info
    disposition: deferred
    title: "Process-wide index cache"
open: 0
total: 7
recorded: 2026-10-12T00:00:00Z
---

# Phase 32: Code Review Disposition

All seven findings are deferred. The reviewer found that none affects the soundness of the recorded receipt (run 38315799626, codedigest c1b4e85d…). Each one either fails closed or is a hardening or coverage gap. The code is frozen at 079e112f, so fixing any of them now would touch digest paths and reopen the M006 receipts (D-10a). They are carried as one pending todo, `.planning/todos/pending/2026-10-12-phase32-review-hardening-after-m006.md`, for the next milestone.

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | deferred | Fails closed: it only rejects a post-freeze `.md` edit. Tracked in todo 2026-10-12-phase32-review-hardening-after-m006. |
| WR-02 | warning | deferred | Both real callers cd to the repo root first, so the recorded digest is correct. Tracked in the same todo: add a root cd, an empty-listing check, and an armed control. |
| IN-01 | info | deferred | Hardening: say in the pass line where the digest came from. Same todo. |
| IN-02 | info | deferred | Coverage gap in the D-10a marker scan. Same todo. |
| IN-03 | info | deferred | Coverage: pin the match, nested-arm and if/while contexts. Same todo. |
| IN-04 | info | deferred | The digest binding is unaffected. Same todo: tie the check to the citing row. |
| IN-05 | info | deferred | Latent only. Same todo. |
