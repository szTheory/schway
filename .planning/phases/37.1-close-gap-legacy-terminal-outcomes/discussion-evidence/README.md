# Phase 37.1 bounded discussion probes — 2026-10-16

These are decision evidence, not permanent acceptance regressions. Repository source/tests were unchanged. The original audit and its probes remain untouched.

| Probe | What actually ran | Meaning of PASS | Limit |
| --- | --- | --- | --- |
| `return-probe.go.txt` | Two checked/corevalidated Phase 24 modeled failure cases; one modified-core candidate | UseError and AcquireError traces reached `foreign.failed` then `OpFail`; the candidate `OpReturn` rewrite was refused by core validation | Model-only, `actual_host_io=false`; no universal raw-core impossibility theorem and no repaired peer acceptance |
| `reason-probe.go.txt` | Three checked no-call sources, native/application emission, actual retained native application build/run | Reproduced stderr corruption for bytes 01 41, 00 41 and 00 5A while preserving exit 65/empty stdout | This deliberately diagnoses defects; PASS does not mean exact-reason semantics passed; native conformance optimization lanes were not run |

The `*.command` files retain exact original bounded commands; `*.log` files retain output. `bindings.json` records the source revision, selected direct source hashes, probe/log hashes and locally observed Go/Clang versions. This is a selected binding inventory, not a full compiler/SDK/runtime closure. No full local package/project suite, hosted or Linux receipt follows.

To reproduce from repository root, create temporary overlays that add these inert text files as Go tests without changing repository tests:

```sh
python3 - <<'PY'
import json, pathlib
root = pathlib.Path.cwd()
base = root / '.planning/phases/37.1-close-gap-legacy-terminal-outcomes/discussion-evidence'
for name, package in [('return', 'session'), ('reason', 'cgen')]:
    overlay = {'Replace': {
        str(root / ('internal/compiler/' + package + '/zz_schway371_' + name + '_probe_test.go')):
        str(base / (name + '-probe.go.txt'))
    }}
    pathlib.Path('/tmp/schway-371-' + name + '-overlay.json').write_text(json.dumps(overlay))
PY
GOCACHE=/tmp/ai-lang-verification-gocache go test -overlay=/tmp/schway-371-return-overlay.json ./internal/compiler/session -run '^TestSchway371(ReturnFailureReachabilityProbe|RawCoreCanReachModeledFailureOpReturn)$' -count=1 -v -timeout=120s
GOCACHE=/tmp/ai-lang-verification-gocache go test -overlay=/tmp/schway-371-reason-overlay.json ./internal/compiler/cgen -run '^TestSchway371ReasonByteProbe$' -count=1 -v -timeout=120s
```

The raw-core candidate's test name predates its result. Its asserted passing result is refusal (`core.owner_transfer_cleanup`), not valid modeled-return admission. The reason probe similarly asserts that the pre-repair mismatch is reached. Execution must write permanent regressions with the correct positive expected behavior and show actual-source RED/clean GREEN evidence; these diagnostics must not be counted as those regressions.

## UTF-8 decision probe — 2026-10-16

`utf8-probe.go.txt` runs full source admission/native emission for a no-call
reason containing FF 41, observes existing JSON projection to EF BF BD 41, and
compares invalid UTF-8 outside a quote (existing syntax.invalid_utf8). The
original command/output and selected source hashes are in utf8-probe.command,
utf8-probe.log and utf8-bindings.json. It is a reached counterexample, not a
production fix or a general JSON encoding experiment. To reproduce, add an
overlay mapping internal/compiler/cgen/zz_schway371_utf8_probe_test.go to this
text file and run the exact named root in the command file. D-21 in CONTEXT.md
owns the resulting narrow source/raw-core admission decision.
