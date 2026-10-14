package session_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 30 single-IO-site scan (D-10, T-30-29). Outside internal/compiler/modload
// and test files, no route reads a module source program: os.OpenRoot appears
// only in modload, the remaining readBoundedFile(…, syntax.MaxSourceBytes)
// sites are the classified single-file and corpus list, and every os.Open,
// os.OpenFile, os.ReadFile and os.Stat call under session and cmd is pinned to
// a purpose. A new unclassified read fails the scan.

const (
	phase30OSImport     = "os"
	phase30SyntaxImport = "github.com/szTheory/schway/internal/compiler/syntax"
)

type phase30IOSites struct {
	openRoot []string // "label:func" for every os.OpenRoot call
	bounded  []string // "label:func" for readBoundedFile(_, syntax.MaxSourceBytes)
	osCalls  []string // "label:func" for os.Open, OpenFile, ReadFile and Stat
}

// phase30ScanIO parses one file and records its IO call sites keyed by the
// enclosing top-level function. It resolves import aliases and fails closed on
// a dot import of os or syntax.
func phase30ScanIO(t *testing.T, label, filename string, src any) phase30IOSites {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	osNames := map[string]bool{}
	syntaxNames := map[string]bool{}
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, "\"")
		if path != phase30OSImport && path != phase30SyntaxImport {
			continue
		}
		if spec.Name != nil && spec.Name.Name == "." {
			t.Fatalf("fail-closed: dot import of %s in %s", path, label)
		}
		name := filepath.Base(path)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if path == phase30OSImport {
			osNames[name] = true
		} else {
			syntaxNames[name] = true
		}
	}

	var sites phase30IOSites
	scan := func(function string, body ast.Node) {
		ast.Inspect(body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			key := label + ":" + function
			switch fun := call.Fun.(type) {
			case *ast.SelectorExpr:
				receiver, isIdent := fun.X.(*ast.Ident)
				if !isIdent || !osNames[receiver.Name] {
					return true
				}
				switch fun.Sel.Name {
				case "OpenRoot":
					sites.openRoot = append(sites.openRoot, key)
				case "Open", "OpenFile", "ReadFile", "Stat":
					sites.osCalls = append(sites.osCalls, key)
				}
			case *ast.Ident:
				if fun.Name != "readBoundedFile" || len(call.Args) != 2 {
					return true
				}
				limit, isSelector := call.Args[1].(*ast.SelectorExpr)
				if !isSelector || limit.Sel.Name != "MaxSourceBytes" {
					return true
				}
				if receiver, isIdent := limit.X.(*ast.Ident); isIdent && syntaxNames[receiver.Name] {
					sites.bounded = append(sites.bounded, key)
				}
			}
			return true
		})
	}
	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			scan(typed.Name.Name, typed)
		case *ast.GenDecl:
			scan("<package>", typed)
		}
	}
	return sites
}

// phase30Unclassified returns the observed sites missing from the pinned set,
// and the pinned sites no longer observed.
func phase30Unclassified(observed []string, pinned map[string]string) (extra, missing []string) {
	seen := map[string]bool{}
	for _, site := range observed {
		seen[site] = true
		if _, ok := pinned[site]; !ok {
			extra = append(extra, site)
		}
	}
	for site := range pinned {
		if !seen[site] {
			missing = append(missing, site)
		}
	}
	sort.Strings(extra)
	sort.Strings(missing)
	return extra, missing
}

