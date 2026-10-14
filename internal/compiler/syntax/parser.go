package syntax

import (
	"strings"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/surface"
)

const (
	maxPrimaryDiagnostics = 20
	// MaxSourceBytes and MaxTokens bound every untrusted source entry point.
	MaxSourceBytes = 1 << 20
	MaxTokens      = 1 << 17
	// MaxDeclarations and MaxFunctions are the per-file caps (unchanged); plan
	// 30-04 reuses them as the merged-program caps across every loaded module
	// (D-21).
	MaxDeclarations = 4096
	MaxFunctions    = 1024
	// MaxImportsPerModule bounds the import declarations in one file; the
	// comparison goes through exceeds (D-13, D-25).
	MaxImportsPerModule = 16
	maxAlternatives     = 4096
	maxLinearBindings   = 1 << 16
)

// exceedsOffByOneForTest turns exceeds' `>` into `>=`; the exact-N control row
// of TestPhase30ImportsLimitParse engages it to observe the cap's seam.
var exceedsOffByOneForTest bool

// admitModulePathFormForTest is a test-only table-driven seam ("", "slash",
// "absolute", "dotdot"). Production leaves it "". Each non-empty value makes
// exactly one rejected module-path shape parse, so each path-syntax refusal
// has its own armed control (D-25).
var admitModulePathFormForTest string

// exceeds is the single comparison behind every import-count cap.
func exceeds(n, limit int) bool {
	if exceedsOffByOneForTest {
		return n >= limit
	}
	return n > limit
}

type ParseResult struct {
	Tree        Tree
	Program     ast.Program
	Diagnostics []diagnostic.Diagnostic
}

type parser struct {
	source                    []byte
	tokens                    []Token
	position                  int
	diagnostics               []diagnostic.Diagnostic
	truncated                 bool
	typeDepth                 int
	typeNodes                 int
	claimedRecoverySpans      map[recoveryTokenSpan]struct{}
	claimedRecoveryBoundaries map[recoveryTokenSpan]struct{}
}

type recoveryTokenSpan struct {
	start int
	end   int
}

const (
	maxTypeDepth = 64
	maxTypeNodes = 4096
)

// AuthoringBounds exposes the parser's current structural caps to a named
// source-free authoring consumer without making those private checks mutable.
type AuthoringBounds struct {
	PrimaryDiagnostics int
	TypeDepth          int
	TypeNodes          int
	Alternatives       int
	LinearBindings     int
}

func CurrentAuthoringBounds() AuthoringBounds {
	return AuthoringBounds{
		PrimaryDiagnostics: maxPrimaryDiagnostics,
		TypeDepth:          maxTypeDepth,
		TypeNodes:          maxTypeNodes,
		Alternatives:       maxAlternatives,
		LinearBindings:     maxLinearBindings,
	}
}

func Parse(source []byte) ParseResult {
	if len(source) > MaxSourceBytes {
		return ParseResult{Diagnostics: []diagnostic.Diagnostic{inputLimit(MaxSourceBytes)}}
	}
	tokens, diagnostics, truncated := lex(source)
	if len(diagnostics) == 1 && diagnostics[0].Code == "syntax.input_limit" {
		return ParseResult{Tree: Tree{Source: append([]byte(nil), source...), Tokens: tokens}, Diagnostics: diagnostics}
	}
	normalizeOwnershipTokens(tokens)
	diagnostics = filterNormalizedTokenDiagnostics(diagnostics, tokens)
	if truncated && !hasUnknownToken(tokens) {
		// The Phase 1 lexer intentionally does not know provisional generic
		// punctuation. If every unknown token normalized at this parser boundary,
		// its pre-normalization diagnostic cap is not a real truncation.
		truncated = false
	}
	p := parser{
		source:                    source,
		tokens:                    tokens,
		diagnostics:               diagnostics,
		truncated:                 truncated,
		claimedRecoverySpans:      make(map[recoveryTokenSpan]struct{}),
		claimedRecoveryBoundaries: make(map[recoveryTokenSpan]struct{}),
	}
	program := p.parseProgram()
	if limit := programLimit(program); limit != "" {
		p.diagnostics = []diagnostic.Diagnostic{inputLimit(len(source))}
		p.truncated = false
	}
	if p.truncated {
		p.diagnostics = append(p.diagnostics, tooManyErrors(len(source)))
	}
	p.diagnostics = filterClaimedUnexpectedBytes(p.diagnostics, p.claimedRecoverySpans)
	return ParseResult{Tree: Tree{Source: append([]byte(nil), source...), Tokens: tokens}, Program: program, Diagnostics: p.diagnostics}
}

func filterClaimedUnexpectedBytes(diagnostics []diagnostic.Diagnostic, claimed map[recoveryTokenSpan]struct{}) []diagnostic.Diagnostic {
	if len(claimed) == 0 {
		return diagnostics
	}
	kept := diagnostics[:0]
	for _, problem := range diagnostics {
		if problem.Code == "syntax.unexpected_byte" {
			key := recoveryTokenSpan{start: problem.Primary.Start, end: problem.Primary.End}
			if _, ok := claimed[key]; ok {
				continue
			}
		}
		kept = append(kept, problem)
	}
	return kept
}

func programLimit(program ast.Program) string {
	if len(program.Data)+len(program.Funcs) > MaxDeclarations || len(program.Funcs) > MaxFunctions {
		return "declarations"
	}
	alternatives, bindings := 0, 0
	for _, declaration := range program.Data {
		alternatives += len(declaration.Alternatives)
	}
	for _, function := range program.Funcs {
		if function.Body.Linear != nil {
			bindings += len(function.Body.Linear.Bindings)
		}
	}
	if alternatives > maxAlternatives || bindings > maxLinearBindings {
		return "body facts"
	}
	return ""
}

func inputLimit(offset int) diagnostic.Diagnostic {
	return diagnostic.Error("syntax.input_limit", diagnostic.Span{Start: offset, End: offset}, "source exceeds a compiler input limit")
}

func hasUnknownToken(tokens []Token) bool {
	for _, token := range tokens {
		if token.Kind == TokenUnknown {
			return true
		}
	}
	return false
}

func filterNormalizedTokenDiagnostics(diagnostics []diagnostic.Diagnostic, tokens []Token) []diagnostic.Diagnostic {
	kept := diagnostics[:0]
	for _, problem := range diagnostics {
		if problem.Code == "syntax.unexpected_byte" {
			normalized := false
			for _, token := range tokens {
				if token.Span == problem.Primary && (token.Kind == TokenLAngle || token.Kind == TokenRAngle || token.Kind == TokenComma || (token.Kind == TokenUnknown && (token.Text == "^" || token.Text == "&"))) {
					normalized = true
					break
				}
			}
			if normalized {
				continue
			}
		}
		kept = append(kept, problem)
	}
	return kept
}

func normalizeOwnershipTokens(tokens []Token) {
	for index := range tokens {
		switch tokens[index].Text {
		case "let":
			if tokens[index].Kind == TokenIdentifier {
				tokens[index].Kind = TokenLet
			}
		case "take":
			if tokens[index].Kind == TokenIdentifier {
				tokens[index].Kind = TokenTake
			}
		case "borrow":
			if tokens[index].Kind == TokenIdentifier {
				tokens[index].Kind = TokenBorrow
			}
		case "mut":
			if tokens[index].Kind == TokenIdentifier {
				tokens[index].Kind = TokenMut
			}
		case "<":
			if tokens[index].Kind == TokenUnknown {
				tokens[index].Kind = TokenLAngle
			}
		case ">":
			if tokens[index].Kind == TokenUnknown {
				tokens[index].Kind = TokenRAngle
			}
		case ",":
			if tokens[index].Kind == TokenUnknown {
				tokens[index].Kind = TokenComma
			}
		}
	}
}

