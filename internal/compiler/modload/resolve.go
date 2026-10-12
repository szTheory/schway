package modload

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// Refusal codes. Every code is a literal at its emit site as well; there is no
// central registry.
const (
	CodeInvalidSegment = "module.invalid_segment"
	CodeEscape         = "module.escape"
	CodeSymlink        = "module.symlink"
	CodeNotFound       = "module.not_found"
	CodeCaseMismatch   = "module.case_mismatch"
	CodeNotRegular     = "module.not_regular"
	CodeUnstableFile   = "module.unstable_file"
	CodeDirectoryLimit = "module.directory_limit"
	CodeReadFailed     = "module.read_failed"
	codeInputLimit     = "syntax.input_limit"
)

const (
	// MaxSegmentBytes bounds one module path segment (D-18).
	MaxSegmentBytes = 64
	// MaxDirectoryEntries bounds one exact-case directory listing (D-24).
	MaxDirectoryEntries = 65536

	// moduleExtension is the exact lowercase source suffix of a module file.
	moduleExtension = ".schway"
	listBatch       = 256
)

// Fault-injection seams. Each makes the resolver weaker in exactly one way so a
// control can observe the matching test fail. They are unexported, default
// false, and restored by tests with defer.
var (
	// skipComponentLstatForTest skips the per-directory Lstat, so an in-root
	// intermediate symlink is followed.
	skipComponentLstatForTest bool
	// openBeforeLstatForTest opens the final path before checking its mode,
	// which blocks forever on a FIFO.
	openBeforeLstatForTest bool
	// caseFoldAcceptForTest makes the exact-case check accept an
	// strings.EqualFold match.
	caseFoldAcceptForTest bool
	// bypassRootForTest makes the real filesystem resolve host paths outside
	// the confinement root, following links.
	bypassRootForTest bool
	// skipSameFileForTest accepts an opened file whose identity differs from
	// the pre-open Lstat result.
	skipSameFileForTest bool
	// sizeFromStatForTest takes the per-module size from Stat instead of the
	// bytes actually read.
	sizeFromStatForTest bool
	// byteCapOffByOneForTest refuses a module of exactly the maximum size.
	byteCapOffByOneForTest bool
	// formatErrorTextForTest appends the operating-system error text to
	// escape and read_failed messages, reproducing the leak anti-pattern.
	// Production leaves it false.
	formatErrorTextForTest bool
)

// moduleFile is the opened module file the resolver reads.
type moduleFile interface {
	Stat() (fs.FileInfo, error)
	Read(p []byte) (int, error)
	Close() error
}

// fileSystem is the single IO seam of the resolver. Names are slash-separated
// and relative to the entry directory.
type fileSystem interface {
	Lstat(name string) (fs.FileInfo, error)
	Open(name string) (moduleFile, error)
	// ListDir lists dir, returning at most max names and truncated=true when
	// more than max entries exist.
	ListDir(dir string, max int) (names []string, truncated bool, err error)
}

// rootFileSystem implements fileSystem over an *os.Root.
type rootFileSystem struct {
	root *os.Root
	dir  string
}

func (f *rootFileSystem) Lstat(name string) (fs.FileInfo, error) {
	if bypassRootForTest {
		return os.Stat(filepath.Join(f.dir, filepath.FromSlash(name)))
	}
	return f.root.Lstat(name)
}

func (f *rootFileSystem) Open(name string) (moduleFile, error) {
	if bypassRootForTest {
		return os.Open(filepath.Join(f.dir, filepath.FromSlash(name)))
	}
	return f.root.Open(name)
}

// ListDir has no Root.ReadDir to call, so the directory is opened through the
// root and listed in bounded batches; it never reads with a negative count.
func (f *rootFileSystem) ListDir(dir string, max int) ([]string, bool, error) {
	handle, err := f.root.Open(dir)
	if err != nil {
		return nil, false, err
	}
	defer handle.Close()
	var names []string
	for {
		entries, err := handle.ReadDir(listBatch)
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		if len(names) > max {
			return names[:max], true, nil
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return names, false, nil
			}
			return nil, false, err
		}
	}
}

func (f *rootFileSystem) close() error { return f.root.Close() }

// Source is one module file read under the entry directory.
type Source struct {
	// Module is the dotted module name that was requested.
	Module string
	// RelPath is the slash-separated path under the entry directory.
	RelPath string
	Bytes   []byte
	info    fs.FileInfo
}

// resolver maps validated module names to file bytes under one entry directory.
type resolver struct {
	fsys     fileSystem
	sameFile func(a, b fs.FileInfo) bool
	// dir is the entry directory. It is rendered only under
	// formatErrorTextForTest.
	dir string
}

// newResolver opens dir as the entry directory.
func newResolver(dir string) (*resolver, error) {
	root, err := openModuleRoot(dir)
	if err != nil {
		return nil, err
	}
	return &resolver{fsys: &rootFileSystem{root: root, dir: dir}, sameFile: os.SameFile, dir: dir}, nil
}

