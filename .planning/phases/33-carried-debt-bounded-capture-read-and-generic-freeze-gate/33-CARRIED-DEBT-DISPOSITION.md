# Phase 33 Carried Debt Disposition

Date: 2026-10-13. Phase 33 binds no hosted receipt.

| Item | Source | Disposition | Witness | Plan |
| ------ | ------ | ------------- | --------- | ------ |
| 32.1 WR-01 | Phase 32.1 review | fixed | `TestFreezeGate` subdirectory and m006-real-repo rows | 33-03 |
| 32.1 WR-02 | Phase 32.1 review | fixed | `TestPhase33AbortLikeNarrowed`, `TestPhase33AbortLikeWideningMutantKilled`, `TestPhase33DefectParityFixtureRoute`, `TestPhase33DefectParityFileRoute` | 33-02 |
| 32.1 IN-01 | Phase 32.1 review | covered, no code change | `TestPhase32PublicRouteAudit` seeded uncovered-route and set-drift cases | 33-05 |
| 32.1 IN-02 | Phase 32.1 review | fixed | `TestPhase321RunInputRequiredCLI` | 33-01 |
| 32.1 IN-03 | Phase 32.1 review | fixed | `TestFreezeGate` md-only-change-after-freeze-accepted and hash-length-derived rows; `TestPhase32CodeDigest` | 33-03 |
| coverage-kind pin | Phase 32.1 verification | fixed | `TestSummaryCoverageGroundedness` bad-kind control | 33-01 |
| Phase 29 WR-02 | Phase 29 review | fixed in 32.1-03, verified | `TestPhase321ByteViewEntryRefused` | 33-05 |
| Phase 29 WR-03 | Phase 29 review | fixed in 32.1-04, verified | `TestPhase321DefectParityWithLiveOwner` | 33-05 |
| Phase 28 VALIDATION.md | Phase 28 evidence | written, retroactive | `TestPhase20ValidationLifecycle`, `TestVerificationGroundedness` | 33-05 |
| Phase 28 SECURITY.md | Phase 28 evidence | written, retroactive, no new tests | `TestPhase20ValidationLifecycle`, `TestVerificationGroundedness`; no Go test parses SECURITY documents, so the witnesses are these groundedness checks and the 28-VERIFICATION/28-REVIEW evidence cited by each threat | 33-05 |
| QLT-13 bounded capture read | QLT-13 | fixed | `TestPhase33CaptureReadBounded` | 33-01 |
| EVD-16 generic gate | EVD-16 | fixed | `TestReceiptGateArchivedM006`, `TestNoCheckPhase32Caller` | 33-03/33-04 |
