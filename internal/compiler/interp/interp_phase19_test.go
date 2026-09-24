package interp

import "testing"

func TestPhase19ScalarProjection(t *testing.T) {
	tests := []struct {
		name  string
		value value
		want  string
	}{
		{name: "u64 zero", value: value{isU64: true}, want: "0"},
		{name: "u64 42", value: value{u64: 42, isU64: true}, want: "42"},
		{name: "u64 maximum", value: value{u64: ^uint64(0), isU64: true}, want: "18446744073709551615"},
		{name: "legacy scalar", value: value{payload: "01020304"}, want: "01020304"},
		{name: "tagged payload", value: value{tag: "some", payload: "42"}, want: "some"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
