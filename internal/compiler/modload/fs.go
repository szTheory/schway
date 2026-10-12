// Package modload is the only Phase 30 package that opens module files (D-10).
//
// Modules are whole-program composition, not separate compilation: the loader
// resolves a closed set of source files under one entry directory, and the
// result is checked, validated and emitted as one program. This file holds the
// filesystem primitive that every module read goes through. It never wires
// itself into a CLI route; later Phase 30 plans compose it.
//
// The primitive refuses, by name and without leaking error text: any path
// component that is a symlink (module.symlink), anything the root layer
// rejects as leaving the entry directory (module.escape), a final component
// that is not a regular file before it is opened (module.not_regular), a name
// that differs from the on-disk spelling only by case (module.case_mismatch),
// a missing file (module.not_found), and a file whose identity changes between
// the pre-open Lstat and the post-open Stat (module.unstable_file).
package modload

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/szTheory/schway/internal/compiler/syntax"
)

// Refusal codes returned by readModule. A successful read returns "".
const (
	CodeEscape       = "module.escape"
	CodeSymlink      = "module.symlink"
	CodeNotFound     = "module.not_found"
	CodeCaseMismatch = "module.case_mismatch"
	CodeNotRegular   = "module.not_regular"
	CodeUnstableFile = "module.unstable_file"
)

// moduleExtension is the source suffix of a module file.
const moduleExtension = ".schway"

// Fault-injection seams for the gate tests. Each makes the primitive weaker in
// exactly one way so a control can observe the matching test fail. They are
// unexported, default false, and restored by tests with defer.
var (
	// skipComponentLstatForTest makes the primitive Lstat only the final path,
	// which follows in-root intermediate symlinks.
	skipComponentLstatForTest bool
	// openBeforeLstatForTest makes the primitive open the final path before
	// checking its mode, which blocks forever on a FIFO.
	openBeforeLstatForTest bool
	// caseFoldAcceptForTest makes the exact-case check accept an
	// strings.EqualFold match.
	caseFoldAcceptForTest bool
)

// rootFS is a module file reader confined to one entry directory.
type rootFS struct {
	root *os.Root
}

// openModuleRoot opens dir as the confinement root for module reads.
func openModuleRoot(dir string) (*rootFS, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	return &rootFS{root: root}, nil
}

func (r *rootFS) close() error {
	return r.root.Close()
}

// validSegment reports whether a module path segment can name a directory
// entry. Anything else is refused as an escape before the root layer sees it.
func validSegment(segment string) bool {
	if segment == "" || segment == "." || segment == ".." {
		return false
	}
	return !strings.ContainsAny(segment, "/\\\x00")
}

// listNames returns the entry names of the directory at path (relative to the
// root; "." for the root itself). There is no Root.ReadDir, so the directory is
// opened through the root and listed with (*os.File).ReadDir.
func (r *rootFS) listNames(path string) ([]string, error) {
	dir, err := r.root.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

// exactCase classifies name against the entries of parent: "" for a byte-equal
// match, CodeCaseMismatch when only a case-folded match exists, CodeNotFound
// when neither does. Any listing failure is CodeEscape: the parent was already
// proven to be an in-root directory, so only the root layer can refuse it now.
func (r *rootFS) exactCase(parent, name string) string {
	names, err := r.listNames(parent)
	if err != nil {
		return CodeEscape
	}
	folded := false
	for _, candidate := range names {
		if candidate == name {
			return ""
		}
		if strings.EqualFold(candidate, name) {
			folded = true
		}
	}
	if folded {
		if caseFoldAcceptForTest {
			return ""
		}
		return CodeCaseMismatch
	}
	return CodeNotFound
}

// notExistCode classifies a name whose Lstat reported fs.ErrNotExist: a
// case-folded sibling is a case mismatch on every OS, otherwise not found.
func (r *rootFS) notExistCode(parent, name string) string {
	if code := r.exactCase(parent, name); code == CodeCaseMismatch {
		return code
	}
	return CodeNotFound
}

// joinRel joins relative root paths with a slash; "" is the root.
func joinRel(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}

// readModule reads the module file named by segments (for example
// ["a", "b"] reads a/b.schway). It returns the bytes and "" on success, or
// nil and a refusal code. At most syntax.MaxSourceBytes+1 bytes are read, and
// the size is never taken from Stat.
func (r *rootFS) readModule(segments []string) ([]byte, string) {
	if len(segments) == 0 {
		return nil, CodeNotFound
	}
	for _, segment := range segments {
		if !validSegment(segment) {
			return nil, CodeEscape
		}
	}
	dirs := segments[:len(segments)-1]
	finalName := segments[len(segments)-1] + moduleExtension

	parent := ""
	if skipComponentLstatForTest {
		if len(dirs) > 0 {
			parent = strings.Join(dirs, "/")
		}
	} else {
		for _, dir := range dirs {
			path := joinRel(parent, dir)
			info, err := r.root.Lstat(path)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return nil, r.notExistCode(listParent(parent), dir)
				}
				return nil, CodeEscape
			}
			if info.Mode()&fs.ModeSymlink != 0 {
				return nil, CodeSymlink
			}
			if !info.IsDir() {
				return nil, CodeNotFound
			}
			if code := r.exactCase(listParent(parent), dir); code != "" {
				return nil, code
			}
			parent = path
		}
	}

	finalPath := joinRel(parent, finalName)
	var early *os.File
	if openBeforeLstatForTest {
		file, err := r.root.Open(finalPath)
		if err == nil {
			early = file
			defer early.Close()
		}
	}
	linfo, err := r.root.Lstat(finalPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, r.notExistCode(listParent(parent), finalName)
		}
		return nil, CodeEscape
	}
	if linfo.Mode()&fs.ModeSymlink != 0 {
		return nil, CodeSymlink
	}
	if code := r.exactCase(listParent(parent), finalName); code != "" {
		return nil, code
	}
	if !linfo.Mode().IsRegular() {
		return nil, CodeNotRegular
	}
	file := early
	if file == nil {
		opened, err := r.root.Open(finalPath)
		if err != nil {
			return nil, CodeEscape
		}
		defer opened.Close()
		file = opened
	}
	oinfo, err := file.Stat()
	if err != nil || !os.SameFile(linfo, oinfo) {
		return nil, CodeUnstableFile
	}
	data, err := io.ReadAll(io.LimitReader(file, int64(syntax.MaxSourceBytes)+1))
	if err != nil {
		return nil, CodeEscape
	}
	return data, ""
}

// listParent maps a parent path ("" for the root) to the path listNames opens.
func listParent(parent string) string {
	if parent == "" {
		return "."
	}
	return parent
}
