---
phase: 34-agent-diagnostics-floor-and-surface-manifest
plan: "07"
subsystem: compiler-authoring-surface
tags: [go, checkcommandfile, manifest, go-ast, independent-witnesses]

requires:
  - phase: 34-06
    provides: Typed source-free surface manifest, current limits, and route inventory.
  - phase: 34-05
    provides: Seven literal `surface.unsupported_*` refusals and hints.
provides:
  - Independently authored complete-source witnesses for published manifest contexts.
  - Bounded AST branch census for seven named compiler source files.
  - Bidirectional manifest comparisons and target-specific drift controls.
affects: [phase-34-08, phase-35, agent-authoring, surface-manifest]

actuals:
  tokens: 119182
  tasks: 3
  commits: 6
commits: 6
plan_head_before: 2706b089bf342677c7faebd43e9fa9d08375d2b5
plan_head_after: 3fbc55a29eea0371ba6525f82861ce6cc778d612

tech-stack:
  added: []
  patterns:
    - Literal expected source outcomes and constraints remain independent from producer registry output.
    - AST census rows use bounded structural descriptors with explicit surface classifications and reasons.
    - Drift controls mutate copied manifests and source strings, then prove a clean baseline remains.

key-files:
  created:
    - internal/compiler/session/session_phase34_surface_test.go
    - internal/compiler/session/session_phase34_limit_consumers_test.go
    - internal/compiler/syntax/parser_phase34_limit_scope_test.go
    - testdata/phase34/surface-witnesses.json
    - testdata/phase34/surface-census.json
  modified:
    - internal/compiler/surfacemanifest/manifest.go
    - internal/compiler/surfacemanifest/manifest_test.go
    - docs/agent-surface.md
    - .planning/WINDOWS.md

key-decisions:
  - "Publish a manifest context only when an independent complete program passes CheckCommandFile; ability derivation alone is not admission."
  - "Keep the AST census finite to the seven named files and state that it does not prove arbitrary source or files outside the boundary."
  - "Keep drift controls on copied artifacts so test mutations never alter production registry state or source files."

patterns-established:
  - "Source witnesses: author expected forms, contexts, statuses, codes, hints, arities, constraints and limits as literal independent data."
  - "Control assertions: require the expected distinct reason/facet and reject unrelated findings, then re-run clean baselines."

requirements-completed: [DX-23]

coverage:
  - id: D1
    description: The current manifest's admitted facets and contexts are compared against complete CheckCommandFile witnesses with exact refusal codes, hints and constraints.
    requirement: DX-23
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestWitnesses
        status: pass
      - kind: unit
        ref: .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-07-task3-post-commit-final.json
        status: pass
    human_judgment: false
  - id: D2
    description: Every branch descriptor inside the seven-file census is explicitly classified and linked to independent witnesses or a reason for exclusion.
    requirement: DX-23
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestCensus
        status: pass
      - kind: unit
        ref: .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-07-task2-attempt6.json
        status: pass
    human_judgment: false
  - id: D3
    description: Admission, refusal, hint, arity, call-context, loop, limit and source-branch mutations produce their named failure and restore clean baselines.
    requirement: DX-23
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestControls
        status: pass
      - kind: unit
        ref: .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-07-task3-post-commit-final.json
        status: pass
    human_judgment: false

duration: 33min
completed: 2026-10-13
status: complete
---

# Phase 34 Plan 07: Independent Surface Manifest Witnesses Summary

**Schway's surface manifest is now checked against independently authored complete-source witnesses and a bounded compiler-branch census.**

## Performance

- **Duration:** 33 minutes, from the pinned execution start through summary creation.
- **Started:** 2026-10-13T23:52:25Z.
- **Completed:** 2026-10-14T00:25:48Z.
- **Tasks:** 3.
- **Files changed:** 40 plan and evidence files.
- **Actual-size method:** 476,727 characters in the measured task diff divided by four, or 119,181.75; the stored whole-token count is rounded to 119,182.

## Accomplishments

