---
phase: 32
review: 32-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "Freeze gate pathspec includes `.md`, contradicting the digest definition it claims to follow"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "`codedigest.sh` is cwd-dependent and cannot detect an empty listing"
  - id: IN-01
    severity: info
    disposition: fixed
    title: "Env override of the expected digest in the production gate"
  - id: IN-02
    severity: info
    disposition: fixed
    title: "Living-claim marker only matches at column 0"
  - id: IN-03
    severity: info
    disposition: fixed
    title: "Formatter guard has no negative-context tests"
  - id: IN-04
    severity: info
    disposition: fixed
    title: "Retirement citation check is satisfied by any `| superseded |` line mentioning the name"
  - id: IN-05
    severity: info
    disposition: fixed
    title: "Process-wide index cache"
open: 0
total: 7
recorded: 2026-10-12T00:00:00Z
---

# Phase 32: Code Review Disposition

All seven findings were fixed in Phase 32.1 (plan 32.1-06) after the M006 receipts were reopened, so the digest paths could change. Each row names the test that witnesses it. The earlier carrier todo, `2026-10-12-phase32-review-hardening-after-m006.md`, is now in `.planning/todos/completed/`.

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | fixed | Fixed in 32.1-06; witnessed by TestPhase32FreezeGate. |
| WR-02 | warning | fixed | Fixed in 32.1-06; witnessed by TestPhase32CodeDigest. |
| IN-01 | info | fixed | Fixed in 32.1-06; witnessed by TestPhase32ReceiptGate. |
| IN-02 | info | fixed | Fixed in 32.1-06; witnessed by TestPhase32ReceiptLivingClaims. |
| IN-03 | info | fixed | Fixed in 32.1-06; witnessed by TestPhase32FormatVarBody. |
| IN-04 | info | fixed | Fixed in 32.1-06; witnessed by TestValidationRowGradesAreEarnedOverArchivedCorpus. |
| IN-05 | info | fixed | Fixed in 32.1-06; witnessed by TestVerificationGroundedness. |
