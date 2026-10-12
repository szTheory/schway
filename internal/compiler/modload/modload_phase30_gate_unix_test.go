//go:build unix

package modload

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type gateResult struct {
	code string
}

// TestPhase30GateFIFONonBlocking is the FIFO half of flag (a).
func TestPhase30GateFIFONonBlocking(t *testing.T) {
	rootDir := t.TempDir()
	pipe := filepath.Join(rootDir, "pipe.schway")
	if err := syscall.Mkfifo(pipe, 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	r := gateOpen(t, rootDir)

	run := func() chan gateResult {
		done := make(chan gateResult, 1)
		go func() {
			_, code := gateRead(r, []string{"pipe"})
			done <- gateResult{code}
		}()
		return done
	}

	select {
	case got := <-run():
		if got.code != CodeNotRegular {
			t.Fatalf("FIFO: code %q, want %q", got.code, CodeNotRegular)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FIFO read blocked past the 2 s deadline")
	}

	t.Run("control open before Lstat", func(t *testing.T) {
		openBeforeLstatForTest = true
		done := run()
		select {
		case got := <-done:
			openBeforeLstatForTest = false
			t.Fatalf("control returned %q within the deadline; it must block", got.code)
		case <-time.After(2 * time.Second):
		}
		// Release the blocked open by opening the write side.
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
