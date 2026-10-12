package modload

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
	"time"
)

// fakeEntry is one in-memory path.
type fakeEntry struct {
	mode    fs.FileMode
	content []byte
	// id is the file identity. identityAfterOpen, when non-zero, is the
	// identity the opened file reports, simulating a swap after Lstat.
	id                int
	identityAfterOpen int
	// statSize, when non-zero, is the size Stat and Lstat report regardless of
	// the real content length.
	statSize int64
	// extraNames synthesizes that many additional sibling names in this
	// directory without allocating content.
	extraNames int
}

// fakeFileSystem is an in-memory fileSystem with controllable lies.
type fakeFileSystem struct {
	entries map[string]*fakeEntry
	opens   int
}

func newFakeFileSystem() *fakeFileSystem {
	return &fakeFileSystem{entries: map[string]*fakeEntry{
		".": {mode: fs.ModeDir | 0o755, id: 1},
	}}
}

func (f *fakeFileSystem) addFile(path, content string, id int) *fakeEntry {
	return f.addBytes(path, []byte(content), id)
}

func (f *fakeFileSystem) addBytes(path string, content []byte, id int) *fakeEntry {
	entry := &fakeEntry{mode: 0o644, content: content, id: id}
	f.entries[path] = entry
	return entry
}

func (f *fakeFileSystem) addDir(path string, id int) *fakeEntry {
	entry := &fakeEntry{mode: fs.ModeDir | 0o755, id: id}
	f.entries[path] = entry
	return entry
}

type fakeInfo struct {
	name string
	mode fs.FileMode
	size int64
	id   int
}

func (i *fakeInfo) Name() string       { return i.name }
func (i *fakeInfo) Size() int64        { return i.size }
func (i *fakeInfo) Mode() fs.FileMode  { return i.mode }
func (i *fakeInfo) ModTime() time.Time { return time.Time{} }
func (i *fakeInfo) IsDir() bool        { return i.mode.IsDir() }
func (i *fakeInfo) Sys() any           { return i.id }

func fakeSameFile(a, b fs.FileInfo) bool {
	left, lok := a.(*fakeInfo)
	right, rok := b.(*fakeInfo)
	return lok && rok && left.id == right.id
}

func (e *fakeEntry) info(path string, id int) *fakeInfo {
	size := int64(len(e.content))
	if e.statSize != 0 {
		size = e.statSize
	}
	return &fakeInfo{name: path[strings.LastIndex(path, "/")+1:], mode: e.mode, size: size, id: id}
}

func (f *fakeFileSystem) Lstat(name string) (fs.FileInfo, error) {
	entry, ok := f.entries[name]
	if !ok {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: fs.ErrNotExist}
	}
	return entry.info(name, entry.id), nil
}

type fakeFile struct {
	reader io.Reader
	info   *fakeInfo
}

func (f *fakeFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f *fakeFile) Read(p []byte) (int, error) { return f.reader.Read(p) }
func (f *fakeFile) Close() error               { return nil }

func (f *fakeFileSystem) Open(name string) (moduleFile, error) {
	f.opens++
	entry, ok := f.entries[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	id := entry.id
	if entry.identityAfterOpen != 0 {
		id = entry.identityAfterOpen
	}
	return &fakeFile{reader: bytes.NewReader(entry.content), info: entry.info(name, id)}, nil
}

func (f *fakeFileSystem) ListDir(dir string, max int) ([]string, bool, error) {
	owner, ok := f.entries[dir]
	if !ok || !owner.mode.IsDir() {
		return nil, false, &fs.PathError{Op: "open", Path: dir, Err: fs.ErrNotExist}
	}
	var names []string
	for path := range f.entries {
		if path == "." {
			continue
		}
		parent, base := ".", path
		if slash := strings.LastIndex(path, "/"); slash >= 0 {
			parent, base = path[:slash], path[slash+1:]
		}
		if parent == dir {
			names = append(names, base)
		}
	}
	sort.Strings(names)
	for index := 0; index < owner.extraNames; index++ {
		names = append(names, fmt.Sprintf("zz%06d.txt", index))
	}
	if len(names) > max {
		return names[:max], true, nil
	}
	return names, false, nil
}

func newFakeResolver(fsys *fakeFileSystem) *resolver {
	return &resolver{fsys: fsys, sameFile: fakeSameFile}
}
