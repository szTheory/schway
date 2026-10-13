package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPhase34CheckDiagnosticPresentation(t *testing.T) {
	source := []byte("module draft.presentation\nexport { fn main }\nfn main(n: U64) -> U64 { n + }\n")
	path := filepath.Join(t.TempDir(), "broken.schway")
	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}

	textCode, textOutput, textError := captureLangRun(t, []string{"check", path})
	if textCode != 2 || textOutput != "" || textError == "" {
		t.Fatalf("text check code=%d stdout=%q stderr=%q, want invalid source on stderr", textCode, textOutput, textError)
	}
	textOutput = textError

	jsonCode, jsonOutput, jsonError := captureLangRun(t, []string{"--json", "check", path})
	if jsonCode != 2 || jsonError != "" {
		t.Fatalf("JSON check code=%d stderr=%q stdout=%q, want invalid source JSON on stdout", jsonCode, jsonError, jsonOutput)
	}
	var result struct {
		Status      string `json:"status"`
		Diagnostics []struct {
			ID      string `json:"id"`
			Code    string `json:"code"`
			Primary struct {
				Start int `json:"start"`
				End   int `json:"end"`
				Line  int `json:"line"`
				Col   int `json:"col"`
			} `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		t.Fatalf("decode JSON diagnostic: %v; output=%q", err, jsonOutput)
	}
	if result.Status != "invalid" || len(result.Diagnostics) == 0 {
		t.Fatalf("JSON status=%q diagnostics=%d, want a source error", result.Status, len(result.Diagnostics))
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.ID == "" || diagnostic.Code == "" || diagnostic.Primary.Start >= diagnostic.Primary.End {
		t.Fatalf("JSON diagnostic lost semantic identity or byte span: %+v", diagnostic)
	}
	if diagnostic.Primary.Line != 3 || diagnostic.Primary.Col < 1 {
		t.Fatalf("JSON source location=%d:%d, want 1-based line 3 and a positive scalar column", diagnostic.Primary.Line, diagnostic.Primary.Col)
	}
	if strings.Contains(jsonOutput, string(source)) || strings.Contains(jsonOutput, "1 + }") {
		t.Fatalf("JSON diagnostic leaked source bytes: %q", jsonOutput)
	}
	if !strings.Contains(textOutput, path+":3:"+strconv.Itoa(diagnostic.Primary.Col)) {
		t.Fatalf("text output lacks the matching file:line:col %s:3:%d: %q", path, diagnostic.Primary.Col, textOutput)
	}
	if !strings.Contains(textOutput, "n + }") || !strings.Contains(textOutput, "^") || !strings.Contains(textOutput, diagnostic.ID) || !strings.Contains(textOutput, diagnostic.Code) {
		t.Fatalf("text output lacks source snippet, caret, code, or stable ID: %q", textOutput)
	}
}
