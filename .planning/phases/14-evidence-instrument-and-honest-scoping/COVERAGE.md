# Phase 14 — API Coverage Declaration

**Detector status:** the `api-coverage` detector verb is not available in the
installed `gsd-tools.cjs` (`query api-coverage scan 14 --json` → `Unknown command`),
so the result is **skipped**, not a confirmed negative. Scope was therefore
re-read directly (14-CONTEXT.md `<domain>`, 14-RESEARCH.md §"Standard Stack",
§"Package Legitimacy Audit") to decide the branch.

No external API integration: Phase 14 is Go stdlib `testing` guards over this
repository's own `.planning/**` markdown, its own diagnostic identity, and its
own `lang-repair` subprocess protocol — `go.mod` is untouched, zero external
packages are installed or called, and the project carries a hard
zero-external-production-dependency verdict (`.planning/STANDING-VERDICTS.md`).
