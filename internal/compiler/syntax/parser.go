package syntax

import (
	"strings"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

const (
	maxPrimaryDiagnostics = 20
	// MaxSourceBytes and MaxTokens bound every untrusted source entry point.
	MaxSourceBytes    = 1 << 20
	MaxTokens         = 1 << 17
	maxDeclarations   = 4096
	maxFunctions      = 1024
	maxAlternatives   = 4096
	maxLinearBindings = 1 << 16
)

type ParseResult struct {
	Tree        Tree
	Program     ast.Program
	Diagnostics []diagnostic.Diagnostic
}

type parser struct {
	tokens      []Token
	position    int
	diagnostics []diagnostic.Diagnostic
	truncated   bool
	typeDepth   int
	typeNodes   int
}

const (
	maxTypeDepth = 64
	maxTypeNodes = 4096
)

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
	p := parser{tokens: tokens, diagnostics: diagnostics, truncated: truncated}
	program := p.parseProgram()
	if limit := programLimit(program); limit != "" {
		p.diagnostics = []diagnostic.Diagnostic{inputLimit(len(source))}
		p.truncated = false
	}
	if p.truncated {
		p.diagnostics = append(p.diagnostics, tooManyErrors(len(source)))
	}
	return ParseResult{Tree: Tree{Source: append([]byte(nil), source...), Tokens: tokens}, Program: program, Diagnostics: p.diagnostics}
}

func programLimit(program ast.Program) string {
	if len(program.Data)+len(program.Funcs) > maxDeclarations || len(program.Funcs) > maxFunctions {
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
				if token.Span == problem.Primary && (token.Kind == TokenLAngle || token.Kind == TokenRAngle || token.Kind == TokenComma) {
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
	program.Exports = p.exports()
	for p.peek().Kind != TokenEOF {
		switch p.peek().Kind {
		case TokenData:
			program.Data = append(program.Data, p.dataDecl())
		case TokenFn:
			program.Funcs = append(program.Funcs, p.funcDecl())
		default:
			p.problem("syntax.expected_declaration", p.peek(), "expected `data` or `fn` declaration")
			p.recoverUntil(TokenData, TokenFn, TokenEOF)
		}
	}
	return program
}

func (p *parser) modulePath() string {
	parts := []string{p.identifier("syntax.expected_module_name").Text}
	for p.accept(TokenDot) {
		parts = append(parts, p.identifier("syntax.expected_module_name").Text)
	}
	return strings.Join(parts, ".")
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
		decl.Alternatives = append(decl.Alternatives, ast.Alternative{Name: alternative.Text, Span: alternative.Span})
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
	p.expect(TokenRParen, "syntax.expected_rparen")
	p.expect(TokenArrow, "syntax.expected_arrow")
	returnOrigin := p.borrowOrigin()
	returnType := p.typeRef()
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	var body ast.Body
	if p.peek().Kind == TokenMatch {
		match := p.matchExpr()
		body.MatchExpr = match
	} else {
		linear := p.linearBody()
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
	for !p.atAny(TokenRAngle, TokenRParen, TokenLBrace, TokenRBrace, TokenData, TokenFn, TokenEOF) {
		before := p.position
		result.Arguments = append(result.Arguments, p.typeRef())
		if !p.accept(TokenComma) {
			break
		}
		p.assertProgressOrBoundary(before, TokenRAngle, TokenRParen, TokenLBrace, TokenRBrace, TokenData, TokenFn, TokenEOF)
	}
	if !p.accept(TokenRAngle) {
		p.problem("syntax.expected_type_close", p.peek(), "expected `>`")
		if !p.atAny(TokenRParen, TokenLBrace, TokenRBrace, TokenData, TokenFn, TokenEOF) {
			p.advance()
		}
	}
	return result
}

func (p *parser) linearBody() ast.LinearBody {
	start := p.peek()
	body := ast.LinearBody{Span: start.Span}
	for p.peek().Kind == TokenLet {
		bindingStart := p.advance()
		name := p.identifier("syntax.expected_binding_name")
		p.expect(TokenEqual, "syntax.expected_equal")
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
		body.Bindings = append(body.Bindings, ast.Binding{
			Name: name.Text, RHS: ast.RHS{Kind: kind, Source: source.Text, Span: source.Span}, Span: spanFrom(bindingStart, source),
		})
		body.Span.End = source.Span.End
	}
	result := p.identifier("syntax.expected_linear_result")
	body.Result = result.Text
	body.Span.End = result.Span.End
	return body
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
	for !p.atAny(TokenRBrace, TokenData, TokenFn, TokenEOF) {
		startPosition := p.position
		pattern := p.identifier("syntax.expected_pattern")
		if !p.accept(TokenFatArrow) {
			p.problem("syntax.expected_fat_arrow", p.peek(), "expected `=>`")
			p.recoverUntil(TokenIdentifier, TokenRBrace, TokenData, TokenFn, TokenEOF)
			p.assertProgressOrBoundary(startPosition, TokenRBrace, TokenData, TokenFn, TokenEOF)
			continue
		}
		if len(expression.Arms) >= maxArmsPerMatch {
			p.problem("syntax.arm_limit", pattern, "match exceeds the declared arm limit")
			p.recoverUntil(TokenIdentifier, TokenRBrace, TokenData, TokenFn, TokenEOF)
			p.assertProgressOrBoundary(startPosition, TokenRBrace, TokenData, TokenFn, TokenEOF)
			continue
		}
		if p.peek().Kind == TokenLBrace {
			p.advance()
			body := p.linearBody()
			end := p.expect(TokenRBrace, "syntax.expected_rbrace")
			if pattern.Kind == TokenIdentifier {
				expression.Arms = append(expression.Arms, ast.MatchArm{Pattern: pattern.Text, Body: &body, Span: spanFrom(pattern, end)})
				expression.Span.End = end.Span.End
			}
		} else {
			value := p.identifier("syntax.expected_value")
			if pattern.Kind == TokenIdentifier && value.Kind == TokenIdentifier {
				expression.Arms = append(expression.Arms, ast.MatchArm{Pattern: pattern.Text, Value: value.Text, Span: spanFrom(pattern, value)})
				expression.Span.End = value.Span.End
			}
		}
		p.assertProgressOrBoundary(startPosition, TokenRBrace, TokenData, TokenFn, TokenEOF)
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
	p.problem(code, token, "expected `"+string(kind)+"`")
	if token.Kind != TokenEOF && !isDeclarationBoundary(token.Kind) && token.Kind != TokenRBrace {
		return p.advance()
	}
	return token
}

func (p *parser) identifier(code string) Token {
	return p.expect(TokenIdentifier, code)
}

func (p *parser) problem(code string, token Token, message string) {
	if len(p.diagnostics) >= maxPrimaryDiagnostics {
		p.truncated = true
		return
	}
	p.diagnostics = append(p.diagnostics, diagnostic.Error(code, token.Span, message))
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

// recoverUntil is the only token-skipping recovery primitive. If it is not
// already at a caller-owned boundary, it must consume at least one token.
func (p *parser) recoverUntil(boundaries ...Kind) {
	if p.atAny(boundaries...) {
		return
	}
	start := p.position
	for !p.atAny(boundaries...) {
		p.advance()
	}
	if p.position <= start {
		panic("parser recovery made no progress")
	}
}

func (p *parser) assertProgressOrBoundary(start int, boundaries ...Kind) {
	if p.position > start || p.atAny(boundaries...) {
		return
	}
	panic("parser branch made no progress")
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
