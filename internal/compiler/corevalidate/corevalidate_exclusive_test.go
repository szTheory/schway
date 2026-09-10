package corevalidate_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

const exclusiveBorrowValidateSource = `module owned.exclusive_borrow_validate

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let reviewed = borrow view
  view
}
`

func validExclusiveBorrowProgram(t *testing.T) core.Program {
	t.Helper()
	checked := session.Check([]byte(exclusiveBorrowValidateSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	if result := corevalidate.Validate(checked.Program); !result.Valid {
		t.Fatalf("baseline exclusive-borrow program rejected: %+v", result.Problems)
	}
	return cloneCoreProgram(t, checked.Program)
}

// TestExclusiveBorrowAuthorizedIndependently is 03-02-01's independent-
// validator falsifier: corevalidate.replay authorizes core.OpBorrowExclusive
// through its own dispatch arm and its own separately-written ability
// check, not by trusting the checker's admission. Denying the ability fact
// the validator itself re-derives — without touching the checker at all —
// must independently reject the operation.
func TestExclusiveBorrowAuthorizedIndependently(t *testing.T) {
	baseline := validExclusiveBorrowProgram(t)
	if result := corevalidate.Validate(baseline); !result.Valid {
		t.Fatalf("valid exclusive-borrow core rejected: %+v", result.Problems)
	}

	denied := cloneCoreProgram(t, baseline)
	linear := denied.Functions[0].Linear
	for index, fact := range linear.Types {
		granted := make([]core.Ability, 0, len(fact.Abilities))
		for _, ability := range fact.Abilities {
			if ability != core.AbilityShare {
				granted = append(granted, ability)
			}
		}
		linear.Types[index].Abilities = granted
	}
	result := corevalidate.Validate(denied)
	if result.Valid {
		t.Fatal("validator admitted an exclusive borrow whose re-derived ability fact denies share")
	}
}

// TestCorevalidateIndependentlyRejectsBorrowConflict is corevalidate's own
// D-12 falsifier for the five-row conflict matrix (T-03-07): the checker
// never emits a core artifact that violates the matrix (it rejects at the
// frontend first), so to exercise the validator's OWN conflict re-derivation
// in isolation this test takes a checker-approved baseline and mutates it at
// the CORE level — retargeting the second borrow's SourceID from the
// exclusive loan's own target place back to the shared owner place, without
// touching any ID/ordinal — so both loans now originate from the same
// owner and genuinely overlap (the Return operation's later read of the
// exclusive loan's target already extends its lastUse across the second
// borrow). This must be rejected by replay's own ownerLiveSharedUntil/
// ownerLiveExclusiveUntil bookkeeping, not by anything the checker decided.
func TestCorevalidateIndependentlyRejectsBorrowConflict(t *testing.T) {
	baseline := validExclusiveBorrowProgram(t)
	linear := baseline.Functions[0].Linear
	if len(linear.Operations) < 2 || linear.Operations[0].Kind != core.OpBorrowExclusive || linear.Operations[1].Kind != core.OpBorrowShared {
		t.Fatalf("unexpected baseline operation shape: %+v", linear.Operations)
	}
	ownerPlaceID := linear.Operations[0].SourceID // the exclusive loan's owner (buffer)
	mutated := cloneCoreProgram(t, baseline)
	mutated.Functions[0].Linear.Operations[1].SourceID = ownerPlaceID
	result := corevalidate.Validate(mutated)
	if result.Valid {
		t.Fatal("validator admitted an exclusive loan overlapping a shared loan on the same owner")
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.borrow_conflict" {
		t.Fatalf("code=%v want=core.borrow_conflict", result.Problems)
	}
}

// TestUnknownOperationStillRejected confirms replay's fail-closed default
// arm is still intact after this phase's new core.OpBorrowExclusive case is
// added: an operation kind the validator has not been explicitly taught is
// rejected as core.unknown_operation rather than silently admitted.
func TestUnknownOperationStillRejected(t *testing.T) {
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	program := cloneCoreProgram(t, checked.Program)
	linear := program.Functions[0].Linear
	if len(linear.Operations) == 0 {
		t.Fatal("expected at least one operation")
	}
	linear.Operations[0].Kind = core.OperationKind("borrow_super_exclusive")
	result := corevalidate.Validate(program)
	if result.Valid {
		t.Fatal("validator admitted an operation kind it has not been taught")
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.unknown_operation" {
		t.Fatalf("code=%v want=core.unknown_operation", result.Problems)
	}
}

// restrictBorrowCheckedProgram checks the Phase 5 tracer fixture, reused by
// the D-05-04 attribute-justification falsifiers below.
func restrictBorrowCheckedProgram(t *testing.T) session.CheckResult {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	return checked
}

func mustBeAttributeUnjustified(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	var attrErr *corevalidate.AttributeUnjustifiedError
	if !errors.As(err, &attrErr) || attrErr.Code != "core.attribute_unjustified" {
		t.Fatalf("want core.attribute_unjustified, got %v", err)
	}
}

// TestAttributeJustificationIsIndependentlyRederived is D-05-04's own
// independence falsifier: the validator produces the correct justification
// for restrict_borrow.lang purely from core.Program -- it is never given
// the sidecar's own JustifiedBy as input to its own re-derivation
// (recomputeAliasJustifications takes only a core.Program). Blanking the
// claim's JustifiedBy before validating still fails for the RIGHT reason (a
// mismatch against the independently re-derived value, not "no
// justification found at all"), proving a nonempty answer was computed from
// core alone.
func TestAttributeJustificationIsIndependentlyRederived(t *testing.T) {
	checked := restrictBorrowCheckedProgram(t)
	function := checked.Program.Functions[0]
	realLoanID := function.Linear.Operations[0].LoanID
	if realLoanID == "" {
		t.Fatal("fixture's first operation carries no loan id")
	}
	claim := corevalidate.AttributeClaim{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: realLoanID}
	if err := corevalidate.ValidateEmittedAttributes(checked.Program, []corevalidate.AttributeClaim{claim}); err != nil {
		t.Fatalf("expected the independently re-derived justification to match, got: %v", err)
	}

	blanked := claim
	blanked.JustifiedBy = ""
	mustBeAttributeUnjustified(t, corevalidate.ValidateEmittedAttributes(checked.Program, []corevalidate.AttributeClaim{blanked}))
}

// TestUnjustifiedAttributeIsRefused is D-05-03b's own falsifier: an
// emitted_attributes entry with an empty or wrong justified_by is refused
// with core.attribute_unjustified, a hard build failure never a warning.
func TestUnjustifiedAttributeIsRefused(t *testing.T) {
	checked := restrictBorrowCheckedProgram(t)
	function := checked.Program.Functions[0]

	wrong := corevalidate.AttributeClaim{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: "bogus-loan-id"}
	mustBeAttributeUnjustified(t, corevalidate.ValidateEmittedAttributes(checked.Program, []corevalidate.AttributeClaim{wrong}))

	empty := corevalidate.AttributeClaim{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: ""}
	mustBeAttributeUnjustified(t, corevalidate.ValidateEmittedAttributes(checked.Program, []corevalidate.AttributeClaim{empty}))
}

// TestUnprovenAttributeNameIsRefused proves the only justifiable attribute
// name is "restrict" -- an entry claiming any other optimizer attribute
// (e.g. "noalias") is refused outright, regardless of its JustifiedBy value.
func TestUnprovenAttributeNameIsRefused(t *testing.T) {
	checked := restrictBorrowCheckedProgram(t)
	function := checked.Program.Functions[0]
	claim := corevalidate.AttributeClaim{
		Attr: "noalias", CoreNode: function.ID, Parameter: function.Parameter.ID,
		JustifiedBy: function.Linear.Operations[0].LoanID,
	}
	mustBeAttributeUnjustified(t, corevalidate.ValidateEmittedAttributes(checked.Program, []corevalidate.AttributeClaim{claim}))
}

// TestAttributeValidatorImportsStayIndependent asserts corevalidate.go's own
// source text references neither "compiler/check" nor "compiler/cgen" --
// D-12's independence invariant, extended by this plan to the new attribute
// re-derivation (TestValidatorImportsStayIndependent, the package-internal
// sibling of this test, already covers check/ast; this test additionally
// covers cgen, which that internal test does not).
func TestAttributeValidatorImportsStayIndependent(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "corevalidate", "corevalidate.go"))
	if err != nil {
		t.Fatal(err)
	}
	// Phase 09 (D-09-05): also scan corevalidate_peer_liveness.go, the new
	// sibling production file (D-09-04) -- never assume a new file is
	// automatically covered by an existing literal-file-list scan.
	peerLivenessSource, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "corevalidate", "corevalidate_peer_liveness.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"compiler/check", "compiler/cgen"} {
		if strings.Contains(string(source), forbidden) {
			t.Fatalf("corevalidate.go must never reference %s", forbidden)
		}
		if strings.Contains(string(peerLivenessSource), forbidden) {
			t.Fatalf("corevalidate_peer_liveness.go must never reference %s", forbidden)
		}
	}
}
