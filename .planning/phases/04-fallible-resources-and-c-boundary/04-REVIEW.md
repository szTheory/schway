---
phase: 04-fallible-resources-and-c-boundary
reviewed: 2026-09-05T00:00:00Z
depth: standard
files_reviewed: 46
files_reviewed_list:
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/evidence/evidence.go
  - internal/compiler/evidence/evidence_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/foreign_nonlocal.go
  - internal/compiler/native/foreign_resource.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_conformance_test.go
  - internal/compiler/native/native_test.go
  - internal/compiler/native/symbols.go
  - internal/compiler/native/symbols_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/syntax_test.go
  - internal/compiler/syntax/token.go
  - native/lang_foreign_nonlocal.c
  - native/lang_foreign_resource.c
  - native/lang_foreign_resource_private.h
  - scripts/verify-phase4.sh
  - testdata/phase4/acquire_three_fail_second.lang
  - testdata/phase4/acquire_three_fail_third.lang
  - testdata/phase4/acquire_three_success.lang
  - testdata/phase4/defect_terminal.lang
  - testdata/phase4/discard_because.lang
  - testdata/phase4/fallible_call_unconsumed.lang
  - testdata/phase4/foreign_acquire_one.lang
  - testdata/phase4/foreign_call_target_not_foreign.lang
  - testdata/phase4/foreign_layout_mismatch.golden.c
  - testdata/phase4/foreign_origin_omitted.lang
  - testdata/phase4/foreign_policy_value_injection.lang
  - testdata/phase4/foreign_unwind_undeclared.lang
  - testdata/phase4/nonlocal_exit_probe.lang
findings:
  critical: 0
  warning: 1
  info: 2
  total: 3
status: issues_found
---

# Phase 04: Code Review Report

**Reviewed:** 2026-09-05T00:00:00Z
**Depth:** standard
**Files Reviewed:** 46
**Status:** issues_found

## Summary

This is the fifth and final gap-closure round for Phase 4, following plan 04-13's
three-layer audit of every `core.ForeignContract` string field that `cgen`
splices into generated C. The focus of this review was verifying that round's
core claim: that every field reaching emitted C text is now validated, that the
three layers (`check.foreign_policy_value_unsafe` / admission-time audit in
`check.go`; `foreign.policy_value_not_identifier` / `foreign.contract_field_not_c_safe`
in `corevalidate.go`; `commentSafeForeignField` / `validForeignCType` /
`unsafeForeignContractField` in `cgen.go`) are genuinely independent, and that
none of the new predicates over-rejects a legitimate value.

I traced every `core.ForeignContract` field (`Symbol`, `Allocator`, `Unwind`,
`NonlocalExit`, `Fails`, `InitializedState`, `Capture`, `Retention`, `Aliasing`,
`Alias`, and `Layout.{ForeignTypeName,Fields[].Name,Fields[].CType}`) forward
from its origin (source-level `ast.ForeignPolicy.Value` or a check.go-emitted
compiler constant) to every place it is spliced into generated C text
(`cgen.EmitForeignHeader`'s comment block, `cgen.EmitForeignConformance`'s
`_Static_assert` operands, `cgen.EmitForeignManifest`'s JSON, and
`emitLinearForeign`'s extern declaration). Every splice site funnels through
`singleForeignFunction`'s `validForeignSymbol`/`unsafeForeignContractField`
gate (cgen layer) and is independently re-derived in `corevalidate.linear`
(`validCIdentifier` / `foreignContractFieldsCSafe`) and, for source-controlled
policy values, in `check.collectForeignSymbols` (`validForeignPolicyValue`,
applied uniformly to every policy key's value before the key-specific switch,
so `fails` and `alias` values are audited too even though only `unwind`,
`nonlocal_exit`, and `allocator` are ever spliced raw). The three
byte-loop C-identifier predicates (`check.validForeignPolicyValue`,
`corevalidate.validCIdentifier`, `cgen.validForeignSymbol`) are
byte-for-byte identical in logic but textually separate implementations with
no shared helper or import between the three packages — a genuine
triple-independence posture, not one layer delegating to another. The
comment-safety and C-type-expression predicates (`commentSafe`/
`validCTypeExpression` in corevalidate; `commentSafeForeignField`/
`validForeignCType` in cgen) are likewise separately implemented and agree on
edge cases (empty string valid, leading/trailing/double space rejected in the
CType splitter, `"*/"` and `"/*"` both rejected, non-ASCII and control bytes
rejected).

`Layout`-shaped fields (`ForeignTypeName`, `Fields[].Name`, `Fields[].CType`)
are exercised only through `EmitForeignConformance`/`EmitForeignHeader`/
`EmitForeignManifest`, all of which route through `singleForeignFunction`;
the main per-function code generator (`emitLinearForeign`, used by `Emit`/
`EmitNative`) only ever splices `Symbol` raw into C (validated there directly
via a second `validForeignSymbol` call, independent of `corevalidate.Validate`
having run) and never touches `Allocator`/`Unwind`/`NonlocalExit`/`Fails`/
`Layout` in the emitted function body, so there is no unguarded splice path in
the normal compile pipeline. I did not find a value shape that a legitimate
foreign declaration would need but that these predicates reject: the sole
honest `CType`, `"unsigned char"`, round-trips through both
`validCTypeExpression` and `validForeignCType` correctly, and ordinary
identifier-shaped allocator/unwind/nonlocal_exit/fails/alias values pass.
`go build ./...`, `go vet ./...`, and the full `go test ./internal/compiler/...`
suite pass clean at review time.

I found no BLOCKER-level defects in this round's work. I did find one
WARNING-level completeness gap in the field enumeration's own documentation
(the `Alias` field is silently excluded from every audit layer, which is
currently safe only because no emitter splices it, and nothing enforces that
invariant going forward) and two INFO-level polish items.

## Warnings

### WR-01: `core.ForeignContract.Alias` is excluded from every C-injection audit layer with no enforcing test

**File:** `internal/compiler/corevalidate/corevalidate.go:1756-1781` (and `internal/compiler/cgen/cgen.go:856-912`)
**Issue:** `foreignContractFieldsCSafe` (corevalidate) and `unsafeForeignContractField`
(cgen) both enumerate every `core.ForeignContract` string field their own
package's emitters splice into C, and both omit `Alias` (distinct from
`Aliasing`, which *is* covered). This is correct *today* — grep confirms no
emitter in `cgen.go` ever references `contract.Alias` (it is consumed only by
`originvalidate.go` in a strict `switch` over `"borrow"`/`"retain"`, never
spliced into text). However:
- `check.go`'s admission-time `validForeignPolicyValue` loop *does* happen to
  validate the raw `alias` policy value (it runs before the per-key switch,
  so every key's value is audited, including one the switch does not
  otherwise consume specially) — but this is incidental coverage from a loop
  structured for a different purpose, not a field the corevalidate/cgen doc
  comments claim to protect.
