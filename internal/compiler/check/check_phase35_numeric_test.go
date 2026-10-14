package check_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase35NumericOwnershipDeltaFile struct {
	Schema              string                       `json:"schema"`
	DiagnosticOwnership phase35NumericOwnershipDelta `json:"diagnostic_ownership"`
}

type phase35NumericOwnershipDelta struct {
	Date    string                         `json:"date"`
	Literal string                         `json:"literal"`
	Reason  string                         `json:"reason"`
	Before  []phase35NumericOwnershipRoute `json:"before"`
	After   []phase35NumericOwnershipRoute `json:"after"`
}

type phase35NumericOwnershipRoute struct {
	Route           string                           `json:"route"`
	DiagnosticCount int                              `json:"diagnostic_count"`
	Diagnostics     []phase35NumericOwnershipProblem `json:"diagnostics"`
}

type phase35NumericOwnershipProblem struct {
	Code        string          `json:"code"`
	PrimarySpan diagnostic.Span `json:"primary_span"`
}

func TestPhase35NumericRoutePolicy(t *testing.T) {
	phase35AssertNumericOwnershipDelta(t)
	maximum := "18446744073709551615"
	longLeadingZeros := "0X" + strings.Repeat("0", 4096) + "1"
	cases := []struct {
		name      string
		route     string
		literal   string
		wantValue string
		wantCode  string
	}{
		{name: "legacy/decimal-leading-zero", route: "legacy", literal: "010", wantValue: "10"},
		{name: "legacy/binary-prefix-policy", route: "legacy", literal: "0B10", wantCode: "check.literal_out_of_range"},
		{name: "legacy/uppercase-hex", route: "legacy", literal: "0X1A", wantValue: "26"},
		{name: "legacy/mixed-case-hex", route: "legacy", literal: "0XfF", wantValue: "255"},
		{name: "legacy/hex-separators", route: "legacy", literal: "0XA_B", wantValue: "171"},
		{name: "legacy/maximum", route: "legacy", literal: "0XFFFFFFFFFFFFFFFF", wantValue: maximum},
		{name: "legacy/long-leading-zero-value", route: "legacy", literal: "0X" + strings.Repeat("0", 4096) + "1", wantValue: "1"},
		{name: "legacy/hex-overflow", route: "legacy", literal: "0X10000000000000000", wantCode: "check.literal_out_of_range"},
		{name: "legacy/decimal-overflow", route: "legacy", literal: "18446744073709551616", wantCode: "check.literal_out_of_range"},
		{name: "scalar/decimal-leading-zero", route: "scalar", literal: "010", wantValue: "8"},
		{name: "scalar/binary-prefix", route: "scalar", literal: "0B10", wantValue: "2"},
		{name: "scalar/uppercase-hex", route: "scalar", literal: "0X1A", wantValue: "26"},
		{name: "scalar/mixed-case-hex", route: "scalar", literal: "0XfF", wantValue: "255"},
		{name: "scalar/hex-separators", route: "scalar", literal: "0XA_B", wantValue: "171"},
		{name: "scalar/maximum", route: "scalar", literal: "0XFFFFFFFFFFFFFFFF", wantValue: maximum},
		{name: "scalar/long-leading-zero-value", route: "scalar", literal: longLeadingZeros, wantValue: "1"},
		{name: "scalar/hex-overflow", route: "scalar", literal: "0X10000000000000000", wantCode: "type.invalid_u64"},
		{name: "scalar/decimal-overflow", route: "scalar", literal: "18446744073709551616", wantCode: "type.invalid_u64"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := phase35NumericRouteSource(tc.route, tc.literal)
			checked := session.Check(source)
			if tc.wantCode != "" {
				if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != tc.wantCode {
					t.Fatalf("%s %q diagnostics=%+v, want exactly %s", tc.route, tc.literal, checked.Diagnostics, tc.wantCode)
				}
				problem := checked.Diagnostics[0]
				if string(source[problem.Primary.Start:problem.Primary.End]) != tc.literal {
					t.Fatalf("%s %q range span %+v covers %q", tc.route, tc.literal, problem.Primary, source[problem.Primary.Start:problem.Primary.End])
				}
				return
			}
			if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 || checked.Program.Functions[0].Linear == nil {
				t.Fatalf("%s %q was refused: diagnostics=%+v program=%+v", tc.route, tc.literal, checked.Diagnostics, checked.Program)
			}
			constants := numericConstOperations(checked.Program)
			if len(constants) != 1 || constants[0] != tc.wantValue {
				t.Fatalf("%s %q constants=%v, want %q", tc.route, tc.literal, constants, tc.wantValue)
			}
		})
	}

	for _, route := range []string{"legacy", "scalar"} {
		for _, literal := range []string{"0x", "0X", "0x1G", "0b2", "0x_FF", "0xFF_", "0xFF__0", "0xFFu64", "0xFFλ"} {
			t.Run(route+"/malformed/"+literal, func(t *testing.T) {
				source := phase35NumericRouteSource(route, literal)
				checked := session.Check(source)
				if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "syntax.malformed_numeric_literal" {
					t.Fatalf("%s malformed %q produced duplicate or unrelated causes: %+v", route, literal, checked.Diagnostics)
				}
				problem := checked.Diagnostics[0]
				if string(source[problem.Primary.Start:problem.Primary.End]) != literal || problem.Primary.End-problem.Primary.Start != len(literal) {
					t.Fatalf("%s malformed %q diagnostic span %+v does not cover the full candidate", route, literal, problem.Primary)
				}
			})
		}
	}
}