func phase30WalkIO(t *testing.T) (all phase30IOSites, sessionAndCmd phase30IOSites) {
	t.Helper()
	root := testsupport.ProjectPath()
	for _, dir := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			relative, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			label := filepath.ToSlash(relative)
			sites := phase30ScanIO(t, label, path, nil)
			all.openRoot = append(all.openRoot, sites.openRoot...)
			inScope := strings.HasPrefix(label, "internal/compiler/session/") || strings.HasPrefix(label, "cmd/")
			if inScope {
				sessionAndCmd.bounded = append(sessionAndCmd.bounded, sites.bounded...)
				sessionAndCmd.osCalls = append(sessionAndCmd.osCalls, sites.osCalls...)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return phase30Dedupe(all), phase30Dedupe(sessionAndCmd)
}

// phase30Dedupe sorts each list and drops repeats: a site is a function, not a
// call, so several calls in one function are one site.
func phase30Dedupe(sites phase30IOSites) phase30IOSites {
	unique := func(values []string) []string {
		sort.Strings(values)
		var out []string
		for _, value := range values {
			if len(out) == 0 || out[len(out)-1] != value {
				out = append(out, value)
			}
		}
		return out
	}
	return phase30IOSites{openRoot: unique(sites.openRoot), bounded: unique(sites.bounded), osCalls: unique(sites.osCalls)}
}

// phase30ModuleRouteFunctions are the functions that must never read a source
// file themselves: every module-aware route goes through checkSourcePath.
var phase30ModuleRouteFunctions = []string{
	"CheckFile", "RunInterpreterFile", "runInterpreterCommand", "RunNativeFile", "runNativeCommand",
	"BuildApplicationFile", "VerifyApplicationCasesFile", "DebugMapCommandFile",
	"checkSourcePath", "loadSourcePath", "checkLoaded", "buildApplicationChecked",
}

// phase30BoundedWant is the classified single-file and corpus set of
// readBoundedFile(_, syntax.MaxSourceBytes) call sites.
var phase30BoundedWant = map[string]string{
	"internal/compiler/session/session.go:EvidenceCommandFile":                               "single-file evidence (imports refused)",
	"internal/compiler/session/session.go:FormatFile":                                        "single-file format (imports formatted, never loaded)",
	"internal/compiler/session/session.go:InterfaceCoreCommandFile":                          "single-file interface core (imports refused)",
	"internal/compiler/session/session.go:InterfaceExportCommandFile":                        "single-file interface export (imports refused)",
	"internal/compiler/session/session.go:Phase4CheckedProgram":                              "repo corpus fixture, not a user module",
	"internal/compiler/session/session.go:ValidateEvidenceCommandFile":                       "single-file evidence validation",
	"internal/compiler/session/session.go:VerifyCorpus":                                      "corpus lane",
	"internal/compiler/session/session.go:verifyBorrowedCorpus":                              "corpus lane",
	"internal/compiler/session/session.go:verifyForeignCorpus":                               "corpus lane",
	"internal/compiler/session/session.go:verifyOwnedCorpus":                                 "corpus lane",
	"internal/compiler/session/session_phase6_evidence.go:ValidateEvidenceExpanded":          "single-file evidence expansion",
	"internal/compiler/session/session_phase6_explain.go:ExplainCommandFile":                 "single-file explain (imports refused)",
	"internal/compiler/session/session_phase6_query.go:resolveQueryDebugMapFacts":            "single-file query (imports refused)",
	"internal/compiler/session/session_phase6_query.go:resolveQueryDiagnosticFactsPresented": "single-file query (imports refused)",
	"internal/compiler/session/session_phase7.go:VerifyPhase7ControlsAndWork":                "corpus lane",
}

// phase30OSWant pins every os.Open, os.OpenFile, os.ReadFile and os.Stat call
// under internal/compiler/session and cmd to its purpose.
var phase30OSWant = map[string]string{
	"cmd/schway-repair/repair.go:applyRepair":                                               "schway-repair single-file source (imports not loaded)",
	"cmd/schway/main.go:isPhase5Corpus":                                                     "corpus presence probe",
	"cmd/schway/main.go:isPhase6Corpus":                                                     "corpus presence probe",
	"cmd/schway/main.go:isPhase7Corpus":                                                     "corpus presence probe",
	"internal/compiler/session/qlt01.go:AuditQLT01Registry":                                 "corpus presence probe",
	"internal/compiler/session/session.go:VerifyCorpus":                                     "corpus presence probe",
	"internal/compiler/session/session.go:readBoundedFile":                                  "single-file bounded read primitive",
	"internal/compiler/session/session_app_verify.go:removeStaleVerificationReport":         "output path guard",
	"internal/compiler/session/session_phase5.go:VerifyPhase5ControlsAndWork":               "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase5.go:phase5RunInterpreterO0O3LTOLane":           "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase5_alias.go:VerifyAliasFalseNoAlias":             "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase5_alias.go:assertAliasFalseNoAliasMovesAxis":    "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase5_alias.go:assertNonlocalExitMovesAxis":         "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase5_alias.go:assertReleaseTranspositionMovesAxis": "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase5_escapes.go:VerifyCoordinatedLieEscape":        "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase5_mismatch.go:ReduceSeededAliasMismatch":        "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase5_sanitize.go:compilePhase5SanitizeFixture":     "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase6.go:VerifyPhase6ControlsAndWork":               "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase6.go:phase6RunBorrowInjectionLane":              "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase6.go:phase6RunCleanupInjectionLane":             "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase6.go:phase6RunMatchInjectionLane":               "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase6.go:phase6RunMoveInjectionLane":                "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase6.go:phase6RunStaleEvidenceInjectionLane":       "repo corpus fixture, not a user module",
	"internal/compiler/session/session_phase6_budget.go:qlt02ManifestNeverWritten":          "non-source artifact",
	"internal/compiler/session/session_phase6_verify.go:VerifyPhase6ChangedRisk":            "corpus presence probe",
	"internal/compiler/session/session_phase6_verify.go:classifyPhase6FixtureKind":          "corpus presence probe",
	"internal/compiler/session/session_phase6_verify.go:phase6CompileBinary":                "non-source artifact",
}

func TestPhase30SingleIOSite(t *testing.T) {
	all, scoped := phase30WalkIO(t)

	t.Run("open-root-only-in-modload", func(t *testing.T) {
		want := []string{"internal/compiler/modload/fs.go:openModuleRoot"}
		if !reflect.DeepEqual(all.openRoot, want) {
			t.Fatalf("os.OpenRoot call sites = %v, want %v", all.openRoot, want)
		}
	})

	t.Run("bounded-reads-classified", func(t *testing.T) {
		extra, missing := phase30Unclassified(scoped.bounded, phase30BoundedWant)
		if len(extra) != 0 || len(missing) != 0 {
			t.Fatalf("readBoundedFile(…, syntax.MaxSourceBytes) sites: unclassified %v, pinned but gone %v\nobserved: %v", extra, missing, scoped.bounded)
		}
		for _, site := range scoped.bounded {
			function := site[strings.LastIndex(site, ":")+1:]
			for _, route := range phase30ModuleRouteFunctions {
				if function == route {
					t.Errorf("module-aware route %s reads a source file directly (%s)", route, site)
				}
			}
		}
	})

	t.Run("os-calls-classified", func(t *testing.T) {
		extra, missing := phase30Unclassified(scoped.osCalls, phase30OSWant)
		if len(extra) != 0 || len(missing) != 0 {
			t.Fatalf("os.Open/OpenFile/ReadFile/Stat sites: unclassified %v, pinned but gone %v\nobserved: %v", extra, missing, scoped.osCalls)
		}
		for site, purpose := range phase30OSWant {
			if strings.TrimSpace(purpose) == "" {
				t.Errorf("%s has no purpose", site)
			}
			function := site[strings.LastIndex(site, ":")+1:]
			if strings.HasSuffix(site, "session_modules.go:"+function) {
				t.Errorf("session_modules.go function %s must not open files", function)
			}
			for _, route := range phase30ModuleRouteFunctions {
				if function == route {
					t.Errorf("module-aware route %s is classified as opening a file (%s)", route, site)
				}
			}
		}
	})

	t.Run("negative-control-bounded-read", func(t *testing.T) {
		synthetic := "package session\nimport \"" + phase30SyntaxImport + "\"\nfunc CheckSomethingFile(path string) { _, _ = readBoundedFile(path, syntax.MaxSourceBytes) }\n"
		sites := phase30ScanIO(t, "synthetic.go", "synthetic.go", synthetic)
		if extra, _ := phase30Unclassified(sites.bounded, phase30BoundedWant); len(extra) != 1 {
			t.Fatalf("an unclassified readBoundedFile call must be reported, got %v", extra)
		}
	})

	t.Run("negative-control-open-root", func(t *testing.T) {
		synthetic := "package session\nimport \"os\"\nfunc probe(dir string) { _, _ = os.OpenRoot(dir) }\n"
		sites := phase30ScanIO(t, "internal/compiler/session/synthetic.go", "synthetic.go", synthetic)
		if len(sites.openRoot) != 1 || sites.openRoot[0] == "internal/compiler/modload/fs.go:openModuleRoot" {
			t.Fatalf("an os.OpenRoot call outside modload must be reported, got %v", sites.openRoot)
		}
	})

	t.Run("negative-control-os-call", func(t *testing.T) {
		synthetic := "package session\nimport osx \"os\"\nfunc CheckSomethingFile(path string) { _, _ = osx.ReadFile(path) }\n"
		sites := phase30ScanIO(t, "synthetic.go", "synthetic.go", synthetic)
		if extra, _ := phase30Unclassified(sites.osCalls, phase30OSWant); len(extra) != 1 {
			t.Fatalf("an unclassified os call (through an alias) must be reported, got %v", extra)
		}
	})
}
