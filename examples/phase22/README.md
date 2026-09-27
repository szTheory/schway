# Phase 22 native application example

`identity.lang` is a checked `U64 -> U64` program. The expected values in
`identity.cases.json` are authored literals; they are not copied from an
interpreter or native run.

## Build and run

```sh
lang build examples/phase22/identity.lang --output ./identity
lang app run ./identity -- 7
lang app run ./identity -- 42
```

For declared local C build inputs, use
`lang build SOURCE --manifest MANIFEST --output ARTIFACT`; `--manifest` and
`--output` may appear in either order. The checked-in example can be built with:

```sh
lang build examples/phase22/identity.lang \
  --manifest examples/phase22/identity.bindings.json --output ./identity
```

The closed `lang.local-c/1` manifest requires local `.c` sources, every local
header (including transitive includes), ordered manifest-relative include
directories, per-symbol `{name, header, function_type}` declarations using C
function typedefs, and exactly `runtime_dependencies: ["platform-c-runtime"]`.
Paths cannot escape the manifest directory. The manifest is bounded to 64 KiB;
each source/header is at most 4 MiB, and all declared source/header bytes are
at most 16 MiB. Build uses fixed C17 flags and no user-supplied compiler or link
flags. The example's `support.c` is linked but uncalled: this authority does
not admit Lang foreign calls or pointers. Declared C, the installed compiler,
and host SDK are trusted inputs; this is not a sandbox for hostile C.

The retained build receipt identifies known source, C, ABI, toolchain and
artifact inputs. Host SDK/linker/runtime closure remains `incomplete` on macOS
and Linux, and builds are not cacheable. Run checks the receipt and executable
digest before launching. See [`BINDINGS.md`](./BINDINGS.md) for the full path,
header, ABI-probe, and identity rules.

`lang app run ARTIFACT [--report REPORT] [--evidence=events] -- INPUT` passes
one opaque argument to one retained application process. The `--` separator
ends Lang option parsing, so values that resemble flags remain input bytes.
Canonical decimal `U64` input has at most 20 digits and is limited by the
4,096-byte transport bound. The child has a 30-second default timeout. Its stdout and
stderr go directly to the caller's streams without an added output cap; exit,
signal, timeout, and launch outcomes remain distinct.
The app's caller-owned stdout and stderr streams are passed through without a
byte cap. Conformance runs used by `app verify` cap each execution document at
16 MiB and compiler/process diagnostic streams at 64 KiB.

`--report` by itself writes a `lang.app-evidence/1` record with
`capture_status: "disabled"`. Adding `--evidence=events` asks the same child
process to write its checked compiler events through a private file channel.
The event/report bound is 64 KiB. A report includes build, artifact, source,
input, and child-process identities. Capture status is `complete`, `incomplete`,
or `capacity_exhausted` when requested. Missing or malformed capture and report
write errors fail the tool operation. Every app evidence report has
`verified: false`: a complete trace records one execution and is not a
differential verdict.

## Explicit replay verification

```sh
lang app verify examples/phase22/identity.lang \
  --cases examples/phase22/identity.cases.json \
  --report ./identity-verification.json
```

The closed `lang.replay-cases/1` file holds one to sixteen isolated cases. Each
`source` case supplies a canonical decimal input, an independently authored
expected `U64` result, and an explicit empty `foreign_outcomes` array. The
verifier checks the source once, then compares each case's ordered execution
trace across the interpreter and generated-C conformance programs at `-O0` and
`-O3`, as well as comparing all three terminal values with the fixture answer.
It does not launch the retained application. Case inputs are bounded to 4,096
bytes; the cases file and report are each bounded to 64 KiB; execution documents
are bounded to 16 MiB. The report records source, generated-C, build, case-set,
and per-input identities. Host toolchain dependency closure remains
`incomplete`, and reports are not cacheable.

Ordinary `source` cases cannot consume scripted foreign outcomes. A non-empty
script is reported as unsupported, and source-level foreign contracts are
refused. `app verify` accepts no local-C manifest option, so replay never
compiles, links, or executes manifest C.

The closed `verifier_model_only` case kind is for verifier unit controls. Its
single supported synthetic operation, `fixture.value`, consumes a declared
`U64` outcome exactly once and compares it with an independent expected value.
It never compiles Lang foreign calls or executes C. The report names this kind
and sets `actual_host_io` and `physical_cleanup` to `false`; a scripted result
does not establish host IO or physical resource cleanup.

Application verification writes `lang.app-verification/1`. Only a report whose
every case passed the engine comparisons and independent expected answer has
`verified: true`. This is controlled local evidence, not a claim of complete
toolchain provenance or behavior on another host.
