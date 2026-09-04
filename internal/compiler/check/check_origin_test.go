package check_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
)

const publicOriginSource = `module owned.public_origin_lowering

export {
  fn view
}

fn view(buffer: Buffer) -> borrow(buffer) Buffer {
  let observed = borrow buffer
  observed
}
`

// TestPublicOriginFactLowered is 03-06-01's checker falsifier: checking a
// function whose return type carries a borrow(path) annotation must produce
// a PublicOrigin fact naming the declared path and access mode, carried
// independently of the type's own abilities (Match/Linear stay unaffected).
func TestPublicOriginFactLowered(t *testing.T) {
	checked := session.Check([]byte(publicOriginSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 {
		t.Fatalf("want one function, got %d", len(checked.Program.Functions))
	}
	function := checked.Program.Functions[0]
	if function.PublicOrigin == nil {
		t.Fatalf("expected a PublicOrigin fact: %+v", function)
	}
	if len(function.PublicOrigin.Paths) != 1 || function.PublicOrigin.Paths[0] != "buffer" {
		t.Fatalf("expected origin path [buffer], got %+v", function.PublicOrigin.Paths)
	}
	if function.PublicOrigin.Access != "shared" {
		t.Fatalf("expected shared access, got %q", function.PublicOrigin.Access)
	}

	exclusive := `module owned.public_origin_exclusive

export {
  fn view
}

fn view(buffer: Buffer) -> borrow mut(buffer) Buffer {
  let observed = borrow mut buffer
  observed
}
`
	checkedExclusive := session.Check([]byte(exclusive))
	if len(checkedExclusive.Diagnostics) != 0 {
		t.Fatalf("unexpected exclusive diagnostics: %+v", checkedExclusive.Diagnostics)
	}
	exclusiveFunction := checkedExclusive.Program.Functions[0]
	if exclusiveFunction.PublicOrigin == nil || exclusiveFunction.PublicOrigin.Access != "exclusive" {
		t.Fatalf("expected exclusive PublicOrigin, got %+v", exclusiveFunction.PublicOrigin)
	}

	// A declared path other than the function's own (single) parameter is a
	// causal, span-bearing rejection — this reduced language has no other
	// legal path.
	wrongPath := `module owned.public_origin_wrong_path

export {
  fn view
}

fn view(buffer: Buffer) -> borrow(other) Buffer {
  let observed = borrow buffer
  observed
}
`
	checkedWrongPath := session.Check([]byte(wrongPath))
	if len(checkedWrongPath.Diagnostics) != 1 || checkedWrongPath.Diagnostics[0].Code != "origin.unknown_path" {
		t.Fatalf("expected origin.unknown_path, got %+v", checkedWrongPath.Diagnostics)
	}
}

// TestBorrowedReturnWithoutOriginRejected is 03-06-01's checker falsifier for
// OWN-04's "the origin annotation is mandatory" rule: `borrow` with no
// following `(path)` is a parse-level, causal, span-bearing rejection, not an
// inference — a borrowed return can never reach check.go without a declared
// origin, because the grammar itself makes the path structurally mandatory
// wherever `borrow` appears in return-type position.
func TestBorrowedReturnWithoutOriginRejected(t *testing.T) {
	source := `module owned.public_origin_missing_path

export {
  fn view
}

fn view(buffer: Buffer) -> borrow Buffer {
  let observed = borrow buffer
  observed
}
`
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) == 0 {
		t.Fatal("expected a parse diagnostic for a borrow annotation missing its origin path")
	}
	found := false
	for _, problem := range checked.Diagnostics {
		if problem.Code == "syntax.expected_lparen" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected syntax.expected_lparen among diagnostics, got %+v", checked.Diagnostics)
	}
}

// TestMatchBodyBorrowedReturnRejectedWithCause is 03-06-01's checker
// falsifier for the S1 scope fence: a match-bodied function (bare-arm or
// arm-body/branch) must not return a borrowed view this phase. It must be
// rejected with a named, span-bearing cause naming the scope limit, not with
// the generic type.return_mismatch a bare-match function's normal return-type
// gate would otherwise produce for an origin-annotated return type it does
// not recognize.
func TestMatchBodyBorrowedReturnRejectedWithCause(t *testing.T) {
	source := `module owned.public_origin_match_scope

export {
  type Switch
  fn direction
}

data Switch =
  | On
  | Off

fn direction(flag: Switch) -> borrow(flag) Switch {
  match flag {
    On => Off
    Off => On
  }
}
`
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "ownership.match_borrowed_return_unsupported" {
		t.Fatalf("expected exactly ownership.match_borrowed_return_unsupported, got %+v", checked.Diagnostics)
	}
	if checked.Diagnostics[0].Primary.Start == 0 && checked.Diagnostics[0].Primary.End == 0 {
		t.Fatalf("expected a real span, got %+v", checked.Diagnostics[0].Primary)
	}

	branch := `module owned.public_origin_branch_scope

export {
  type Switch
  fn direction
}

data Switch =
  | On
  | Off

fn direction(flag: Switch) -> borrow(flag) Switch {
  match flag {
    On => {
      let held = take flag
      held
    }
    Off => {
      let held = take flag
      held
    }
  }
}
`
	checkedBranch := session.Check([]byte(branch))
	if len(checkedBranch.Diagnostics) != 1 || checkedBranch.Diagnostics[0].Code != "ownership.match_borrowed_return_unsupported" {
		t.Fatalf("expected exactly ownership.match_borrowed_return_unsupported for the branch case, got %+v", checkedBranch.Diagnostics)
	}
	// The bare and branch fixtures above are the falsifier: confirm neither
	// path ever admitted a function into the checked core.
	if len(checked.Program.Functions) != 0 || len(checkedBranch.Program.Functions) != 0 {
		t.Fatalf("expected no admitted functions: bare=%+v branch=%+v", checked.Program.Functions, checkedBranch.Program.Functions)
	}
}
