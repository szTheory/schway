package interp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

type phase37ReaderInfo struct {
	mode os.FileMode
	size int64
}

func (phase37ReaderInfo) Name() string        { return "test" }
func (i phase37ReaderInfo) Size() int64       { return i.size }
func (i phase37ReaderInfo) Mode() os.FileMode { return i.mode }
func (phase37ReaderInfo) ModTime() time.Time  { return time.Time{} }
func (i phase37ReaderInfo) IsDir() bool       { return i.mode.IsDir() }
func (phase37ReaderInfo) Sys() any            { return nil }

type phase37ReaderDescriptor struct {
	data         []byte
	mode         os.FileMode
	reportedSize int64
	statErr      error
	readErr      error
	closeErr     error
	interrupts   int
	chunkLimit   int
	zeroReads    int
	offset       int
	readCalls    int
	closeCalls   int
	requests     []int
	onRead       func(int)
}

func (f *phase37ReaderDescriptor) Stat() (os.FileInfo, error) {
	if f.statErr != nil {
		return nil, f.statErr
	}
	return phase37ReaderInfo{mode: f.mode, size: f.reportedSize}, nil
}

func (f *phase37ReaderDescriptor) Read(p []byte) (int, error) {
	f.readCalls++
	f.requests = append(f.requests, len(p))
	if f.onRead != nil {
		f.onRead(len(p))
	}
	if f.interrupts > 0 {
		f.interrupts--
		return 0, syscall.EINTR
	}
	if f.zeroReads > 0 {
		f.zeroReads--
		return 0, nil
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

func TestPhase37FileReaderFaults(t *testing.T) {
	for _, row := range []struct {
		name      string
		file      *phase37ReaderDescriptor
		openErr   error
		want      string
		wantBytes string
		wantReads int
		wantClose int
	}{
		{name: "open-error", openErr: errors.New("open"), want: fileBytesFileUnavailable},
		{name: "stat-and-close-error", file: &phase37ReaderDescriptor{mode: 0o600, statErr: errors.New("stat"), closeErr: errors.New("close")}, want: fileBytesFileUnavailable, wantClose: 1},
		{name: "nonregular-and-close-error", file: &phase37ReaderDescriptor{mode: os.ModeDevice, closeErr: errors.New("close")}, want: fileBytesFileUnavailable, wantClose: 1},
		{name: "read-and-close-error", file: &phase37ReaderDescriptor{mode: 0o600, readErr: errors.New("read"), closeErr: errors.New("close")}, want: fileBytesFileUnavailable, wantReads: 1, wantClose: 1},
		{name: "zero-progress-is-not-eof", file: &phase37ReaderDescriptor{mode: 0o600, data: []byte("bounded"), zeroReads: 1}, want: fileBytesFileUnavailable, wantReads: 1, wantClose: 1},
		{name: "short-reads-and-eintr", file: &phase37ReaderDescriptor{mode: 0o600, data: []byte("bounded"), interrupts: 1, chunkLimit: 2}, wantBytes: "bounded", wantReads: 6, wantClose: 1},
		{name: "success-close-error", file: &phase37ReaderDescriptor{mode: 0o600, data: []byte("ok"), closeErr: errors.New("close")}, want: fileBytesFileUnavailable, wantReads: 2, wantClose: 1},
		{name: "oversize-remains-primary-over-close-error", file: &phase37ReaderDescriptor{mode: 0o600, data: bytes.Repeat([]byte{1}, FileBytesLimit+1), closeErr: errors.New("close")}, want: fileBytesFileTooLarge, wantReads: 1, wantClose: 1},
	} {
		t.Run(row.name, func(t *testing.T) {
			opens := 0
			got, alternative := readFileBytesWith("opaque", func(string) (fileBytesDescriptor, error) {
				opens++
				return row.file, row.openErr
			})
			if opens != 1 || alternative != row.want || string(got) != row.wantBytes {
				t.Fatalf("opened=%d result=(%d,%q,%q), want one open and (%q,%q)", opens, len(got), string(got), alternative, row.wantBytes, row.want)
			}
			reads, closes := 0, 0
			if row.file != nil {
				reads, closes = row.file.readCalls, row.file.closeCalls
			}
			if reads != row.wantReads || closes != row.wantClose {
				t.Fatalf("reads=%d closes=%d, want reads=%d closes=%d", reads, closes, row.wantReads, row.wantClose)
			}
		})
	}
}

func TestPhase37FileReaderBoundControls(t *testing.T) {
	for _, size := range []int{FileBytesLimit - 1, FileBytesLimit, FileBytesLimit + 1} {
		t.Run(fmt.Sprintf("boundary-%d", size), func(t *testing.T) {
			data := bytes.Repeat([]byte{0xa5}, size)
			file := &phase37ReaderDescriptor{mode: 0o600, reportedSize: int64(size), data: data}
			allocated := 0
			allocationAtRead := 0
			file.onRead = func(int) { allocationAtRead = allocated }
			opens := 0
			got, alternative := readFileBytesWithAllocator("opaque", func(string) (fileBytesDescriptor, error) {
				opens++
				return file, nil
			}, func(length int) []byte {
				allocated = length
				return make([]byte, length)
			})
			if size <= FileBytesLimit {
				if alternative != "" || !bytes.Equal(got, data) {
					t.Fatalf("size=%d result=(%d,%q), want exact bytes", size, len(got), alternative)
				}
			} else if got != nil || alternative != fileBytesFileTooLarge {
				t.Fatalf("size=%d result=(%d,%q), want FileTooLarge", size, len(got), alternative)
			}
			if opens != 1 || file.closeCalls != 1 || file.offset > FileBytesLimit+1 {
				t.Fatalf("opened=%d closes=%d bytes-read=%d, want one descriptor and at most %d bytes", opens, file.closeCalls, file.offset, FileBytesLimit+1)
			}
			if len(file.requests) == 0 || file.requests[0] != FileBytesLimit+1 {
				t.Fatalf("read requests=%v, want initial cap+1 request %d", file.requests, FileBytesLimit+1)
			}
			if allocated != FileBytesLimit+1 || allocationAtRead != FileBytesLimit+1 {
				t.Fatalf("allocation=%d at first read=%d, want exactly cap+1 before read", allocated, allocationAtRead)
			}
			for _, request := range file.requests {
				if request > FileBytesLimit+1 {
					t.Fatalf("reader requested %d bytes, exceeds cap+1", request)
				}
			}
		})
	}

	t.Run("forged-small-stat-cannot-bypass-read-ceiling", func(t *testing.T) {
		file := &phase37ReaderDescriptor{mode: 0o600, reportedSize: 1, data: bytes.Repeat([]byte{0xff}, FileBytesLimit+1)}
		got, alternative := readFileBytesWith("opaque", func(string) (fileBytesDescriptor, error) { return file, nil })
		if got != nil || alternative != fileBytesFileTooLarge || file.offset != FileBytesLimit+1 || file.closeCalls != 1 {
			t.Fatalf("forged size=%d result=(%d,%q) read=%d close=%d, want actual cap+1 refusal", file.reportedSize, len(got), alternative, file.offset, file.closeCalls)
		}
	})
	t.Run("forged-large-stat-does-not-change-short-read", func(t *testing.T) {
		file := &phase37ReaderDescriptor{mode: 0o600, reportedSize: 1 << 40, data: []byte("short")}
		got, alternative := readFileBytesWith("opaque", func(string) (fileBytesDescriptor, error) { return file, nil })
		if alternative != "" || string(got) != "short" || file.offset != len(got) || file.closeCalls != 1 {
			t.Fatalf("forged size=%d result=(%q,%q) read=%d close=%d, want actual short content", file.reportedSize, got, alternative, file.offset, file.closeCalls)
		}
	})

	t.Run("replacement-after-open-stays-on-original-descriptor", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "input.bin")
		if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
			t.Fatal(err)
		}
		opened := 0
		got, alternative := readFileBytesWith(path, func(name string) (fileBytesDescriptor, error) {
			opened++
			file, err := openFileBytesDescriptor(name)
			if err != nil {
				return nil, err
			}
			if err := os.Rename(name, name+".old"); err != nil {
				_ = file.Close()
				return nil, err
			}
			if err := os.WriteFile(name, []byte("replacement"), 0o600); err != nil {
				_ = file.Close()
				return nil, err
			}
			return file, nil
		})
		if opened != 1 || alternative != "" || string(got) != "original" {
			t.Fatalf("opened=%d result=(%q,%q), want one original descriptor", opened, got, alternative)
		}
	})

	t.Run("symlink-to-fifo-refused-without-blocking", func(t *testing.T) {
		dir := t.TempDir()
		fifo := filepath.Join(dir, "input.pipe")
		link := filepath.Join(dir, "input.link")
		if err := syscall.Mkfifo(fifo, 0o600); err != nil {
			t.Skipf("env:mkfifo unavailable on this host: %v", err)
		}
		if err := os.Symlink(fifo, link); err != nil {
			t.Fatal(err)
		}
		started := time.Now()
		got, alternative := ReadFileBytes(link)
		if elapsed := time.Since(started); elapsed > time.Second {
			t.Fatalf("symlink-to-FIFO inspection took %s; want nonblocking refusal", elapsed)
		}
		if got != nil || alternative != fileBytesFileUnavailable {
			t.Fatalf("symlink-to-FIFO result=(%v,%q), want FileUnavailable", got, alternative)
		}
	})
	t.Run("device-refused", func(t *testing.T) {
		got, alternative := ReadFileBytes("/dev/null")
		if got != nil || alternative != fileBytesFileUnavailable {
			t.Fatalf("device result=(%v,%q), want FileUnavailable", got, alternative)
		}
	})
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
