package execution_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/execution"
)

func TestInvocationGrammar(t *testing.T) {
	segments := []execution.InvocationSegment{{OpCallID: "op:call/%#\u00e9", Ordinal: 0}, {OpCallID: "next", Ordinal: 12}}
	got, err := execution.FormatInvocation("entry:%\u00e9", segments)
	if err != nil {
		t.Fatal(err)
	}
	const want = "inv:entry:entry:%25%C3%A9/op:call%2F%25%23%C3%A9#0/next#12"
	if got != want {
		t.Fatalf("FormatInvocation() = %q, want %q", got, want)
	}
	parsed, err := execution.ParseInvocation(got)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.EntryID != "entry:%\u00e9" || len(parsed.Segments) != 2 || parsed.Segments[0] != segments[0] || parsed.Segments[1] != segments[1] {
		t.Fatalf("ParseInvocation() = %#v, want entry and segments %#v", parsed, segments)
	}
	for _, invalid := range []string{
		"", "inv:entry:", "inv:entry:entry/op", "inv:entry:entry/op#", "inv:entry:entry/op#00",
		"inv:entry:entry/op#-1", "inv:entry:entry%2fchild#0", "inv:entry:entry%41", "inv:entry:entry/%zz#0",
		"inv:entry:entry/%C3#0", "inv:entry:entry//#0", "inv:entry:entry/%2F#0/next#",
	} {
		if _, err := execution.ParseInvocation(invalid); err == nil {
			t.Fatalf("ParseInvocation(%q) unexpectedly succeeded", invalid)
		}
	}
}
