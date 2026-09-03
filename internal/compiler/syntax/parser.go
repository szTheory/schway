package syntax

import (
	"strings"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
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
}

func Parse(source []byte) ParseResult {
	tokens, diagnostics := Lex(source)
	p := parser{tokens: tokens, diagnostics: diagnostics}
	program := p.parseProgram()
	return ParseResult{Tree: Tree{Source: append([]byte(nil), source...), Tokens: tokens}, Program: program, Diagnostics: p.diagnostics}
}

func (p *parser) parseProgram() ast.Program {
	var program ast.Program
	p.expect(TokenModule, "syntax.expected_module")
	program.Module = p.modulePath()
	program.Exports = p.exports()
	for p.peek().Kind == TokenData {
		program.Data = append(program.Data, p.dataDecl())
	}
	for p.peek().Kind == TokenFn {
		program.Funcs = append(program.Funcs, p.funcDecl())
	}
	if p.peek().Kind != TokenEOF {
		p.problem("syntax.trailing_tokens", p.peek(), "unexpected tokens after declarations")
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
	for p.peek().Kind != TokenRBrace && p.peek().Kind != TokenEOF {
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
	parameterType := p.identifier("syntax.expected_parameter_type")
	p.expect(TokenRParen, "syntax.expected_rparen")
	p.expect(TokenArrow, "syntax.expected_arrow")
	returnType := p.identifier("syntax.expected_return_type")
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	body := p.matchExpr()
	end := p.expect(TokenRBrace, "syntax.expected_rbrace")
	return ast.FuncDecl{
		Name:       name.Text,
		Parameter:  ast.Parameter{Name: parameterName.Text, Type: parameterType.Text, Span: spanFrom(parameterName, parameterType)},
		ReturnType: returnType.Text,
		Body:       body,
		Span:       spanFrom(start, end),
	}
}

func (p *parser) matchExpr() ast.MatchExpr {
	start := p.expect(TokenMatch, "syntax.expected_match")
	scrutinee := p.identifier("syntax.expected_scrutinee")
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	expression := ast.MatchExpr{Scrutinee: scrutinee.Text, Span: spanFrom(start, scrutinee)}
	for p.peek().Kind != TokenRBrace && p.peek().Kind != TokenEOF {
		pattern := p.identifier("syntax.expected_pattern")
		p.expect(TokenFatArrow, "syntax.expected_fat_arrow")
		value := p.identifier("syntax.expected_value")
		expression.Arms = append(expression.Arms, ast.MatchArm{Pattern: pattern.Text, Value: value.Text, Span: spanFrom(pattern, value)})
		expression.Span.End = value.Span.End
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
	if token.Kind != TokenEOF {
		return p.advance()
	}
	return token
}

func (p *parser) identifier(code string) Token {
	return p.expect(TokenIdentifier, code)
}

func (p *parser) problem(code string, token Token, message string) {
	if len(p.diagnostics) >= 20 {
		return
	}
	p.diagnostics = append(p.diagnostics, diagnostic.Error(code, token.Span, message))
}

func spanFrom(first, last Token) diagnostic.Span {
	return diagnostic.Span{Start: first.Span.Start, End: last.Span.End}
}
