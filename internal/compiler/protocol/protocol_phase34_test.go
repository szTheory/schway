package protocol_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func phase34Presented(t *testing.T, source []byte, problem diagnostic.Diagnostic) (string, json.RawMessage) {
	t.Helper()
	result := protocol.New("check", protocol.StatusInvalid)
	result.Diagnostics = []diagnostic.Diagnostic{problem}
	result = protocol.WithPresentation(result, protocol.PresentationContext{Sources: map[string]protocol.CapturedSource{"": {Label: "draft.swy", Bytes: source}}})
	encoded, err := protocol.JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Diagnostics []json.RawMessage `json:"diagnostics"`
	}
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	if len(wire.Diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1: %s", len(wire.Diagnostics), encoded)
	}
	return string(encoded), wire.Diagnostics[0]
}

func TestPhase34DiagnosticLocations(t *testing.T) {
	t.Run("exact coordinate goldens", func(t *testing.T) {
		cases := []struct {
			name   string
			source []byte
			span   diagnostic.Span
			origin bool
			want   string
		}{
			{"adjacent first", []byte("abcd"), diagnostic.Span{Start: 0, End: 1}, false, `{"start":0,"end":1,"line":1,"col":1}`},
			{"adjacent second", []byte("abcd"), diagnostic.Span{Start: 1, End: 2}, false, `{"start":1,"end":2,"line":1,"col":2}`},
			{"cross line", []byte("a\nbc"), diagnostic.Span{Start: 0, End: 4}, false, `{"start":0,"end":4,"line":1,"col":1}`},
			{"known zero start", []byte("abc"), diagnostic.Span{Start: 0, End: 0}, true, `{"start":0,"end":0,"line":1,"col":1}`},
			{"synthetic zero", []byte("abc"), diagnostic.Span{Start: 0, End: 0}, false, `{"start":0,"end":0}`},
			{"empty source EOF", []byte{}, diagnostic.Span{Start: 0, End: 0}, true, `{"start":0,"end":0,"line":1,"col":1}`},
			{"CR byte", []byte("a\r\nb"), diagnostic.Span{Start: 1, End: 1}, true, `{"start":1,"end":1,"line":1,"col":2}`},
			{"LF byte", []byte("a\r\nb"), diagnostic.Span{Start: 2, End: 2}, true, `{"start":2,"end":2,"line":1,"col":2}`},
			{"after CRLF", []byte("a\r\nb"), diagnostic.Span{Start: 3, End: 3}, true, `{"start":3,"end":3,"line":2,"col":1}`},
			{"before LF", []byte("a\nb"), diagnostic.Span{Start: 1, End: 1}, true, `{"start":1,"end":1,"line":1,"col":2}`},
			{"after LF", []byte("a\nb"), diagnostic.Span{Start: 2, End: 2}, true, `{"start":2,"end":2,"line":2,"col":1}`},
			{"tab is one scalar", []byte("\tX"), diagnostic.Span{Start: 1, End: 2}, false, `{"start":1,"end":2,"line":1,"col":2}`},
			{"lone CR is content", []byte("a\rb"), diagnostic.Span{Start: 2, End: 3}, false, `{"start":2,"end":3,"line":1,"col":3}`},
			{"inside UTF-8 scalar", []byte("éx"), diagnostic.Span{Start: 1, End: 2}, false, `{"start":1,"end":2,"line":1,"col":1}`},
			{"after UTF-8 scalar", []byte("éx"), diagnostic.Span{Start: 2, End: 3}, false, `{"start":2,"end":3,"line":1,"col":2}`},
			{"after invalid byte", []byte{0xff, 'x'}, diagnostic.Span{Start: 1, End: 2}, false, `{"start":1,"end":2,"line":1,"col":2}`},
			{"last line EOF", []byte("a\nb"), diagnostic.Span{Start: 3, End: 3}, true, `{"start":3,"end":3,"line":2,"col":2}`},
			{"final newline EOF", []byte("a\n"), diagnostic.Span{Start: 2, End: 2}, true, `{"start":2,"end":2,"line":2,"col":1}`},
			{"negative span refused", []byte("abc"), diagnostic.Span{Start: -1, End: 0}, true, `{"start":-1,"end":0}`},
			{"inverted span refused", []byte("abc"), diagnostic.Span{Start: 2, End: 1}, true, `{"start":2,"end":1}`},
			{"out of bounds refused", []byte("abc"), diagnostic.Span{Start: 2, End: 4}, true, `{"start":2,"end":4}`},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				problem := diagnostic.Error("test.location", tc.span, "location")
				if tc.origin {
					problem = problem.AtSource()
				}
				_, encoded := phase34Presented(t, tc.source, problem)
				var got struct {
					Primary json.RawMessage `json:"primary_span"`
				}
				if err := json.Unmarshal(encoded, &got); err != nil {
					t.Fatal(err)
				}
				if string(got.Primary) != tc.want {
					t.Fatalf("primary_span = %s, want exact wire golden %s", got.Primary, tc.want)
				}
			})
		}
	})

	t.Run("nested nil and effective units", func(t *testing.T) {
		entry := []byte("abcd")
		library := []byte("xy\nz")
		causeDefault := diagnostic.Span{Start: 1, End: 2}
		causeLibrary := diagnostic.Span{Start: 3, End: 4}
		repairDefault := diagnostic.Span{Start: 2, End: 3}
		repairLibrary := diagnostic.Span{Start: 0, End: 1}
		problem := diagnostic.ErrorWithRepairs("test.nested", diagnostic.Span{Start: 0, End: 1}, "nested",
			[]diagnostic.Cause{{Kind: "nil"}, {Kind: "default", Span: &causeDefault}, {Kind: "library", Unit: "lib.a", Span: &causeLibrary}},
			diagnostic.Repair{Kind: "default", Span: &repairDefault},
			diagnostic.Repair{Kind: "library", Unit: "lib.a", Span: &repairLibrary},
		).InUnit("main")
		result := protocol.New("check", protocol.StatusInvalid)
		result.Diagnostics = []diagnostic.Diagnostic{problem}
		result = protocol.WithPresentation(result, protocol.PresentationContext{Sources: map[string]protocol.CapturedSource{
			"main":  {Label: "main.swy", Bytes: entry},
			"lib.a": {Label: "lib/a.swy", Bytes: library},
		}})
		beforeID := result.Finalize().ID
		first, err := protocol.JSON(result)
		if err != nil {
			t.Fatal(err)
		}
		second, err := protocol.JSON(result)
		if err != nil {
			t.Fatal(err)
		}
		if string(first) != string(second) || result.Finalize().ID != beforeID {
			t.Fatalf("repeated enrichment changed bytes or result identity:\n%s\n%s", first, second)
		}
		if result.Diagnostics[0].Primary.Line != 0 || result.Diagnostics[0].Causes[1].Span.Line != 0 || result.Diagnostics[0].Repairs[0].Span.Line != 0 {
			t.Fatalf("JSON enrichment mutated semantic input: %+v", result.Diagnostics[0])
		}
		var wire struct {
			Diagnostics []struct {
				Primary json.RawMessage `json:"primary_span"`
				Causes  []struct {
					Span json.RawMessage `json:"span"`
				} `json:"causes"`
				Repairs []struct {
					Span json.RawMessage `json:"span"`
				} `json:"repairs"`
			} `json:"diagnostics"`
		}
		if err := json.Unmarshal(first, &wire); err != nil {
			t.Fatal(err)
		}
		problemWire := wire.Diagnostics[0]
		if string(problemWire.Primary) != `{"start":0,"end":1,"line":1,"col":1}` {
			t.Fatalf("primary used the wrong unit: %s", problemWire.Primary)
		}
		if len(problemWire.Causes) != 3 || len(problemWire.Causes[0].Span) != 0 {
			t.Fatalf("nil cause span was not preserved: %+v", problemWire.Causes)
		}
		if string(problemWire.Causes[1].Span) != `{"start":1,"end":2,"line":1,"col":2}` || string(problemWire.Causes[2].Span) != `{"start":3,"end":4,"line":2,"col":1}` {
			t.Fatalf("cause effective units or coordinates changed: %+v", problemWire.Causes)
		}
		if len(problemWire.Repairs) != 2 || string(problemWire.Repairs[0].Span) != `{"start":2,"end":3,"line":1,"col":3}` || string(problemWire.Repairs[1].Span) != `{"start":0,"end":1,"line":1,"col":1}` {
			t.Fatalf("repair effective units or order changed: %+v", problemWire.Repairs)
		}
		if strings.Contains(string(first), "main.swy") || strings.Contains(string(first), "lib/a.swy") || strings.Contains(string(first), "abcd") {
			t.Fatalf("presentation source leaked into JSON: %s", first)
		}
	})

	t.Run("parser marks real empty EOF and limits stay synthetic", func(t *testing.T) {
		parsed := syntax.Parse(nil)
		if len(parsed.Diagnostics) == 0 || !parsed.Diagnostics[0].SourcePosition {
			t.Fatalf("empty source EOF was not marked as a source position: %+v", parsed.Diagnostics)
		}
		_, encoded := phase34Presented(t, parsed.Tree.Source, parsed.Diagnostics[0])
		var wire struct {
			Primary json.RawMessage `json:"primary_span"`
		}
		if err := json.Unmarshal(encoded, &wire); err != nil {
			t.Fatal(err)
		}
		if string(wire.Primary) != `{"start":0,"end":0,"line":1,"col":1}` {
			t.Fatalf("empty source location = %s, want EOF 1:1", wire.Primary)
		}
		limited := syntax.Parse(make([]byte, syntax.MaxSourceBytes+1))
		if len(limited.Diagnostics) != 1 || limited.Diagnostics[0].SourcePosition {
			t.Fatalf("synthetic source limit acquired source provenance: %+v", limited.Diagnostics)
		}
	})

	t.Run("lexer source errors carry provenance", func(t *testing.T) {
		for _, source := range [][]byte{[]byte("0x"), []byte("\"open"), {0xff}, []byte("@")} {
			_, diagnostics := syntax.Lex(source)
			if len(diagnostics) == 0 || !diagnostics[0].SourcePosition {
				t.Errorf("lexer error for %q lacks source provenance: %+v", source, diagnostics)
			}
		}
	})

	t.Run("InUnit before and after presentation keeps identity", func(t *testing.T) {
		base := diagnostic.Error("test.unit", diagnostic.Span{Start: 0, End: 1}, "unit")
		before := base.InUnit("lib.a")
		result := protocol.New("check", protocol.StatusInvalid)
		result.Diagnostics = []diagnostic.Diagnostic{base}
		result = protocol.WithPresentation(result, protocol.PresentationContext{Sources: map[string]protocol.CapturedSource{"lib.a": {Label: "lib.swy", Bytes: []byte("x")}}})
		result.Diagnostics[0] = result.Diagnostics[0].InUnit("lib.a")
		for attempt := 0; attempt < 2; attempt++ {
			if _, err := protocol.JSON(result); err != nil {
				t.Fatal(err)
			}
			if result.Diagnostics[0].ID != before.ID || result.Finalize().Diagnostics[0].ID != before.ID {
				t.Fatalf("InUnit/enrichment attempt %d changed diagnostic identity", attempt)
			}
		}
	})
}

func TestPhase34ResultFinalizeIgnoresPresentation(t *testing.T) {
	result := protocol.New("check", protocol.StatusInvalid)
	result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("test.presentation", diagnostic.Span{Start: 1, End: 2}, "message")}
	plainID := result.Finalize().ID
	withSource := protocol.WithPresentation(result, protocol.PresentationContext{Sources: map[string]protocol.CapturedSource{"": {Label: "x", Bytes: []byte("abc")}}})
	if got := withSource.Finalize().ID; got != plainID {
		t.Fatalf("Result.Finalize included presentation source: got %s, want %s", got, plainID)
	}
	if fmt.Sprint(withSource.Diagnostics[0].ID) != result.Diagnostics[0].ID {
		t.Fatal("presentation changed diagnostic ID")
	}
}
