package check

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

// The library pair for the body-blind guard: identical declared signature
// `mix(value: Byte) -> Byte`, different valid bodies. Both libraries define
// the private `other` so the function sets match and only mix's body differs.
const (
	phase30MixA = "module lib\n\nexport {\n  fn mix\n}\n\nfn other(value: Byte) -> Byte {\n  value\n}\n\nfn mix(value: Byte) -> Byte {\n  value\n}\n"
	phase30MixB = "module lib\n\nexport {\n  fn mix\n}\n\nfn other(value: Byte) -> Byte {\n  value\n}\n\nfn mix(value: Byte) -> Byte {\n  let again = other(value)\n  again\n}\n"

	phase30MixEntry = "module app.main\n\nimport lib {\n  fn mix\n}\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  let result = mix(value)\n  result\n}\n"
)

func phase30MixProgram(t *testing.T, lib string) (resultJSON string, contract calleeContract, importer string) {
	t.Helper()
	program := mergeForTest(t, []byte(phase30MixEntry), map[string][]byte{"lib": []byte(lib)})
	result := Program(program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected a clean check, got %+v", result.Diagnostics)
	}
	return phase30JSON(t, result.Diagnostics), buildScopedCalleeContracts(program)[""]["mix"],
		phase30JSON(t, phase30FunctionByID(t, result.Program, "s1:app.main:fn:main"))
}

// TestPhase30CallAdmissionBodyBlind pins SC2 / MOD-04: changing an imported
// function's body without changing its signature leaves the importer's
// diagnostics, its scoped contracts and its core JSON byte-identical, with no
// callee body read. Two armed controls are observed to break the equality.
func TestPhase30CallAdmissionBodyBlind(t *testing.T) {
	defer func() {
		verifyCallableRefusalBodyReadSeam = false
		scopedContractBodyReadSeam = false
		calleeBodyReadObserved = nil
	}()

	bodyReads := 0
	calleeBodyReadObserved = func(string) { bodyReads++ }
	diagnosticsA, contractA, importerA := phase30MixProgram(t, phase30MixA)
	diagnosticsB, contractB, importerB := phase30MixProgram(t, phase30MixB)
	calleeBodyReadObserved = nil

	if diagnosticsA != diagnosticsB {
		t.Fatalf("importer diagnostics differ across the pair:\n%s\n%s", diagnosticsA, diagnosticsB)
	}
	if contractA != contractB {
		t.Fatalf("scoped contracts differ across the pair: %+v vs %+v", contractA, contractB)
	}
	if contractA.ID != "s1:lib:fn:mix" || contractA.ParameterType != "Byte" || contractA.ReturnType != "Byte" || contractA.ReturnsBorrowOfParam {
		t.Fatalf("unexpected contract %+v", contractA)
	}
	if importerA != importerB {
		t.Fatalf("the importer's core function differs across the pair:\n%s\n%s", importerA, importerB)
	}
	if bodyReads != 0 {
		t.Fatalf("expected 0 callee body reads at the production default, got %d", bodyReads)
	}

	t.Run("control: a contract that reads the body breaks the equality", func(t *testing.T) {
		scopedContractBodyReadSeam = true
		defer func() { scopedContractBodyReadSeam = false }()
		a := buildScopedCalleeContracts(mergeForTest(t, []byte(phase30MixEntry), map[string][]byte{"lib": []byte(phase30MixA)}))[""]["mix"]
		b := buildScopedCalleeContracts(mergeForTest(t, []byte(phase30MixEntry), map[string][]byte{"lib": []byte(phase30MixB)}))[""]["mix"]
		if a == b {
			t.Fatalf("the body-read seam must make the contracts differ, both %+v", a)
		}
	})

	t.Run("control: the existing admission-arm body-read seam fires on the module fixture", func(t *testing.T) {
		reads := 0
		calleeBodyReadObserved = func(string) { reads++ }
		verifyCallableRefusalBodyReadSeam = true
		defer func() {
			verifyCallableRefusalBodyReadSeam = false
			calleeBodyReadObserved = nil
		}()
		Program(mergeForTest(t, []byte(phase30MixEntry), map[string][]byte{"lib": []byte(phase30MixA)}))
		if reads == 0 {
			t.Fatal("expected the body-read seam to make at least one callee body read observable")
		}
	})
}

const (
	phase30RelayEntry = "module app.main\n\nimport lib {\n  fn relay\n}\n\nexport {\n  fn main\n}\n\nfn main(buffer: Buffer) -> Buffer {\n  let result = relay(buffer)\n  result\n}\n"
	// Same declared signature as the callable relay, but the body returns a
	// borrowed view with no declared origin, which fails publication.
	phase30RelayCallable   = "module lib\n\nexport {\n  fn relay\n}\n\nfn relay(buffer: Buffer) -> Buffer {\n  buffer\n}\n"
	phase30RelayUncallable = "module lib\n\nexport {\n  fn relay\n}\n\nfn relay(buffer: Buffer) -> Buffer {\n  let view = borrow mut buffer\n  let reviewed = borrow view\n  view\n}\n"
)

