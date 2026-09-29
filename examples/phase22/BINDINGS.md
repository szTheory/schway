# Local C build inputs

From the repository root, build the CLI and retained identity application:

```sh
go build -o /tmp/schway ./cmd/schway
/tmp/schway build examples/phase22/identity.schway \
  --manifest examples/phase22/identity.bindings.json --output /tmp/identity
/tmp/schway app run /tmp/identity -- 7
/tmp/schway app run /tmp/identity -- 42
```

The two runs print `7` and `42`, each followed by a newline. Build launches
Clang only. Run validates the adjacent `/tmp/identity.schway-build.json` receipt
and starts the selected application once. Both retained files may be moved
together; C intermediates and the original checkout are unnecessary at run time.
`--json` on build prints the receipt. `--manifest` and `--output` may occur in
either order after SOURCE. Omitting the manifest builds the pure Lang app.

`support.c` is trusted local C, linked but **uncalled** by this Lang body.
The manifest does not admit Lang foreign calls or pointer operations. Existing
emitter refusals remain in force. Successful compilation/linking proves that
the declared header types agree and the symbol resolves; it does not prove
arbitrary C behavior, physical cleanup, ownership, or implementation ABI safety.

## Manifest contract: `schway.local-c/1`

Every field shown in `identity.bindings.json` is required. Unknown fields,
case aliases, duplicate JSON keys (including nested objects), duplicate paths
or symbols, and trailing JSON are errors. There are no arbitrary compiler flags,
library search flags, package names, or shell command fields.

- `sources`: nonempty list of local `.c` translation units.
- `headers`: nonempty list of all local headers, including transitive includes.
- `include_dirs`: ordered list of declared local include search directories;
  an empty list is allowed and `.` denotes the manifest directory.
- `symbols`: nonempty list of `{name, header, function_type}`. The header must
  be declared. Name and function_type are C identifiers. Function_type must name
  a **function typedef**, such as `typedef uint64_t identity_fn(uint64_t)`;
  an object type or pointer typedef is rejected. Macro symbols are rejected.
- `runtime_dependencies`: exactly `["platform-c-runtime"]` in this version.

Paths are manifest-relative, canonical slash-separated names. Their portable
alphabet is letters, digits, underscore, dot, slash, space, and hyphen. Absolute
paths, drive paths, backslashes, `.` components, `..` components and symlink
escapes are rejected. In-root symlinks are resolved and snapshotted; two declared
names may not resolve to the same input. The manifest is bounded to 64 KiB,
each source/header to 4 MiB, all source/header bytes together to 16 MiB, and
each declaration list to 64 entries (runtime has the single value above).

The builder compiles a private snapshot using C17, `-Wall -Wextra -Werror
-pedantic -O0`, fixed argv, and the declared include directories. A generated
probe checks the function typedef and retains a volatile function-pointer
reference for each symbol so link resolution cannot be skipped. Clang's full
`-M` output is checked before compiling each unit. Every local included header
must belong to the snapshot and be declared, even if marked as a system header.
Installed compiler system include roots are discovered separately. Ambient
`CPATH`, language include-path variables, `LIBRARY_PATH`, driver override and
dependency-output variables cannot add build authority. The installed compiler,
host SDK and audited C inputs are trusted; this is not a sandbox for hostile C
or a hostile compiler. Input trees must not be concurrently replaced during
resolution. After resolution the exact captured bytes are compiled and hashed.

## Build and evidence identity

The adjacent `schway.app-build/1` receipt now contains `input_id`, `build_id`,
`identity`, the canonical `bindings`, the artifact digest and closure status.
SHA-256 hashes a length-framed schema domain and canonical JSON fields:

- Lang source bytes and emitted C bytes;
- canonical manifest and normalized source/header names and byte digests;
- symbol/typedef declarations, actual local includes and generated ABI probe;
- compiler executable digest and version, target triple and host OS/architecture;
- fixed flags, actual compile/link argv with relative staged filenames;
- runtime dependency declarations;
- final executable bytes, for `build_id` only.

`input_id` identifies the known build inputs before artifact bytes are added.
JSON whitespace and key order do not matter; source/header/symbol lists are
sorted, but include search order remains significant. Output and compiler paths
are recorded outside these identity fields. Staging and checkout paths never
enter the identity. Relocation of the same known inputs preserves `input_id`;
`build_id` also matches when the compiler produces the same binary bytes. Any
different binary has a different build ID, including differences from unknown
host dependencies or nondeterministic C constructs such as time macros.

On **both macOS and Linux**, these builds currently report
`dependency_closure: "incomplete"` and `cacheable: false`. The known inventory
is useful for invalidation and artifact integrity, but the installed driver's
SDK/system headers, linker, startup objects, implicit libraries, environment,
dynamic-loader configuration and runtime library bytes are not fully hashed.
On macOS this includes the SDK and libSystem resolution; on Linux it includes
libc, startup objects and the loader. A future **complete** closure would have
to discover, confine or record, and hash all those inputs and their resolution
order. This version never emits a complete closure or reuses a cached app.

Run refuses a swapped executable or a receipt whose recorded identity fields
no longer match their hashes. A receipt is an integrity record, not a signature
against a malicious author who can rewrite both files. Legacy pure-source v1
receipts remain runnable, but lack the new build identity; they cannot supply
the new provenance claim. Execution evidence must bind to this build ID and
validate the artifact digest separately. Neither a receipt nor a successful
link is an execution-verification result.

## Executable controls

```sh
GOCACHE=/tmp/ai-schway-verification-gocache go test ./internal/compiler/native \
  -run '^TestPhase22(Bindings|BuildIdentity|Relocated)' -count=1 -v
```

These tests exercise public CLI builds from relocated trees, retained runs,
input/toolchain identity mutations, negative declarations, receipt/binary swaps,
and a marker plus child recorder proving zero application launches at build
and exactly one at run. Passing evidence is scoped to the executing host;
macOS and Linux are separate host lanes.