func (p *parser) parseProgram() ast.Program {
	var program ast.Program
	p.expect(TokenModule, "syntax.expected_module")
	program.Module = p.modulePath()
	limitReported := false
	addImport := func(declaration ast.ImportDecl) {
		if exceeds(len(program.Imports)+1, MaxImportsPerModule) {
			// Checked at parse time, before any resolution (D-13, D-25). The
			// excess declarations are consumed for recovery but not kept.
			if !limitReported {
				limitReported = true
				p.report(diagnostic.Error("module.imports_limit", declaration.Span, "module declares more imports than the per-module limit"))
			}
			return
		}
		program.Imports = append(program.Imports, declaration)
	}
	for p.peek().Kind == TokenImport {
		addImport(p.importDecl())
	}
	exportStart, hadExport := p.peek().Span.Start, p.peek().Kind == TokenExport
	program.Exports = p.exports()
	for p.peek().Kind != TokenEOF {
		switch p.peek().Kind {
		case TokenImport:
			declaration := p.importDecl()
			p.report(p.importAfterExport(declaration, exportStart, hadExport))
			addImport(declaration)
		case TokenData:
			program.Data = append(program.Data, p.dataDecl())
		case TokenFn:
			program.Funcs = append(program.Funcs, p.funcDecl())
		case TokenForeign:
			program.Foreign = append(program.Foreign, p.foreignBlock())
		default:
			unexpected := p.peek()
			p.advance()
			skipped, discarded := p.recoverRegion(TokenData, TokenFn, TokenForeign, TokenImport, TokenEOF)
			if discarded {
				p.problem("syntax.expected_declaration", unexpected, "expected `data` or `fn` declaration", diagnostic.Cause{Kind: "skipped_region", Span: &skipped})
			} else {
				p.problem("syntax.expected_declaration", unexpected, "expected `data` or `fn` declaration")
			}
		}
	}
	p.duplicateImports(program.Imports)
	return program
}

// report appends an already-built diagnostic under the primary-diagnostic cap.
func (p *parser) report(value diagnostic.Diagnostic) {
	if len(p.diagnostics) >= maxPrimaryDiagnostics {
		p.truncated = true
		return
	}
	p.diagnostics = append(p.diagnostics, value.AtSource())
}

// importAfterExport refuses an import written after the export block (D-02).
// When the export block was actually parsed it carries the one deterministic
// repair that rotates the declaration above `export` (D-04).
func (p *parser) importAfterExport(declaration ast.ImportDecl, exportStart int, hadExport bool) diagnostic.Diagnostic {
	const message = "import declarations must precede the export block"
	if !hadExport || exportStart > declaration.Span.Start || declaration.Span.End > len(p.source) {
		return diagnostic.Error("syntax.import_after_export", declaration.Span, message)
	}
	span := diagnostic.Span{Start: exportStart, End: declaration.Span.End}
	exportText := strings.TrimRight(string(p.source[exportStart:declaration.Span.Start]), " \t\r\n")
	replacement := string(p.source[declaration.Span.Start:declaration.Span.End]) + "\n\n" + exportText
	return diagnostic.ErrorWithRepairs("syntax.import_after_export", declaration.Span, message, nil, diagnostic.Repair{
		Kind:          "move_import_above_export",
		Detail:        declaration.Module,
		Span:          &span,
		Replacement:   replacement,
		Applicability: diagnostic.ApplicabilityMachineApplicable,
	})
}

// duplicateImports refuses a module imported twice in one file (D-04, D-19).
// The delete repair is offered only when the later item set is a subset of the
// earlier declaration's, so applying it can never drop an import.
func (p *parser) duplicateImports(imports []ast.ImportDecl) {
	first := make(map[string]int, len(imports))
	for index, declaration := range imports {
		earlierIndex, seen := first[declaration.Module]
		if !seen {
			first[declaration.Module] = index
			continue
		}
		earlier := imports[earlierIndex]
		const message = "module is imported more than once in this file"
		causes := []diagnostic.Cause{{Kind: "earlier_import", Span: &earlier.Span}}
		if !importItemsSubset(declaration.Items, earlier.Items) || declaration.Span.End > len(p.source) {
			p.report(diagnostic.Error("module.duplicate_import", declaration.Span, message, causes...))
			continue
		}
		end := declaration.Span.End
		if end < len(p.source) && p.source[end] == '\r' {
			end++
		}
		if end < len(p.source) && p.source[end] == '\n' {
			end++
		}
		span := diagnostic.Span{Start: declaration.Span.Start, End: end}
		p.report(diagnostic.ErrorWithRepairs("module.duplicate_import", declaration.Span, message, causes, diagnostic.Repair{
			Kind:          "delete_duplicate_import",
			Detail:        declaration.Module,
			Span:          &span,
			Replacement:   "\n",
			Applicability: diagnostic.ApplicabilityMachineApplicable,
		}))
	}
}

