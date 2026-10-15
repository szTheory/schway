# Phase 37 — UI Review

**Audited:** 2026-10-15  
**Baseline:** Canonical `ui.safety-gate37` applicability result; no UI-SPEC.md  
**Screenshots:** Not captured — `frontend=false`, `hasUiFiles=false`, `hasUiSpec=false`, `block=false`  
**Interaction captures:** off (`interaction_capture=false`)

## Applicability

Phase 37 adds compiler and interpreter behavior, public CLI routes, CRC source examples, and an executable text guide. The explicit changed-file inventory contains no frontend source, and the repository has no UI-SPEC.md, React component files, stylesheet files, `components.json`, or frontend package manifest. No browser surface exists to score. Visual and interaction pillars are therefore not applicable; this review does not compute an overall score out of 24.

The canonical gate and prior wave records are applicability evidence only, not evidence of visual quality. No screenshot or browser interaction was attempted. The phase's CLI behavior and guide have objective route, tuple, and parser checks, which are summarized separately below.

## Pillar Applicability

| Pillar | Score | Basis |
|--------|-------|-------|
| 1. Copywriting | N/A | No UI copy surface. CLI/documentation consistency is checked below without scoring subjective readability. |
| 2. Visuals | N/A | No frontend view or visual component exists in this phase. |
| 3. Color | N/A | No UI color tokens or rendered UI exist in this phase. |
| 4. Typography | N/A | No UI type scale or rendered text layout exists in this phase. |
| 5. Spacing | N/A | No UI layout or spacing system exists in this phase. |
| 6. Experience Design | N/A | This is a CLI feature, not a screen-based interaction flow; CLI outcomes have separate automated evidence. |

**Overall:** Not scored; UI pillars are not applicable to this phase.

## Applicable Objective Finding

### WARNING — Guide overstates where generated outputs are written

[examples/crc32/README.md:16] says all generated files and native artifacts stay under `.schway/crc32/`, but the setup fence at [examples/crc32/README.md:21] writes the CLI to `./schway` at the repository root. `.gitignore` ignores that root executable at [.gitignore:2], so it is not exposed to accidental commits, but the location statement is inaccurate. The guide's own clean-checkout harness pins this root output in [main_phase37_guide_test.go:247] and checks for the executable at the copied checkout root in [main_phase37_guide_test.go:579].

**Fix:** Clarify that the ignored CLI executable is written to the repository root, while fixtures and retained native artifacts are under `.schway/crc32/`; alternatively, move the CLI output into `.schway/crc32/` and update the marked run commands and guide harness together.

## Evidence Boundaries

- **Source inspection in this review:** The public usage string lists both file routes at [main.go:927], the specific file-run usage is present at [main.go:183], and the guide's route templates and outcome table are at [examples/crc32/README.md:62] and [examples/crc32/README.md:86]. The setup-output mismatch above was found by comparing guide text, the actual setup command, the harness's asserted output path, and `.gitignore`.
- **Existing recorded checks, not rerun for this review:** Phase 37 records `TestPhase37GuideCleanCheckout` passing at C22 and C24, plus JSON/parser controls at C23 and C24; see [37-REMEDIATION-EVIDENCE.md:96] and [37-REMEDIATION-EVIDENCE.md:134]. The phase summary reports 78 guide tuples checked across both consumers and routes at [37-08-SUMMARY.md:94] and [37-08-SUMMARY.md:107]. These receipts support CLI/documentation behavior, not visual quality.
- **Phase state:** The plan summaries retain APP-13, MOD-07, APP-14, and DX-24 as pending parent acceptance; this review does not change that status.

## Files Audited

- `AGENTS.md`, `CLAUDE.md`
- `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-CONTEXT.md`
- `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-09-PLAN.md`
- `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-07-SUMMARY.md`
- `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-08-SUMMARY.md`
- `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-09-SUMMARY.md`
- `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-WAVE-CHECKS.md`
- `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-REMEDIATION-EVIDENCE.md`
- `/tmp/schway37-review-files-explicit.txt`
- `README.md`, `examples/crc32/README.md`, `.gitignore`
- `cmd/schway/main.go`, `cmd/schway/main_phase37_guide_test.go`

## UI REVIEW COMPLETE

**Phase:** 37 — CRC-32 Consumer, Interpreter File Route, and Guide  
**Applicability:** No frontend/UI surface; visual pillars not applicable  
**Screenshots:** Not captured  
**Interaction captures:** off  
**Objective findings:** 1 warning  
**File created:** `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-UI-REVIEW.md`

## Finding disposition —2026-10-15

The single objective warning is fixed: guide prose now names ignored ./schway for the CLI and .schway/crc32/ for inputs/retained artifacts. Executable fences are unchanged. Parent focused JSON/parser checks verify consistency after the prose edit; no visual score or humanUAT added.
