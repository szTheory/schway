package modload

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const (
	// MaxModules bounds the distinct modules of a program, the entry included
	// (D-13).
	MaxModules = 16
	// MaxImportDepth bounds the import chain below the entry (entry = 0).
	MaxImportDepth = 8
	// MaxTotalSourceBytes bounds the bytes read across every module.
	MaxTotalSourceBytes = 4194304
)

// Fault-injection seams (unexported, default off, restored with defer). Each
// weakens exactly one defence so a control can observe the matching test fail.
var (
	// capOffByOneForTest refuses a count, depth or byte total exactly at its
	// limit.
	capOffByOneForTest bool
	// countAfterOpenForTest moves the module-count check after the read, so
	// the refused module is opened.
	countAfterOpenForTest bool
	// diamondTwiceForTest enqueues an already-visited module once more, so a
	// shared module is read and counted twice.
	diamondTwiceForTest bool
	// visitedInsteadOfColorForTest treats any already-visited node as a cycle.
	visitedInsteadOfColorForTest bool
	// rereadForTest reads each imported module a second time after its parse,
	// reproducing the identity-hashing re-read anti-pattern (D-10).
	rereadForTest bool
	// skipGraphCheckForTest names one refusal code (module.not_found,
	// module.duplicate or module.path_mismatch) whose check is skipped; "" in
	// production.
	skipGraphCheckForTest string
	// readObservedForTest, when set, is called for every file read, with module
	// "" for the entry.
	readObservedForTest func(module string)
	// skipParseUnitStampForTest skips the InUnit stamping of an imported
	// module's parse diagnostics.
	skipParseUnitStampForTest bool
)

// SetDiamondTwiceForTest arms the diamond-twice mutant and returns a restore
// func. It exists for the session tests of plan 30-08.
func SetDiamondTwiceForTest(on bool) (restore func()) {
	previous := diamondTwiceForTest
	diamondTwiceForTest = on
	return func() { diamondTwiceForTest = previous }
}

// SetReadObserverForTest installs a per-file read observer.
func SetReadObserverForTest(observer func(module string)) (restore func()) {
	previous := readObservedForTest
	readObservedForTest = observer
	return func() { readObservedForTest = previous }
}

// SetMergeInImportOrderForTest arms the merge-in-import-order mutant.
func SetMergeInImportOrderForTest(on bool) (restore func()) {
	previous := mergeInImportOrderForTest
	mergeInImportOrderForTest = on
	return func() { mergeInImportOrderForTest = previous }
}

// SetSkipParseUnitStampForTest arms the skip-parse-unit-stamp mutant.
func SetSkipParseUnitStampForTest(on bool) (restore func()) {
	previous := skipParseUnitStampForTest
	skipParseUnitStampForTest = on
	return func() { skipParseUnitStampForTest = previous }
}

// SetFormatErrorTextForTest arms the error-text leak mutant of the resolver.
func SetFormatErrorTextForTest(on bool) (restore func()) {
	previous := formatErrorTextForTest
	formatErrorTextForTest = on
	return func() { formatErrorTextForTest = previous }
}

// Set is one loaded multi-module program.
type Set struct {
	// Program is the merged whole program (D-12). For an entry with no import
	// declarations it is exactly the single-file parse.
	Program ast.Program
	// Trees maps module name to its lossless tree, "" for the entry.
	Trees map[string]syntax.Tree
	// Sources lists every module read, entry first then sorted by name.
	Sources []Source
	// Work is the summed token count across every tree.
	Work int
}

// entryAliasAtProcessForTest restores the pre-fix behaviour where an import of
// the entry file is counted at enqueue time and uncounted only when processed.
var entryAliasAtProcessForTest bool

// repeatCapDiagForTest restores the pre-fix behaviour where every import of a
// cap-refused target repeats the cap diagnostic.
var repeatCapDiagForTest bool

func capExceeded(n, limit int) bool {
	if capOffByOneForTest {
		return n >= limit
	}
	return n > limit
}

func observeRead(module string) {
	if readObservedForTest != nil {
		readObservedForTest(module)
	}
}

