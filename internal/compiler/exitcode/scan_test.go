package exitcode_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

var (
	exitConstNameRE  = regexp.MustCompile(`(?i)^exit|exit(code|status)$`)
	exitOperandRE    = regexp.MustCompile(`(?i)exit(code|status)?$`)
	exitSwitchTagRE  = regexp.MustCompile(`(?i)(status|exitcode)$`)
	repairMirrorFile = "cmd/schway-repair/main.go"
	// repairMirrorNames are the four D-22 pinned-mirror declarations of
	// cmd/schway-repair/main.go; nothing else in that file is exempt.
	repairMirrorNames = map[string]bool{
		"exitSuccess": true, "exitUnrepaired": true, "exitOperational": true, "exitUsage": true,
	}
)

func isIntLit(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && lit.Kind == token.INT
}

// finalName returns the last identifier of an identifier, selector or call.
func finalName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.CallExpr:
		return finalName(e.Fun)
	case *ast.ParenExpr:
		return finalName(e.X)
	}
	return ""
}

func isExitFlowFunc(name string, typ *ast.FuncType, cmdMain bool) bool {
	if typ.Results == nil || typ.Results.NumFields() != 1 {
		return false
	}
	if ident, ok := typ.Results.List[0].Type.(*ast.Ident); !ok || ident.Name != "int" {
		return false
	}
	return cmdMain || name == "ExitCode" || strings.Contains(name, "Exit")
}

