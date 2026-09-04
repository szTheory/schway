package testsupport

import "testing"

// TestCLIStreamCeilingBoundary is D-02-06's boundary falsifier: writing
// exactly MaxCLIStreamBytes must not overflow, and writing one byte more
// must, proving the tightened ceiling (D-02-06) is exact rather than merely
// "roughly right."
func TestCLIStreamCeilingBoundary(t *testing.T) {
	t.Run("exact", func(t *testing.T) {
		var writer boundedWriter
		data := make([]byte, MaxCLIStreamBytes)
		if _, err := writer.Write(data); err != nil {
			t.Fatal(err)
		}
		if writer.overflowed() {
			t.Fatalf("exactly MaxCLIStreamBytes reported overflow: total=%d", writer.total)
		}
		if len(writer.bytes()) != MaxCLIStreamBytes {
			t.Fatalf("captured %d bytes, want %d", len(writer.bytes()), MaxCLIStreamBytes)
		}
	})
	t.Run("one over", func(t *testing.T) {
		var writer boundedWriter
		data := make([]byte, MaxCLIStreamBytes+1)
		if _, err := writer.Write(data); err != nil {
			t.Fatal(err)
		}
		if !writer.overflowed() {
			t.Fatalf("MaxCLIStreamBytes+1 did not report overflow: total=%d", writer.total)
		}
	})
	t.Run("split writes across the boundary", func(t *testing.T) {
		var writer boundedWriter
		if _, err := writer.Write(make([]byte, MaxCLIStreamBytes)); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write([]byte{'x'}); err != nil {
			t.Fatal(err)
		}
		if !writer.overflowed() {
			t.Fatal("split writes crossing the boundary did not report overflow")
		}
	})
}