- Neither `foreignContractFieldsCSafe`'s nor `unsafeForeignContractField`'s
  doc comment lists `Alias` as "deliberately excluded because it is never
  spliced" the way `Fails`/`InitializedState`/etc. are enumerated as
  "covered". A future engineer adding a comment-block line for `Alias` in
  `EmitForeignHeader` (a natural documentation addition, since every other
  contract field already gets one) would silently reintroduce exactly the
  injection class this whole gap-closure round exists to close, and neither
  `TestForeignContractCommentSafetyRefused` nor
  `TestForeignPolicyValueInjectionNeverReachesGeneratedC`'s
  `hostileForeignContractFields` table would catch it, because `Alias` is not
  a member of either table.
**Fix:** Add `Alias` to `hostileForeignContractFields` (cgen_test.go) and the
corevalidate equivalent even though no current emitter splices it, so a future
addition of an `/* alias: %s */` comment line is caught by the existing
falsifier tables rather than requiring a sixth gap-closure round. Alternatively,
add an explicit doc-comment line to `foreignContractFieldsCSafe`/
`unsafeForeignContractField` naming `Alias` as "deliberately unaudited because
no emitter splices it; must be added here first if that ever changes" so the
omission is a documented decision rather than a silent gap.

## Info

### IN-01: `validForeignPolicyValue`'s admission-time audit incidentally covers `Fails`/`Alias`, but this is not called out in the corevalidate/cgen doc comments' "the third field" framing

**File:** `internal/compiler/check/check.go:1080-1174`
**Issue:** The doc comment on `validForeignPolicyValue` (check.go:1080-1094)
and the surrounding plan documentation frame this round's work as auditing
"the three flat string fields" (Allocator/Unwind/NonlocalExit). In fact
`collectForeignSymbols`'s loop applies `validForeignPolicyValue` to *every*
declared policy's value regardless of key — including `fails` and `alias`,
and even an unrecognized key — before the key-specific switch discards
anything it doesn't recognize. This is stronger than what the docs claim (a
good thing), but the discrepancy between "three fields" in the narrative and
"every declared policy value" in the actual code could mislead a future
reader auditing for completeness into thinking `Fails`/`Alias` are unguarded
at the source-admission layer when they are not.
**Fix:** Update `validForeignPolicyValue`'s doc comment to state explicitly
that the admission-time gate covers every declared policy value uniformly,
not just the three named policy keys, so a future reader does not need to
re-derive this from the loop structure.

### IN-02: `foreignFailureLiteral`'s and `errorLiteral`'s C-string splice site is outside this round's audited field set but shares the same splice shape

**File:** `internal/compiler/cgen/cgen.go:741-742` (via `foreignFailureLiteral`, cgen.go:766-776)
**Issue:** `emitForeignReleasesAndFail` splices `foreignFailureLiteral`'s
return value (a resolved ADT alternative name, `dataType.Alternatives[0]`)
into generated C via `strconv.Quote(errorLiteral)` as an argument to
`lang_write_json_string`. This value is not part of `core.ForeignContract`
and so was correctly out of scope for 04-13's audit (which is scoped to
"every `core.ForeignContract` string field"), and it is safe today because
ADT alternative names are constrained to identifier lexical tokens by the
parser before ever reaching `core.DataType.Alternatives`. It is not currently
exploitable. Noting only because a future contributor grepping cgen.go for
"every splice site" to extend this audit might reasonably expect this call
site to already be covered by the `ForeignContract`-scoped predicates and be
surprised to find it isn't — its safety instead rests entirely on the parser's
identifier-token invariant holding for `data` declarations, a fact not
documented at this call site the way the `ForeignContract` fields' safety is.
**Fix:** Optional: add a one-line comment at `foreignFailureLiteral` or its
call site noting that its safety depends on `ast.Program.Data[].Alternatives`
names being lexer-level identifiers, so this invariant is documented at the
same standard as the audited `ForeignContract` fields, in case a future phase
loosens ADT alternative naming rules.

---

_Reviewed: 2026-09-05T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
