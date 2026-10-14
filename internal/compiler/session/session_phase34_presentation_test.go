package session_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase34ImportedDiagnosticPresentation(t *testing.T) {
	entry := writeModuleFixture(t, map[string]string{
		"main.schway": "module app.main\n\nimport lib {\n  fn calc\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = calc(n)\n  r\n}\n",
		"lib.schway":  "module lib\n\nexport {\n  fn calc\n}\n\nfn calc(n: U64) -> U64 {\n  let d = n + @\n  d\n}\n",
	})

	reads := map[string]int{}
	restore := modload.SetReadObserverForTest(func(module string) { reads[module]++ })
	defer restore()

	result, err := session.CheckCommandFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusInvalid || len(result.Diagnostics) == 0 {
		t.Fatalf("CheckCommandFile status=%q diagnostics=%v, want an imported syntax error", result.Status, result.Diagnostics)
	}
	if reads[""] != 1 || reads["lib"] != 1 || len(reads) != 2 {
		t.Fatalf("loader reads = %v, want entry and lib exactly once", reads)
	}

	encoded := phase30ResultJSON(t, result)
	var wire struct {
		Diagnostics []struct {
			Unit        string          `json:"unit"`
			PrimarySpan diagnostic.Span `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(encoded), &wire); err != nil {
		t.Fatal(err)
	}
	if len(wire.Diagnostics) == 0 {
		t.Fatalf("JSON has no diagnostics: %s", encoded)
	}
	problem := wire.Diagnostics[0]
	if problem.Unit != "lib" {
		t.Fatalf("JSON unit = %q, want dotted module name lib: %s", problem.Unit, encoded)
	}
	if problem.PrimarySpan.Line != 8 || problem.PrimarySpan.Col != 15 {
		t.Fatalf("imported primary location = %d:%d, want 8:15", problem.PrimarySpan.Line, problem.PrimarySpan.Col)
	}
	line := "  let d = n + @"
	wantStart := strings.Index(line, "@")
	if problem.PrimarySpan.Start != wantStart+len("module lib\n\nexport {\n  fn calc\n}\n\nfn calc(n: U64) -> U64 {\n") || problem.PrimarySpan.End != problem.PrimarySpan.Start+1 {
		t.Fatalf("imported byte span = [%d:%d], want the @ byte", problem.PrimarySpan.Start, problem.PrimarySpan.End)
	}

	human, err := protocol.Human(result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(human, "lib.schway:8:15") || !strings.Contains(human, line) {
		t.Fatalf("human output does not show imported path, position, and source line:\n%s", human)
	}
	if strings.Contains(human, filepath.Dir(entry)) || strings.Contains(encoded, filepath.Dir(entry)) {
		t.Fatalf("diagnostic presentation leaked a canonical absolute path:\n%s\n%s", human, encoded)
	}
}

func TestPhase34PresentationUsesCapturedSource(t *testing.T) {
	entrySource := "module app.main\n\nimport lib {\n  fn relay\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = n\n  r\n}\n"
	libSource := "module lib\n\nexport {\n  fn relay\n}\n\nfn relay(value: U64) -> U64 {\n  let result = missing_name\n  result\n}\n"
	entry := writeModuleFixture(t, map[string]string{"main.schway": entrySource, "lib.schway": libSource})

	reads := map[string]int{}
	restoreReads := modload.SetReadObserverForTest(func(module string) { reads[module]++ })
	defer restoreReads()

	result, err := session.CheckCommandFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	unknown, ok := phase30FindCode(result.Diagnostics, "name.unknown_place")
	if !ok || unknown.Unit != "lib" || result.Presentation == nil {
		t.Fatalf("result = %+v, want imported checker failure with captured presentation", result)
	}
	if reads[""] != 1 || reads["lib"] != 1 || len(reads) != 2 {
		t.Fatalf("loader reads = %v, want entry and lib exactly once", reads)
	}
	if got := string(result.Presentation.Sources["lib"].Bytes[unknown.Primary.Start:unknown.Primary.End]); got != "missing_name" {
		t.Fatalf("captured imported primary selects %q, want missing_name", got)
	}

	beforeHuman, err := protocol.Human(result)
	if err != nil {
		t.Fatal(err)
	}
	beforeJSON, err := protocol.JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	var beforeWire struct {
		Diagnostics []struct {
			ID          string          `json:"id"`
			Unit        string          `json:"unit"`
			PrimarySpan diagnostic.Span `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal(beforeJSON, &beforeWire); err != nil {
		t.Fatal(err)
	}
	var located struct {
		id   string
		span diagnostic.Span
	}
	for _, item := range beforeWire.Diagnostics {
		if item.ID == unknown.ID {
			located.id, located.span = item.ID, item.PrimarySpan
			break
		}
	}
	if located.id == "" || located.span.Line < 1 || located.span.Col < 1 || located.span.Start != unknown.Primary.Start || located.span.End != unknown.Primary.End {
		t.Fatalf("imported checker JSON span = %+v, want the same byte span with a source location", located)
	}
	wantLine, wantCol := phase34LocationAt(result.Presentation.Sources["lib"].Bytes, unknown.Primary.Start)
	if located.span.Line != wantLine || located.span.Col != wantCol {
		t.Fatalf("imported checker JSON location = %d:%d, want source-derived %d:%d", located.span.Line, located.span.Col, wantLine, wantCol)
	}
	if !strings.Contains(beforeHuman, fmt.Sprintf("lib.schway:%d:%d", wantLine, wantCol)) || !strings.Contains(beforeHuman, "missing_name") || strings.Contains(beforeHuman, filepath.Dir(entry)) {
		t.Fatalf("imported checker output should use captured relative source without host path:\n%s", beforeHuman)
	}

	// A deliberate wrong-unit inventory substitution is a negative control:
	// the exact imported-source assertion must stop matching, then the
	// original captured inventory is restored before any later check.
	imported := result.Presentation.Sources["lib"]
	result.Presentation.Sources["lib"] = result.Presentation.Sources[""]
	wrongUnitHuman, err := protocol.Human(result)
	result.Presentation.Sources["lib"] = imported
	if err != nil {
		t.Fatal(err)
	}
	if wrongUnitHuman == beforeHuman || strings.Contains(wrongUnitHuman, "lib.schway:") || !strings.Contains(beforeHuman, "lib.schway:") {
		t.Fatalf("wrong-unit control did not break the exact imported label assertion:\n%s", wrongUnitHuman)
	}

	if err := os.Remove(entry); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(filepath.Dir(entry), "lib.schway")); err != nil {
		t.Fatal(err)
	}
	afterHuman, err := protocol.Human(result)
	if err != nil {
		t.Fatal(err)
	}
	afterJSON, err := protocol.JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	if beforeHuman != afterHuman || string(beforeJSON) != string(afterJSON) {
		t.Fatalf("rendering changed after source files were removed:\nbefore human:\n%s\nafter human:\n%s\nbefore JSON:\n%s\nafter JSON:\n%s", beforeHuman, afterHuman, beforeJSON, afterJSON)
	}
	if reads[""] != 1 || reads["lib"] != 1 || len(reads) != 2 {
		t.Fatalf("rendering performed extra source reads: %v", reads)
	}
	if result.Diagnostics[0].ID != unknown.ID || result.Diagnostics[0].Primary.Line != 0 || result.Diagnostics[0].Primary.Col != 0 {
		t.Fatalf("presentation mutated semantic diagnostic identity or input location: %+v", result.Diagnostics[0])
	}

	t.Run("path mismatch and nested effective units", func(t *testing.T) {
		const importer = "module app.main\n\nimport lib {\n  fn twice\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = twice(n)\n  r\n}\n"
		mismatchEntry := writeModuleFixture(t, map[string]string{
			"main.schway": importer,
			"lib.schway":  strings.Replace(phase30TwiceLib, "module lib", "module other", 1),
		})
		mismatch, err := session.CheckCommandFile(mismatchEntry)
		if err != nil {
			t.Fatal(err)
		}
		pathProblem, ok := phase30FindCode(mismatch.Diagnostics, "module.path_mismatch")
		if !ok || len(pathProblem.Causes) == 0 || pathProblem.Causes[0].Span == nil || pathProblem.Causes[0].Unit != "lib" {
			t.Fatalf("path mismatch = %+v, want an imported-unit cause", pathProblem)
		}
		encoded, err := protocol.JSON(mismatch)
		if err != nil {
			t.Fatal(err)
		}
		var wire struct {
			Diagnostics []struct {
				ID     string `json:"id"`
				Unit   string `json:"unit"`
				Causes []struct {
					Unit string          `json:"unit"`
					Span diagnostic.Span `json:"span"`
				} `json:"causes"`
			} `json:"diagnostics"`
		}
		if err := json.Unmarshal(encoded, &wire); err != nil {
			t.Fatal(err)
		}
		var found *struct {
			ID     string `json:"id"`
			Unit   string `json:"unit"`
			Causes []struct {
				Unit string          `json:"unit"`
				Span diagnostic.Span `json:"span"`
			} `json:"causes"`
		}
		for index := range wire.Diagnostics {
			if wire.Diagnostics[index].ID == pathProblem.ID {
				found = &wire.Diagnostics[index]
				break
			}
		}
		if found == nil || len(found.Causes) == 0 || found.Causes[0].Unit != "lib" || found.Causes[0].Span.Line < 1 || found.Causes[0].Span.Col < 1 {
			t.Fatalf("path mismatch cause was not located from the imported unit: %s", encoded)
		}
		causeBytes := mismatch.Presentation.Sources["lib"].Bytes
		wantLine, wantCol := phase34LocationAt(causeBytes, pathProblem.Causes[0].Span.Start)
		if found.Causes[0].Span.Start != pathProblem.Causes[0].Span.Start || found.Causes[0].Span.End != pathProblem.Causes[0].Span.End || found.Causes[0].Span.Line != wantLine || found.Causes[0].Span.Col != wantCol {
			t.Fatalf("path mismatch cause = %+v, want imported byte span at %d:%d", found.Causes[0].Span, wantLine, wantCol)
		}
		if pathProblem.ID != found.ID || pathProblem.Unit != found.Unit {
			t.Fatalf("presentation changed path mismatch identity or unit: before=%+v wire=%+v", pathProblem, found)
		}
	})

	t.Run("origin peer and synthetic nested spans", func(t *testing.T) {
		originEntry := writeModuleFixture(t, map[string]string{
			"main.schway": "module app.main\n\nimport lib {\n  fn relay\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = n\n  r\n}\n",
			"lib.schway":  "module lib\n\nexport {\n  fn relay\n}\n\nfn relay(value: Buffer) -> Buffer {\n  let result = borrow value\n  result\n}\n",
		})
		origin, err := session.CheckCommandFile(originEntry)
		if err != nil {
			t.Fatal(err)
		}
		refusal, ok := phase30FindCode(origin.Diagnostics, "core.origin_omitted")
		if !ok || refusal.Unit != "lib" || refusal.Primary.Line != 0 || refusal.Causes[0].Unit != "" {
			t.Fatalf("origin peer refusal = %+v, want imported unit with a same-unit cause", refusal)
		}
		originJSON, err := protocol.JSON(origin)
		if err != nil {
			t.Fatal(err)
		}
		if origin.Presentation == nil || origin.Presentation.Sources[""].Label != originEntry {
			t.Fatalf("entry presentation label does not preserve the CLI-supplied path %q", originEntry)
		}
		if !strings.Contains(string(originJSON), `"unit":"lib"`) || !strings.Contains(phase30ResultHuman(t, origin), "lib.schway:") {
			t.Fatalf("origin peer refusal lost its captured imported location: %s", originJSON)
		}
		var originWire struct {
			Diagnostics []struct {
				ID      string          `json:"id"`
				Unit    string          `json:"unit"`
				Primary diagnostic.Span `json:"primary_span"`
				Causes  []struct {
					Span diagnostic.Span `json:"span"`
				} `json:"causes"`
			} `json:"diagnostics"`
		}
		if err := json.Unmarshal(originJSON, &originWire); err != nil {
			t.Fatal(err)
		}
		var locatedOrigin *struct {
			ID      string          `json:"id"`
			Unit    string          `json:"unit"`
			Primary diagnostic.Span `json:"primary_span"`
			Causes  []struct {
				Span diagnostic.Span `json:"span"`
			} `json:"causes"`
		}
		for index := range originWire.Diagnostics {
			if originWire.Diagnostics[index].ID == refusal.ID {
				locatedOrigin = &originWire.Diagnostics[index]
				break
			}
		}
		if locatedOrigin == nil || locatedOrigin.Unit != "lib" || locatedOrigin.Primary.Start != refusal.Primary.Start || locatedOrigin.Primary.End != refusal.Primary.End || locatedOrigin.Primary.Line < 1 || len(locatedOrigin.Causes) != 1 || locatedOrigin.Causes[0].Span.Start != refusal.Causes[0].Span.Start || locatedOrigin.Causes[0].Span.End != refusal.Causes[0].Span.End || locatedOrigin.Causes[0].Span.Line < 1 {
			t.Fatalf("origin refusal primary or cause lost its imported source coordinates: %s", originJSON)
		}

		entryBytes := origin.Presentation.Sources[""].Bytes
		libBytes := origin.Presentation.Sources["lib"].Bytes
		entryStart := strings.Index(string(entryBytes), "n: U64")
		libCause := strings.Index(string(libBytes), "borrow value")
		libRepair := strings.Index(string(libBytes), "result")
		if entryStart < 0 || libCause < 0 || libRepair < 0 {
			t.Fatal("fixture tokens for nested presentation spans are missing")
		}
		cause, repair := diagnostic.Span{Start: libCause, End: libCause + len("borrow value")}, diagnostic.Span{Start: libRepair, End: libRepair + len("result")}
		unmapped := diagnostic.Span{Start: 0, End: 1}
		invalid := diagnostic.Span{Start: len(libBytes) + 1, End: len(libBytes) + 4}
		crossUnit := diagnostic.ErrorWithRepairs("test.cross_unit", diagnostic.Span{Start: entryStart, End: entryStart + len("n: U64")}, "cross unit",
			[]diagnostic.Cause{{Kind: "nil"}, {Kind: "imported", Unit: "lib", Span: &cause}, {Kind: "missing", Unit: "missing", Span: &unmapped}, {Kind: "invalid", Unit: "lib", Span: &invalid}},
			diagnostic.Repair{Kind: "imported", Unit: "lib", Span: &repair, Replacement: "fixed"},
			diagnostic.Repair{Kind: "missing", Unit: "missing", Span: &unmapped},
			diagnostic.Repair{Kind: "invalid", Unit: "lib", Span: &invalid},
			diagnostic.Repair{Kind: "nil"},
		)
		crossID := crossUnit.ID
		crossResult := protocol.New("check", protocol.StatusInvalid)
		crossResult.Diagnostics = []diagnostic.Diagnostic{crossUnit}
		crossResult = protocol.WithPresentation(crossResult, *origin.Presentation)
		crossJSON, err := protocol.JSON(crossResult)
		if err != nil {
			t.Fatal(err)
		}
		var nested struct {
			Diagnostics []struct {
				ID      string          `json:"id"`
				Primary diagnostic.Span `json:"primary_span"`
				Causes  []struct {
					Kind string           `json:"kind"`
					Span *diagnostic.Span `json:"span"`
				} `json:"causes"`
				Repairs []struct {
					Kind string           `json:"kind"`
					Span *diagnostic.Span `json:"span"`
				} `json:"repairs"`
			} `json:"diagnostics"`
		}
		if err := json.Unmarshal(crossJSON, &nested); err != nil {
			t.Fatal(err)
		}
		if len(nested.Diagnostics) != 1 || nested.Diagnostics[0].ID != crossID {
			t.Fatalf("cross-unit presentation changed identity or count: %s", crossJSON)
		}
		item := nested.Diagnostics[0]
		if item.Primary.Start != entryStart || item.Primary.Line < 1 || item.Primary.Col < 1 || len(item.Causes) != 4 || item.Causes[0].Kind != "nil" || item.Causes[0].Span != nil || item.Causes[1].Kind != "imported" || item.Causes[1].Span == nil || item.Causes[1].Span.Start != cause.Start || item.Causes[1].Span.End != cause.End || item.Causes[1].Span.Line < 1 || item.Causes[2].Kind != "missing" || item.Causes[2].Span == nil || item.Causes[2].Span.Line != 0 || item.Causes[2].Span.Start != unmapped.Start || item.Causes[3].Kind != "invalid" || item.Causes[3].Span == nil || item.Causes[3].Span.Line != 0 || item.Causes[3].Span.Start != invalid.Start || len(item.Repairs) != 4 || item.Repairs[0].Kind != "imported" || item.Repairs[0].Span == nil || item.Repairs[0].Span.Start != repair.Start || item.Repairs[0].Span.Line < 1 || item.Repairs[1].Kind != "invalid" || item.Repairs[1].Span == nil || item.Repairs[1].Span.Line != 0 || item.Repairs[1].Span.Start != invalid.Start || item.Repairs[2].Kind != "missing" || item.Repairs[2].Span == nil || item.Repairs[2].Span.Line != 0 || item.Repairs[2].Span.Start != unmapped.Start || item.Repairs[3].Kind != "nil" || item.Repairs[3].Span != nil {
			t.Fatalf("nested effective units, invalid spans, or nil spans changed: %s", crossJSON)
		}
		primaryLine, primaryCol := phase34LocationAt(entryBytes, entryStart)
		causeLine, causeCol := phase34LocationAt(libBytes, cause.Start)
		repairLine, repairCol := phase34LocationAt(libBytes, repair.Start)
		if item.Primary.Line != primaryLine || item.Primary.Col != primaryCol || item.Causes[1].Span.Line != causeLine || item.Causes[1].Span.Col != causeCol || item.Repairs[0].Span.Line != repairLine || item.Repairs[0].Span.Col != repairCol {
			t.Fatalf("nested spans used wrong effective source coordinates: %s", crossJSON)
		}

		peer := session.PeerRefusalDiagnosticForTest(corevalidate.Result{Valid: false}, "app.main")
		peerResult := protocol.New("check", protocol.StatusInvalid)
		peerResult.Diagnostics = []diagnostic.Diagnostic{peer}
		peerResult = protocol.WithPresentation(peerResult, *origin.Presentation)
		peerJSON, err := protocol.JSON(peerResult)
		if err != nil {
			t.Fatal(err)
		}
		var peerWire struct {
			Diagnostics []struct {
				Primary diagnostic.Span `json:"primary_span"`
			} `json:"diagnostics"`
		}
		if err := json.Unmarshal(peerJSON, &peerWire); err != nil {
			t.Fatal(err)
		}
		peerHuman, err := protocol.Human(peerResult)
		if err != nil {
			t.Fatal(err)
		}
		if peerResult.Status != protocol.StatusInvalid || len(peerWire.Diagnostics) != 1 || peerWire.Diagnostics[0].Primary.Line != 0 || peerWire.Diagnostics[0].Primary.Col != 0 || strings.Contains(peerHuman, "app.main.schway:") {
			t.Fatalf("unmappable peer diagnostic should remain invalid and unlocated: %s\n%s", peerJSON, peerHuman)
		}
	})
}

func phase34LocationAt(source []byte, offset int) (line, column int) {
	if offset < 0 || offset > len(source) {
		return 0, 0
	}
	line, column = 1, 1
	for index := 0; index < offset; index++ {
		if source[index] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