func (r *resolver) close() error {
	if closer, ok := r.fsys.(interface{ close() error }); ok {
		return closer.close()
	}
	return nil
}

// validateModulePath reports whether module is a dotted name whose every
// segment is 1..64 ASCII bytes matching [a-z][a-z0-9]*(_[a-z0-9]+)* (D-07,
// D-18). It is a byte loop on purpose: no pattern engine, no unicode tables.
func validateModulePath(module string) bool {
	if module == "" {
		return false
	}
	start := 0
	for index := 0; index <= len(module); index++ {
		if index < len(module) && module[index] != '.' {
			continue
		}
		if !validModuleSegment(module[start:index]) {
			return false
		}
		start = index + 1
	}
	return true
}

func validModuleSegment(segment string) bool {
	if len(segment) == 0 || len(segment) > MaxSegmentBytes {
		return false
	}
	if segment[0] < 'a' || segment[0] > 'z' {
		return false
	}
	for index := 1; index < len(segment); index++ {
		c := segment[index]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '_':
			// An underscore must be followed by an alphanumeric: no "__",
			// no trailing "_".
			if index+1 >= len(segment) {
				return false
			}
			next := segment[index+1]
			if !(next >= 'a' && next <= 'z' || next >= '0' && next <= '9') {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// refusal is one named resolver refusal. err is the underlying operation error;
// it is only ever rendered under formatErrorTextForTest.
type refusal struct {
	code string
	rel  string
	err  error
}

// read resolves module to its file under the entry directory. Every refusal is
// one diagnostic with primary span at (the import declaration in importer),
// attributed to importer, and naming the module and a root-relative path only.
//
// When entry is non-nil and the file's Lstat identity is the entry's
// (sameFile), isEntry is true and the file is not opened: the loader records an
// edge to the entry instead of reading it a second time (D-24).
func (r *resolver) read(module, importer string, at diagnostic.Span, entry fs.FileInfo) (Source, bool, []diagnostic.Diagnostic) {
	if !validateModulePath(module) {
		return Source{}, false, r.diagnose(&refusal{code: CodeInvalidSegment}, module, importer, at)
	}
	return r.readEntryAware(strings.Split(module, "."), importer, at, entry)
}

// readSegments performs resolution without the grammar check; read calls it
// after validateModulePath. The root layer is still the second defence.
func (r *resolver) readSegments(segments []string, importer string, at diagnostic.Span) (Source, []diagnostic.Diagnostic) {
	source, _, diags := r.readEntryAware(segments, importer, at, nil)
	return source, diags
}

func (r *resolver) readEntryAware(segments []string, importer string, at diagnostic.Span, entry fs.FileInfo) (Source, bool, []diagnostic.Diagnostic) {
	module := strings.Join(segments, ".")
	source, isEntry, refused := r.load(module, segments, entry)
	if refused != nil {
		return Source{}, false, r.diagnose(refused, module, importer, at)
	}
	if isEntry {
		return source, true, nil
	}
	size := int64(len(source.Bytes))
	if sizeFromStatForTest {
		size = source.info.Size()
	}
	if size > int64(syntax.MaxSourceBytes) || byteCapOffByOneForTest && size == int64(syntax.MaxSourceBytes) {
		return Source{}, false, []diagnostic.Diagnostic{
			diagnostic.Error(codeInputLimit, at, "source exceeds a compiler input limit").InUnit(module),
		}
	}
	return source, false, nil
}

func (r *resolver) diagnose(refused *refusal, module, importer string, at diagnostic.Span) []diagnostic.Diagnostic {
	where := ""
	if refused.rel != "" {
		where = " at " + refused.rel
	}
	var message string
	switch refused.code {
	case CodeInvalidSegment:
		message = "module \"" + module + "\" has an invalid name: each segment must be lowercase ASCII letters, digits and single underscores, start with a letter, and be at most 64 bytes"
	case CodeEscape:
		message = "module \"" + module + "\" resolves outside the entry directory" + where
	case CodeSymlink:
		message = "module \"" + module + "\" crosses a symbolic link" + where
	case CodeNotFound:
		message = "module \"" + module + "\" was not found" + where
	case CodeCaseMismatch:
		message = "module \"" + module + "\" differs from the on-disk name only by case" + where
	case CodeNotRegular:
		message = "module \"" + module + "\" is not a regular file" + where
	case CodeUnstableFile:
		message = "module \"" + module + "\" changed while it was being read" + where
	case CodeDirectoryLimit:
		message = "module \"" + module + "\" is in a directory with more entries than the listing limit" + where
	default:
		message = "module \"" + module + "\" could not be read" + where
	}
	if formatErrorTextForTest {
		if refused.err != nil {
			message += ": " + refused.err.Error()
		}
		if refused.code == CodeNotFound {
			message += ": " + filepath.Join(r.dir, filepath.FromSlash(refused.rel))
		}
	}
	return []diagnostic.Diagnostic{diagnostic.Error(refused.code, at, message).InUnit(importer)}
}

// load runs the filesystem pipeline: per-component Lstat and exact-case
// listing, regular-file check before Open, SameFile after Open, one bounded
// read.
func (r *resolver) load(module string, segments []string, entry fs.FileInfo) (Source, bool, *refusal) {
	if len(segments) == 0 {
		return Source{}, false, &refusal{code: CodeNotFound}
	}
	for _, segment := range segments {
		if !pathSafeSegment(segment) {
			return Source{}, false, &refusal{code: CodeEscape, rel: segment}
		}
	}
	dirs := segments[:len(segments)-1]
	finalName := segments[len(segments)-1] + moduleExtension

	parent := ""
	if skipComponentLstatForTest {
		parent = strings.Join(dirs, "/")
	} else {
		for _, dir := range dirs {
			path := joinRel(parent, dir)
			info, err := r.fsys.Lstat(path)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return Source{}, false, r.notExist(parent, dir, path)
				}
				return Source{}, false, &refusal{code: CodeEscape, rel: path, err: err}
			}
			if info.Mode()&fs.ModeSymlink != 0 {
				return Source{}, false, &refusal{code: CodeSymlink, rel: path}
			}
			if !info.IsDir() {
				return Source{}, false, &refusal{code: CodeNotFound, rel: path}
			}
			if refused := r.exactCase(parent, dir, path); refused != nil {
				return Source{}, false, refused
			}
			parent = path
		}
	}

	finalPath := joinRel(parent, finalName)
	var early moduleFile
	if openBeforeLstatForTest {
		if file, err := r.fsys.Open(finalPath); err == nil {
			early = file
			defer early.Close()
		}
	}
	linfo, err := r.fsys.Lstat(finalPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Source{}, false, r.notExist(parent, finalName, finalPath)
		}
		return Source{}, false, &refusal{code: CodeEscape, rel: finalPath, err: err}
	}
	if linfo.Mode()&fs.ModeSymlink != 0 {
		return Source{}, false, &refusal{code: CodeSymlink, rel: finalPath}
	}
	if refused := r.exactCase(parent, finalName, finalPath); refused != nil {
		return Source{}, false, refused
	}
	if !linfo.Mode().IsRegular() {
		return Source{}, false, &refusal{code: CodeNotRegular, rel: finalPath}
	}
	if entry != nil && r.sameFile(linfo, entry) {
		return Source{Module: module, RelPath: finalPath}, true, nil
	}
	file := early
	if file == nil {
		opened, err := r.fsys.Open(finalPath)
		if err != nil {
			return Source{}, false, &refusal{code: CodeEscape, rel: finalPath, err: err}
		}
		defer opened.Close()
		file = opened
	}
	oinfo, err := file.Stat()
	if err != nil || !skipSameFileForTest && !r.sameFile(linfo, oinfo) {
		return Source{}, false, &refusal{code: CodeUnstableFile, rel: finalPath}
	}
	data, err := io.ReadAll(io.LimitReader(file, int64(syntax.MaxSourceBytes)+1))
	if err != nil {
		return Source{}, false, &refusal{code: CodeReadFailed, rel: finalPath, err: err}
	}
	return Source{Module: module, RelPath: finalPath, Bytes: data, info: oinfo}, false, nil
}

