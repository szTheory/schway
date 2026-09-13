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
		case TokenForeign:
			program.Foreign = append(program.Foreign, p.foreignBlock())
		default:
			p.problem("syntax.expected_declaration", p.peek(), "expected `data` or `fn` declaration")
			p.recoverUntil(TokenData, TokenFn, TokenForeign, TokenEOF)
		}
	}
	return program
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
			arguments, end := p.callArguments()
			body.Bindings = append(body.Bindings, ast.Binding{
				Name: name.Text, RHS: ast.RHS{Kind: "call", Callee: source.Text, Arguments: arguments, Span: source.Span}, Span: spanFrom(bindingStart, source),
			})
			body.Span.End = end
			continue
		}
		body.Bindings = append(body.Bindings, ast.Binding{
			Name: name.Text, RHS: ast.RHS{Kind: kind, Source: source.Text, Span: source.Span}, Span: spanFrom(bindingStart, source),
		})
		body.Span.End = source.Span.End
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
	result := p.identifier("syntax.expected_linear_result")
	body.Result = result.Text
	body.Span.End = result.Span.End
	return body
}

// tryCallBinding parses `try <callee>(<arg>, ...)` following a binding's `=`
// (D-04-06's sole admissible fallible-call consumer this phase).
func (p *parser) tryCallBinding(bindingStart, name Token) (ast.Binding, int) {
	tryToken := p.advance()
	callee := p.identifier("syntax.expected_foreign_callee")
	arguments, end := p.callArguments()
	rhs := ast.RHS{Kind: "try_call", Callee: callee.Text, Arguments: arguments, Span: spanFrom(tryToken, callee)}
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
	arguments, _ := p.callArguments()
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
	rhs := ast.RHS{Kind: "discard_call", Callee: callee.Text, Arguments: arguments, Rationale: rationale, Span: spanFrom(start, callee)}
	return ast.Binding{Name: "", RHS: rhs, Span: diagnostic.Span{Start: start.Span.Start, End: end}}, end
}

// callArguments parses `(arg, arg, ...)`, where each argument is a bare
// identifier naming an already-bound place, returning the argument names in
// source order and the end offset of the closing paren.
func (p *parser) callArguments() ([]string, int) {
	p.expect(TokenLParen, "syntax.expected_lparen")
	var arguments []string
	if p.peek().Kind != TokenRParen {
		for {
			argument := p.identifier("syntax.expected_call_argument")
			arguments = append(arguments, argument.Text)
			if !p.accept(TokenComma) {
				break
			}
		}
	}
	end := p.expect(TokenRParen, "syntax.expected_rparen")
	return arguments, end.Span.End
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
		binder := p.optionalPayloadBinder()
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
				expression.Arms = append(expression.Arms, ast.MatchArm{Pattern: pattern.Text, Body: &body, Binder: binder, Span: spanFrom(pattern, end)})
				expression.Span.End = end.Span.End
			}
		} else {
			value := p.identifier("syntax.expected_value")
			// D-12-11: the arm-value position's own optional parenthesized
			// place name is the construction site (`Ok(v)`). Per this
			// language's move-on-bind semantics (D-12-14), the constructed
			// argument and the pattern's own destructuring binder are the
			// same declared place this phase (there is no second place
			// available to construct from), so a non-empty value-side
			// binder overrides (and, in every legal program, simply
			// restates) the pattern-side binder captured above.
			valueBinder := p.optionalPayloadBinder()
			if valueBinder != "" {
				binder = valueBinder
			}
			if pattern.Kind == TokenIdentifier && value.Kind == TokenIdentifier {
				expression.Arms = append(expression.Arms, ast.MatchArm{Pattern: pattern.Text, Value: value.Text, Binder: binder, Span: spanFrom(pattern, value)})
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
