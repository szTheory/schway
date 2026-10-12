# Phase 30 local modules: the Adler-32 library as its own file

The Phase 29 Adler-32 program with the library moved into its own module file
and imported by the entry (D-17). The checksum code exists in exactly one place.

## Module layout

- `adler32.schway` is the library. Its header is `module adler32`, its export
  list is `fn adler32`, and it is pure scalar: no data type, no foreign block,
  no `main`.
- `adler32_cli.schway` is the entry. Its header is
  `module examples.phase30.adler32_cli`. It imports `fn adler32` from `adler32`,
  keeps the Phase 29 `foreign C` owner and `data FixtureError`, and exports
  `fn main`.

A module header must match the file's path under the entry's directory: the
import `adler32` loads `adler32.schway` beside the entry, and a mismatch is a
refusal with a `module.*` code, not a silent load. The entry's own header is
free, as it is for every single-file program.

## Imports

The entry writes an import block that names the module and then lists the
exported functions it uses, one per line. The imported function is then called
by its bare name, `adler32(view)`. Only names in the library's export list are
visible; a private library name is refused.

## Build and run

Run these commands from the repository root with Go 1.24 and Clang installed.
The build reuses the Phase 29 binding manifest in place, because that manifest
names its adapter files relative to itself and carries no module name.

```sh
go build -o ./schway ./cmd/schway
./schway build examples/phase30/adler32_cli.schway \
  --manifest examples/phase29/fixture_bytes.bindings.json --output ./adler32-module
./schway app run ./adler32-module -- 3
./schway app run ./adler32-module -- 10
```

Case 3 prints `300286872` with exit 0 and empty stderr. Case 10 is outside the
table: nothing is printed and the exit code is 65. The case table and the
expected checksums are exactly those of `examples/phase29/README.md`; they come
from Python `zlib.adler32` and Go `hash/adler32`, never from Schway.

`schway run` keeps the exit class the Phase 29 fixture gives under the same
command, because the program owns a foreign resource. The interpreter route for
the same ten answers is the session interpreter API with the Phase 29 fixture
bytes.

The tests that pin this are `TestPhase30AdlerModuleAppCLI` (public build and app
run, with a `mutated-library` control), `TestPhase30AdlerModuleInterpreter` and
`TestPhase30LibraryIsSingleCopy`.

## Resolution rules

Resolution is closed. There is no search path and no environment lookup: an
import resolves only to a file under the entry's directory. A path segment that
is `..`, an absolute path, and a symlink are all refused.

## Non-claims

- This is whole-program composition: the entry and its imports are merged into
  one program and checked and emitted together. It is not separate compilation.
- There are no per-module interface artifacts.
- Module hashing in build identity and evidence is Phase 31 (MOD-06); the build
  receipt digests the entry file only.
