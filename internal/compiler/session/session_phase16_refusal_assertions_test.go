package session_test

import (
	"strings"
	"testing"
)

func requirePhase16M004Refusal(t testing.TB, err error, family string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected terminal Phase 16 M004 %s refusal, got nil", family)
	}
	want := map[string]string{
		"foreign":    "multi-function foreign-call bodies are not supported by native emission this phase",
		"by-pointer": "by-pointer bodies are not supported by whole-program native emission this phase",
	}[family]
	if want == "" {
		t.Fatalf("unknown M004 refusal family %q", family)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected terminal Phase 16 M004 %s refusal containing %q, got %v", family, want, err)
	}
}

func requirePhase16AnyM004Refusal(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected terminal Phase 16 M004 refusal, got nil")
	}
	message := err.Error()
	if !strings.Contains(message, "multi-function foreign-call bodies are not supported by native emission this phase") &&
		!strings.Contains(message, "by-pointer bodies are not supported by whole-program native emission this phase") {
		t.Fatalf("expected terminal Phase 16 M004 refusal, got %v", err)
	}
}