// exitLiteralFindings reports every integer literal (including 0) sitting in
// an exit-code position of file. Negative literals (UnaryExpr) are out of scope.
func exitLiteralFindings(fset *token.FileSet, file *ast.File, path string) []string {
	slash := filepath.ToSlash(path)
	cmdMain := strings.Contains(slash, "cmd/") && strings.HasSuffix(slash, "/main.go")
	isRepairMain := strings.HasSuffix(slash, repairMirrorFile)
	var findings []string
	report := func(pos token.Pos, what string) {
		findings = append(findings, fmt.Sprintf("%s:%d: %s", slash, fset.Position(pos).Line, what))
	}
	var walk func(root ast.Node, exitFlow bool)
	walk = func(root ast.Node, exitFlow bool) {
		ast.Inspect(root, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.FuncLit:
				walk(n.Body, isExitFlowFunc("", n.Type, cmdMain))
				return false
			case *ast.ReturnStmt:
				if exitFlow {
					for _, result := range n.Results {
						if isIntLit(result) {
							report(result.Pos(), "integer literal returned from exit-flow func")
						}
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Exit" {
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "os" {
						for _, arg := range n.Args {
							if isIntLit(arg) {
								report(arg.Pos(), "integer literal passed to os.Exit")
							}
						}
					}
				}
			case *ast.ValueSpec:
				for index, name := range n.Names {
					if !exitConstNameRE.MatchString(name.Name) {
						continue
					}
					if isRepairMain && repairMirrorNames[name.Name] {
						continue
					}
					if index < len(n.Values) && isIntLit(n.Values[index]) {
						report(n.Values[index].Pos(), "integer literal bound to exit-named "+name.Name)
					}
				}
			case *ast.KeyValueExpr:
				if key, ok := n.Key.(*ast.Ident); ok && (key.Name == "ExitCode" || strings.HasSuffix(key.Name, "ExitStatus")) && isIntLit(n.Value) {
					report(n.Value.Pos(), "integer literal in "+key.Name+" field")
				}
			case *ast.BinaryExpr:
				switch n.Op {
				case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
					if isIntLit(n.X) && exitOperandRE.MatchString(finalName(n.Y)) {
						report(n.X.Pos(), "integer literal compared with "+finalName(n.Y))
					}
					if isIntLit(n.Y) && exitOperandRE.MatchString(finalName(n.X)) {
						report(n.Y.Pos(), "integer literal compared with "+finalName(n.X))
					}
				case token.ADD:
					if isIntLit(n.X) && strings.Contains(finalName(n.Y), "Signal") {
						report(n.X.Pos(), "signal arithmetic with integer literal")
					}
					if isIntLit(n.Y) && strings.Contains(finalName(n.X), "Signal") {
						report(n.Y.Pos(), "signal arithmetic with integer literal")
					}
				}
			case *ast.SwitchStmt:
				if n.Tag != nil && exitSwitchTagRE.MatchString(finalName(n.Tag)) {
					for _, stmt := range n.Body.List {
						if clause, ok := stmt.(*ast.CaseClause); ok {
							for _, value := range clause.List {
								if isIntLit(value) {
									report(value.Pos(), "integer literal in case of "+finalName(n.Tag))
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Body != nil {
				walk(fn.Body, isExitFlowFunc(fn.Name.Name, fn.Type, cmdMain))
			}
			continue
		}
		walk(decl, false)
	}
	sort.Strings(findings)
	return findings
}

// repairMirrorFindings verifies the schway-repair exit constants against the table.
func repairMirrorFindings(fset *token.FileSet, file *ast.File) []string {
	want := map[string]int{
		"exitSuccess":     exitcode.Success,
		"exitUnrepaired":  exitcode.Unrepaired,
		"exitOperational": exitcode.Operational,
		"exitUsage":       exitcode.Usage,
	}
	found := map[string]bool{}
	var findings []string
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for index, name := range spec.Names {
			if !exitConstNameRE.MatchString(name.Name) {
				continue
			}
			line := fset.Position(name.Pos()).Line
			expected, known := want[name.Name]
			if !known {
				findings = append(findings, fmt.Sprintf("line %d: extra exit-named declaration %s", line, name.Name))
				continue
			}
			found[name.Name] = true
			if index >= len(spec.Values) || !isIntLit(spec.Values[index]) {
				findings = append(findings, fmt.Sprintf("line %d: %s is not an integer literal", line, name.Name))
				continue
			}
			value, err := strconv.Atoi(spec.Values[index].(*ast.BasicLit).Value)
			if err != nil || value != expected {
				findings = append(findings, fmt.Sprintf("line %d: %s = %s, want %d", line, name.Name, spec.Values[index].(*ast.BasicLit).Value, expected))
			}
		}
		return true
	})
	for name := range want {
		if !found[name] {
			findings = append(findings, "missing mirror declaration "+name)
		}
	}
	sort.Strings(findings)
	return findings
}

func TestNoDuplicatedExitLiterals(t *testing.T) {
	roots := []string{
		testsupport.ProjectPath("internal", "compiler", "session"),
		testsupport.ProjectPath("internal", "compiler", "native"),
		testsupport.ProjectPath("internal", "compiler", "protocol"),
		testsupport.ProjectPath("internal", "compiler", "cgen"),
		testsupport.ProjectPath("cmd", "schway"),
		testsupport.ProjectPath("cmd", "schway-repair"),
	}
	repairDir := testsupport.ProjectPath("cmd", "schway-repair")
	var findings []string
	scanned := 0
	sawMain := false
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			if filepath.Dir(path) == repairDir && filepath.Base(path) != "main.go" {
				return nil
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			rel, relErr := filepath.Rel(testsupport.ProjectPath(), path)
			if relErr != nil {
				rel = path
			}
			scanned++
			if filepath.ToSlash(rel) == "cmd/schway/main.go" {
				sawMain = true
			}
			findings = append(findings, exitLiteralFindings(fset, file, rel)...)
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", root, err)
		}
	}
	if scanned < 20 {
		t.Fatalf("scanned %d files, want at least 20", scanned)
	}
	if !sawMain {
		t.Fatal("cmd/schway/main.go was not scanned")
	}
	if len(findings) > 0 {
		t.Fatalf("exit-code literals outside the exitcode table:\n%s", strings.Join(findings, "\n"))
	}
}

func TestExitScanDetectorFires(t *testing.T) {
	rows := []struct {
		name string
		path string
		src  string
	}{
		{"os-exit-argument", "x.go", "package p\nfunc f() { os.Exit(3) }\n"},
		{"cmd-main-int-return", "cmd/x/main.go", "package main\nfunc run() int { return 2 }\n"},
		{"exit-named-func-return", "x.go", "package p\nfunc ExitCode(s string) int { return 0 }\n"},
		{"exit-named-const", "x.go", "package p\nconst exitFoo = 7\n"},
		{"exit-code-composite-field", "x.go", "package p\nvar x = T{ExitCode: 4}\n"},
		{"comparison-field", "x.go", "package p\nfunc f() { if r.ExitCode == 1 {} }\n"},
		{"comparison-call", "x.go", "package p\nfunc f() { if exitErr.ExitCode() != 0 {} }\n"},
		{"signal-arithmetic", "x.go", "package p\nfunc f() int { return 128 + outcome.SignalNumber }\n"},
		{"switch-case-status", "x.go", "package p\nfunc f(status int) { switch status { case 5: } }\n"},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, row.path, row.src, 0)
			if err != nil {
				t.Fatal(err)
			}
			if got := exitLiteralFindings(fset, file, row.path); len(got) < 1 {
				t.Fatalf("detector found nothing in %q", row.src)
			}
		})
	}
	t.Run("clean-control", func(t *testing.T) {
		src := `package p
func f(xs []int) (int, error) {
	const limit = 10
	for i := 0; i < 3; i++ {
		_ = xs[1]
	}
	sentinel := -1
	_ = sentinel
	if len(xs) == 2 {
		return 1, nil
	}
	os.Exit(exitcode.Usage)
	return exitcode.Success, nil
}
func g() int { return exitcode.Operational }
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "clean.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}
		if got := exitLiteralFindings(fset, file, "clean.go"); len(got) != 0 {
			t.Fatalf("clean control produced findings: %v", got)
		}
	})
}

func TestRepairExitMirrorMatchesTable(t *testing.T) {
	path := testsupport.ProjectPath("cmd", "schway-repair", "main.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := repairMirrorFindings(fset, file); len(got) != 0 {
		t.Fatalf("schway-repair exit mirror drifted from the table:\n%s", strings.Join(got, "\n"))
	}
	t.Run("mismatched-value-is-reported", func(t *testing.T) {
		src := "package main\nconst (\n\texitSuccess = 0\n\texitUnrepaired = 9\n\texitOperational = 3\n\texitUsage = 64\n)\n"
		synthetic := token.NewFileSet()
		parsed, err := parser.ParseFile(synthetic, "main.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}
		got := repairMirrorFindings(synthetic, parsed)
		if len(got) != 1 || !strings.Contains(got[0], "exitUnrepaired") {
			t.Fatalf("findings = %v, want one exitUnrepaired mismatch", got)
		}
	})
	t.Run("missing-and-extra-are-reported", func(t *testing.T) {
		src := "package main\nconst (\n\texitSuccess = 0\n\texitExtra = 5\n)\n"
		synthetic := token.NewFileSet()
		parsed, err := parser.ParseFile(synthetic, "main.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}
		if got := repairMirrorFindings(synthetic, parsed); len(got) < 4 {
			t.Fatalf("findings = %v, want extra plus three missing", got)
		}
	})
}
