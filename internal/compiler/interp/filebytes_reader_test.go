package interp

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

type phase37ReaderInfo struct{ mode os.FileMode }

func (phase37ReaderInfo) Name() string        { return "test" }
func (phase37ReaderInfo) Size() int64         { return 0 }
func (i phase37ReaderInfo) Mode() os.FileMode { return i.mode }
func (phase37ReaderInfo) ModTime() time.Time  { return time.Time{} }
func (i phase37ReaderInfo) IsDir() bool       { return i.mode.IsDir() }
func (phase37ReaderInfo) Sys() any            { return nil }

type phase37ReaderDescriptor struct {
	data       []byte
	mode       os.FileMode
	statErr    error
	readErr    error
	closeErr   error
	interrupts int
	chunkLimit int
	offset     int
	readCalls  int
	closeCalls int
}

func (f *phase37ReaderDescriptor) Stat() (os.FileInfo, error) {
	if f.statErr != nil {
		return nil, f.statErr
	}
	return phase37ReaderInfo{mode: f.mode}, nil
}

func (f *phase37ReaderDescriptor) Read(p []byte) (int, error) {
	f.readCalls++
	if f.interrupts > 0 {
		f.interrupts--
		return 0, syscall.EINTR
	}
	if f.readErr != nil {
		return 0, f.readErr
	}
	if f.offset == len(f.data) {
		return 0, io.EOF
	}
	count := len(p)
	if f.chunkLimit > 0 && count > f.chunkLimit {
		count = f.chunkLimit
	}
	if count > len(f.data)-f.offset {
		count = len(f.data) - f.offset
	}
	copy(p, f.data[f.offset:f.offset+count])
	f.offset += count
	return count, nil
}

func (f *phase37ReaderDescriptor) Close() error {
	f.closeCalls++
	return f.closeErr
}

func TestPhase37FileReaderBasic(t *testing.T) {
	t.Run("regular-file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "input.bin")
		want := []byte{0x00, 0xff, 'x'}
		if err := os.WriteFile(path, want, 0o600); err != nil {
			t.Fatal(err)
		}
		got, alternative := ReadFileBytes(path)
		if alternative != "" || !bytes.Equal(got, want) {
			t.Fatalf("ReadFileBytes = (%x, %q), want (%x, empty)", got, alternative, want)
		}
	})
	t.Run("symlink-follows-target", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "target.bin")
		link := filepath.Join(dir, "input.bin")
		if err := os.WriteFile(target, []byte("linked"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		got, alternative := ReadFileBytes(link)
		if alternative != "" || string(got) != "linked" {
			t.Fatalf("ReadFileBytes(symlink) = (%q, %q), want linked bytes", got, alternative)
		}
	})
	t.Run("missing-path", func(t *testing.T) {
		got, alternative := ReadFileBytes(filepath.Join(t.TempDir(), "missing"))
		if got != nil || alternative != fileBytesFileUnavailable {
			t.Fatalf("missing path = (%v, %q), want FileUnavailable", got, alternative)
		}
	})
	t.Run("directory-refused", func(t *testing.T) {
		got, alternative := ReadFileBytes(t.TempDir())
		if got != nil || alternative != fileBytesFileUnavailable {
			t.Fatalf("directory = (%v, %q), want FileUnavailable", got, alternative)
		}
	})
	t.Run("fifo-refused-with-nonblocking-open", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "input.pipe")
		if err := syscall.Mkfifo(path, 0o600); err != nil {
			t.Fatal(err)
		}
		started := time.Now()
		got, alternative := ReadFileBytes(path)
		if elapsed := time.Since(started); elapsed > time.Second {
			t.Fatalf("FIFO inspection took %s; want nonblocking refusal", elapsed)
		}
		if got != nil || alternative != fileBytesFileUnavailable {
			t.Fatalf("FIFO = (%v, %q), want FileUnavailable", got, alternative)
		}
	})
	t.Run("content-cap", func(t *testing.T) {
		dir := t.TempDir()
		within := filepath.Join(dir, "within.bin")
		oversize := filepath.Join(dir, "oversize.bin")
		if err := os.WriteFile(within, bytes.Repeat([]byte{0xff}, FileBytesLimit), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(oversize, bytes.Repeat([]byte{0xff}, FileBytesLimit+1), 0o600); err != nil {
			t.Fatal(err)
		}
		got, alternative := ReadFileBytes(within)
		if alternative != "" || len(got) != FileBytesLimit {
			t.Fatalf("exact-cap read = len %d, alternative %q", len(got), alternative)
		}
		got, alternative = ReadFileBytes(oversize)
		if got != nil || alternative != fileBytesFileTooLarge {
			t.Fatalf("cap+1 read = (%v, %q), want FileTooLarge", got, alternative)
		}
	})
	t.Run("descriptor-faults-and-cleanup", func(t *testing.T) {
		for _, row := range []struct {
			name      string
			file      *phase37ReaderDescriptor
			want      string
			wantReads int
		}{
			{name: "stat-error", file: &phase37ReaderDescriptor{mode: 0o600, statErr: errors.New("stat")}, want: fileBytesFileUnavailable, wantReads: 0},
			{name: "nonregular", file: &phase37ReaderDescriptor{mode: os.ModeNamedPipe}, want: fileBytesFileUnavailable, wantReads: 0},
			{name: "read-error", file: &phase37ReaderDescriptor{mode: 0o600, readErr: errors.New("read")}, want: fileBytesFileUnavailable, wantReads: 1},
			{name: "close-error", file: &phase37ReaderDescriptor{mode: 0o600, data: []byte("ok"), closeErr: errors.New("close")}, want: fileBytesFileUnavailable, wantReads: 2},
			{name: "retry-eintr-and-short-reads", file: &phase37ReaderDescriptor{mode: 0o600, data: []byte("bounded"), interrupts: 1, chunkLimit: 2}, want: "", wantReads: 6},
			{name: "too-large-is-primary-over-close-error", file: &phase37ReaderDescriptor{mode: 0o600, data: bytes.Repeat([]byte{1}, FileBytesLimit+1), closeErr: errors.New("close")}, want: fileBytesFileTooLarge, wantReads: 1},
		} {
			t.Run(row.name, func(t *testing.T) {
				opens := 0
				got, alternative := readFileBytesWith("opaque", func(string) (fileBytesDescriptor, error) {
					opens++
					return row.file, nil
				})
				if opens != 1 || alternative != row.want || row.file.closeCalls != 1 || row.file.readCalls != row.wantReads {
					t.Fatalf("opened=%d reads=%d closes=%d result=(%d,%q), want one open/close, %d reads, %q", opens, row.file.readCalls, row.file.closeCalls, len(got), alternative, row.wantReads, row.want)
				}
				if row.want == "" && string(got) != "bounded" {
					t.Fatalf("read bytes=%q, want bounded", got)
				}
			})
		}
	})
	t.Run("invalid-paths-do-not-open", func(t *testing.T) {
		for _, path := range []string{"", "bad\x00path", strings.Repeat("p", maxFileBytesPathBytes+1)} {
			opens := 0
			got, alternative := readFileBytesWith(path, func(string) (fileBytesDescriptor, error) {
				opens++
				return nil, errors.New("must not open")
			})
			if opens != 0 || got != nil || alternative != fileBytesFileUnavailable {
				t.Fatalf("path length %d opened=%d result=(%v,%q), want pre-open FileUnavailable", len(path), opens, got, alternative)
			}
		}
	})
}