func importItemsSubset(later, earlier []ast.ImportItem) bool {
	for _, item := range later {
		found := false
		for _, candidate := range earlier {
			if candidate.Kind == item.Kind && candidate.Name == item.Name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// maxForeignSymbolsPerBlock and maxForeignPoliciesPerSymbol bound Phase 4's
// foreign declaration surface (T-04-05), sized generously above anything
// this phase's fixtures need, mirroring maxArmsPerMatch's fail-closed-above-
// the-cap shape rather than truncating silently.
const (
	maxForeignSymbolsPerBlock   = 64
	maxForeignPoliciesPerSymbol = 32
)

// foreignBlock parses one `foreign C { ... }` declaration (D-04-01). Each
// declared symbol's admission (missing unwind/nonlocal_exit policy, a call
// target that resolves to a Lang function) is check.go's job, not the
// parser's -- the grammar accepts any number of `key: value` policy lines so
// check.go can decide which keys are missing (D-04-16: no default value).
func (p *parser) foreignBlock() ast.ForeignBlock {
	start := p.expect(TokenForeign, "syntax.expected_foreign")
	language := p.identifier("syntax.expected_foreign_language")
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	block := ast.ForeignBlock{Language: language.Text, Span: spanFrom(start, language)}
	for p.peek().Kind == TokenFn {
		if len(block.Symbols) >= maxForeignSymbolsPerBlock {
			p.problem("syntax.foreign_symbol_limit", p.peek(), "foreign block exceeds the declared symbol limit")
			p.recoverUntil(TokenRBrace, TokenData, TokenFn, TokenForeign, TokenEOF)
			break
		}
		block.Symbols = append(block.Symbols, p.foreignSymbol())
	}
	end := p.expect(TokenRBrace, "syntax.expected_rbrace")
	block.Span.End = end.Span.End
	return block
}

func (p *parser) foreignSymbol() ast.ForeignSymbol {
	start := p.expect(TokenFn, "syntax.expected_fn")
	name := p.identifier("syntax.expected_function_name")
	p.expect(TokenLParen, "syntax.expected_lparen")
	parameterName := p.identifier("syntax.expected_parameter_name")
	p.expect(TokenColon, "syntax.expected_colon")
	parameterType := p.typeRef()
	p.refuseExtraDeclarationParameter()
	p.expect(TokenRParen, "syntax.expected_rparen")
	p.expect(TokenArrow, "syntax.expected_arrow")
	returnType := p.typeRef()
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	var policies []ast.ForeignPolicy
	for p.peek().Kind == TokenIdentifier {
		if len(policies) >= maxForeignPoliciesPerSymbol {
			p.problem("syntax.foreign_policy_limit", p.peek(), "foreign symbol exceeds the declared policy limit")
			p.recoverUntil(TokenRBrace, TokenData, TokenFn, TokenForeign, TokenEOF)
			break
		}
		key := p.advance()
		p.expect(TokenColon, "syntax.expected_colon")
		value := p.peek()
		isString := value.Kind == TokenString
		if isString || value.Kind == TokenIdentifier {
			p.advance()
		} else {
			p.problem("syntax.expected_foreign_policy_value", value, "expected an identifier or string policy value")
		}
		text := value.Text
		if isString && len(text) >= 2 {
			text = text[1 : len(text)-1]
		}
		policies = append(policies, ast.ForeignPolicy{Key: key.Text, Value: text, IsString: isString, Span: spanFrom(key, value)})
	}
	end := p.expect(TokenRBrace, "syntax.expected_rbrace")
	return ast.ForeignSymbol{
		Name:       name.Text,
		Parameter:  ast.Parameter{Name: parameterName.Text, Type: parameterType, Span: parameterName.Span},
		ReturnType: returnType,
		Policies:   policies,
		Span:       spanFrom(start, end),
	}
}

func (p *parser) modulePath() string {
	text, _ := p.modulePathSpan()
	return text
}

// modulePathSpan parses a dotted module path and also returns the byte extent
// of the tokens it consumed.
func (p *parser) modulePathSpan() (string, diagnostic.Span) {
	opening := p.peek()
	prefix := ""
	if admitModulePathFormForTest == "dotdot" {
		for p.peek().Kind == TokenDot {
			prefix += "."
			p.advance()
		}
	}
	first := p.identifier("syntax.expected_module_name")
	last := first
	parts := []string{prefix + first.Text}
	for p.accept(TokenDot) {
		last = p.identifier("syntax.expected_module_name")
		parts = append(parts, last.Text)
	}
	return strings.Join(parts, "."), spanFrom(opening, last)
}

// importDecl parses one `import a.b { fn x  fn y }` declaration (D-02):
// importDecl := "import" modulePath "{" ("fn" IDENT | "type" IDENT)+ "}".
// Items are whitespace/newline separated; `type` items are kept so check
// refuses them by name.
func (p *parser) importDecl() ast.ImportDecl {
	start := p.expect(TokenImport, "syntax.expected_import")
	module, moduleSpan := p.modulePathSpan()
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	var items []ast.ImportItem
	reported := false
	for !p.atAny(TokenRBrace, TokenData, TokenEOF, TokenExport, TokenImport, TokenForeign) {
		if p.peek().Kind == TokenFn && p.looksLikeFunctionDecl() {
			break
		}
		kind := p.peek()
		if kind.Kind != TokenType && kind.Kind != TokenFn {
			// Commas, `as`, `*` and every other stray token: items are
			// whitespace separated `fn`/`type` names and nothing else.
			p.problem("syntax.expected_import_name", kind, "expected `fn` or `type` followed by a name in import block")
			reported = true
			p.advance()
			continue
		}
		p.advance()
		name := p.identifier("syntax.expected_import_name")
		if name.Kind != TokenIdentifier {
			reported = true
			continue
		}
		items = append(items, ast.ImportItem{Kind: kind.Text, Name: name.Text, Span: spanFrom(kind, name)})
	}
	if len(items) == 0 && !reported {
		p.problem("syntax.expected_import_name", p.peek(), "import block needs at least one `fn` item")
	}
	end := p.expect(TokenRBrace, "syntax.expected_rbrace")
	return ast.ImportDecl{Module: module, Items: items, Span: spanFrom(start, end), ModuleSpan: moduleSpan}
}

func (p *parser) exports() []ast.Export {
	p.expect(TokenExport, "syntax.expected_export")
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	var exports []ast.Export
	for !p.atAny(TokenRBrace, TokenData, TokenEOF) {
		if p.peek().Kind == TokenFn && p.looksLikeFunctionDecl() {
			break
		}
		kind := p.peek()
		if kind.Kind != TokenType && kind.Kind != TokenFn {
			p.problem("syntax.expected_export_kind", kind, "expected `type` or `fn` in export block")
			p.advance()
			continue
		}
		p.advance()
		name := p.identifier("syntax.expected_export_name")
		exports = append(exports, ast.Export{Kind: kind.Text, Name: name.Text, Span: spanFrom(kind, name)})
	}
	p.expect(TokenRBrace, "syntax.expected_rbrace")
	return exports
}

func (p *parser) dataDecl() ast.DataDecl {
	start := p.expect(TokenData, "syntax.expected_data")
	name := p.identifier("syntax.expected_type_name")
	p.expect(TokenEqual, "syntax.expected_equal")
	decl := ast.DataDecl{Name: name.Text, Span: spanFrom(start, name)}
	for p.accept(TokenPipe) {
		alternative := p.identifier("syntax.expected_alternative")
		payloadType := p.optionalPayloadBinder()
		decl.Alternatives = append(decl.Alternatives, ast.Alternative{Name: alternative.Text, PayloadType: payloadType, Span: alternative.Span})
		decl.Span.End = alternative.Span.End
	}
	if len(decl.Alternatives) == 0 {
		p.problem("syntax.expected_alternative", p.peek(), "data declaration needs at least one alternative")
	}
	return decl
}

func (p *parser) funcDecl() ast.FuncDecl {
	start := p.expect(TokenFn, "syntax.expected_fn")
	name := p.identifier("syntax.expected_function_name")
	p.expect(TokenLParen, "syntax.expected_lparen")
	parameterName := p.identifier("syntax.expected_parameter_name")
	p.expect(TokenColon, "syntax.expected_colon")
	parameterType := p.typeRef()
	p.refuseExtraDeclarationParameter()
	p.expect(TokenRParen, "syntax.expected_rparen")
	p.expect(TokenArrow, "syntax.expected_arrow")
	returnOrigin := p.borrowOrigin()
	returnType := p.typeRef()
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	var body ast.Body
	if p.peek().Kind == TokenMatch {
		match := p.matchExpr()
		body.MatchExpr = match
	} else if p.scalarBodyStart(parameterType, returnType) {
		linear := p.scalarLinearBody(returnType.Constructor == "Unit")
		body.Linear = &linear
	} else {
		linear := p.linearBody(true)
		body.Linear = &linear
	}
	end := p.expect(TokenRBrace, "syntax.expected_rbrace")
	return ast.FuncDecl{
		Name:         name.Text,
		Parameter:    ast.Parameter{Name: parameterName.Text, Type: parameterType, Span: parameterName.Span},
		ReturnOrigin: returnOrigin,
		ReturnType:   returnType,
		Body:         body,
		Span:         spanFrom(start, end),
	}
}

// scalarLetStart distinguishes scalar value bindings from the generic `let`
// form, whose right-hand side may be a call or an ownership expression. The
// scalar grammar only claims a literal/place with an optional scalar operator.
func (p *parser) scalarLetStart() bool {
	if p.peekNonTrivia(0).Kind != TokenLet || p.peekNonTrivia(1).Kind != TokenIdentifier || p.peekNonTrivia(2).Kind != TokenEqual {
		return false
	}
	value := p.peekNonTrivia(3)
	if value.Kind == TokenUnknown && value.Text == "!" {
		return p.peekNonTrivia(4).Kind == TokenIdentifier
	}
	if value.Kind != TokenNumber && value.Kind != TokenIdentifier && value.Kind != TokenLParen {
		return false
	}
	next := p.peekNonTrivia(4)
	if value.Kind == TokenIdentifier && (value.Text == "byte_len" || value.Text == "byte_at") && next.Kind == TokenLParen {
		return true
	}
	_, precedence, _ := p.scalarOperatorAt(4)
	if precedence != 0 {
		return true
	}
	return value.Kind == TokenLParen || (next.Kind != TokenLParen && next.Kind != TokenDot)
}

func (p *parser) scalarBodyStart(parameterType, returnType ast.TypeRef) bool {
	if p.atAny(TokenVar, TokenIf, TokenWhile) {
		return true
	}
	if (parameterType.Constructor == "U64" || parameterType.Constructor == "ByteView") && (returnType.Constructor == "U64" || returnType.Constructor == "Unit") && p.scalarLetStart() {
		return true
	}
	return p.isWriteCallStart(p.peek())
}

func (p *parser) isWriteCallStart(token Token) bool {
	if token.Kind != TokenIdentifier || p.peekNonTrivia(1).Kind != TokenLParen {
		return false
	}
	return token.Text == "write_text" || token.Text == "write_u64" || token.Text == "write_newline"
}

func (p *parser) scalarLinearBody(unitResult bool) ast.LinearBody {
	body := ast.LinearBody{Span: p.peek().Span}
	body.Statements = p.scalarStatements(false)
	if unitResult && p.peek().Kind == TokenRBrace {
		if len(body.Statements) > 0 {
			body.Span.End = body.Statements[len(body.Statements)-1].Span.End
		}
	} else if p.peek().Kind == TokenIdentifier {
		result := p.advance()
		body.Result = result.Text
		body.ResultSpan = result.Span
		body.Span.End = result.Span.End
		if !p.atAny(TokenRBrace, TokenEOF, TokenData, TokenFn, TokenForeign, TokenImport) {
			p.problem("syntax.expected_rbrace", p.peek(), "expected `}` after the scalar result")
			p.recoverDelimited(TokenRBrace, TokenData, TokenFn, TokenForeign, TokenImport, TokenEOF)
			if isCallerOwnedBoundary(p.peek().Kind) {
				p.claimRecoveryBoundary(p.peek())
			}
		}
	} else {
		boundary := p.peek()
		p.problemAtBoundary("syntax.expected_linear_result", boundary, "expected a scalar result")
		if !p.atAny(TokenRBrace, TokenEOF, TokenData, TokenFn, TokenForeign, TokenImport) {
			p.recoverDelimited(TokenRBrace, TokenData, TokenFn, TokenForeign, TokenImport, TokenEOF)
			if isCallerOwnedBoundary(p.peek().Kind) {
				p.claimRecoveryBoundary(p.peek())
			}
		}
	}
	return body
}

func (p *parser) scalarStatements(inBlock bool) []ast.Statement {
	statements := []ast.Statement{}
	for p.peek().Kind != TokenEOF && p.peek().Kind != TokenRBrace {
		start := p.peek()
		if !inBlock && start.Kind == TokenIdentifier && p.peekNonTrivia(1).Kind != TokenEqual && !p.isWriteCallStart(start) {
			return statements
		}
		switch start.Kind {
		case TokenLet:
			p.advance()
			name := p.identifier("syntax.expected_binding_name")
			p.expect(TokenEqual, "syntax.expected_equal")
			expr := p.scalarExpr()
			statements = append(statements, ast.Statement{Kind: "let", Name: name.Text, Expr: expr, Span: diagnostic.Span{Start: start.Span.Start, End: expr.Span.End}})
		case TokenVar:
			p.advance()
			name := p.identifier("syntax.expected_binding_name")
			p.expect(TokenEqual, "syntax.expected_equal")
			expr := p.scalarExpr()
			statements = append(statements, ast.Statement{Kind: "var", Name: name.Text, Expr: expr, Span: diagnostic.Span{Start: start.Span.Start, End: expr.Span.End}})
		case TokenIdentifier:
			if p.isWriteCallStart(start) {
				statements = append(statements, p.scalarWriteStatement())
				continue
			}
			name := p.advance()
			p.expect(TokenEqual, "syntax.expected_equal")
			expr := p.scalarExpr()
			statements = append(statements, ast.Statement{Kind: "assign", Name: name.Text, Expr: expr, Span: diagnostic.Span{Start: name.Span.Start, End: expr.Span.End}})
		case TokenIf, TokenWhile:
			p.advance()
			expr := p.scalarExpr()
			thenStart := p.expect(TokenLBrace, "syntax.expected_lbrace")
			thenBody := p.scalarStatements(true)
			end := p.expect(TokenRBrace, "syntax.expected_rbrace")
			statement := ast.Statement{Kind: map[Kind]string{TokenIf: "if", TokenWhile: "while"}[start.Kind], Expr: expr, Then: thenBody, Span: diagnostic.Span{Start: start.Span.Start, End: end.Span.End}}
			if start.Kind == TokenIf {
				if p.accept(TokenElse) {
					p.expect(TokenLBrace, "syntax.expected_lbrace")
					statement.Else = p.scalarStatements(true)
					elseEnd := p.expect(TokenRBrace, "syntax.expected_rbrace")
					statement.Span.End = elseEnd.Span.End
				} else if thenStart.Kind == TokenLBrace && end.Kind == TokenRBrace &&
					!p.atAny(TokenEOF, TokenData, TokenFn, TokenForeign, TokenImport) {
					p.reportSurfaceRefusal("surface.unsupported_if_without_else", Token{
						Kind: TokenUnknown,
						Span: diagnostic.Span{Start: start.Span.Start, End: end.Span.End},
					})
				} else {
					p.problem("syntax.expected_else", p.peek(), "if requires an else block")
				}
			}
			statements = append(statements, statement)
		case TokenDefect:
			p.advance()
			reason := p.expect(TokenString, "syntax.expected_defect_reason")
			value := reason.Text
			if len(value) >= 2 {
				value = value[1 : len(value)-1]
			}
			statements = append(statements, ast.Statement{Kind: "defect", Value: value, Span: diagnostic.Span{Start: start.Span.Start, End: reason.Span.End}})
		default:
			if inBlock {
				return statements
			}
			return statements
		}
	}
	return statements
}

func (p *parser) scalarWriteStatement() ast.Statement {
	name := p.advance()
	p.expect(TokenLParen, "syntax.expected_lparen")
	statement := ast.Statement{Kind: name.Text}
	if name.Text == "write_u64" || name.Text == "write_text" {
		if p.peek().Kind != TokenRParen {
			statement.Expr = p.scalarExpr()
		} else {
			if name.Text == "write_text" {
				p.problem("syntax.write_text_arity", p.peek(), "write_text requires one quoted text literal")
			} else {
				p.problem("syntax.write_u64_arity", p.peek(), "write_u64 requires one U64 operand")
			}
		}
	} else if p.peek().Kind != TokenRParen {
		argument := p.peek()
		for !p.atAny(TokenRParen, TokenEOF) {
			p.advance()
		}
		p.problem("syntax.write_newline_arity", argument, "write_newline takes no operands")
	}
	end := p.expect(TokenRParen, "syntax.expected_rparen")
	statement.Span = diagnostic.Span{Start: name.Span.Start, End: end.Span.End}
	if statement.Expr != nil {
		statement.Span.End = end.Span.End
	}
	return statement
}

func (p *parser) scalarExpr() *ast.Expr {
	return p.scalarBinaryExpr(1)
}

func scalarBinaryOperator(operator Token) int {
	switch operator.Kind {
	case TokenLAngle, TokenEqualEqual:
		return 1
	case TokenPipe:
		return 2
	case TokenUnknown:
		switch operator.Text {
		case "^":
			return 3
		case "&":
			return 4
		case "<<", ">>":
			return 5
		}
	case TokenPlus:
		return 6
	case TokenPercent:
		return 7
	}
	return 0
}

func (p *parser) scalarOperatorAt(offset int) (Token, int, int) {
	operator := p.peekNonTrivia(offset)
	if operator.Kind == TokenPipe {
		second := p.peekNonTrivia(offset + 1)
		if second.Kind == TokenPipe && operator.Span.End == second.Span.Start {
			return operator, 0, 1
		}
	}
	if operator.Kind == TokenLAngle || operator.Kind == TokenRAngle {
		second := p.peekNonTrivia(offset + 1)
		if second.Kind == operator.Kind && operator.Span.End == second.Span.Start {
			combined := Token{
				Kind: TokenUnknown,
				Text: operator.Text + second.Text,
				Span: diagnostic.Span{Start: operator.Span.Start, End: second.Span.End},
			}
			if combined.Text == "<<" || combined.Text == ">>" {
				return combined, scalarBinaryOperator(combined), 2
			}
		}
	}
	return operator, scalarBinaryOperator(operator), 1
}

func (p *parser) scalarBinaryExpr(minPrecedence int) *ast.Expr {
	left := p.scalarExprAtom()
	for {
		if operator, code, ok := p.consumeUnsupportedBinaryOperator(); ok {
			p.reportSurfaceRefusal(code, operator)
			if !isCallerOwnedBoundary(p.peek().Kind) {
				p.scalarBinaryExpr(minPrecedence)
			}
			return left
		}
		operator, precedence, width := p.scalarOperatorAt(0)
		if precedence < minPrecedence || precedence == 0 {
			return left
		}
		for range width {
			p.advance()
		}
		var right *ast.Expr
		if (operator.Text == "<<" || operator.Text == ">>") && p.peek().Kind == TokenUnknown && p.peek().Text == "-" {
			minus := p.advance()
			amount := p.peek()
			end := minus.Span.End
			value := "-"
			if amount.Kind == TokenNumber {
				amount = p.advance()
				value += amount.Text
				end = amount.Span.End
			}
			p.reportSurfaceRefusal("surface.unsupported_shift_amount", minus)
			right = &ast.Expr{Kind: "number", Value: value, Span: diagnostic.Span{Start: minus.Span.Start, End: end}}
		} else {
			right = p.scalarBinaryExpr(precedence + 1)
		}
		kind := ""
		switch {
		case operator.Kind == TokenUnknown && operator.Text == "^":
			kind = "bit_xor"
		case operator.Kind == TokenUnknown && operator.Text == "&":
			kind = "bit_and"
		case operator.Kind == TokenUnknown && operator.Text == "<<":
			kind = "shift_left"
		case operator.Kind == TokenUnknown && operator.Text == ">>":
			kind = "shift_right"
		case operator.Kind == TokenPipe:
			kind = "bit_or"
		default:
			switch operator.Kind {
			case TokenPlus:
				kind = "add"
			case TokenLAngle:
				kind = "less"
			case TokenEqualEqual:
				kind = "equal"
			case TokenPercent:
				kind = "remainder"
			}
		}
		left = &ast.Expr{Kind: kind, Left: left, Right: right, Span: diagnostic.Span{Start: left.Span.Start, End: right.Span.End}}
	}
}

func (p *parser) consumeUnsupportedBinaryOperator() (Token, string, bool) {
	operator := p.peek()
	if operator.Kind == TokenUnknown {
		switch operator.Text {
		case "&&":
			p.advance()
			return operator, "surface.unsupported_logical_and", true
		case "-":
			p.advance()
			return operator, "surface.unsupported_subtraction", true
		}
	}
	if operator.Kind == TokenPipe {
		second := p.peekNonTrivia(1)
		if second.Kind == TokenPipe && operator.Span.End == second.Span.Start {
			p.advance()
			p.advance()
			return Token{
				Kind: TokenUnknown,
				Text: "||",
				Span: diagnostic.Span{Start: operator.Span.Start, End: second.Span.End},
			}, "surface.unsupported_logical_or", true
		}
	}
	return Token{}, "", false
}

func (p *parser) consumeUnsupportedLogicalNot() bool {
	operator := p.peek()
	if operator.Kind != TokenUnknown || operator.Text != "!" || p.peekNonTrivia(1).Kind != TokenIdentifier {
		return false
	}
	p.advance()
	p.reportSurfaceRefusal("surface.unsupported_logical_not", operator)
	return true
}

func (p *parser) reportSurfaceRefusal(code string, token Token) {
	refusal, ok := surface.LookupRefusal(code)
	if !ok {
		panic("syntax parser requested an unknown surface refusal: " + code)
	}
	p.problem(refusal.Code, token, refusal.Hint)
}

func (p *parser) scalarExprAtom() *ast.Expr {
	leftToken := p.peek()
	if p.consumeUnsupportedLogicalNot() {
		return p.scalarExprAtom()
	}
	if isCallerOwnedBoundary(leftToken.Kind) {
		p.problemAtBoundary("syntax.expected_scalar_operand", leftToken, "expected a scalar operand")
		return &ast.Expr{Kind: "name", Name: leftToken.Text, Span: leftToken.Span}
	}
	leftToken = p.advance()
	switch leftToken.Kind {
	case TokenLParen:
		expression := p.scalarExpr()
		end := p.expect(TokenRParen, "syntax.expected_rparen")
		expression.Span = diagnostic.Span{Start: leftToken.Span.Start, End: end.Span.End}
		return expression
	case TokenNumber:
		return &ast.Expr{Kind: "number", Value: leftToken.Text, Span: leftToken.Span}
	case TokenString:
		return &ast.Expr{Kind: "text", Value: leftToken.Text, Span: leftToken.Span}
	default:
		if leftToken.Kind != TokenIdentifier {
			p.problem("syntax.expected_scalar_operand", leftToken, "expected a scalar operand")
		}
		if leftToken.Kind == TokenIdentifier && p.peek().Kind == TokenLParen {
			switch leftToken.Text {
			case "byte_len":
				p.advance()
				if p.peek().Kind != TokenIdentifier {
					p.byteReadArity("syntax.byte_len_arity", "byte_len requires one byte view operand")
					return &ast.Expr{Kind: "byte_len", Span: leftToken.Span}
				}
				view := p.advance()
				if p.peek().Kind != TokenRParen {
					p.byteReadArity("syntax.byte_len_arity", "byte_len requires one byte view operand")
					return &ast.Expr{Kind: "byte_len", Name: view.Text, Span: diagnostic.Span{Start: leftToken.Span.Start, End: view.Span.End}}
				}
				end := p.advance()
				return &ast.Expr{Kind: "byte_len", Name: view.Text, Span: diagnostic.Span{Start: leftToken.Span.Start, End: end.Span.End}}
			case "byte_at":
				p.advance()
				const message = "byte_at requires a byte view and one U64 index"
				if p.peek().Kind != TokenIdentifier {
					p.byteReadArity("syntax.byte_at_arity", message)
					return &ast.Expr{Kind: "byte_at", Span: leftToken.Span}
				}
				view := p.advance()
				if p.peek().Kind != TokenComma {
					p.byteReadArity("syntax.byte_at_arity", message)
					return &ast.Expr{Kind: "byte_at", Name: view.Text, Span: diagnostic.Span{Start: leftToken.Span.Start, End: view.Span.End}}
				}
				p.advance()
				if p.peek().Kind == TokenRParen {
					p.byteReadArity("syntax.byte_at_arity", message)
					return &ast.Expr{Kind: "byte_at", Name: view.Text, Span: diagnostic.Span{Start: leftToken.Span.Start, End: view.Span.End}}
				}
				index := p.scalarExpr()
				if p.peek().Kind != TokenRParen {
					p.byteReadArity("syntax.byte_at_arity", message)
					return &ast.Expr{Kind: "byte_at", Name: view.Text, Right: index, Span: diagnostic.Span{Start: leftToken.Span.Start, End: index.Span.End}}
				}
				end := p.advance()
				return &ast.Expr{Kind: "byte_at", Name: view.Text, Right: index, Span: diagnostic.Span{Start: leftToken.Span.Start, End: end.Span.End}}
			}
		}
		return &ast.Expr{Kind: "name", Name: leftToken.Text, Span: leftToken.Span}
	}
}

// byteReadArity reports a named byte-read arity refusal at the current token,
// then recovers to and consumes the closing paren so cascades stay minimal.
func (p *parser) byteReadArity(code, message string) {
	p.problem(code, p.peek(), message)
	for !p.atAny(TokenRParen, TokenRBrace, TokenEOF) {
		p.advance()
	}
	p.accept(TokenRParen)
}

// borrowOrigin parses the optional Phase 3 return-type annotation preceding
// an ordinary return TypeRef: `borrow(path)` (shared access) or
// `borrow mut(path)` (exclusive access). Its presence is the syntactic
// discriminant for the borrowed-view return case (OWN-04) — resolved here,
// before sameType ever runs, so every other return-type spelling takes the
// unchanged identity path. A `borrow` token with no following `(path)` is a
// parse-level rejection (syntax.expected_lparen / syntax.expected_origin_path),
// not an inference: a borrowed return with no declared origin never reaches
// check.go as a legal program.
func (p *parser) borrowOrigin() *ast.BorrowOrigin {
	if p.peek().Kind != TokenBorrow {
		return nil
	}
	start := p.advance()
	access := "shared"
	if p.accept(TokenMut) {
		access = "exclusive"
	}
	p.expect(TokenLParen, "syntax.expected_lparen")
	path := p.identifier("syntax.expected_origin_path")
	end := p.expect(TokenRParen, "syntax.expected_rparen")
	return &ast.BorrowOrigin{Path: path.Text, Access: access, Span: spanFrom(start, end)}
}

func (p *parser) typeRef() ast.TypeRef {
	start := p.identifier("syntax.expected_type")
	p.typeDepth++
	p.typeNodes++
	defer func() { p.typeDepth-- }()
	result := ast.TypeRef{Constructor: start.Text}
	if p.typeDepth > maxTypeDepth || p.typeNodes > maxTypeNodes {
		p.problem("syntax.type_limit", start, "type expression exceeds parser limits")
		return result
	}
	if !p.accept(TokenLAngle) {
		return result
	}
	for !p.atAny(TokenRAngle, TokenRParen, TokenLBrace, TokenRBrace, TokenData, TokenFn, TokenEOF, TokenArrow, TokenFatArrow, TokenElse) {
		before := p.position
		result.Arguments = append(result.Arguments, p.typeRef())
		if !p.accept(TokenComma) {
			break
		}
		p.assertProgressOrBoundary(before, TokenRAngle, TokenRParen, TokenLBrace, TokenRBrace, TokenData, TokenFn, TokenEOF, TokenArrow, TokenFatArrow, TokenElse)
	}
	if !p.accept(TokenRAngle) {
		p.problem("syntax.expected_type_close", p.peek(), "expected `>`")
		if !isCallerOwnedBoundary(p.peek().Kind) {
			p.advance()
		}
	}
	return result
}

func (p *parser) linearBody(allowTerminalMatch bool) ast.LinearBody {
	start := p.peek()
	body := ast.LinearBody{Span: start.Span}
	for p.peek().Kind == TokenLet || p.peek().Kind == TokenDiscard {
		if p.peek().Kind == TokenDiscard {
			binding, end := p.discardBecause()
			body.Bindings = append(body.Bindings, binding)
			body.Span.End = end
			continue
		}
		bindingStart := p.advance()
		name := p.identifier("syntax.expected_binding_name")
		p.expect(TokenEqual, "syntax.expected_equal")
		if p.peek().Kind == TokenTry {
			binding, end := p.tryCallBinding(bindingStart, name)
			body.Bindings = append(body.Bindings, binding)
			body.Span.End = end
			continue
		}
		if number := p.peek(); number.Kind == TokenNumber {
			p.advance()
			body.Bindings = append(body.Bindings, ast.Binding{
				Name: name.Text,
				RHS:  ast.RHS{Kind: "numeric_literal", Source: number.Text, Span: number.Span},
				Span: diagnostic.Span{Start: bindingStart.Span.Start, End: number.Span.End},
			})
			body.Span.End = number.Span.End
			continue
		}
		kind := "read"
		if p.accept(TokenTake) {
			kind = "take"
		} else if p.accept(TokenBorrow) {
			kind = "borrow"
			if p.accept(TokenMut) {
				kind = "borrow_mut"
			}
		}
		source := p.identifier("syntax.expected_binding_source")
		if p.peek().Kind == TokenLParen {
			// D-07-01/D-07-40: a bare call's callee identity (Lang function,
			// foreign symbol, or unresolved) is a check-time fact, not a
			// parse-time one. The parser used to unconditionally refuse this
			// shape with syntax.fallible_call_not_consumed -- that refusal
			// is still correct for a foreign callee, but it has relocated to
			// check, where foreignSymbols and functionNames are both
			// available. The parser now accepts the shape unconditionally
			// and emits an ast.RHS{Kind: "call"} binding for check to admit
			// or refuse.
			arguments, argumentSpans, end := p.callArguments()
			// 13-01 Task 1 (Rule 1 bugfix): Binding.Span for a call binding
			// must cover the WHOLE "let NAME = callee(ARGS)" statement,
			// exactly like every other binding kind's Span already does
			// (spanFrom(bindingStart, source) below, where `source` is the
			// bound expression's own last token) -- but here `source` is
			// the CALLEE token, parsed before the "(" is even seen, so
			// spanFrom(bindingStart, source) previously stopped short of
			// the arguments and closing paren. `end` (the closing paren's
			// own End, already computed by callArguments and already used
			// two lines below for body.Span.End) is the correct bound. This
			// was latent until check.go's own ":stmt" span-widening (same
			// plan) started reading Binding.Span for call bindings; no
			// existing pinned diagnostic ID depends on the old, truncated
			// value (verified by a full `go test ./...` run before and
			// after this change).
			body.Bindings = append(body.Bindings, ast.Binding{
				Name: name.Text, RHS: ast.RHS{Kind: "call", Callee: source.Text, Arguments: arguments, ArgumentSpans: argumentSpans, Span: source.Span}, Span: diagnostic.Span{Start: bindingStart.Span.Start, End: end},
			})
			body.Span.End = end
			continue
		}
		body.Bindings = append(body.Bindings, ast.Binding{
			Name: name.Text, RHS: ast.RHS{Kind: kind, Source: source.Text, Span: source.Span}, Span: spanFrom(bindingStart, source),
		})
		body.Span.End = source.Span.End
	}
	if allowTerminalMatch && p.peek().Kind == TokenMatch {
		match := p.matchExpr()
		body.TerminalMatch = &match
		body.Span.End = match.Span.End
		return body
	}
	if p.peek().Kind == TokenDefect {
		// D-04-15: `defect "<reason>"` is the other admissible terminal form
		// alongside a bare Result identifier -- a real, reachable, abort-only
		// terminal outcome needing no call surface. The reason is a required
		// non-empty string literal, exactly like discard's own rationale.
		start := p.advance()
		reasonToken := p.expect(TokenString, "syntax.expected_defect_reason")
		reason := reasonToken.Text
		if len(reason) >= 2 && reasonToken.Kind == TokenString {
			reason = reason[1 : len(reason)-1]
		}
		if reason == "" {
			p.problem("syntax.defect_reason_empty", reasonToken, "defect requires a non-empty reason string")
		}
		body.DefectReason = reason
		end := reasonToken.Span.End
		if start.Kind != TokenDefect {
			end = start.Span.End
		}
		body.Span.End = end
		return body
	}
	if p.consumeUnsupportedLogicalNot() {
		if !isCallerOwnedBoundary(p.peek().Kind) {
			p.scalarBinaryExpr(1)
		}
		return body
	}
	result := p.peek()
	if result.Kind != TokenIdentifier {
		p.problemAtBoundary("syntax.expected_linear_result", result, "expected `identifier`")
		if !isCallerOwnedBoundary(result.Kind) {
			p.recoverDelimited(TokenRBrace, TokenData, TokenFn, TokenForeign, TokenImport, TokenEOF)
			if isCallerOwnedBoundary(p.peek().Kind) {
				p.claimRecoveryBoundary(p.peek())
			}
		}
		return body
	}
	p.advance()
	body.Result = result.Text
	body.ResultSpan = result.Span
	body.Span.End = result.Span.End
	if operator, code, ok := p.consumeUnsupportedBinaryOperator(); ok {
		p.reportSurfaceRefusal(code, operator)
		if !isCallerOwnedBoundary(p.peek().Kind) {
			p.scalarBinaryExpr(1)
		}
	}
	return body
}

// refuseExtraDeclarationParameter consumes the first valid extra parameter
// and refuses its name-through-type span. The declaration's closing paren
// remains owned by funcDecl/foreignSymbol, and malformed parameter lists keep
// their existing expectation diagnostics.
func (p *parser) refuseExtraDeclarationParameter() {
	if p.peek().Kind != TokenComma ||
		p.peekNonTrivia(1).Kind != TokenIdentifier ||
		p.peekNonTrivia(2).Kind != TokenColon ||
		p.peekNonTrivia(3).Kind != TokenIdentifier {
		return
	}
	p.advance() // comma after the declared parameter
	name := p.advance()
	p.advance() // colon before the extra parameter type
	p.typeRef()
	last := p.tokens[p.position-1]
	span := diagnostic.Span{Start: name.Span.Start, End: last.Span.End}
	p.reportSurfaceRefusal("surface.unsupported_parameter_arity", Token{Kind: TokenUnknown, Span: span})
	if !p.atAny(TokenRParen, TokenEOF, TokenData, TokenFn, TokenForeign, TokenImport) {
		p.recoverDelimited(TokenRParen, TokenData, TokenFn, TokenForeign, TokenImport, TokenEOF)
	}
}

// tryCallBinding parses `try <callee>(<arg>, ...)` following a binding's `=`
// (D-04-06's sole admissible fallible-call consumer this phase).
func (p *parser) tryCallBinding(bindingStart, name Token) (ast.Binding, int) {
	tryToken := p.advance()
	callee := p.identifier("syntax.expected_foreign_callee")
	arguments, argumentSpans, end := p.callArguments()
	rhs := ast.RHS{Kind: "try_call", Callee: callee.Text, Arguments: arguments, ArgumentSpans: argumentSpans, Span: spanFrom(tryToken, callee)}
	return ast.Binding{Name: name.Text, RHS: rhs, Span: diagnostic.Span{Start: bindingStart.Span.Start, End: end}}, end
}

// discardBecause parses `discard <callee>(<arg>, ...) because "<rationale>"`
// (D-04-06's second and only other admissible fallible-call consumer). The
// rationale is a required non-empty string literal; an empty or absent
// rationale is a parse-level rejection, so the core IR never has to encode a
// discard whose rationale is missing.
func (p *parser) discardBecause() (ast.Binding, int) {
	start := p.expect(TokenDiscard, "syntax.expected_discard")
	callee := p.identifier("syntax.expected_foreign_callee")
	arguments, argumentSpans, _ := p.callArguments()
	because := p.expect(TokenBecause, "syntax.expected_because")
	rationaleToken := p.expect(TokenString, "syntax.expected_discard_rationale")
	rationale := rationaleToken.Text
	if len(rationale) >= 2 && rationaleToken.Kind == TokenString {
		rationale = rationale[1 : len(rationale)-1]
	}
	if rationale == "" {
		p.problem("syntax.discard_rationale_empty", rationaleToken, "discard ... because requires a non-empty rationale string")
	}
	end := rationaleToken.Span.End
	if because.Kind != TokenBecause {
		end = because.Span.End
	}
	rhs := ast.RHS{Kind: "discard_call", Callee: callee.Text, Arguments: arguments, ArgumentSpans: argumentSpans, Rationale: rationale, Span: spanFrom(start, callee)}
	return ast.Binding{Name: "", RHS: rhs, Span: diagnostic.Span{Start: start.Span.Start, End: end}}, end
}

// callArguments parses `(arg, arg, ...)`, where each argument is a bare
// identifier naming an already-bound place, returning the argument names in
// source order and the end offset of the closing paren.
func (p *parser) callArguments() ([]string, []diagnostic.Span, int) {
	p.expect(TokenLParen, "syntax.expected_lparen")
	var arguments []string
	var argumentSpans []diagnostic.Span
	if p.peek().Kind != TokenRParen {
		for {
			argument := p.peek()
			if argument.Kind != TokenIdentifier {
				p.problem("syntax.expected_call_argument", argument, "expected an identifier call argument")
				if !isCallerOwnedBoundary(argument.Kind) {
					p.recoverDelimited(TokenRParen, TokenRBrace, TokenData, TokenFn, TokenForeign, TokenImport, TokenEOF)
					if isCallerOwnedBoundary(p.peek().Kind) {
						p.claimRecoveryBoundary(p.peek())
					}
				}
				break
			}
			p.advance()
			arguments = append(arguments, argument.Text)
			argumentSpans = append(argumentSpans, argument.Span)
			if !p.accept(TokenComma) {
				break
			}
		}
	}
	end := p.expect(TokenRParen, "syntax.expected_rparen")
	return arguments, argumentSpans, end.Span.End
}

// maxArmsPerMatch bounds T-03-01's CFG-shape denial-of-service surface at the
// syntax layer, sized generously above any fixture this phase ships (single
// digits of alternatives) while still rejecting fail-closed rather than
// admitting an unbounded arm list.
const maxArmsPerMatch = 64

func (p *parser) matchExpr() ast.MatchExpr {
	start := p.expect(TokenMatch, "syntax.expected_match")
	scrutinee := p.identifier("syntax.expected_scrutinee")
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	expression := ast.MatchExpr{Scrutinee: scrutinee.Text, Span: spanFrom(start, scrutinee)}
	for !p.atAny(TokenRBrace, TokenData, TokenFn, TokenEOF, TokenRParen, TokenArrow, TokenFatArrow, TokenElse) {
		startPosition := p.position
		pattern := p.identifier("syntax.expected_pattern")
		patternBinder := p.optionalPayloadBinder()
		if !p.accept(TokenFatArrow) {
			p.problem("syntax.expected_fat_arrow", p.peek(), "expected `=>`")
			p.recoverUntil(TokenIdentifier, TokenRBrace, TokenData, TokenFn, TokenEOF, TokenRParen, TokenArrow, TokenFatArrow, TokenElse)
			p.assertProgressOrBoundary(startPosition, TokenRBrace, TokenData, TokenFn, TokenEOF, TokenRParen, TokenArrow, TokenFatArrow, TokenElse)
			continue
		}
		if len(expression.Arms) >= maxArmsPerMatch {
			p.problem("syntax.arm_limit", pattern, "match exceeds the declared arm limit")
			p.recoverUntil(TokenIdentifier, TokenRBrace, TokenData, TokenFn, TokenEOF, TokenRParen, TokenArrow, TokenFatArrow, TokenElse)
			p.assertProgressOrBoundary(startPosition, TokenRBrace, TokenData, TokenFn, TokenEOF, TokenRParen, TokenArrow, TokenFatArrow, TokenElse)
			continue
		}
		if p.peek().Kind == TokenLBrace {
			p.advance()
			body := p.linearBody(false)
			end := p.expect(TokenRBrace, "syntax.expected_rbrace")
			if pattern.Kind == TokenIdentifier {
				expression.Arms = append(expression.Arms, ast.MatchArm{Pattern: pattern.Text, Body: &body, Binder: patternBinder, Span: spanFrom(pattern, end)})
				expression.Span.End = end.Span.End
			}
		} else {
			value := p.identifier("syntax.expected_value")
			// D-12-11: the arm-value position's own optional parenthesized
			// place name is the construction site (`Ok(v)`). Per this
			// language's move-on-bind semantics (D-12-14), the constructed
			// argument and the pattern's own destructuring binder are the
			// same declared place this phase -- but Plan 03 (D-12-15's
			// third refusal) keeps the two names SEPARATE here rather than
			// having one silently override the other: check.go's
			// analyzePayloadArm is the single place that decides whether a
			// differing pattern-side/value-side name is the arity-mismatch
			// refusal or (when they agree, or one is empty) the one shared
			// place D-12-14 describes.
			valueBinder := p.optionalPayloadBinder()
			if pattern.Kind == TokenIdentifier && value.Kind == TokenIdentifier {
				expression.Arms = append(expression.Arms, ast.MatchArm{Pattern: pattern.Text, Value: value.Text, Binder: patternBinder, ConstructBinder: valueBinder, Span: spanFrom(pattern, value)})
				expression.Span.End = value.Span.End
			}
		}
		p.assertProgressOrBoundary(startPosition, TokenRBrace, TokenData, TokenFn, TokenEOF, TokenRParen, TokenArrow, TokenFatArrow, TokenElse)
	}
	p.expect(TokenRBrace, "syntax.expected_rbrace")
	return expression
}

func (p *parser) peek() Token {
	for p.position < len(p.tokens) && p.tokens[p.position].Trivia() {
		p.position++
	}
	if p.position >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.position]
}

func (p *parser) advance() Token {
	token := p.peek()
	if p.position < len(p.tokens)-1 {
		p.position++
	}
	return token
}

func (p *parser) accept(kind Kind) bool {
	if p.peek().Kind != kind {
		return false
	}
	p.advance()
	return true
}

func (p *parser) expect(kind Kind, code string) Token {
	token := p.peek()
	if token.Kind == kind {
		return p.advance()
	}
	if isCallerOwnedBoundary(token.Kind) {
		p.problemAtBoundary(code, token, "expected `"+string(kind)+"`")
	} else {
		p.problem(code, token, "expected `"+string(kind)+"`")
		if kind == TokenRParen || kind == TokenRBrace {
			p.recoverDelimited(kind, TokenData, TokenFn, TokenForeign, TokenImport, TokenEOF)
			if p.peek().Kind == kind {
				return p.advance()
			}
			if isCallerOwnedBoundary(p.peek().Kind) {
				p.claimRecoveryBoundary(p.peek())
			}
			return p.peek()
		}
	}
	if !isCallerOwnedBoundary(token.Kind) {
		return p.advance()
	}
	return token
}

func (p *parser) identifier(code string) Token {
	return p.expect(TokenIdentifier, code)
}

// optionalPayloadBinder parses an optional "(identifier)" immediately
// following a pattern or alternative name (D-12-11/D-12-13): shared by a
// data declaration's payload type, a match arm's destructuring binder, and
// a match arm's construction argument. Returns "" when no parenthesis
// follows (the nullary case) -- never consumed, so callers see zero token
// movement. Reports syntax.expected_binder when the parentheses are empty
// and syntax.expected_rparen when unclosed, using the same p.problem shape
// every other named refusal in this file uses.
func (p *parser) optionalPayloadBinder() string {
	if !p.accept(TokenLParen) {
		return ""
	}
	if p.peek().Kind != TokenIdentifier {
		p.problem("syntax.expected_binder", p.peek(), "expected an identifier inside `(...)`")
		p.accept(TokenRParen)
		return ""
	}
	name := p.advance()
	if !p.accept(TokenRParen) {
		p.problem("syntax.expected_rparen", p.peek(), "expected `)`")
	}
	return name.Text
}

func (p *parser) problem(code string, token Token, message string, causes ...diagnostic.Cause) {
	if len(p.diagnostics) >= maxPrimaryDiagnostics {
		p.truncated = true
		return
	}
	p.claimRecoveryToken(token)
	p.diagnostics = append(p.diagnostics, diagnostic.Error(code, token.Span, message, causes...).AtSource())
}

// problemAtBoundary gives one parser expectation ownership of an exact
// caller-owned token. Follow-up expectations at the same byte span are
// consequences of that recovery boundary and do not create a second cause.
func (p *parser) problemAtBoundary(code string, token Token, message string, causes ...diagnostic.Cause) {
	key := recoveryTokenSpan{start: token.Span.Start, end: token.Span.End}
	if _, owned := p.claimedRecoveryBoundaries[key]; owned {
		return
	}
	before := len(p.diagnostics)
	p.problem(code, token, message, causes...)
	if len(p.diagnostics) > before {
		p.claimRecoveryBoundary(token)
	}
}

func (p *parser) claimRecoveryBoundary(token Token) {
	span := recoveryTokenSpan{start: token.Span.Start, end: token.Span.End}
	p.claimedRecoveryBoundaries[span] = struct{}{}
}

func (p *parser) claimRecoveryToken(token Token) {
	if token.Kind != TokenUnknown {
		return
	}
	span := recoveryTokenSpan{start: token.Span.Start, end: token.Span.End}
	p.claimedRecoverySpans[span] = struct{}{}
}

func (p *parser) atAny(kinds ...Kind) bool {
	current := p.peek().Kind
	for _, kind := range kinds {
		if current == kind {
			return true
		}
	}
	return false
}

func (p *parser) looksLikeFunctionDecl() bool {
	return p.peekNonTrivia(0).Kind == TokenFn &&
		p.peekNonTrivia(1).Kind == TokenIdentifier &&
		p.peekNonTrivia(2).Kind == TokenLParen
}

func (p *parser) peekNonTrivia(ordinal int) Token {
	seen := 0
	for index := p.position; index < len(p.tokens); index++ {
		if p.tokens[index].Trivia() {
			continue
		}
		if seen == ordinal {
			return p.tokens[index]
		}
		seen++
	}
	return p.tokens[len(p.tokens)-1]
}

// recoverUntil is the only token-skipping recovery primitive. It delegates to
// delimiter-aware recovery so a nested closer cannot be mistaken for the
// caller's boundary.
func (p *parser) recoverUntil(boundaries ...Kind) {
	p.recoverDelimited(boundaries...)
}

// recoverDelimited scans the bounded token stream with a small delimiter
// stack. Requested boundaries stop recovery only at depth zero; matching
// nested closers are consumed, while unmatched closers and valid declaration
// siblings remain available to their caller.
func (p *parser) recoverDelimited(boundaries ...Kind) (diagnostic.Span, bool) {
	if p.atAny(boundaries...) || p.peek().Kind == TokenEOF {
		return diagnostic.Span{}, false
	}
	startPosition := p.position
	startSpan := p.peek().Span
	var last Token
	stack := make([]Kind, 0, 4)
	for {
		token := p.peek()
		if token.Kind == TokenEOF {
			break
		}
		if len(stack) == 0 && p.atAny(boundaries...) {
			break
		}
		if isDeclarationRecoveryBoundary(p, token.Kind) {
			break
		}
		stopped := false
		switch token.Kind {
		case TokenRParen, TokenRBrace:
			if len(stack) == 0 || stack[len(stack)-1] != token.Kind {
				stopped = true
			} else {
				stack = stack[:len(stack)-1]
				last = p.advance()
			}
		case TokenLParen:
			stack = append(stack, TokenRParen)
			last = p.advance()
		case TokenLBrace:
			stack = append(stack, TokenRBrace)
			last = p.advance()
		default:
			last = p.advance()
		}
		if stopped {
			break
		}
	}
	if p.position <= startPosition {
		return diagnostic.Span{}, false
	}
	return diagnostic.Span{Start: startSpan.Start, End: last.Span.End}, true
}

// recoverRegion additionally returns the exact byte extent consumed, for use
// as a skipped_region diagnostic cause.
func (p *parser) recoverRegion(boundaries ...Kind) (diagnostic.Span, bool) {
	return p.recoverDelimited(boundaries...)
}

func (p *parser) assertProgressOrBoundary(start int, boundaries ...Kind) {
	if p.position > start || p.atAny(boundaries...) {
		return
	}
	panic("parser branch made no progress")
}

func isCallerOwnedBoundary(kind Kind) bool {
	switch kind {
	case TokenEOF, TokenData, TokenFn, TokenForeign, TokenImport, TokenRParen, TokenRBrace,
		TokenArrow, TokenFatArrow, TokenElse:
		return true
	default:
		return false
	}
}

func isDeclarationRecoveryBoundary(p *parser, kind Kind) bool {
	switch kind {
	case TokenData, TokenForeign, TokenImport:
		return true
	case TokenFn:
		return p.looksLikeFunctionDecl()
	default:
		return false
	}
}

func isDeclarationBoundary(kind Kind) bool {
	return kind == TokenData || kind == TokenFn
}

func tooManyErrors(offset int) diagnostic.Diagnostic {
	return diagnostic.Error(
		"syntax.too_many_errors",
		diagnostic.Span{Start: offset, End: offset},
		"additional syntax diagnostics were suppressed",
	)
}

func spanFrom(first, last Token) diagnostic.Span {
	return diagnostic.Span{Start: first.Span.Start, End: last.Span.End}
}
