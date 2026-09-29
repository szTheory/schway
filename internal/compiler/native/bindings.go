package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const BindingSchema = "schway.local-c/1"

const (
	maxBindingManifestBytes = 64 * 1024
	maxBindingFileBytes     = 4 * 1024 * 1024
	maxBindingTotalBytes    = 16 * 1024 * 1024
	maxBindingEntries       = 64
)

// BindingManifest grants local build authority, not Lang foreign-call authority.
// FunctionType names a function typedef in Header, never a free-form C fragment.
type BindingManifest struct {
	Schema              string          `json:"schema"`
	Sources             []string        `json:"sources"`
	Headers             []string        `json:"headers"`
	IncludeDirs         []string        `json:"include_dirs"`
	Symbols             []BindingSymbol `json:"symbols"`
	RuntimeDependencies []string        `json:"runtime_dependencies"`
}

type BindingSymbol struct {
	Name         string `json:"name"`
	Header       string `json:"header"`
	FunctionType string `json:"function_type"`
}

type BindingInput struct {
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Digest string `json:"digest"`
}

// ResolvedBindings retains the exact bytes hashed during resolution. Compilation
// uses a private copy, so later source-tree edits cannot silently change inputs.
type ResolvedBindings struct {
	Manifest       BindingManifest
	ManifestDigest string
	Inventory      []BindingInput
	files          map[string][]byte
}

func bindingError(format string, args ...any) error {
	return &ToolError{Code: "native.bindings_invalid", Err: fmt.Errorf(format, args...)}
}

func readBindingFile(name string, limit int64) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("input is not a regular file")
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("input exceeds size limit")
	}
	return data, nil
}

