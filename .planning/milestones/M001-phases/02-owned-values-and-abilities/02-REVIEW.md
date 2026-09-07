---
phase: 02-owned-values-and-abilities
reviewed: 2026-09-04T02:40:00Z
depth: deep
files_reviewed: 24
files_reviewed_list:
  - internal/compiler/ability/ability.go
  - internal/compiler/ability/ability_test.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_names_test.go
  - internal/compiler/cgen/export_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/evidence/canonical_test.go
  - internal/compiler/evidence/evidence.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_test.go
  - internal/compiler/protocol/protocol.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/syntax_test.go
  - internal/compiler/testsupport/cli_test.go
  - internal/compiler/testsupport/testsupport.go
  - testdata/phase1/generated.golden.c
  - testdata/phase2/evidence.golden.json
  - testdata/phase2/reborrow_while_moved.lang
findings:
  critical: 0
  warning: 4
  info: 4
  total: 9
status: resolved_with_accepted_debt
resolution:
  critical_fixed_in: 3399ddc
  warnings_accepted_as_debt: .planning/phases/02-owned-values-and-abilities/02-DEBT.md
  reconciled_at: 2026-09-04
  reconciled_head: 3399ddc
---

# Phase 02: Code Review Report (post-fix-wave re-derivation)

**Reviewed:** 2026-09-04T02:40:00Z
**Depth:** deep
**Files Reviewed:** 24 (nine commits `f1206bb..da75e95`, re-derived at `HEAD = da75e95`)
**Status:** resolved with accepted debt (critical fixed in `3399ddc`)

## Summary

Eight of the nine prior findings are genuinely closed, and I proved each closure by
re-deriving it rather than accepting the commit messages. **The ninth — CR-02 — is not
closed: its fix (`890bcd0`) is a net regression.** It repaired the generic-return-type
case it targeted and simultaneously broke a case that worked correctly before the wave.

`890bcd0` replaced the old `f.previous`-based brace classifier with a `formatter.header`
state field. `formatter.newline()` clears `header` unconditionally (format.go:232-235), and
`TokenComment` routes through `newline()` (format.go:38-45). Therefore **any `//` comment
on a function's header line erases the classification**, the brace opens a `"block"`
context, and the exact CR-02 corruption returns — now for *every* parameter type, not just
generic ones. On a two-function module the damage crosses declaration boundaries: the
second function is swallowed into the first function's body. `lang format` exits 0 while
emitting source that does not reparse.

The old formatter handled both of these correctly. This is a strict regression, verified by
building `890bcd0^`'s `format.go` against the current tree and diffing the outputs.

Everything else in the wave is sound. The five blocking constraints all hold at HEAD.

### Verification evidence (commands actually run, `GOCACHE=/tmp/ai-lang-rereview`)

```
go build ./...                                                        # ok
go vet ./...                                                          # exit 0
go test ./... -count=1                                                # all ok
go test -race ./... -count=1                                          # all ok
bash scripts/verify-phase2.sh                                         # PASS, 9/9 controls
git diff 91a6206^ --exit-code -- testdata/phase1/generated.golden.c \
    testdata/phase1/evidence.golden.json \
    testdata/phase2/owned_transfer.golden.c                           # exit 0
```

### Requested claims, verified independently

