---
type: quick
quick_id: 261013-hht
status: planned
autonomous: true
files_modified:
  - .planning/config.json
  - .planning/STATE.md
---

# Switch OpenGSD to Codex and inspect model drift

User scope: switch Schway and the shared OpenGSD default runtime from Claude to
Codex; use `gpt-6.1-sol` for Schway's top Codex tier; report other projects whose
Codex settings differ. Execute inline under the quick workflow.

## Tasks

1. Change `.planning/config.json` through `gsd-tools query config-set`:
   `runtime=codex` and `model_profile_overrides.codex.opus=gpt-6.1-sol`.
   Change only `runtime` in the shared GSD defaults, with a temporary backup.
   Preserve adaptive routing, lower-tier models, effort, and workflow settings.
2. Inspect GSD configs, workstream configs, and local Codex model/agent settings
   under the user's projects directory. Report both active Codex mismatches and
   dormant Codex overrides in other runtimes. Other repositories are read-only.
3. Verify serialized settings and live GSD model/effort resolution, retain a
   scoped audit report and summary, update the quick-task table, and commit the
   Schway configuration and planning artifacts.

## Verification

- Compare before/after JSON: Schway changes exactly the two requested keys;
  shared defaults change exactly `runtime`.
- Resolve planning, execution, checking, verification, research, and review
  agents through the installed OpenGSD CLI and inspect returned models/effort.
- Parse each discovered config and identify explicit role/tier overrides;
  distinguish source settings from live resolution.
- Inspect the scoped Git diff and whitespace errors. No language code changes
  or project test-suite execution are required.

<threat_model>
Configuration-only scope. Preserve unrelated keys and secrets; inspect model
fields only in other projects; do not change other repository configurations.
No external messages or model inference requests are required.
</threat_model>