// Load is the single IO site for a module program (D-10). It reads the entry
// once (the entry path is user authority, as before) and every imported module
// once through the resolver; parsing, merging and identity use those in-memory
// bytes only. The error result is non-nil only when the entry itself cannot be
// opened or read; every refusal is a diagnostic.
func Load(entryPath string) (Set, []diagnostic.Diagnostic, error) {
	file, err := os.Open(entryPath)
	if err != nil {
		return Set{}, nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return Set{}, nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, int64(syntax.MaxSourceBytes)+1))
	if err != nil {
		return Set{}, nil, err
	}
	observeRead("")

	parsed := syntax.Parse(data)
	entry := parsedModule{
		program: parsed.Program,
		tree:    parsed.Tree,
		source:  Source{RelPath: filepath.Base(entryPath), Bytes: data},
	}
	if len(parsed.Diagnostics) != 0 || len(parsed.Program.Imports) == 0 {
		return Set{
			Program: parsed.Program,
			Trees:   map[string]syntax.Tree{"": parsed.Tree},
			Sources: []Source{entry.source},
			Work:    len(parsed.Tree.Tokens),
		}, parsed.Diagnostics, nil
	}

	resolve, err := newResolver(filepath.Dir(entryPath))
	if err != nil {
		return Set{}, nil, err
	}
	defer resolve.close()
	set, diags := loadGraph(entry, info, resolve)
	return set, diags, nil
}

// importEdge is one import declaration, recorded from importer to target.
type importEdge struct {
	from string
	to   string
	decl ast.ImportDecl
}

// workItem is one module to read: an enqueue made by an import declaration.
type workItem struct {
	name     string
	depth    int
	importer string
	decl     ast.ImportDecl
	ordinal  int
}

type graphLoader struct {
	resolve   *resolver
	entryInfo fs.FileInfo
	entry     parsedModule

	nodes     map[string]parsedModule
	read      map[string]parsedModule
	edges     map[string][]importEdge
	alias     map[string]string
	visited   map[string]bool
	refused   map[string]bool
	processed map[string]bool
	reloaded  map[string]bool
	queue     []workItem
	order     []string

	count int
	total int

	diags    []diagnostic.Diagnostic
	capDiags []diagnostic.Diagnostic
}

// loadGraph runs the iterative worklist over the entry's imports, then the
// cycle walk and the merge. entryInfo identifies the entry file for the
// via-entry cycle check and may be nil.
func loadGraph(entry parsedModule, entryInfo fs.FileInfo, resolve *resolver) (Set, []diagnostic.Diagnostic) {
	l := &graphLoader{
		resolve:   resolve,
		entryInfo: entryInfo,
		entry:     entry,
		nodes:     map[string]parsedModule{"": entry},
		read:      map[string]parsedModule{"": entry},
		edges:     map[string][]importEdge{},
		alias:     map[string]string{},
		visited:   map[string]bool{},
		refused:   map[string]bool{},
		processed: map[string]bool{"": true},
		reloaded:  map[string]bool{},
		count:     1,
		total:     len(entry.source.Bytes),
	}
	l.enqueueImports("", entry, 0)

	for len(l.queue) > 0 {
		item := l.queue[0]
		l.queue = l.queue[1:]
		if !l.process(item) {
			break
		}
	}

	if len(l.capDiags) > 0 {
		// The entry program is kept, imports and all, so a caller that tests
		// len(Program.Imports) takes the module-aware path and reports the cap
		// diagnostic (D-13) instead of re-checking the entry as a single file.
		return l.result(entry.program), l.capDiags
	}
	l.findCycles()

	// Handed over in first-import order: merge, not the loader, owns the
	// canonical order (D-12).
	imported := make([]parsedModule, 0, len(l.order))
	for _, name := range l.order {
		imported = append(imported, l.nodes[name])
	}
	program, mergeDiags := merge(entry, imported)
	l.diags = append(l.diags, mergeDiags...)
	return l.result(program), l.diags
}

func (l *graphLoader) result(program ast.Program) Set {
	set := Set{Program: program, Trees: map[string]syntax.Tree{}}
	names := make([]string, 0, len(l.read))
	for name := range l.read {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		module := l.read[name]
		set.Trees[name] = module.tree
		set.Sources = append(set.Sources, module.source)
		set.Work += len(module.tree.Tokens)
	}
	return set
}

// readModule reads one module through the resolver and reports it to the read
// observer. An entry hit is not a read.
func (l *graphLoader) readModule(name, importer string, at diagnostic.Span) (Source, bool, []diagnostic.Diagnostic) {
	source, isEntry, diags := l.resolve.read(name, importer, at, l.entryInfo)
	if len(diags) == 0 && !isEntry {
		observeRead(name)
	}
	return source, isEntry, diags
}

