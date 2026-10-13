---
type: quick
quick_id: 261013-hht
status: complete
completed: 2026-10-13
---

# OpenGSD runtime and model routing updated

Schway and the shared OpenGSD default now select Codex. Schway's highest Codex
tier is `gpt-6.1-sol`; standard and light tiers remain `gpt-6-luna`. Adaptive
routing and the existing heavy/high, standard/medium, and light/low effort
settings are preserved. The global Codex session config remains 6.1 Sol/xhigh.

Fresh verification compared both JSON configurations with their pre-change
copies and confirmed exactly the requested keys changed. Nine live OpenGSD
resolver queries confirmed 6.1 Sol/high for planning and debugging, Luna/medium
for execution, research, verification, and code review, and Luna/low for plan
checking, integration checking, and codebase mapping.

A read-only source audit inspected 122 configuration files across 32 project
directories. All parsed successfully. 29 of the 32 primary GSD configurations
have the requested top Codex model. Three primary settings differ; 14 local
Codex agent files and 36 nested checkout configurations retain older settings.
Source settings in other repositories were inspected without querying their
live runtime or changing their files. The detailed cross-project report is
kept outside this repository in a temporary local artifact for the user.

Schway commit `25dc503f` switched the runtime to Claude on October 10 while
preserving the old Sol Codex tier. A later Codex-only update may have skipped
this checkout; that explanation is an inference.

No language source changed. No local language test suite or model inference
request was run. The milestone remains at Phase 33 context gathered.
