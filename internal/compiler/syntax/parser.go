package syntax

import (
	"strings"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

const maxPrimaryDiagnostics = 20

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
	tokens, diagnostics, truncated := lex(source)
	p := parser{tokens: tokens, diagnostics: diagnostics, truncated: truncated}
	program := p.parseProgram()
	if p.truncated {
		p.diagnostics = append(p.diagnostics, tooManyErrors(len(source)))
	}
	return ParseResult{Tree: Tree{Source: append([]byte(nil), source...), Tokens: tokens}, Program: program, Diagnostics: p.diagnostics}
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
		Name:       name.Text,
		Parameter:  ast.Parameter{Name: parameterName.Text, Type: parameterType, Span: parameterName.Span},
		ReturnType: returnType,
		Body:       body,
		Span:       spanFrom(start, end),
	}
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
	for !p.atAny(TokenRAngle, TokenRParen, TokenEOF) {
		before := p.position
		result.Arguments = append(result.Arguments, p.typeRef())
		if !p.accept(TokenComma) {
			break
		}
		p.assertProgressOrBoundary(before, TokenRAngle, TokenRParen, TokenEOF)
	}
	p.expect(TokenRAngle, "syntax.expected_type_close")
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
		value := p.identifier("syntax.expected_value")
		if pattern.Kind == TokenIdentifier && value.Kind == TokenIdentifier {
			expression.Arms = append(expression.Arms, ast.MatchArm{Pattern: pattern.Text, Value: value.Text, Span: spanFrom(pattern, value)})
			expression.Span.End = value.Span.End
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