// enqueueImports records an edge for every import declaration of module and
// enqueues each target not yet visited, enforcing the count and depth caps
// before any open (D-13).
func (l *graphLoader) enqueueImports(importer string, module parsedModule, depth int) {
	for _, decl := range module.program.Imports {
		for _, item := range decl.Items {
			if !validImportName(item.Name) {
				l.diags = append(l.diags, diagnostic.Error(codeInvalidImportName, item.Span,
					"imported name \""+item.Name+"\" is not allowed: an imported name cannot have a leading underscore, a double underscore or a trailing underscore",
				).InUnit(importer))
			}
		}
		target := decl.Module
		l.edges[importer] = append(l.edges[importer], importEdge{from: importer, to: target, decl: decl})
		if l.namesEntry(target) && !entryAliasAtProcessForTest {
			// The import names the entry file itself: an edge to the entry,
			// never a module, so it must not count toward the module cap.
			l.alias[target] = ""
			continue
		}
		if l.visited[target] && !(diamondTwiceForTest && !l.reloaded[target]) {
			continue
		}
		if l.visited[target] {
			l.reloaded[target] = true
		}
		if l.refused[target] && !repeatCapDiagForTest {
			continue // already reported once; do not repeat the cap diagnostic
		}
		next := depth + 1
		if !countAfterOpenForTest && capExceeded(l.count+1, MaxModules) {
			l.capDiags = append(l.capDiags, diagnostic.Error(codeCountLimit, decl.Span,
				"the program exceeds the limit of 16 modules, the entry included",
			).InUnit(importer))
			l.refused[target] = true
			continue
		}
		if capExceeded(next, MaxImportDepth) {
			l.capDiags = append(l.capDiags, diagnostic.Error(codeDepthLimit, decl.Span,
				"the import chain exceeds the limit of 8 modules below the entry",
			).InUnit(importer))
			l.refused[target] = true
			continue
		}
		l.count++
		l.visited[target] = true
		l.queue = append(l.queue, workItem{name: target, depth: next, importer: importer, decl: decl, ordinal: l.count})
	}
}

// namesEntry reports whether module is the path-derived name of the entry
// file, which resolves to the entry itself. Aliases that only the file identity
// reveals (a hard link) are still caught by process.
func (l *graphLoader) namesEntry(module string) bool {
	name, ok := strings.CutSuffix(l.entry.source.RelPath, ".schway")
	return ok && name != "" && name == module
}

// validImportName reports whether an imported name is free of a leading
// underscore, a double underscore and a trailing underscore (D-18).
func validImportName(name string) bool {
	return !strings.HasPrefix(name, "_") && !strings.HasSuffix(name, "_") && !strings.Contains(name, "__")
}

// process reads, parses and checks one module. It returns false when the
// loader must stop.
func (l *graphLoader) process(item workItem) bool {
	source, isEntry, diags := l.readModule(item.name, item.importer, item.decl.ModuleSpan)
	if isEntry {
		// The import names the entry file itself: an edge to the entry, never a
		// second read (D-24). The extra enqueue count is returned.
		l.alias[item.name] = ""
		l.count--
		return true
	}
	if len(diags) > 0 {
		if skipGraphCheckForTest == CodeNotFound && diags[0].Code == CodeNotFound {
			return true
		}
		l.diags = append(l.diags, diags...)
		return true
	}

	l.total += len(source.Bytes)
	if capExceeded(l.total, MaxTotalSourceBytes) {
		l.capDiags = append(l.capDiags, diagnostic.Error(codeTotalBytesLimit, item.decl.Span,
			"the program's module sources exceed the total limit of 4194304 bytes",
		).InUnit(item.importer))
		return false
	}
	if countAfterOpenForTest && capExceeded(item.ordinal, MaxModules) {
		l.capDiags = append(l.capDiags, diagnostic.Error(codeCountLimit, item.decl.Span,
			"the program exceeds the limit of 16 modules, the entry included",
		).InUnit(item.importer))
		return true
	}

	parsed := syntax.Parse(source.Bytes)
	module := parsedModule{name: item.name, program: parsed.Program, tree: parsed.Tree, source: source}
	l.read[item.name] = module
	if len(parsed.Diagnostics) > 0 {
		for _, problem := range parsed.Diagnostics {
			if !skipParseUnitStampForTest {
				problem = problem.InUnit(item.name)
			}
			l.diags = append(l.diags, problem)
		}
		return true
	}
	if !l.headerValid(item, module) {
		return true
	}

	if _, again := l.nodes[item.name]; !again {
		l.order = append(l.order, item.name)
	}
	l.nodes[item.name] = module
	l.processed[item.name] = true
	if rereadForTest {
		l.readModule(item.name, item.importer, item.decl.ModuleSpan)
	}
	l.enqueueImports(item.name, module, item.depth)
	return true
}