- Added a literal independent matrix that checks complete source programs through `session.CheckCommandFile`, compares both directions with the typed manifest, and pins forms, contexts, arities, inputs, results, constraints, refusal codes and hints. It also independently records all 21 scoped limits and their meanings.
- Added an AST census for exactly the seven planned parser, checker, ability, route and refusal files. Stable branch descriptors require explicit admitted, refused or internal classifications and reasons, while the document states the finite boundary.
- Added copied-artifact controls for missing and invented admissions, every refusal and hint, builtin/call arity, scalar calls, ownership context, required `else`, loop carry restrictions, numeric limits, units, scope/category, an unaccounted operator branch and an emitted refusal-code branch.
- Corrected producer claims exposed by complete-source checks: Box and Pair ability derivations are not executable type admissions, and `if_else` has only a checked scalar-CFG context. No parser or checker semantics changed.

## Task Commits

1. `518b3886` — `test(34-07): add independent U64-add surface witness`.
2. `175b3dca` — `test(34-07): record U64-add tracer verification`.
3. `8b7d15dd` — `test(34-07): capture red for branch census coverage` (valid RED before the matrix implementation).
4. `425200db` — `feat(34-07): complete independent surface census`.
5. `7594ae7b` — `test(34-07): cover surface manifest drift controls`.
6. `3fbc55a2` — `test(34-07): record final control verification`.

The measured task count is six commits from `2706b089bf342677c7faebd43e9fa9d08375d2b5` through `3fbc55a29eea0371ba6525f82861ce6cc778d612`. The tasks and evidence touch 40 unique paths. No task commit deleted tracked files.

## Verification Evidence

All runs used the standard-library `/tmp/schway-go-junit.py` reporter, Go JSON events, `/tmp/ai-lang-verification-gocache`, and focused selectors. The committed evidence records retain raw events, actual exit codes and durations, JUnit reports, reporter source hashes, freshness, matched test names and semantic assessments.

- **Task 1:** `TestPhase34SurfaceManifestWitnesses` passed, exit 0, 1.347 s; the post-commit tracer rerun passed, exit 0, 0.303 s. The accepted U64 addition was checked as a complete source program, and a copied manifest-only operator produced `invented_admission`.
- **Task 2 RED:** `TestPhase34SurfaceManifestCensus` failed on unclassified source branches while its companion U64 witness passed, exit 1, 1.398 s. `gsd_run check tdd-red-evidence` returned `RED_EVIDENCE_OK` before the implementation commit.
- **Task 2 GREEN:** `TestPhase34SurfaceManifestWitnesses` and `TestPhase34SurfaceManifestCensus` both passed, exit 0, 1.719 s. Their complete post-commit rerun with all three selectors passed in Task 3.
- **Task 3 controls:** `TestPhase34SurfaceManifestControls` passed, exit 0, 1.106 s. The post-commit run of `TestPhase34SurfaceManifestControls`, `TestPhase34SurfaceManifestWitnesses` and `TestPhase34SurfaceManifestCensus` passed, exit 0, 0.756 s.
- **Summary metadata:** `TestSummaryCoverageGroundedness` passed, exit 0, 0.973 s, validating this summary's coverage entries against the actual selectors and passing evidence.
- The first five Task 2 matrix attempts and the initial Task 3 control probe are preserved with inspected intermediate-failure assessments; none are represented as successful evidence. The Task 3 probe counted a duplicate `constraint_changed` finding from the witness comparator and the literal call-contract comparator as two distinct failures. The assertion was corrected to compare unique reason/facet pairs; no behavior change was warranted.

No package-wide suite, broad build, native matrix, hosted dispatch or receipt binding ran.

## TDD Gate Compliance