| # | Claim | Verdict | How I proved it |
|---|---|---|---|
| 1 | Five blocking constraints hold | **HOLDS** | **Causality:** `verify-phase2.sh` emits `lane:owned-backend-causality controls=[control:backend.runtime_causality] work=2`. `session.go:66-90` still wraps the concrete `native.Runner` (not an interface), fails closed on `strings.Count(cSource, site) != 1` → `native.backend_control_invalid`, and `session.go:661-671` additionally requires `reflect.DeepEqual(Optimizations(), []string{"-O0","-O3"})` before the lane is credited, with `EngineMismatch` → `StatusMismatch` → exit 4. **Bounds:** independent grep (not the repo's own guard) found the only spawns are `native.go:134`, `evidence.go:96`, `testsupport.go:92,135` — all `exec.CommandContext` under `context.WithTimeout`, all with paired bounded writers; the only `io.ReadAll` in production is `session.go:400` under `io.LimitReader`. **C namespace:** 20 adversarial programs whose parameter is named after an emitted identifier (`lang_events`, `lang_record_event`, `lang_write_buffer_hex`, `lang_value_x`, `value`, `index`, `hex`, `encoded`, `event`, `escape`, `length`, `bytes`, `main`, `argc`, `argv`, `input`, `output`, `stdout`, `LANG_BUFFER`) all `run --engine=native` → `run pass` under `-Wall -Wextra -Werror -pedantic`. `dd9c0a8` only widened the reserved sets and added documentation, which is monotone-safe. **Goldens:** the three frozen files diff empty. |
| 2 | Oracle independence is real, not remirrored | **HOLDS** | I spliced `890bcd0^`'s `discoverLoanLastUses` body back into the current `check.go` and ran the differential: `--- FAIL: TestOwnershipSequenceExhaustive ... length=2 case=330 support mismatch` — exactly as claimed, `LoanFinalUses OperationIndex:1` (got) vs `:2` (want). The two derivations encode the same *law* by structurally incompatible *methods*: production (check.go:366-398) is a single forward pass carrying inherited association lists and never materializes the relation; the oracle (check_test.go:264-306) materializes the whole derivation edge set, closes each loan by order-independent fixed-point iteration, then reduces to the maximum ordinal. Neither is obtainable from the other by renaming. I ran the mirror experiment on the second layer too: reverting `corevalidate.go` to the one-hop `loanForTarget` form fails `TestTransitiveLoanBlocksMove`. |
| 3 | Buffer gaining `share` weakened nothing | **HOLDS** | `implicit_noncopy.lang` still yields `ownership.transfer_requires_take`; `move_while_borrowed.lang` is byte-unchanged (`git diff 91a6206^ --exit-code` → 0) and still fires `control:ownership.move_while_borrowed`. All nine controls fire in the live gate run. Only `copy` was ever the Buffer restriction that Phase 2 gates on: `AbilityShare` has exactly two consumers (`check.go:282`, `corevalidate.go:281`), both borrow gates, and `AbilityDrop`/`AbilitySend`/`AbilityEscape` are unconsumed. The two ability derivations remain independent implementations and `corevalidate.go:167` fails with `core.ability_mismatch` if they ever disagree. I re-ran the transitive-loan matrix by CLI: direct / 1-hop reborrow / copy-of-loan / 3-hop chain all correctly reject with `ownership.move_while_borrowed`, and the no-use-after-move control correctly passes (no over-blocking). |
| 4 | The new borrow gate is unreachable from source | **CONFIRMED** | I enumerated the type surface through the shipped CLI rather than through the ability package: `Byte`, `Buffer`, `Box<Byte>`, `Box<Buffer>`, `Box<Box<Buffer>>`, `Pair<Buffer, Buffer>`, `Pair<Box<Buffer>, Byte>` all `check pass` on a `borrow`; every other constructor spelling (`Unit`, `Foo`, `Box`, `Box<Byte, Byte>`, `Pair<Byte>`) is rejected earlier with `type.unknown` and never reaches the gate. Byte and Buffer both grant `share` (ability.go:112,124); `Box`/`Pair` route to `deriveStructural`, which conjoins. By structural induction `share` is universally granted. The fixer's reasoning is correct. See IN-01 for the posture judgement. |
| 5 | Golden compatibility | **HOLDS** | Three-file diff exits 0. `testdata/phase2/evidence.golden.json` moved in exactly two fields — I compared it as parsed JSON, key by key, not as text: `/core_digest` `f06bd365…` → `e17fe549…` and `/id` `evidence:fe86f7e0…` → `evidence:cea9b525…`. `source_digest`, `c_digest`, `execution_digests`, all schema/identity/flag/policy fields are byte-identical. This is the expected causal consequence of Buffer's `TypeFact` changing. |

### Prior-finding closure ledger

| Prior | Status | Basis |
|---|---|---|
| CR-01 `borrow` without `share` | **CLOSED** | Gate present at check.go:277-292; the previously-reproducing program now checks and runs cleanly instead of dying as a spanless exit-3. |
| CR-02 formatter corruption | **NOT CLOSED — REGRESSED** | Base generic case fixed; header-comment case newly broken for all types. See CR-01 below. |
| WR-01 evidence canonical round trip | **CLOSED** (residual: WR-02) | `evidence.go:184-189` fails closed on both reparse diagnostics and fixed-point failure; end-to-end `lang evidence` on a corrupting input now refuses instead of emitting a fabricated `name.unknown`. |
| WR-02 non-transitive loan liveness | **CLOSED** | Both admission layers made transitive and both mutation-proven by me (above). New fixture + ninth control wired at session and CLI layers. |
| WR-03 property generator gap | **PARTIALLY CLOSED** | Type × kind × binding-count axis now covered by `generatedLinearProgram` (syntax_test.go:424-452) plus two fuzz seeds. The comment-*placement* axis is still uncovered — which is precisely why the new BLOCKER shipped. See WR-03. |
| WR-04 unbounded testsupport spawns | **CLOSED** | `exec.CommandContext` + `context.WithTimeout` + paired `boundedWriter`s + typed `CLIError` on both helpers; `boundedWriter` arithmetic is correct (max+1 detection byte, `total` is the true count). Race detector clean over the converted concurrent test. |
| IN-01 dishonest cgen reserved sets | **CLOSED** | Sets lifted to `matchFixedNames`/`linearFixedNames` and made supersets; `cgen_names_test.go` derives the fixed-identifier set from generated output by disjoint-name differencing rather than restating a list, and asserts the prefix-confinement regexes directly. Non-vacuous. |
| IN-02 causality control coupled to a literal line | **OPEN** | Unchanged (`session.go:77`). Carried forward. |
| IN-03 `protocol.Human` unconverged | **OPEN** | Unchanged (`protocol.go:132-142`). Carried forward. |

Known and not re-reported: `Box`/`Pair` are checkable for ability derivation but not
executable in Phase 2, so `evidence`/`run` on a `Box`-typed function is an honest
`operational_failure`.

## Narrative Findings (AI reviewer)

### Critical

#### CR-01: `syntax.Format` destroys any module whose function header line carries a comment — a regression introduced by the CR-02 fix

**File:** `internal/compiler/syntax/format.go:232-235` (root cause), triggered from
`internal/compiler/syntax/format.go:38-45`, damage surfacing at
`internal/compiler/syntax/format.go:95-107` and `internal/compiler/syntax/format.go:137-147`
**Severity:** BLOCKER

**Issue.** `890bcd0` correctly identified that `f.previous` is not a sound brace classifier
and replaced it with a `formatter.header` field set at `TokenFn`/`TokenMatch`. But it clears
that field inside `newline()`:

```go
func (f *formatter) newline() {
	// A declaration header never spans a line in the canonical projection, so
	// the pending header classification dies with the line that carried it.
	f.header = ""
```

The comment's premise is false in the *input* projection. `TokenComment` (format.go:38-45)
calls `f.newline()`, so a `//` comment anywhere on the header line erases `header` before
the brace is seen. `TokenLBrace` then falls through to `context = "block"`, and the two
downstream handlers take the same wrong branches the original CR-02 described: `TokenEqual`
(format.go:142-146) takes the `data`-declaration branch (`" ="` + newline + `indent++` + an
unpopped `"data"` context) and `TokenIdentifier` (format.go:88-91) skips the newline that
terminates a binding, fusing the binding's RHS identifier with the body result.

**This is strictly worse than the bug it replaced.** The old classifier did *not* have this
failure: `f.previous` is not updated for comment tokens (format.go:176-178), so the old code
still saw `TokenIdentifier` and classified correctly. I proved the regression by building
`890bcd0^`'s `format.go` against the current tree:

```
$ cat /tmp/lc/c2.lang                    # plain Byte, no generics at all
module p.d
export { fn f }
fn f(v: Byte) -> Byte // c
{
  let w = take v
  w
}

$ lang check /tmp/lc/c2.lang             # exit 0 — valid program

$ langbin_old format /tmp/lc/c2.lang     # PRE-890bcd0: correct
fn f(v: Byte) -> Byte // c
 {
  let w = take v
  w
}

$ lang format /tmp/lc/c2.lang            # HEAD: corrupted, exit 0
fn f(v: Byte) -> Byte // c
 {
  let w =
    take vw                              # <-- `v` and `w` fused; result destroyed
  }
```

**Blast radius is larger than a single function.** With two functions, the unpopped `"data"`
context and leaked `indent++` swallow the *next* declaration into the first function's body:

```
$ lang format /tmp/lc/n2.lang            # exit 0
fn f(x: Byte) -> Byte // hdr
 {
  let a =
    take xa
  }

  fn g(y: Byte) -> Byte {                # <-- g is now nested inside f
    let b = take y
    b
  }
$ lang check <that output>
check invalid  syntax.expected_linear_result
```

**Impact.** (1) `lang format` writes non-reparsable source at **exit 0** — redirecting format
output over the input is data loss, and the loss now crosses declaration boundaries.
(2) `format --check` is permanently unsatisfiable for this class, so any "formatting is
enforced" CI rule is unsatisfiable. (3) The class *widened*: pre-wave only generic parameter
types were affected; now every type is, whenever a header comment is present. (4) The
generic + header-comment combination remains broken, so the original CR-02 is not fully
closed either. (5) The fixed point is destroyed: re-formatting the output produces different
bytes again, so `SourceDigest` would bind a non-idempotent projection.

Mitigating: `evidence` no longer launders this (WR-01's fix catches it and refuses,
verified: `lang evidence /tmp/lc/c1.lang` → `operational_failure`, exit 3), and I could not
construct a corrupted form that reparses into a *different valid* program — the result token
is always consumed, so the reparse always fails. It is data destruction, not silent semantic
drift.

**Fix.** A header classification must survive a comment-induced newline, because a comment
does not end the declaration. Preserve it across the comment:

```go
case TokenComment:
	if f.lineOpen {
		f.out.WriteByte(' ')
	} else {
		f.writeIndent()
	}
	f.out.WriteString(token.Text)
	pending := f.header
	f.newline()
	f.header = pending
```

I applied exactly this patch and verified it: all five probe programs (`n0`, `n2`, `c1`,
`c2`, `g0` — plain/generic × comment/no-comment × 0/1/2 bindings) format to source that
reparses and checks clean, `go test ./internal/compiler/syntax/ ./internal/compiler/session/`
stays green, and the three frozen goldens still diff empty. The working tree was restored;
no source file was modified by this review.

Then add the regression to the corpus: a fixture with a header-line comment plus at least
one binding, in `TestOwnershipRoundTrip`, in `FuzzParseFormat`'s seeds (syntax_test.go:249),
and via the generator axis described in WR-03.

### Warnings

#### WR-01: `TestSourceNeverSpawnsUnboundedProcesses` enforces two literal spellings, not the discipline it claims — I defeated it four ways

**File:** `internal/compiler/native/native_test.go:66-119` (needles at
`internal/compiler/native/native_test.go:73-76`)
**Severity:** WARNING

The guard's own documentation states that "every Go file in this module outside
`.planning/spikes/` must spawn processes through `exec.CommandContext` with independently
bounded stdout and stderr streams." It verifies neither property. It is a substring scan for
`"Combined"+"Output"` and `"exec."+"Command("`, so it forbids two spellings and nothing else.
I added a package under `internal/compiler/` containing four unbounded, deadline-free spawn
patterns and the scan passed clean:

```go
// B: no deadline at all, plain unbounded bytes.Buffer streams
c := exec.CommandContext(context.Background(), "sh", "-c", "yes")
var out, errb bytes.Buffer
c.Stdout, c.Stderr = &out, &errb

// C: unbounded stdout buffer via Output()
exec.CommandContext(context.Background(), "sh", "-c", "yes").Output()

// D: fully unbounded pipe drain
p, _ := c.StdoutPipe(); c.Start(); io.ReadAll(p)

// E: function-value indirection defeats the substring needle outright
var spawn = exec.Command
spawn("sh", "-c", "yes").Output()
```

`go test ./internal/compiler/native/ -run TestSourceNeverSpawnsUnboundedProcesses` → `ok`.
Pattern B is the most dangerous: `context.Background()` satisfies the needle while providing
*no* deadline, which is the precise hazard the guard names. `os.StartProcess` and
`syscall.Exec` are outside the scan entirely. (The probe package was removed; the tree is
clean.)

Production code is genuinely bounded today — I verified that independently by enumerating
every `exec.` site and every `ReadAll`/`ReadFile` by hand, not via this guard. So this is a
control-strength defect, not a live vulnerability. But SEC-02-C is recorded as closed on the
strength of a control that a one-line refactor silently disarms.

**Fix.** Scan the AST, not the text. Parse each file with `go/parser`, resolve the `os/exec`
import under any alias, and assert on the call graph: every `exec.Command`/`CommandContext`
call must (a) be `CommandContext` with a context derived from `WithTimeout`/`WithDeadline`,
and (b) have both `.Stdout` and `.Stderr` assigned a type implementing a bounded-writer
marker interface, with `.Output()`, `.CombinedOutput()`, `.StdoutPipe()`, and `.StderrPipe()`
forbidden. Also extend the scan to `os.StartProcess` and `syscall.Exec`. Until then, soften
the guard's doc comment so it does not overclaim.

---

#### WR-02: `evidence.canonical_unstable` can never be observed by any consumer of the shipped product

**File:** `internal/compiler/evidence/evidence.go:185` and
`internal/compiler/evidence/evidence.go:188`; mapping lost at
`internal/compiler/session/session.go:355-357` and `internal/compiler/session/session.go:678-681`
**Severity:** WARNING

`d9b370f` introduced the fail-closed code, and the fail-closed *behaviour* is real and
load-bearing (I confirmed it end to end against CR-01's corrupting input). But
`evidence.ErrorCode` — the only function that unwraps `*ValidationError` into its stable code
— has **zero production call sites**:

```
$ grep -rn "ErrorCode(" --include='*.go' .
internal/compiler/evidence/canonical_test.go:48   ← test
internal/compiler/evidence/evidence_test.go:72,73,83,88,92  ← tests
internal/compiler/evidence/evidence.go:424        ← the definition
```

`session.RunEvidence` returns the raw error upward and the CLI collapses it to a generic
code. Observed:

```
$ lang evidence /tmp/lc/c1.lang
evidence operational_failure
diagnostic evidence.operation_failed [0:0]: unable to construct evidence
exit=3
```

`verifyOwnedCorpus` (session.go:678-681) does the same, folding it into
`verify.evidence_build_failed`. So a user or CI consumer cannot distinguish "the formatter
corrupted your program, the manifest was refused" from "clang is missing" — which is exactly
the error-taxonomy distinction the protocol exists to make, and the same class of complaint
that the prior CR-01 raised about exit-3 misclassification. The commit's claim that Build
"fails closed **with** `evidence.canonical_unstable`" is true only inside the package.

**Fix.** Route the code through:

```go
product, diagnostics, err := evidence.Build(source, facts)
if err != nil {
	result := protocol.New("evidence", protocol.StatusOperational)
	result.Diagnostics = []diagnostic.Diagnostic{
		diagnostic.Error(evidence.ErrorCode(err), diagnostic.Span{}, "canonical projection is unstable"),
	}
	return evidence.Product{}, completeCommand(result, started, 1), nil
}
```

and assert the surfaced code at the CLI layer in `testsupport/cli_test.go`, so the observable
is covered where consumers actually read it.

---

#### WR-03: the new property generator never places a comment on a declaration header line — the exact blind spot that let CR-01 ship

**File:** `internal/compiler/syntax/syntax_test.go:424-452` (`generatedLinearProgram`),
mirrored at `internal/compiler/syntax/syntax_test.go:454-...` (`generatedProgram`)
**Severity:** WARNING

`2d98a78` closed the axis the prior WR-03 named — type × binding kind × binding count now
spans `{Byte, Buffer, Box<Byte>, Box<Box<Byte>>, Pair<Byte, Buffer>}` × `{copy, take, borrow}`
× `{0,1,2,3}` — and `TestGeneratedLinearRoundTrips` asserts reparse, fixed point, token
identity, semantic order, comment preservation, and diagnostic-code stability. That is a
strong property. It emits comments, too. But only in two positions:

```go
if caseID%3 == 0 {
	fmt.Fprintf(&source, "// generated linear case %d\n", caseID)   // before `module`
}
...
if caseID%7 == 0 && index == 0 {
	source.WriteString("// generated binding\n")                    // on its own line, in-body
}
```

Never *trailing a token on the `fn` header line*, which is the only placement the formatter's
new `header` state machine is sensitive to. Both existing placements are safe precisely
because `header` is already consumed or not yet set. The generator therefore ran 1000 cases
per run and could not see CR-01. The fuzz seeds (syntax_test.go:249-260) have the same hole:
none of the five seeds places a comment on a header line, and unguided mutation is unlikely
to synthesise `// text` between `->` and `{`.

**Fix.** Add a comment-*placement* dimension independent of the type dimension — at minimum
`{none, before-module, before-binding, trailing-header-line, trailing-binding-line,
between-`}`-and-next-`fn`}` — and cross it with the existing axes. Add a trailing-header-line
seed to `FuzzParseFormat`. The formatter's state machine is now stateful across tokens, so
comment placement is a first-class equivalence class, not decoration.

---

#### WR-04: `lang check` burns ~19s of CPU on a 358 KB source before the declared work limit fires; the wave amplified it ~2.6x

**File:** `internal/compiler/check/check.go:366-398`,
`internal/compiler/corevalidate/corevalidate.go:231-249`, limit at
`internal/compiler/session/session.go:36` enforced at `internal/compiler/session/session.go:108`
**Severity:** WARNING

`MaxCheckWork = 100_000` is checked *after* `check.Program` has already run, and
`recomputed_work` is linear in token count while actual cost is superlinear, so the declared
bound does not bound the work performed. `3d9493a` made this materially worse: both layers
now carry per-binding loan *sets* (`loansForBinding[index] = inherited` in check.go;
`carried := append([]string(nil), loansForPlace[...]...)` copied per operation in
corevalidate.go), so an N-link reborrow chain costs Θ(N²) where it previously cost Θ(N).

Measured on a chain of `let a_i = borrow a_{i-1}`, current HEAD vs. a binary built from
`f1206bb`:

| n | source bytes | HEAD | pre-wave `f1206bb` |
|---|---|---|---|
| 2000 | 51 KB | 0.34 s | 0.54 s |
| 6000 | 160 KB | 3.56 s | 1.43 s |
| 12000 | 326 KB | 15.6 s | 6.0 s |
| 13000 | 358 KB | **18.6 s** → `check.work_limit` | ~7 s |

The damage is genuinely **bounded** — `MaxTokens = 1<<17` fires around n≈13000 and `n=15000`
is rejected instantly with `syntax.input_limit` — so this is a fixed ~19s ceiling, not
unbounded resource exhaustion. But an untrusted 358 KB file costing 19s of CPU while the
tool reports `recomputed_work=131073` is a hidden runtime cost on untrusted input, against
two named project constraints ("Untrusted input … stays explicit"; the runtime-posture and
work-metering discipline). It is also a straightforward CI-latency footgun.

**Fix.** Charge the limit as work is consumed rather than after the fact: increment a counter
inside `discoverLoanLastUses`' inner loop and inside `validator.replay`'s `carried` copy, and
abort with `check.work_limit` when it trips. Structurally, both layers can drop to Θ(N) by
representing the inherited association as a parent pointer plus a union-find/`lastUse`
propagation instead of copying slices per binding — the transitive law is unchanged, only the
representation.

### Info

#### IN-01: `ownership.borrow_requires_share` is unreachable from source — the right defence, filed in the wrong shape

**File:** `internal/compiler/check/check.go:277-292`
**Severity:** INFO

I confirmed the fixer's reachability argument independently (claim 4 above): every
well-formed Phase 2 type grants `share`, so this branch cannot execute on any program that
survives type resolution. The gate is dead code today.

**Verdict: keep the gate, but the posture is currently half-right.** Keeping it is correct —
it is symmetric with the `AbilityCopy` gate five lines below, it removes a real
checker/validator incoherence, and `TestShareIsUniversallyGrantedAfterBufferShare`
(ability_test.go) is a good tripwire: it enumerates to depth 4 and fails the instant a
future constructor withholds `share`, with a message that names the consequence.

What is not right is that the *root cause* was resolved by changing the ability model rather
than by fixing the checker, and that decision is now recorded only in a commit message. The
load-bearing claim is "Buffer is shareable" — a language-design decision, not a bug fix — and
it is what made a genuine soundness hole unreachable. It belongs in the phase's decision
record and in a doc comment on `deriveAt`'s Buffer case, not solely in `2549311`'s body.

The second gap: because no negative fixture can exist, the gate has no source-level control,
so `ownership.borrow_requires_share` is a diagnostic code the product can emit but no test
exercises through the CLI. It is covered only by the synthetic-`TypeFact` differential. That
is acceptable — but the tripwire test's failure message should also instruct the reader to
add the CLI-layer negative control at that moment, since that is when it first becomes
possible.

---

#### IN-02: the causality control is still coupled to one exact generated line including a source-derived identifier

**File:** `internal/compiler/session/session.go:77`
**Severity:** INFO

Carried forward unchanged from the prior review. `OwnedBackendMutationRunner.Run` still
matches the literal
`"  LANG_BUFFER lang_value_delivered = lang_value_buffer; /* authority transfer: s1:owned.transfer:fn:relay:op:0 */\n"`.
Correctly fail-closed (`strings.Count != 1` → `native.backend_control_invalid`), but renaming
the binding in `owned_transfer.lang`, reindenting the emitter, or editing the comment text
converts the phase's strongest control into an opaque operational failure. Emit a stable
generated marker (e.g. `/* lang:mutation-site:value */`) and match on that instead.

---

#### IN-03: `protocol.Human` returns unconverged output where `protocol.JSON` errors

**File:** `internal/compiler/protocol/protocol.go:132-142`
**Severity:** INFO

Carried forward unchanged. `JSON` returns `"command output size did not converge"` after four
attempts (protocol.go:116-130); `Human` silently returns `human(result)` with a stale
`Metrics.OutputBytes`, so the human projection can self-report a byte count that does not
match its own length while the JSON projection refuses to. Make the two agree.

---

#### IN-04: a header-line comment leaves a stray leading space before the opening brace

**File:** `internal/compiler/syntax/format.go:102-103`
**Severity:** INFO

Independent of CR-01 and pre-existing (the pre-wave formatter does it too). After a comment
closes the line, `TokenLBrace` runs `f.trimSpace()` — which only strips trailing spaces
already in the buffer — and then unconditionally writes `" {\n"`, producing a line whose
first byte is a space and whose indentation is therefore wrong:

```
fn f(x: Byte) -> Byte // hdr
 {
  x
}
```

It is a fixed point and it reparses, so it is cosmetic, not a correctness issue. But the
canonical projection is meant to be the single normative source form, and this is visibly
not canonical. Write `f.writeIndent()` (or emit `"{"` without the leading space) when the
brace begins a line.

---

_Reviewed: 2026-09-04T02:40:00Z_
_Reviewer: gsd-code-reviewer (independent re-derivation at HEAD da75e95; the nine fix commits and my own prior report were both treated as untrusted)_
_Depth: deep_


---

## Reconciliation (2026-09-04, `3399ddc`)

_Added by the orchestrating agent after acting on this review. The review body above
is preserved verbatim as written at `da75e95`._

### The critical finding is fixed

CR-02's regression was reproduced independently through the shipped CLI before any
edit: `fn f(v: Byte) -> Byte // c` formatted to `take vw` (identifiers `v` and `w`
fused) at exit 0, the output failed to reparse with `syntax.expected_linear_result`,
and with two declarations the leaked indent swallowed `fn g` into `f`'s body. The
review's root cause was confirmed exactly.

Fixed in `3399ddc`:

- `internal/compiler/syntax/format.go` preserves `header` across the comment-induced
  line break — a comment is the one token that ends a line without ending the
  construct that opened it — and emits the brace without a separator space when it
  opens its own line (this also closes IN-04).
- `internal/compiler/syntax/syntax_test.go` closes the WR-03 blind spot the review
  identified: both generators now emit a comment trailing a declaration header
  (`caseID%11`) and trailing a `match` header (`caseID%13`).

**Falsification:** with the generators updated and `format.go` alone reverted,
`TestGeneratedLinearRoundTrips` fails at `case=11239299577418040656` with
`borrow hold1hold2` and `syntax.expected_linear_result` — so the new coverage is
load-bearing rather than decorative. Restored, the full syntax suite passes.

Post-fix at `3399ddc`: `go test ./...`, `go test -race ./...`, `go vet ./...`, and
`sh scripts/verify-phase2.sh` all exit 0. Phase 1 work 18, Phase 2 work 52, nine
controls. `git diff 91a6206^` over the three frozen goldens is empty.

### The four warnings are accepted as dated debt

WR-01 (defeatable spawn guard), WR-02 (`evidence.canonical_unstable` unreachable at
the CLI), WR-03 (residual generator gaps beyond the header placement now covered),
and WR-04 (Θ(N²) checker cost) are **not fixed**. They are recorded with reproduction
detail, non-blocking rationale, and a Phase 03 remedy in
`02-DEBT.md` as D-02-01 through D-02-03, alongside IN-02 and IN-03 as D-02-07 and
D-02-08.

The developer decided on 2026-09-04 to close Phase 02 recording these rather than run
a seventh review wave, on the grounds that all four are control strength, cost, or
observability — none changes what a correct program does — and that waves 5 and 6 each
introduced defects while fixing others.
