package session_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

// Phase 28 plan 03 (EVD-13): app verify keeps the named capacity reason at any
// n by reporting an outcome digest instead of embedded interpreter evidence.

func phase28RequireClang(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("env:clang toolchain unavailable")
	}
}

func phase28SumToN(n int) string { return strconv.Itoa(n * (n + 1) / 2) }

// phase28VerifyFile drives the file entry point so the written report is part
// of the evidence.
func phase28VerifyFile(t *testing.T, cases session.ReplayCases) (session.ReplayReport, string) {
	t.Helper()
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sum_to_n.schway")
	if err := os.WriteFile(sourcePath, phase273SumToNSource(t), 0o600); err != nil {
		t.Fatal(err)
	}
	encodedCases, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	casesPath := filepath.Join(dir, "cases.json")
	if err := os.WriteFile(casesPath, encodedCases, 0o600); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(dir, "report.json")
	report, err := session.VerifyApplicationCasesFile(context.Background(), sourcePath, casesPath, reportPath, native.DefaultRunner())
	if err != nil {
		t.Fatalf("exhausted cases must be per-case results, got error: %v", err)
	}
	return report, reportPath
}

func phase28RequireWrittenReport(t *testing.T, reportPath string, wantCases int) {
	t.Helper()
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("report file not written: %v", err)
	}
	if len(data) > session.MaxReplayCasesBytes {
		t.Fatalf("report is %d bytes, over %d", len(data), session.MaxReplayCasesBytes)
	}
	var decoded session.ReplayReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Schema != "lang.app-verification/2" || len(decoded.Cases) != wantCases || decoded.Verified {
		t.Fatalf("written report schema=%q cases=%d verified=%v", decoded.Schema, len(decoded.Cases), decoded.Verified)
	}
}

func TestAppVerifyExhaustedCases(t *testing.T) {
	phase28RequireClang(t)
	for _, n := range []int{173} {
		t.Run("n="+strconv.Itoa(n), func(t *testing.T) {
			value := phase28SumToN(n)
			report, reportPath := phase28VerifyFile(t, phase273Cases(phase273Case{"exhausted", strconv.Itoa(n), value}))
			if report.Verified || len(report.Cases) != 1 {
				t.Fatalf("verified=%v cases=%d", report.Verified, len(report.Cases))
			}
			got := report.Cases[0]
			if got.Status != session.ReplayStatusUnsupported || !strings.HasPrefix(got.Diagnostic, "native.event_capacity_exhausted") {
				t.Fatalf("status=%q diagnostic=%q", got.Status, got.Diagnostic)
			}
			phase273RequireDigest(t, got, value, "-O0")
			phase28RequireWrittenReport(t, reportPath, 1)
		})
	}
}

// phase28FillWorstCase fills every string and int field of value by reflection
// to its worst case and returns how many fields it visited. Maps stay nil.
func phase28FillWorstCase(t *testing.T, value reflect.Value, id string) int {
	t.Helper()
	visited := 0
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		name := value.Type().Field(index).Name
		switch field.Kind() {
		case reflect.String:
			switch {
			case name == "ID":
				field.SetString(id)
			case name == "Diagnostic":
				longest := session.EventCapacityDiagnosticForTest("-O0")
				if alt := "interpreter produced " + strings.Repeat("k", 20) + ":" + strings.Repeat("9", 20) + ", expected " + strings.Repeat("k", 20) + ":" + strings.Repeat("9", 20); len(alt) > len(longest) {
					longest = alt
				}
				field.SetString(longest)
			case strings.Contains(name, "SHA256") || strings.HasSuffix(name, "Digest") || name == "InputID":
				field.SetString(strings.Repeat("f", 64))
			case name == "Kind" && strings.HasPrefix(value.Type().Name(), "ReplayExpected"), name == "OutcomeKind":
				field.SetString("returned") // closed vocabulary: execution.OutcomeReturned
			case name == "Kind":
				field.SetString(session.ReplayCaseVerifierModel) // longest closed case kind
			case name == "Status":
				field.SetString(session.ReplayStatusUnsupported) // longest closed status
			case name == "EvidenceScope":
				field.SetString(session.ReplayEvidenceSource) // longest closed scope
			case name == "Engine":
				field.SetString("interpreter")
			case name == "ExhaustedLevel":
				field.SetString("-O0")
			case name == "Value" || name == "OutcomeValue":
				field.SetString(strings.Repeat("9", 20))
			default:
				field.SetString(strings.Repeat("x", 128))
			}
			visited++
		case reflect.Int, reflect.Int64:
			field.SetInt(math.MaxInt64)
			visited++
		case reflect.Bool:
			field.SetBool(true)
			visited++
		case reflect.Map, reflect.Slice:
			visited++ // maps stay nil (exhausted cases omit engines); the caller fills Cases
		case reflect.Ptr:
			if field.Type().Elem().Kind() != reflect.Struct {
				t.Fatalf("field %s: unhandled pointer kind", name)
			}
			field.Set(reflect.New(field.Type().Elem()))
			phase28FillWorstCase(t, field.Elem(), id)
			visited++
		case reflect.Struct:
			phase28FillWorstCase(t, field, id)
			visited++
		default:
			t.Fatalf("field %s has kind %s with no worst-case rule", name, field.Kind())
		}
	}
	if visited != value.NumField() {
		t.Fatalf("%s: visited %d of %d fields", value.Type(), visited, value.NumField())
	}
	return visited
}