// pathSafeSegment rejects a segment that cannot name one directory entry.
func pathSafeSegment(segment string) bool {
	if segment == "" || segment == "." || segment == ".." {
		return false
	}
	return !strings.ContainsAny(segment, "/\\\x00")
}

// exactCase requires a byte-equal entry named name in parent. A case-folded
// match is CodeCaseMismatch, none is CodeNotFound, a truncated listing is
// CodeDirectoryLimit. A listing failure is CodeEscape: the parent was already
// proven to be an in-root directory, so only the root layer can refuse it now.
func (r *resolver) exactCase(parent, name, rel string) *refusal {
	names, truncated, err := r.fsys.ListDir(listParent(parent), MaxDirectoryEntries)
	if err != nil {
		return &refusal{code: CodeEscape, rel: rel, err: err}
	}
	if truncated {
		return &refusal{code: CodeDirectoryLimit, rel: rel}
	}
	folded := false
	for _, candidate := range names {
		if candidate == name {
			return nil
		}
		if strings.EqualFold(candidate, name) {
			folded = true
		}
	}
	if folded {
		if caseFoldAcceptForTest {
			return nil
		}
		return &refusal{code: CodeCaseMismatch, rel: rel}
	}
	return &refusal{code: CodeNotFound, rel: rel}
}

// notExist classifies a name whose Lstat reported fs.ErrNotExist: a case-folded
// sibling is a case mismatch on every OS, a truncated listing is the directory
// limit, otherwise not found.
func (r *resolver) notExist(parent, name, rel string) *refusal {
	if refused := r.exactCase(parent, name, rel); refused != nil {
		if refused.code == CodeCaseMismatch || refused.code == CodeDirectoryLimit {
			return refused
		}
	}
	return &refusal{code: CodeNotFound, rel: rel}
}

func joinRel(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}

// listParent maps a parent path ("" for the root) to the path ListDir opens.
func listParent(parent string) string {
	if parent == "" {
		return "."
	}
	return parent
}
