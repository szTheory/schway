package interp

import (
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
)

const (
	FileBytesLimit           = 65536
	maxFileBytesPathBytes    = 4096
	fileBytesFileTooLarge    = "FileTooLarge"
	fileBytesFileUnavailable = "FileUnavailable"
)

type fileBytesDescriptor interface {
	Stat() (os.FileInfo, error)
	Read([]byte) (int, error)
	Close() error
}

type fileBytesOpenFunc func(string) (fileBytesDescriptor, error)
type fileBytesAllocateFunc func(int) []byte

func openFileBytesDescriptor(path string) (fileBytesDescriptor, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_CLOEXEC|syscall.O_NOCTTY, 0)
}

// ReadFileBytes is the fixed interpreter host adapter for the admitted
// FileBytes capability. It follows symlinks, opens before checking the
// descriptor, and reads at most one byte beyond the content limit.
func ReadFileBytes(path string) ([]byte, string) {
	return readFileBytesWith(path, openFileBytesDescriptor)
}

func readFileBytesWith(path string, open fileBytesOpenFunc) ([]byte, string) {
	return readFileBytesWithAllocator(path, open, func(size int) []byte { return make([]byte, size) })
}

func readFileBytesWithAllocator(path string, open fileBytesOpenFunc, allocate fileBytesAllocateFunc) ([]byte, string) {
	if open == nil || allocate == nil || path == "" || strings.IndexByte(path, 0) >= 0 || len(path) > maxFileBytesPathBytes {
		return nil, fileBytesFileUnavailable
	}
	file, err := open(path)
	if err != nil {
		return nil, fileBytesFileUnavailable
	}
	closeFile := func() error { return file.Close() }

	info, err := file.Stat()
	if err != nil {
		_ = closeFile()
		return nil, fileBytesFileUnavailable
	}
	if !info.Mode().IsRegular() {
		_ = closeFile()
		return nil, fileBytesFileUnavailable
	}

	data := allocate(FileBytesLimit + 1)
	if len(data) < FileBytesLimit+1 {
		_ = closeFile()
		return nil, fileBytesFileUnavailable
	}
	data = data[:FileBytesLimit+1]
	used := 0
	var readErr error
	for used < len(data) {
		count, err := file.Read(data[used:])
		if count < 0 || count > len(data)-used {
			readErr = errors.New("file reader returned an invalid byte count")
			break
		}
		used += count
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			readErr = err
			break
		}
		if count == 0 {
			readErr = errors.New("file reader made no progress")
			break
		}
	}
	closeErr := closeFile()
	if used > FileBytesLimit {
		return nil, fileBytesFileTooLarge
	}
	if readErr != nil || closeErr != nil {
		return nil, fileBytesFileUnavailable
	}
	return data[:used], ""
}
