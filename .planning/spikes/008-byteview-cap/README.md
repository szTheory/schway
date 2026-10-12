---
spike: 008
idea: ownership-kernel
name: byteview-cap
type: standard
validates: "Given the Adler-32 fixture with a 65,536-byte host-owned byte view run through the real D-05 fixture path, every engine lane stays within the D-13 budget (native run <= 3 s, interpreter <= 12 s) when scaled linearly to the recorded cap, so the 65,536-byte cap can be kept"
verdict: VALIDATED
related: [005-native-ffi-provenance-cleanup]
tags: [byte-views, cap, native, interpreter, timing, phase-gate, d-12, d-13]
---

# Spike 008: Byte View Cap

## What This Validates

Phase 29 pins a 65,536-byte cap on host-delivered byte views (research
Decision 5). Phase 31's oversize matrix pins cap and cap+1, so the constant is
costly to reverse. This spike measures the 65,536 x 0xFF Adler-32 workload
(expected checksum 2006388466, fixture case 9) through the real fixture owner
(`examples/phase29/adler32_fixture.schway` plus its bindings manifest), with
4,096 (case 7, 2170679522) and 16,384 (case 8, 2967061426) for linear scaling,
on the interpreter and on native builds at -O0, -O3 and -O3 -flto. The D-13
budget is 10% of the real deadlines: native run <= 3 s (deadline 30 s),
interpreter <= 12 s (deadline 2 min). The cap is kept only if local evidence
and one hosted macOS and one hosted Linux strict run (plain and -race) pass.

## Research

Pure measurement against the production compiler, interpreter and native
runner; no external libraries. Decisions: D-12 (keep/lower rule), D-13 (budget,
routed not frozen), D-14 (Stage B is the record, Stage A is only a gate).

## How to Run

```sh
# default mode (what CI runs on every push): case 9, 3 runs, checksums + half-deadline ceiling
go test ./internal/compiler/native -run '^TestSpike008ByteViewCap$' -count=1 -v
# record mode: full measurement, logs OVER BUDGET, never fails on timing
SCHWAY_SPIKE008=1 go test ./internal/compiler/native -run '^TestSpike008ByteViewCap$' -count=1 -v
SCHWAY_SPIKE008=1 go test -race ./internal/compiler/native -run '^TestSpike008ByteViewCap$' -count=1 -v
# strict mode: an overrun fails (hosted: workflow_dispatch input spike008_strict=true)
SCHWAY_SPIKE008=strict go test ./internal/compiler/native -run '^TestSpike008ByteViewCap$' -count=1 -v
```

## What to Expect

Every run asserts the pinned checksum. The `armed-fault` subtest builds a
wrong-modulus (65520) mutant and requires the measurement routine to report a
checksum mismatch, so a fast wrong answer cannot be recorded. Default mode never
evaluates the 10% budget; only SCHWAY_SPIKE008=strict fails on timing, and the
only hosted route to it is the dispatch input.

## Investigation Trail

1. Stage A (29-00-GATE-VERDICT.md) used a proxy reading as a gate only; it is
   not the cap record (D-14).
2. Built the harness over the real fixture path (`session.BuildApplicationFile`
   with the fixture manifest, `native.RunApplication`,
   `interp.RunWithFixtureBytes`), -O3 and -O3 -flto lanes via the D-21 test seam.
3. 7 runs per lane per case, first discarded, median and max of the remaining 6.
4. Ran record mode locally plain and under -race.

## Results

Local record (darwin/arm64, go1.24.0, Apple clang 21.0.0 clang-2100.1.1.101,
target arm64-apple-darwin25.6.0). Wall time appears only here and in test logs.

| Lane | Case (bytes) | Median | Max |
|------|--------------|--------|-----|
| native -O0 | 7 (4096) | 2.08 ms | 2.64 ms |
| native -O0 | 8 (16384) | 2.04 ms | 2.34 ms |
| native -O0 | 9 (65536) | 2.41 ms | 2.66 ms |
| native -O3 | 7 (4096) | 2.08 ms | 2.08 ms |
| native -O3 | 8 (16384) | 1.97 ms | 2.27 ms |
| native -O3 | 9 (65536) | 2.25 ms | 2.66 ms |
| native -O3 -flto | 7 (4096) | 1.98 ms | 2.06 ms |
| native -O3 -flto | 8 (16384) | 1.85 ms | 1.98 ms |
| native -O3 -flto | 9 (65536) | 2.03 ms | 2.12 ms |
| interpreter | 7 (4096) | 5.51 ms | 5.53 ms |
| interpreter | 8 (16384) | 20.44 ms | 21.53 ms |
| interpreter | 9 (65536) | 81.57 ms | 82.12 ms |
| interpreter-race | 7 (4096) | 21.78 ms | 22.04 ms |
| interpreter-race | 8 (16384) | 66.60 ms | 68.78 ms |
| interpreter-race | 9 (65536) | 250.47 ms | 264.72 ms |

Compile time (median / max, 7 builds each): -O0 285.8 / 294.0 ms, 52,200 byte
artifact; -O3 331.1 / 349.9 ms, 51,400 bytes; -O3 -flto 328.5 / 357.5 ms,
34,696 bytes. Content digests are in the test log (`sha256=`).

