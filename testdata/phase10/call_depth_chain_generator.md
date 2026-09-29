# Call-depth chain generator contract

Phase 10 Plan 04 (SEM-08) needs a genuine `.schway` program with a call chain
deeper than `interp.MaxCallDepth` (128) -- specifically 129 chained
functions -- to prove the depth-exceeded refusal fires on a program the
compiler genuinely admits, never a hand-built `core.Program` (D-10-24).

That fixture is **generated in-test**, by
`generateCallDepthChainSource(n int) []byte` in
`internal/compiler/interp/interp_test.go`, rather than committed as a static
129-function `.schway` file in this directory. This document records the
generator's contract so a future reader can reconstruct the fixture without
reading the test.

## Why generated rather than committed

- A 129-line (or longer, for Task 3's subprocess probe) chain fixture is
  pure boilerplate with zero hand-authored content -- every link after the
  first is a mechanical transformation of the previous one. Committing it
  as a static file would add a large, low-signal file that must be kept in
  sync by hand if `MaxCallDepth` or the probe's own chain-depth constants
  ever change.
- The SAME generator is reused, parameterized by `n`, across three
  call sites with three different required depths: Task 1's boundary test
  (`n = MaxCallDepth` and `n = MaxCallDepth + 1`), Task 2's pipeline-
  conformance test (`n = MaxCallDepth + 1`), and Task 3's subprocess probe
  (`n = probeChainDepth`, a separate reduced-cap-derived constant). A single
  generator function keeps all three depths mechanically consistent with
  whatever `MaxCallDepth` is declared as, with no risk of a stale committed
  fixture silently drifting out of sync with the constant it is meant to
  bound.

## Module shape

```
module phase10.call_depth_chain

export {
  fn link0
}

fn link(N-1)(value: Byte) -> Byte {
  value
}

...

fn link1(value: Byte) -> Byte {
  let result = link2(value)
  result
}

fn link0(value: Byte) -> Byte {
  let result = link1(value)
  result
}
```

- Module name: `phase10.call_depth_chain`.
- Export block: exports exactly one function, `link0` -- the chain's sole
  entry point.
- Functions are named `link0` through `link(n-1)` for a chain of `n`
  functions. The generator emits them in REVERSE declaration order
  (`link(n-1)` first, `link0` last), mirroring `testdata/phase07/
  call_from_both_match_arms.schway`'s callee-before-caller convention,
  though declaration order is not load-bearing for either `syntax.Parse`
  or `check.Program`, both of which resolve calls by a two-pass name
  table rather than lexical position.

## Per-link template

Every link except the last (`linkI` for `0 <= I < n-1`) follows this exact
shape, generalizing `testdata/phase07/call_basic.schway`'s two-function
`main`/`identity` template to an arbitrary chain position:

```
fn linkI(value: Byte) -> Byte {
  let result = link(I+1)(value)
  result
}
```

The LAST link (`link(n-1)`) is the base case: no call, a bare parameter
return, following `identity`'s own shape in `call_basic.schway`:

```
fn link(n-1)(value: Byte) -> Byte {
  value
}
```

Every parameter and return type is `Byte`, matching `call_basic.schway`'s own
type choice -- the chain's SHAPE (call depth), not its data type, is what
this fixture exercises.

## Entry point

The exported entry point is always `link0`, regardless of `n`. Driving the
chain means calling `interp.Run(program, "link0", input)`: `link0` calls
`link1`, which calls `link2`, and so on down to `link(n-1)`'s bare return,
unwinding value-for-value back up to `link0`'s own caller.

## Depth semantics

A chain of `n` functions produces a frame-stack peak of exactly `n` frames
(the base frame for `link0`, plus one pushed frame per subsequent call, up
to and including `link(n-1)`'s frame). `MaxCallDepth` is therefore compared
directly against `n`: a chain of `n = MaxCallDepth` functions reaches
exactly the ceiling and returns normally; `n = MaxCallDepth + 1` requires
one more pushed frame than the ceiling permits and refuses.

## Reconstructing the fixture

Given only this document, `n`, and `MaxCallDepth`'s declared value, any
reader can regenerate the exact `.schway` source
`generateCallDepthChainSource(n)` produces by:

1. Writing the module header (`module phase10.call_depth_chain`) and export
   block (`export { fn link0 }`).
2. Emitting `link(n-1)` as the bare-return base case.
3. Emitting `link(n-2)` down through `link0`, each calling the next-higher
   link and returning its result unchanged.
