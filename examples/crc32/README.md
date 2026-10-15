# CRC-32 file examples

Run these commands from the repository root. The guide needs Go 1.24 and Python 3
to build the CLI and create the binary sample files. Clang is also required for
native execution. Python and its `zlib` module are not compiler or runtime
dependencies; the checked-in numeric answers were generated independently with
Python 3.14.4 and zlib 1.2.12 and are recorded in `crc32.expected.json`.

The source library implements CRC-32/ISO-HDLC (the reflected IEEE polynomial
`0xEDB88320`, all-ones initial and final XOR). It treats each input as bytes and
prints an unsigned decimal value followed by one newline. This is not POSIX
`cksum` / CRC-32/CKSUM, CRC-32C, a cryptographic hash, or authentication.
`crc_low` imports the same library and prints the value modulo 65,536; for
`123456789`, the full result is `3421780262` and the low result is `14630`.

Build the tracked source CLI and create the exact binary inputs used by the
table below. All generated files and native artifacts stay under the ignored
`.schway/crc32/` directory.

```sh doc-test=setup
go build -o ./schway ./cmd/schway
mkdir -p .schway/crc32/inputs .schway/crc32
python3 -c 'from pathlib import Path; p=Path(".schway/crc32/inputs"); data={"empty.bin":b"","zero.bin":b"\x00","check.bin":b"123456789","highbit-80ff.bin":b"\x80\xff","all-bytes.bin":bytes(range(256)),"cap-minus-one.bin":b"\xff"*65535,"cap.bin":b"\xff"*65536,"cap-plus-one.bin":b"\xff"*65537,"--json":b"123456789","space input.bin":b"123456789","x ;$(literal).bin":b"123456789"}; [(p/name).write_bytes(value) for name,value in data.items()]; (p/"directory").mkdir(exist_ok=True)'
```

The CRC library is ordinary Schway source with one private `crc_byte` helper;
both entries import that same module. The fixed interpreter FileBytes adapter
reads at most 65,536 bytes. The sample native binding also selects a 65,536-byte
limit, but FileBytes limits are binding-specific rather than a language-wide
cap.

Limit: files up to 65,536 bytes are read; a file of 65,537 bytes is refused.
File-read errors exit 65 with empty stdout and one fixed stderr line.

The interpreter uses Schway's fixed production Go file adapter and does not
accept a manifest or execute arbitrary C. The native route requires the
explicit binding manifest. The separator keeps a path opaque, including values
that look like flags, contain spaces, or contain shell metacharacters.

## Reusable native artifacts

Build each native artifact once, then use `app run` for the case commands below.
The build fences are the retained workflow; they produce files only under the
ignored `.schway/` directory.

```sh doc-test=build app=crc32
./schway build examples/crc32/crc32.schway --manifest examples/crc32/file_bytes.bindings.json --output .schway/crc32/crc32
```

```sh doc-test=build app=crc_low
./schway build examples/crc32/crc_low.schway --manifest examples/crc32/file_bytes.bindings.json --output .schway/crc32/crc_low
```

## Exact case results

Each cell is `(exit, stdout, stderr)`. The paths name the binary fixtures under
`.schway/crc32/inputs/`; `missing.bin` is deliberately absent. The table is
checked against the inert expected tuples in `crc32.expected.json`.

| Case | Fixture path | crc32 tuple | crc_low tuple |
| --- | --- | --- | --- |
| `empty` | `.schway/crc32/inputs/empty.bin` | `(0, "0\n", "")` | `(0, "0\n", "")` |
| `zero` | `.schway/crc32/inputs/zero.bin` | `(0, "3523407757\n", "")` | `(0, "61325\n", "")` |
| `check` | `.schway/crc32/inputs/check.bin` | `(0, "3421780262\n", "")` | `(0, "14630\n", "")` |
| `highbit-80ff` | `.schway/crc32/inputs/highbit-80ff.bin` | `(0, "1465410873\n", "")` | `(0, "25913\n", "")` |
| `all-bytes-00-ff` | `.schway/crc32/inputs/all-bytes.bin` | `(0, "688229491\n", "")` | `(0, "35955\n", "")` |
| `cap-minus-one` | `.schway/crc32/inputs/cap-minus-one.bin` | `(0, "529201972\n", "")` | `(0, "64308\n", "")` |
| `cap` | `.schway/crc32/inputs/cap.bin` | `(0, "3735780942\n", "")` | `(0, "32334\n", "")` |
| `cap-plus-one` | `.schway/crc32/inputs/cap-plus-one.bin` | `(65, "", "FileError.FileTooLarge\n")` | `(65, "", "FileError.FileTooLarge\n")` |
| `missing` | `.schway/crc32/inputs/missing.bin` | `(65, "", "FileError.FileUnavailable\n")` | `(65, "", "FileError.FileUnavailable\n")` |
| `directory` | `.schway/crc32/inputs/directory` | `(65, "", "FileError.FileUnavailable\n")` | `(65, "", "FileError.FileUnavailable\n")` |
| `opaque-json` | `.schway/crc32/inputs/--json` | `(0, "3421780262\n", "")` | `(0, "14630\n", "")` |
| `opaque-space` | `.schway/crc32/inputs/space input.bin` | `(0, "3421780262\n", "")` | `(0, "14630\n", "")` |
| `opaque-shell` | `.schway/crc32/inputs/x ;$(literal).bin` | `(0, "3421780262\n", "")` | `(0, "14630\n", "")` |

The fixture's expected values came from Python `zlib.crc32`; selected CRC values
are also cross-checked by Go's `hash/crc32.ChecksumIEEE`. That independent
calculation is separate from Schway execution.

## Run both public engines

These command templates are executed for every row in the table. Replace `PATH`
with the row's fixture path exactly; the test runs the same marked commands in a
temporary copy made only from `git ls-files` entries.

```sh doc-test=run app=crc32 route=interpreter
./schway run --engine=interpreter examples/crc32/crc32.schway -- PATH
```

```sh doc-test=run app=crc32 route=one_shot
./schway run --engine=native examples/crc32/crc32.schway --manifest examples/crc32/file_bytes.bindings.json -- PATH
```

```sh doc-test=run app=crc32 route=retained
./schway app run .schway/crc32/crc32 -- PATH
```

```sh doc-test=run app=crc_low route=interpreter
./schway run --engine=interpreter examples/crc32/crc_low.schway -- PATH
```

```sh doc-test=run app=crc_low route=one_shot
./schway run --engine=native examples/crc32/crc_low.schway --manifest examples/crc32/file_bytes.bindings.json -- PATH
```

```sh doc-test=run app=crc_low route=retained
./schway app run .schway/crc32/crc_low -- PATH
```

The one-shot native form builds in a private temporary location and runs once;
the retained form reuses the artifacts built above. Both preserve the program's
stdout, stderr, and exit status.

File-read errors exit 65 with empty stdout and one fixed stderr line; the path
is not included. An empty file is a valid input whose checksum is zero. An
empty path token is a separate CLI refusal with `(65, "", "")`.

Optional event and identity capture is bounded evidence, not the application
result. File application execution disables that capture; capture exhaustion
does not change the checksum or terminal outcome, and it does not establish a
complete execution trace. Normal return and typed file errors discharge the
owner once. Checked abort-like defects retain their defined cleanup boundary;
this guide does not promise destructor behavior on arbitrary aborts.

These examples are not claimed to be formatter-canonical. The public parity
claim covers the pinned file-program cases and canonical adapters, not arbitrary
divergent programs: native execution has a process deadline while the current
interpreter route has no general fuel or timeout cancellation.