- Task 1's fixture was corrected after its first complete-source attempt reached an unsupported linear result tail. The accepted scalar-local witness and copied-manifest control then passed. No failure was manufactured to force a production change.
- Task 2's census test produced a valid focused RED accepted by the `tdd-red-evidence` check, and that evidence commit preceded the matrix/census completion commit.
- Task 3 added only controls for comparator behavior already implemented by Tasks 1–2. Its first probe exposed duplicate target findings caused by two comparator layers; inspection showed no missing target behavior, so the probe was classified as an intermediate test-assertion failure rather than valid RED. The final target-specific controls passed without a needless production change.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed unsupported manifest admission contexts.**
- **Found during:** Task 2.
- **Issue:** The manifest advertised `Box<T>` and `Pair<A, B>` type contexts based only on ability derivation, and advertised checked linear `if_else` branches. Complete `CheckCommandFile` witnesses rejected the Box/Pair executable shape as `check.unexecutable_shape`; the attempted linear-if source was rejected as `syntax.expected_scalar_operand` because branch parsing is scalar-CFG-only.
- **Fix:** Removed the Box/Pair entries from the admitted Types list, narrowed `if_else` to scalar CFG control flow, and classified the ability-only constructors as internal with literal rejection probes.
- **Files modified:** `internal/compiler/surfacemanifest/manifest.go`, `testdata/phase34/surface-witnesses.json`, `testdata/phase34/surface-census.json`.
- **Verification:** The focused witness and census selectors passed in `34-07-task2-attempt6.json` and again in the post-commit three-selector run.
- **Committed in:** `425200db`.

**Total deviations:** 1 auto-fixed producer bug (Rule 1). The adjustment removes claims that lacked complete-source admission evidence and does not alter language semantics. The corresponding `.planning/WINDOWS.md` entry 8 is marked fixed.

## Issues Encountered

The first census drafts needed descriptor corrections after AST inspection, and the first witness matrix exposed incorrect source expectations for Bool contexts, builtins, unsupported operators, and ability-only shapes. Those literal expectations were corrected from observed `CheckCommandFile` outcomes; each raw run remains in the phase evidence. No unrelated pre-existing issue was changed.

## Known Stubs

None found in files created or modified by this plan. The literal zero values in unused optional witness fields are not UI data or producer claims.

## User Setup Required

None. No dependency or external service configuration was required.

## Next Phase Readiness

Plan 07 is complete and leaves the Phase 34 final plan ready. Its manifest meta-test now provides the independent witness and drift controls needed before later phases register new operators or call contexts. DX-23 was already marked complete in `REQUIREMENTS.md`; the task close-out revalidates that requirement.

## Self-Check: PASSED

The summary path and all three owned deliverables exist. Each listed task commit is present in the current history. The summary coverage entries point to actual named selectors and passing focused records.

---
*Phase: 34-agent-diagnostics-floor-and-surface-manifest*
*Completed: 2026-10-13*

## Parent follow-up: affected producer contract

The post-plan focused producer check failed because `TestPhase34SurfaceDocument` still expected Box/Pair admission entries after the independent complete-source witnesses had refused them. Commit `396a8ec0` updates that literal contract, pins the parse/ability restriction and scalar-only if/else context, and explains the boundary in docs/agent-surface.md. This is one additional parent follow-up commit beyond the six task commits measured above; their original actuals and ledger remain historical measurements.

Fresh command: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/surfacemanifest ./internal/compiler/session -run '^(TestPhase34SurfaceDocument|TestPhase34SurfaceLimits|TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus|TestPhase34SurfaceManifestControls)$' -count=1 -v`. Exit 0; all five named selectors passed (package times 0.006s and 0.188s). No broad suite ran. Expected admissions remain independently literal; no producer-derived oracle or admitted semantics were added.

## Parent follow-up: reviewed limit consumers

Commit `18844d62` resolves WR-01, WR-02, and WR-03 from the configured source review. Parser routes include interface core; scalar analysis routes include every checked consumer in the finite seven-file reference census. Checker-work routes include both interfaces, verify, and the conditional debug-map/import and query/lane-corpus paths, with those conditions named in the manifest. Linear bindings are described as the aggregate outer function-body lists. No parser/checker admission or enforcement changed.

The new consumer test first failed with the specific missing route classifications against the original producer. After correction, all seven selected session checks passed (1.010s), both producer contract checks passed (0.006s), and the direct aggregate AST guard check and inherited scalar exhaustion control passed (0.016s and 0.008s respectively). The interface probes accept exactly 1024 functions, refuse 1025 and excessive checker work before artifact publication. Route mutations detect omissions and false inclusion, then restore. The reference census is finite and branch-insensitive; executable inherited controls distinguish debug-map with and without imports. The AST binding guard is tested directly because earlier source/token bounds can prevent reaching it through source text. Original plan actuals remain historical; this is additional parent close-out work.
