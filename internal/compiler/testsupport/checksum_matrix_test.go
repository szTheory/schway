package testsupport_test

import (
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase31ExpectedMatrixJSON pins examples/checksum/checksum.expected.json
// to testsupport.ChecksumCaseMatrix and the FileError alternative order (D-08).
func TestPhase31ExpectedMatrixJSON(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "checksum.expected.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Schema    string `json:"schema"`
		Limit     int    `json:"limit"`
		ErrorExit struct {
			Code int    `json:"code"`
			Name string `json:"name"`
		} `json:"error_exit"`
		Outcomes []string `json:"outcomes"`
		Cases    []struct {
			Name    string   `json:"name"`
			Kind    string   `json:"kind"`
			Exit    int      `json:"exit"`
			Stdout  string   `json:"stdout"`
			Stderr  string   `json:"stderr"`
			Outcome string   `json:"outcome"`
			Engines []string `json:"engines"`
			Witness string   `json:"witness"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Schema != "schway.checksum-expected/1" {
		t.Fatalf("schema = %q", doc.Schema)
	}
	if doc.Limit != testsupport.ChecksumLimit {
		t.Fatalf("limit = %d, want %d", doc.Limit, testsupport.ChecksumLimit)
	}
	if doc.ErrorExit.Code != exitcode.AppInputRefused || doc.ErrorExit.Name != "AppInputRefused" {
		t.Fatalf("error_exit = %+v, want {%d AppInputRefused}", doc.ErrorExit, exitcode.AppInputRefused)
	}
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "checksum.schway"))
	if err != nil {
		t.Fatal(err)
	}
	block := regexp.MustCompile(`(?s)data FileError =\n((?:  \| \w+\n)+)`).FindStringSubmatch(string(source))
	if block == nil {
		t.Fatal("checksum.schway has no FileError data declaration")
	}
	var alternatives []string
	for _, line := range strings.Split(strings.TrimSpace(block[1]), "\n") {
		alternatives = append(alternatives, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "|")))
	}
	if !reflect.DeepEqual(doc.Outcomes, alternatives) {
		t.Fatalf("outcomes = %v, want FileError alternatives %v", doc.Outcomes, alternatives)
	}
	byName := map[string]int{}
	for i, c := range doc.Cases {
		byName[c.Name] = i
	}
	for _, want := range testsupport.ChecksumCaseMatrix() {
		i, ok := byName[want.Name]
		if !ok {
			t.Fatalf("matrix row %q missing from the expected JSON", want.Name)
		}
		got := doc.Cases[i]
		if got.Kind != want.Kind || got.Exit != want.Exit || got.Stdout != want.Stdout || got.Stderr != want.Stderr || got.Outcome != want.Outcome {
			t.Fatalf("row %q = %+v, want %+v", want.Name, got, want)
		}
		if !reflect.DeepEqual(got.Engines, []string{"native", "interpreter"}) {
			t.Fatalf("row %q engines = %v", want.Name, got.Engines)
		}
	}
	for _, c := range doc.Cases {
		if !reflect.DeepEqual(c.Engines, []string{"native", "interpreter"}) {
			if !reflect.DeepEqual(c.Engines, []string{"native"}) || c.Witness != "TestPhase31ReleaseExactlyOnce" {
				t.Fatalf("native-only row %q needs engines [native] and its witness", c.Name)
			}
		}
	}
}
