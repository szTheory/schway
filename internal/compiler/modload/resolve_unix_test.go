//go:build unix

package modload

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestPhase30FIFORefusedWithoutBlocking covers D-10: the final component must be
// a regular file before it is opened.
func TestPhase30FIFORefusedWithoutBlocking(t *testing.T) {
	rootDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rootDir, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	pipe := filepath.Join(rootDir, "a", "pipe.schway")
	if err := syscall.Mkfifo(pipe, 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	r := gateOpen(t, rootDir)

	run := func() chan string {
		done := make(chan string, 1)
		go func() {
			_, code := resolveCode(t, r, "a.pipe")
			done <- code
		}()
		return done
	}

	select {
	case code := <-run():
		if code != CodeNotRegular {
			t.Fatalf("FIFO: code %q, want %q", code, CodeNotRegular)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FIFO read blocked past the 2 s deadline")
	}

	t.Run("control open before Lstat", func(t *testing.T) {
		openBeforeLstatForTest = true
		done := run()
		select {
		case code := <-done:
			openBeforeLstatForTest = false
			t.Fatalf("control returned %q within the deadline; it must block", code)
		case <-time.After(2 * time.Second):
		}
		writer, err := os.OpenFile(pipe, os.O_WRONLY, 0)
		if err == nil {
			defer writer.Close()
		}
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("blocked reader did not release")
		}
		openBeforeLstatForTest = false
	})
}

// makeFIFO creates root/pipe.schway and reports whether it could.
func makeFIFO(t *testing.T, root string) bool {
	t.Helper()
	return syscall.Mkfifo(filepath.Join(root, "pipe.schway"), 0o644) == nil
}