// headerValid applies D-06: an imported module's header equals its path-derived
// name, and no imported module repeats the entry header.
func (l *graphLoader) headerValid(item workItem, module parsedModule) bool {
	header := module.program.Module
	where := headerSpan(module.tree)
	if header != item.name && skipGraphCheckForTest != codePathMismatch {
		l.diags = append(l.diags, diagnostic.Error(codePathMismatch, item.decl.Span,
			"module \""+item.name+"\" declares the header \""+header+"\": an imported module's header must equal its path",
			diagnostic.Cause{Kind: causeKindModuleHeader, Detail: "the module header declares \"" + header + "\"", Span: &where, Unit: item.name},
		).InUnit(item.importer))
		return false
	}
	if header == l.entry.program.Module && skipGraphCheckForTest != codeDuplicate {
		l.diags = append(l.diags, diagnostic.Error(codeDuplicate, item.decl.Span,
			"module \""+item.name+"\" repeats the entry module's name \""+header+"\"",
			diagnostic.Cause{Kind: causeKindModuleHeader, Detail: "this header repeats the entry module's name", Span: &where, Unit: item.name},
		).InUnit(item.importer))
		return false
	}
	return true
}

// headerSpan covers the dotted name after the `module` keyword of tree.
func headerSpan(tree syntax.Tree) diagnostic.Span {
	tokens := tree.Tokens
	for index := 0; index < len(tokens); index++ {
		if tokens[index].Kind != syntax.TokenModule {
			continue
		}
		span := diagnostic.Span{}
		started := false
		for _, token := range tokens[index+1:] {
			switch token.Kind {
			case syntax.TokenWhitespace, syntax.TokenComment:
				if started {
					return span
				}
			case syntax.TokenIdentifier, syntax.TokenDot:
				if !started {
					span.Start = token.Span.Start
					started = true
				}
				span.End = token.Span.End
			default:
				return span
			}
		}
		return span
	}
	return diagnostic.Span{}
}

// displayName shows the entry by its header text and every other module by its
// name.
func (l *graphLoader) displayName(name string) string {
	if name == "" {
		return l.entry.program.Module
	}
	return name
}

func (l *graphLoader) target(edge importEdge) (string, bool) {
	to := edge.to
	if alias, ok := l.alias[to]; ok {
		to = alias
	}
	_, ok := l.nodes[to]
	return to, ok
}

// findCycles walks the loaded graph iteratively in three colors, copying the
// shape of callgraph.Order: gray means on the stack (a back edge is a cycle),
// black means done (a shared module reached twice is a diamond, not a cycle).
func (l *graphLoader) findCycles() {
	const (
		white = iota
		gray
		black
	)
	names := make([]string, 0, len(l.nodes))
	for name := range l.nodes {
		if name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	starts := append([]string{""}, names...)

	type frame struct {
		node string
		next int
	}
	color := map[string]int{}
	seen := map[string]bool{}
	for _, start := range starts {
		if color[start] != white {
			continue
		}
		stack := []frame{{node: start}}
		color[start] = gray
		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			outgoing := l.edges[top.node]
			if top.next >= len(outgoing) {
				color[top.node] = black
				stack = stack[:len(stack)-1]
				continue
			}
			edge := outgoing[top.next]
			top.next++
			to, ok := l.target(edge)
			if !ok {
				continue
			}
			switch color[to] {
			case white:
				color[to] = gray
				stack = append(stack, frame{node: to})
			case gray:
				index := 0
				for position, entry := range stack {
					if entry.node == to {
						index = position
						break
					}
				}
				members := make([]string, 0, len(stack)-index)
				for _, entry := range stack[index:] {
					members = append(members, entry.node)
				}
				l.reportCycle(members, edge, seen)
			case black:
				if visitedInsteadOfColorForTest {
					l.reportCycle([]string{to}, edge, seen)
				}
			}
		}
	}
}

// reportCycle emits module.import_cycle once per distinct rotation-normalized
// cycle. closing is the import declaration that closes it.
func (l *graphLoader) reportCycle(members []string, closing importEdge, seen map[string]bool) {
	smallest := 0
	for index, member := range members {
		if l.displayName(member) < l.displayName(members[smallest]) {
			smallest = index
		}
	}
	rotated := append(append([]string(nil), members[smallest:]...), members[:smallest]...)
	labels := make([]string, 0, len(rotated)+1)
	for _, member := range rotated {
		labels = append(labels, l.displayName(member))
	}
	labels = append(labels, labels[0])
	text := strings.Join(labels, " -> ")
	if seen[text] {
		return
	}
	seen[text] = true

	var causes []diagnostic.Cause
	for index, from := range rotated {
		to := rotated[(index+1)%len(rotated)]
		for _, edge := range l.edges[from] {
			if target, ok := l.target(edge); ok && target == to {
				span := edge.decl.Span
				causes = append(causes, diagnostic.Cause{
					Kind:   causeKindImportEdge,
					Detail: l.displayName(from) + " imports " + l.displayName(to),
					Span:   &span,
					Unit:   from,
				})
				break
			}
		}
	}
	l.diags = append(l.diags, diagnostic.Error(codeImportCycle, closing.decl.Span,
		"import cycle: "+text, causes...,
	).InUnit(closing.from))
}