// LoadBindings decodes a bounded, closed JSON document, rejecting duplicate keys
// at every nesting level. Arrays are canonicalized; include search order is kept.
func LoadBindings(manifestPath string) (BindingManifest, error) {
	var manifest BindingManifest
	data, err := readBindingFile(manifestPath, maxBindingManifestBytes)
	if err != nil {
		return manifest, bindingError("manifest %q: %v", manifestPath, err)
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return manifest, bindingError("manifest %q: %v", manifestPath, err)
	}
	// encoding/json accepts case-insensitive field aliases. A closed manifest
	// uses exact wire names so aliases cannot overwrite an already-seen key.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return manifest, bindingError("manifest %q: %v", manifestPath, err)
	}
	for key := range fields {
		switch key {
		case "schema", "sources", "headers", "include_dirs", "symbols", "runtime_dependencies":
		default:
			return manifest, bindingError("unknown manifest field %q", key)
		}
	}
	var symbols []map[string]json.RawMessage
	if raw, ok := fields["symbols"]; ok {
		if err := json.Unmarshal(raw, &symbols); err != nil {
			return manifest, bindingError("symbols: %v", err)
		}
		for _, symbol := range symbols {
			for key := range symbol {
				switch key {
				case "name", "header", "function_type":
				default:
					return manifest, bindingError("unknown symbol field %q", key)
				}
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, bindingError("manifest %q: %v", manifestPath, err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return manifest, bindingError("manifest %q: trailing JSON", manifestPath)
	}
	if err := validateBindings(&manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

var bindingPathPattern = regexp.MustCompile(`^[A-Za-z0-9_./ -]+$`)
var bindingIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validBindingPath(name string, directory bool) bool {
	if name == "." {
		return directory
	}
	if !bindingPathPattern.MatchString(name) || path.IsAbs(name) || path.Clean(name) != name {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." || part == "." || part == "" {
			return false
		}
	}
	return true
}

func validateBindings(m *BindingManifest) error {
	if m.Schema != BindingSchema {
		return bindingError("unsupported schema %q", m.Schema)
	}
	if len(m.Sources) == 0 || len(m.Headers) == 0 || len(m.Symbols) == 0 || m.IncludeDirs == nil {
		return bindingError("sources, headers, symbols and include_dirs must be declared")
	}
	if len(m.RuntimeDependencies) != 1 || m.RuntimeDependencies[0] != "platform-c-runtime" {
		return bindingError("runtime_dependencies must contain only platform-c-runtime")
	}
	seen := map[string]bool{}
	for _, group := range []struct {
		kind  string
		names []string
	}{{"source", m.Sources}, {"header", m.Headers}, {"include directory", m.IncludeDirs}} {
		if len(group.names) > maxBindingEntries {
			return bindingError("too many %s declarations", group.kind)
		}
		for _, name := range group.names {
			if !validBindingPath(name, group.kind == "include directory") {
				return bindingError("%s %q: expected canonical relative path without traversal", group.kind, name)
			}
			if seen[name] {
				return bindingError("duplicate declaration %q", name)
			}
			seen[name] = true
			if group.kind == "source" && !strings.HasSuffix(name, ".c") {
				return bindingError("source %q must end in .c", name)
			}
		}
	}
	headers := map[string]bool{}
	for _, h := range m.Headers {
		headers[h] = true
	}
	symbols := map[string]bool{}
	if len(m.Symbols) > maxBindingEntries {
		return bindingError("too many symbol declarations")
	}
	for _, symbol := range m.Symbols {
		if !bindingIdentifierPattern.MatchString(symbol.Name) || !bindingIdentifierPattern.MatchString(symbol.FunctionType) {
			return bindingError("symbol %q type %q: expected C identifiers", symbol.Name, symbol.FunctionType)
		}
		if !headers[symbol.Header] {
			return bindingError("symbol %q: undeclared header %q", symbol.Name, symbol.Header)
		}
		if symbols[symbol.Name] {
			return bindingError("duplicate symbol %q", symbol.Name)
		}
		symbols[symbol.Name] = true
	}
	sort.Strings(m.Sources)
	sort.Strings(m.Headers)
	sort.Slice(m.Symbols, func(i, j int) bool { return m.Symbols[i].Name < m.Symbols[j].Name })
	return nil
}

func withinDirectory(root, name string) bool {
	rel, err := filepath.Rel(root, name)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// ResolveBindings confines every real path (including symlink targets) to the
// manifest directory and snapshots a bounded inventory under portable names.
func ResolveBindings(manifestPath string) (ResolvedBindings, error) {
	m, err := LoadBindings(manifestPath)
	if err != nil {
		return ResolvedBindings{}, err
	}
	root, err := filepath.Abs(filepath.Dir(manifestPath))
	if err == nil {
		root, err = filepath.EvalSymlinks(root)
	}
	if err != nil {
		return ResolvedBindings{}, bindingError("manifest root: %v", err)
	}
	result := ResolvedBindings{Manifest: m, files: map[string][]byte{}}
	canonical, _ := json.Marshal(m)
	result.ManifestDigest = digestBytes(canonical)
	seen := map[string]string{}
	total := 0
	for _, group := range []struct {
		kind  string
		names []string
	}{{"source", m.Sources}, {"header", m.Headers}, {"include directory", m.IncludeDirs}} {
		for _, name := range group.names {
			resolved, err := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(name)))
			if err != nil {
				return result, bindingError("%s %q: %v", group.kind, name, err)
			}
			if !withinDirectory(root, resolved) {
				return result, bindingError("%s %q: symlink escapes manifest root", group.kind, name)
			}
			if prior, exists := seen[resolved]; exists {
				return result, bindingError("duplicate resolved path %q and %q", prior, name)
			}
			seen[resolved] = name
			info, err := os.Stat(resolved)
			if err != nil {
				return result, bindingError("%q: %v", name, err)
			}
			if group.kind == "include directory" {
				if !info.IsDir() {
					return result, bindingError("include directory %q is not a directory", name)
				}
				continue
			}
			if !info.Mode().IsRegular() {
				return result, bindingError("%s %q is not a regular file", group.kind, name)
			}
			data, err := readBindingFile(resolved, maxBindingFileBytes)
			if err != nil {
				return result, bindingError("%s %q: %v", group.kind, name, err)
			}
			total += len(data)
			if total > maxBindingTotalBytes {
				return result, bindingError("declared input bytes exceed limit")
			}
			result.files[name] = data
			result.Inventory = append(result.Inventory, BindingInput{Name: name, Kind: group.kind, Digest: digestBytes(data)})
		}
	}
	return result, nil
}

func (b ResolvedBindings) stage(directory string) error {
	for _, name := range b.Manifest.IncludeDirs {
		if err := os.MkdirAll(filepath.Join(directory, "local", filepath.FromSlash(name)), 0o700); err != nil {
			return err
		}
	}
	for name, data := range b.files {
		dest := filepath.Join(directory, "local", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func bindingProbeSource(s BindingSymbol, index int) string {
	var out strings.Builder
	// Each symbol is checked in isolation against its own named header. A
	// sibling header must not accidentally supply the missing declaration.
	fmt.Fprintf(&out, "#include \"local/%s\"\n", s.Header)
	fmt.Fprintf(&out, "#ifdef %s\n#error declared symbol %s must not be a macro\n#endif\n", s.Name, s.Name)
	// C17 function-designator conversion distinguishes function typedefs
	// from object and function-pointer typedefs without executing a call.
	fmt.Fprintf(&out, "_Static_assert(_Generic(*(%s *)0, %s *: 1, default: 0), \"%s must name a function type\");\n", s.FunctionType, s.FunctionType, s.FunctionType)
	// Compare the declared symbol itself with the named function type. This
	// makes every manifest entry an independent exact-prototype ABI check,
	// rather than relying on assignment diagnostics to reject a mismatch.
	fmt.Fprintf(&out, "_Static_assert(_Generic(&%s, %s *: 1, default: 0), \"%s prototype does not match %s\");\n", s.Name, s.FunctionType, s.Name, s.FunctionType)
	fmt.Fprintf(&out, "%s *volatile schway_binding_probe_%d = &%s;\n", s.FunctionType, index, s.Name)
	return out.String()
}

// BuildIdentityInputs contains only canonical names/content and toolchain facts.
// Absolute output, checkout, staging and compiler paths are deliberately absent.
type BuildIdentityInputs struct {
	SourceDigest        string          `json:"source_digest"`
	EmittedCDigest      string          `json:"emitted_c_digest"`
	ManifestDigest      string          `json:"manifest_digest"`
	Inventory           []BindingInput  `json:"inventory"`
	Symbols             []BindingSymbol `json:"symbols"`
	IncludedHeaders     []string        `json:"included_headers"`
	ABIProbeDigest      string          `json:"abi_probe_digest"`
	CompilerDigest      string          `json:"compiler_digest"`
	CompilerVersion     string          `json:"compiler_version"`
	Target              string          `json:"target"`
	HostABI             string          `json:"host_abi"`
	Flags               []string        `json:"flags"`
	Commands            [][]string      `json:"commands"`
	RuntimeDependencies []string        `json:"runtime_dependencies"`
	ExecutableDigest    string          `json:"executable_digest"`
}

// ID hashes a domain and canonical JSON with explicit length framing. JSON
// itself frames each named field/list element, avoiding concatenation ambiguity.
func (in BuildIdentityInputs) ID() string {
	h := sha256.New()
	encoded, _ := json.Marshal(in)
	for _, field := range [][]byte{[]byte("lang.build-identity/1"), encoded} {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(field)))
		h.Write(size[:])
		h.Write(field)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (in BuildIdentityInputs) InputID() string { in.ExecutableDigest = ""; return in.ID() }

// Compiler environment cannot inject undeclared local include or library roots.
// Host SDK/driver selection remains explicitly outside complete closure.
func applicationEnvironment() []string {
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "CPATH", "C_INCLUDE_PATH", "CPLUS_INCLUDE_PATH", "OBJC_INCLUDE_PATH", "LIBRARY_PATH", "CCC_OVERRIDE_OPTIONS", "DEPENDENCIES_OUTPUT", "SUNPRO_DEPENDENCIES":
			continue
		}
		env = append(env, entry)
	}
	return env
}

func (r Runner) applicationCommand(parent context.Context, clang, directory string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, r.Timeout)
	defer cancel()
	cmd := r.commandContext(ctx, clang, args...)
	cmd.Dir, cmd.Env = directory, applicationEnvironment()
	var stdout, stderr boundedWriter
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	r.recordCommandLine(clang, args)
	err := cmd.Run()
	if ctx.Err() != nil {
		return "", &ToolError{Code: "native.timeout", Err: ctx.Err()}
	}
	if stdout.overflowed() || stderr.overflowed() {
		return "", streamError("native.compile_diagnostics_truncated")
	}
	if err != nil {
		return "", &ToolError{Code: "native.compile_failed", Err: withStderr(err, stderr.bytes())}
	}
	return string(stdout.bytes()) + string(stderr.bytes()), nil
}

func (r Runner) systemIncludeRoots(ctx context.Context, clang, directory string) ([]string, error) {
	output, err := r.applicationCommand(ctx, clang, directory, "-E", "-v", "-x", "c", "-")
	if err != nil {
		return nil, err
	}
	var roots []string
	active, ended := false, false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "#include <...> search starts here:" {
			active = true
			continue
		}
		if active && line == "End of search list." {
			ended = true
			break
		}
		if active {
			line = strings.TrimSuffix(line, " (framework directory)")
			if !filepath.IsAbs(line) {
				return nil, bindingError("compiler reported a relative system include root %q", line)
			}
			root, err := filepath.EvalSymlinks(line)
			if err != nil {
				return nil, bindingError("system include root %q: %v", line, err)
			}
			roots = append(roots, root)
		}
	}
	if !ended || len(roots) == 0 {
		return nil, bindingError("compiler did not report system include roots")
	}
	return roots, nil
}

// Clang make dependencies escape spaces, # and backslashes, and double $.
// Parsing is separate from the old cache parser, whose stricter grammar stays
// unchanged. Every dependency is resolved and classified before compilation.
func bindingDependencies(output string) ([]string, error) {
	output = strings.ReplaceAll(output, "\\\n", " ")
	_, rest, ok := strings.Cut(output, ":")
	if !ok {
		return nil, bindingError("invalid compiler dependency output")
	}
	var names []string
	var token strings.Builder
	flush := func() {
		if token.Len() > 0 {
			names = append(names, token.String())
			token.Reset()
		}
	}
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case '\\':
			i++
			if i == len(rest) {
				return nil, bindingError("incomplete dependency escape")
			}
			token.WriteByte(rest[i])
		case '$':
			if i+1 < len(rest) && rest[i+1] == '$' {
				i++
			}
			token.WriteByte('$')
		case ' ', '\t', '\r', '\n':
			flush()
		default:
			token.WriteByte(rest[i])
		}
	}
	flush()
	if len(names) == 0 {
		return nil, bindingError("empty compiler dependency inventory")
	}
	return names, nil
}