func phase28ID(i int) string {
	return fmt.Sprintf("%03d", i) + strings.Repeat("i", 125)
}

// TestAppVerifyReportBudget proves the static D-16 budget: sixteen maximal
// exhausted cases encode to at most half of MaxReplayCasesBytes, and a new
// field without a worst-case rule fails the filler.
func TestAppVerifyReportBudget(t *testing.T) {
	report := session.ReplayReport{}
	phase28FillWorstCase(t, reflect.ValueOf(&report).Elem(), "")
	report.Cases = nil
	for index := 0; index < session.MaxReplayCaseCount; index++ {
		var caseReport session.ReplayCaseReport
		phase28FillWorstCase(t, reflect.ValueOf(&caseReport).Elem(), phase28ID(index))
		if caseReport.Engines != nil {
			t.Fatal("worst case must omit engines")
		}
		report.Cases = append(report.Cases, caseReport)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	size := len(encoded) + 1
	t.Logf("sixteen-case worst-case report encodes to %d bytes (budget %d)", size, session.MaxReplayCasesBytes/2)
	if size > session.MaxReplayCasesBytes/2 {
		t.Fatalf("worst-case report is %d bytes, over the %d byte budget", size, session.MaxReplayCasesBytes/2)
	}
}

// TestAppVerifyReportOversize hits the runtime defense deliberately: exactly at
// the limit writes, one byte over returns the named error and leaves no file.
func TestAppVerifyReportOversize(t *testing.T) {
	build := func(extra int) session.ReplayReport {
		report := session.ReplayReport{Schema: session.ApplicationVerifySchema}
		base, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		// Grow SourceDigest so len(encoded)+1 == MaxReplayCasesBytes+extra.
		report.SourceDigest = strings.Repeat("a", session.MaxReplayCasesBytes+extra-1-len(base))
		return report
	}
	atLimit := t.TempDir()
	if err := session.WriteApplicationVerificationReportForTest(filepath.Join(atLimit, "report.json"), build(0)); err != nil {
		t.Fatalf("report exactly at the limit refused: %v", err)
	}
	info, err := os.Stat(filepath.Join(atLimit, "report.json"))
	if err != nil || info.Size() != session.MaxReplayCasesBytes {
		t.Fatalf("at-limit report size=%v err=%v, want %d", info, err, session.MaxReplayCasesBytes)
	}
	over := t.TempDir()
	err = session.WriteApplicationVerificationReportForTest(filepath.Join(over, "report.json"), build(1))
	if !errors.Is(err, session.ErrApplicationVerificationReportOversize) {
		t.Fatalf("over-limit error=%v, want ErrApplicationVerificationReportOversize", err)
	}
	if !strings.Contains(err.Error(), "65537") || !strings.Contains(err.Error(), "65536") {
		t.Fatalf("error %q must name size 65537 and limit 65536", err)
	}
	entries, readErr := os.ReadDir(over)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("oversize left files behind: %v err=%v", entries, readErr)
	}
}

// TestAppVerifyOutcomeDigest re-marshals and re-hashes the omitted interpreter
// execution (the D-13 preimage) and compares it with the reported digest.
func TestAppVerifyOutcomeDigest(t *testing.T) {
	phase28RequireClang(t)
	source := phase273SumToNSource(t)
	report, err := session.VerifyApplicationCases(context.Background(), source,
		phase273Cases(phase273Case{"n15", "15", "120"}), native.DefaultRunner())
	if err != nil {
		t.Fatal(err)
	}
	digest := report.Cases[0].OutcomeDigest
	if digest == nil {
		t.Fatalf("no digest on %+v", report.Cases[0])
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatal(checked.Diagnostics)
	}
	raw, err := interp.Run(checked.Program, "main", "15")
	if err != nil {
		t.Fatal(err)
	}
	projected, err := session.ProjectExecutionSchema2(checked.Program, raw)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(projected)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(encoded)
	if digest.EvidenceSHA256 != hex.EncodeToString(sum[:]) || digest.EvidenceBytes != len(encoded) || digest.EventCount != len(projected.Events) {
		t.Fatalf("digest=%+v, recomputed sha=%x bytes=%d events=%d", digest, sum, len(encoded), len(projected.Events))
	}
}
