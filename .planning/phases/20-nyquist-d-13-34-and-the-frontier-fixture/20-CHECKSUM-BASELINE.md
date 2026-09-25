# Checksum frontier baseline at M003 open

There was **no checked-in `examples/checksum.lang` and no diagnostic pin for it at M003 open**. `git log --all -- examples/checksum.lang` returns no commits. The comparison below is a **historical reconstruction**, made on 2026-09-25 by checking the identical proposed source bytes with the M003-open checker and the current checker. It must not be described as a pin that existed at M003 open.

## Proposed exact fixture source

The source block below, including its final newline, has SHA-256 `ae95e550df67114c3464dda9cc24779f8a2efd69bd2b99dd0ccb2acc235cc154`. Plan 20-01 copies these bytes into `examples/checksum.lang`, then independently checks the digest. `checksum_read` and `checksum_print` are foreign-C boundary intent; `loop`, `next`, and two-argument `add` express the future byte walk and accumulation. The source is intentionally refused today.

```lang
module examples.checksum

export {
  fn main
}

foreign C {
  fn checksum_read(path: Buffer) -> Buffer {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: ReadError
  }
  fn checksum_print(value: U64) -> U64 {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: PrintError
  }
}

data ReadError =
  | ReadFailed

data PrintError =
  | PrintFailed

fn main(path: Buffer) -> U64 {
  let buffer = try checksum_read(path)
  let total = 0
  loop buffer {
    let current_byte = next(buffer)
    let total = add(total, current_byte)
  }
  let printed = try checksum_print(total)
  total
}
```

## Reproduction

Historical revision: `d21db90e67750bb19976c4206f4c23c68cd06207` (`docs: start milestone M003 Computation and Honest Instruments`). Its tree was extracted with `git archive d21db90` into a temporary directory. The same proposed source file was checked from each tree with `GOCACHE=/private/tmp/phase20-gocache go run ./cmd/lang --json check /private/tmp/phase20-checksum-probe.lang`.

| Checker tree | First diagnostic code | Primary span | Meaning |
|---|---|---|---|
| M003 open (`d21db90`) | `syntax.unexpected_byte` | 512–513 | numeric literal `0` was not tokenized |
| Current Phase 19 complete (`9b63b3b`) | `syntax.expected_rbrace` | 521–527 | parser reaches unsupported `loop` |

Both checker commands returned an invalid status, as expected. The historical checker also reported `syntax.expected_binding_source` at 512–513. The current checker also reported `syntax.expected_declaration` after the loop token. The first diagnostic moved from the numeric literal to the loop token for identical bytes. The historical first diagnostic ID was `diagnostic:f2a608b0eae7b9231352194d`; the current first diagnostic ID was `diagnostic:b301b1842cc7471a3b2002a9`.

**Evidence limit:** this demonstrates movement relative to the M003-open compiler, not movement relative to a contemporaneously pinned fixture. If Plan 20-01 changes any source byte or the checked-in fixture's first diagnostic differs, it must replay both checker revisions and update this record and its test from actual output before claiming the gate.

## Plan 20-01 pinned comparison

`TestPhase20ChecksumFrontier` binds its current refusal assertion to the fixture
digest above and records the reconstructed historical identity as
`d21db90e67750bb19976c4206f4c23c68cd06207` + `syntax.unexpected_byte` at
512–513. It independently asserts the current production check returns
`syntax.expected_rbrace` at 521–527 and that these identities differ. No
original M003-open checksum fixture or diagnostic pin existed; this test uses
the later historical replay, not an invented contemporaneous baseline.