func phase35AssertNumericOwnershipDelta(t *testing.T) {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "numeric-delta.json"))
	if err != nil {
		t.Fatalf("read numeric diagnostic delta: %v", err)
	}
	var file phase35NumericOwnershipDeltaFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("decode numeric diagnostic delta: %v", err)
	}
	want := phase35NumericOwnershipDelta{
		Date:    "2026-10-14",
		Literal: "0x_FF",
		Reason:  "The lexer owns the complete malformed numeric candidate. The parser expectation at that exact TokenUnknown span is a duplicate and is suppressed; expectations at other spans and for unrelated unknown tokens remain visible.",
		Before: []phase35NumericOwnershipRoute{
			{Route: "legacy", DiagnosticCount: 2, Diagnostics: []phase35NumericOwnershipProblem{{Code: "syntax.malformed_numeric_literal", PrimarySpan: diagnostic.Span{Start: 93, End: 98}}, {Code: "syntax.expected_binding_source", PrimarySpan: diagnostic.Span{Start: 93, End: 98}}}},
			{Route: "scalar", DiagnosticCount: 2, Diagnostics: []phase35NumericOwnershipProblem{{Code: "syntax.malformed_numeric_literal", PrimarySpan: diagnostic.Span{Start: 92, End: 97}}, {Code: "syntax.expected_binding_source", PrimarySpan: diagnostic.Span{Start: 92, End: 97}}}},
		},
		After: []phase35NumericOwnershipRoute{
			{Route: "legacy", DiagnosticCount: 1, Diagnostics: []phase35NumericOwnershipProblem{{Code: "syntax.malformed_numeric_literal", PrimarySpan: diagnostic.Span{Start: 93, End: 98}}}},
			{Route: "scalar", DiagnosticCount: 1, Diagnostics: []phase35NumericOwnershipProblem{{Code: "syntax.malformed_numeric_literal", PrimarySpan: diagnostic.Span{Start: 92, End: 97}}}},
		},
	}
	if file.Schema != "phase35.numeric-delta/1" || !reflect.DeepEqual(file.DiagnosticOwnership, want) {
		t.Fatalf("literal numeric diagnostic ownership delta changed: %+v", file.DiagnosticOwnership)
	}
	for _, row := range want.After {
		checked := session.Check(phase35NumericRouteSource(row.Route, file.DiagnosticOwnership.Literal))
		actual := make([]phase35NumericOwnershipProblem, 0, len(checked.Diagnostics))
		for _, problem := range checked.Diagnostics {
			actual = append(actual, phase35NumericOwnershipProblem{Code: problem.Code, PrimarySpan: problem.Primary})
		}
		if len(actual) != row.DiagnosticCount || !reflect.DeepEqual(actual, row.Diagnostics) {
			t.Errorf("%s current malformed numeric projection=%+v, want literal after-state %+v", row.Route, actual, row)
		}
	}
}

func phase35NumericRouteSource(route, literal string) []byte {
	parameter := "Byte"
	if route == "scalar" {
		parameter = "U64"
	}
	return []byte(fmt.Sprintf("module phase35.numeric.%s\nexport { fn main }\nfn main(input: %s) -> U64 {\n  let value = %s\n  value\n}\n", route, parameter, literal))
}

func numericConstOperations(program core.Program) []string {
	var values []string
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpConst {
				values = append(values, operation.ConstU64)
			}
		}
	}
	return values
}