func (b ResolvedBindings) checkDependencies(directory, unit, output string, systemRoots []string) ([]string, error) {
	names, err := bindingDependencies(output)
	if err != nil {
		return nil, err
	}
	var headers []string
	for _, name := range names {
		abs := name
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(directory, abs)
		}
		real, err := filepath.EvalSymlinks(abs)
		if err != nil {
			return nil, bindingError("%s includes %q: %v", unit, name, err)
		}
		if withinDirectory(directory, real) {
			rel, _ := filepath.Rel(directory, real)
			rel = filepath.ToSlash(rel)
			if rel == unit {
				continue
			}
			declared := strings.TrimPrefix(rel, "local/")
			found := false
			for _, h := range b.Manifest.Headers {
				if rel == "local/"+h {
					found = true
					headers = append(headers, declared)
					break
				}
			}
			if !found {
				return nil, bindingError("%s includes undeclared local header %q", unit, name)
			}
			continue
		}
		system := false
		for _, root := range systemRoots {
			if withinDirectory(root, real) {
				system = true
				break
			}
		}
		if !system {
			return nil, bindingError("%s includes out-of-root or undeclared header %q", unit, name)
		}
	}
	return headers, nil
}

func (r Runner) compileBindings(ctx context.Context, clang, directory string, b ResolvedBindings) ([]string, [][]string, []string, string, error) {
	if err := b.stage(directory); err != nil {
		return nil, nil, nil, "", err
	}
	var units, probes []string
	for i, symbol := range b.Manifest.Symbols {
		unit := fmt.Sprintf("bindings-probe-%d.c", i)
		probe := bindingProbeSource(symbol, i)
		if err := os.WriteFile(filepath.Join(directory, unit), []byte(probe), 0o600); err != nil {
			return nil, nil, nil, "", err
		}
		units = append(units, unit)
		probes = append(probes, probe)
	}
	roots, err := r.systemIncludeRoots(ctx, clang, directory)
	if err != nil {
		return nil, nil, nil, "", err
	}
	flags := append([]string(nil), applicationFlags...)
	for _, name := range b.Manifest.IncludeDirs {
		flags = append(flags, "-I", "local/"+name)
	}
	for _, source := range b.Manifest.Sources {
		units = append(units, "local/"+source)
	}
	var objects, headers []string
	var commands [][]string
	for i, unit := range units {
		deps := append(append([]string(nil), flags...), "-M", "-MT", "lang-deps", unit)
		output, err := r.applicationCommand(ctx, clang, directory, deps...)
		if err != nil {
			return nil, nil, nil, "", fmt.Errorf("binding input %s: %w", unit, err)
		}
		included, err := b.checkDependencies(directory, unit, output, roots)
		if err != nil {
			return nil, nil, nil, "", err
		}
		headers = append(headers, included...)
		object := fmt.Sprintf("binding-%d.o", i)
		args := append(append([]string(nil), flags...), "-c", unit, "-o", object)
		commands = append(commands, args)
		if _, err := r.applicationCommand(ctx, clang, directory, args...); err != nil {
			return nil, nil, nil, "", fmt.Errorf("binding input %s (symbols %v): %w", unit, b.Manifest.Symbols, err)
		}
		objects = append(objects, object)
	}
	sort.Strings(headers)
	unique := make([]string, 0, len(headers))
	for _, h := range headers {
		if len(unique) == 0 || unique[len(unique)-1] != h {
			unique = append(unique, h)
		}
	}
	probeBytes, _ := json.Marshal(probes)
	return objects, commands, unique, digestBytes(probeBytes), nil
}