Scaling ratios (median 16,384 / 4,096 and 65,536 / 4,096; linear is 4.00 and
16.00): interpreter 3.71 and 14.80; interpreter-race 3.06 and 11.50; native
lanes 0.94-1.02 and 1.02-1.16 (process start dominates, so native time is flat
in this range).

### Budget comparison

Each lane's 65,536 max against the D-13 budget, scaled to the recorded cap
65,536: native -O0 2.66 ms, -O3 2.66 ms, -O3 -flto 2.12 ms (budget 3 s);
interpreter 82.1 ms and interpreter-race 264.7 ms (budget 12 s). No lane is
`OVER BUDGET`; the largest fitting cap is 65,536 in every lane locally. The
keep/lower decision is recorded here and in plan 29-12 Task 3, not in the test.

### Stage-A cross-reference

Stage-A numbers in 29-00-GATE-VERDICT.md are a proxy, gate-only, not the record.

### Hosted runs

Run 38186068052 (workflow_dispatch, `spike008_strict=true`) at head SHA
906d6481ef017157eecc8f8f30528ebdd0428876. Overall conclusion: success.

| Host | go test ./... | go test -race ./... | spike 008 strict budget |
|------|---------------|---------------------|-------------------------|
| ubuntu-latest (job 114608534221) | success | success | success |
| macos-latest (job 114608534241) | success | success | success |

(An earlier dispatch, run 38185331902 at 1327cf1, failed before the strict step
on five repo convention tests unrelated to the cap; fixed in 0364898, 532f0ba,
614841f and 906d648. Only run 38186068052 is the record.)

Strict lane tables (7 runs per lane per case, first dropped, median / max of 6).
Toolchain go1.24.13. ubuntu: Ubuntu clang 18.1.3, linux/amd64. macos: Apple
clang 21.0.0 (clang-2100.1.1.101), darwin/arm64.

ubuntu-latest, plain (step part 1):

| Lane | 4096 med/max | 16384 med/max | 65536 med/max |
|------|--------------|---------------|---------------|
| native -O0 | 0.857 / 0.933 ms | 0.995 / 1.442 ms | 1.466 / 1.531 ms |
| native -O3 | 0.890 / 1.070 ms | 0.818 / 0.886 ms | 0.965 / 1.039 ms |
| native -O3 -flto | 0.875 / 1.041 ms | 0.825 / 1.065 ms | 0.975 / 0.994 ms |
| interpreter | 9.66 / 10.01 ms | 34.75 / 36.56 ms | 136.61 / 138.32 ms |

ubuntu-latest, -race (step part 2):

| Lane | 4096 med/max | 16384 med/max | 65536 med/max |
|------|--------------|---------------|---------------|
| native -O0 | 2.080 / 2.168 ms | 2.209 / 2.764 ms | 2.682 / 2.736 ms |
| native -O3 | 2.081 / 2.198 ms | 2.056 / 2.067 ms | 2.213 / 2.237 ms |
| native -O3 -flto | 2.147 / 2.597 ms | 2.100 / 2.146 ms | 2.300 / 2.612 ms |
| interpreter-race | 32.95 / 35.71 ms | 101.23 / 102.45 ms | 371.94 / 376.48 ms |

macos-latest, plain:

| Lane | 4096 med/max | 16384 med/max | 65536 med/max |
|------|--------------|---------------|---------------|
| native -O0 | 3.525 / 6.547 ms | 7.792 / 8.041 ms | 4.356 / 4.628 ms |
| native -O3 | 3.501 / 4.031 ms | 6.433 / 7.151 ms | 4.710 / 6.515 ms |
| native -O3 -flto | 6.960 / 7.239 ms | 3.578 / 5.925 ms | 7.562 / 7.996 ms |
| interpreter | 8.95 / 9.42 ms | 33.37 / 38.29 ms | 149.83 / 157.20 ms |

macos-latest, -race:

| Lane | 4096 med/max | 16384 med/max | 65536 med/max |
|------|--------------|---------------|---------------|
| native -O0 | 4.244 / 5.118 ms | 4.225 / 4.265 ms | 4.655 / 4.867 ms |
| native -O3 | 4.112 / 4.378 ms | 3.965 / 4.182 ms | 4.181 / 4.927 ms |
| native -O3 -flto | 4.057 / 4.317 ms | 3.974 / 4.157 ms | 4.315 / 5.034 ms |
| interpreter-race | 38.91 / 40.91 ms | 180.30 / 238.71 ms | 351.92 / 356.61 ms |

Every lane logged `within budget ... (largest fitting cap 65536)`; no
`OVER BUDGET` line on either host, plain or -race. Worst hosted native max
8.0 ms against 3 s; worst hosted interpreter max 376.5 ms against 12 s. The
`armed-fault` subtest passed on both hosts.

### Decision (D-12 / D-13)

Local evidence and both hosted strict runs (one Linux, one macOS, plain and
-race) pass inside the D-13 budget with orders of magnitude of headroom.
**cap kept at 65,536.** `spike008RecordedCap` stays 65536; no D-13 reopening and
no Phase 31 literal change. Verdict: VALIDATED.