// TestPhase30CallableFlipIsASignatureChange pins the documented boundary:
// Callable is part of the signature (originvalidate's FunctionSignature,
// check.go declaredContractField), so a body change that flips Callable
// legitimately flips the caller. The flip branch HELD: relay keeps its declared
// `Buffer -> Buffer` signature and flips Callable from true to false by a body
// change alone; the equal-signature mix pair above does not flip anything.
func TestPhase30CallableFlipIsASignatureChange(t *testing.T) {
	check := func(lib string) Result {
		return Program(mergeForTest(t, []byte(phase30RelayEntry), map[string][]byte{"lib": []byte(lib)}))
	}
	if got := check(phase30RelayCallable); len(got.Diagnostics) != 0 {
		t.Fatalf("the callable relay must check clean, got %+v", got.Diagnostics)
	}
	flipped := check(phase30RelayUncallable)
	if len(flipped.Diagnostics) != 1 || flipped.Diagnostics[0].Code != core.CalleeNotCallable {
		t.Fatalf("a Callable-flipping body change must flip the caller to %s, got %+v", core.CalleeNotCallable, flipped.Diagnostics)
	}
	// The refusal is the caller's: its span is the entry function's, so it is
	// attributed to the entry unit.
	if flipped.Diagnostics[0].Unit != "" {
		t.Fatalf("the refusal belongs to the calling unit, got %q", flipped.Diagnostics[0].Unit)
	}

	// Same declared signature on both sides: the contracts agree even though
	// the caller's verdict differs, because Callable is not a calleeContract
	// field and is only consulted from the post-body signature table.
	a := buildScopedCalleeContracts(mergeForTest(t, []byte(phase30RelayEntry), map[string][]byte{"lib": []byte(phase30RelayCallable)}))[""]["relay"]
	b := buildScopedCalleeContracts(mergeForTest(t, []byte(phase30RelayEntry), map[string][]byte{"lib": []byte(phase30RelayUncallable)}))[""]["relay"]
	if a != b {
		t.Fatalf("the pre-body contracts must agree across the Callable flip: %+v vs %+v", a, b)
	}
}

// scopedContractBodyReadViolations returns the body-bearing selectors
// (Body, Linear, Match) used in the named function of src outside an
// `if scopedContractBodyReadSeam { ... }` block.
func scopedContractBodyReadViolations(t *testing.T, filename string, src any, function string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filename, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := map[string]bool{"Body": true, "Linear": true, "Match": true}
	var found []string
	located := false
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name.Name != function {
			continue
		}
		located = true
		var walk func(node ast.Node)
		walk = func(root ast.Node) {
			ast.Inspect(root, func(node ast.Node) bool {
				switch typed := node.(type) {
				case *ast.IfStmt:
					if ident, ok := typed.Cond.(*ast.Ident); ok && ident.Name == "scopedContractBodyReadSeam" {
						// The guarded block is the sanctioned seam; the
						// else branch (if any) is not.
						if typed.Else != nil {
							walk(typed.Else)
						}
						return false
					}
				case *ast.SelectorExpr:
					if forbidden[typed.Sel.Name] {
						found = append(found, typed.Sel.Name)
					}
				}
				return true
			})
		}
		walk(fn.Body)
	}
	if !located {
		t.Fatalf("function %s not found in %s", function, filename)
	}
	return found
}

// TestPhase30ScopedContractReadsNoBody pins by construction that the scoped
// contract builder names no body-bearing selector outside its test seam.
func TestPhase30ScopedContractReadsNoBody(t *testing.T) {
	if got := scopedContractBodyReadViolations(t, "check.go", nil, "buildScopedCalleeContracts"); len(got) != 0 {
		t.Fatalf("buildScopedCalleeContracts reads body-bearing selectors %v outside scopedContractBodyReadSeam", got)
	}

	t.Run("negative-control: an unguarded body read is reported", func(t *testing.T) {
		synthetic := "package check\nfunc buildScopedCalleeContracts(callee FuncDecl) int { return len(callee.Body.Linear.Bindings) }\n"
		if got := scopedContractBodyReadViolations(t, "synthetic.go", synthetic, "buildScopedCalleeContracts"); len(got) == 0 {
			t.Fatal("the scanner must flag an unguarded .Body read")
		}
	})

	t.Run("negative-control: a guarded read is allowed", func(t *testing.T) {
		synthetic := "package check\nfunc buildScopedCalleeContracts(callee FuncDecl) int {\n\tif scopedContractBodyReadSeam {\n\t\treturn len(callee.Body.Linear.Bindings)\n\t}\n\treturn 0\n}\n"
		if got := scopedContractBodyReadViolations(t, "synthetic.go", synthetic, "buildScopedCalleeContracts"); len(got) != 0 {
			t.Fatalf("a read inside the seam block is the sanctioned control, got %v", got)
		}
	})
}
