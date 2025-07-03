package parser

import (
	"github.com/conneroisu/gix/internal/types"
	"github.com/conneroisu/gix/pkg/lexer"
)

// parseUnary parses unary expressions.
func (p *Parser) parseUnary(op types.UnaryOp) types.Expr {
	p.advance()
	expr := p.parseExpression(precedenceCall)

	return &types.UnaryExpr{
		Op:   op,
		Expr: expr,
	}
}

// parseBinary parses binary expressions.
func (p *Parser) parseBinary(left types.Expr, op types.BinaryOp) types.Expr {
	precedence := p.curPrecedence()
	p.advance()
	right := p.parseExpression(precedence)

	return &types.BinaryExpr{
		Left:  left,
		Op:    op,
		Right: right,
	}
}

// parseGrouped parses parenthesized expressions.
func (p *Parser) parseGrouped() types.Expr {
	p.advance() // skip '('

	expr := p.parseExpression(precedenceLowest)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return expr
}

// parseFunction parses function definitions.
func (p *Parser) parseFunction() types.Expr {
	param := p.cur.Literal

	if !p.expectPeek(lexer.TOKEN_COLON) {
		return nil
	}

	p.advance()
	body := p.parseExpression(precedenceLowest)

	return &types.FunctionExpr{
		Param: param,
		Body:  body,
	}
}

// parseFunctionApplication parses function applications.
func (p *Parser) parseFunctionApplication(fn types.Expr) types.Expr {
	arg := p.parseExpression(precedenceCall)

	return &types.ApplyExpr{
		Func: fn,
		Arg:  arg,
	}
}

// parseList parses list literals.
func (p *Parser) parseList() types.Expr {
	p.advance() // skip '['

	list := &types.ListExpr{
		Elements: []types.Expr{},
	}

	if p.curIs(lexer.TOKEN_RBRACKET) {
		return list
	}

	// Parse first element
	list.Elements = append(list.Elements, p.parseExpression(precedenceCall+1))

	// Parse remaining elements
	for !p.peekIs(lexer.TOKEN_RBRACKET) && !p.peekIs(lexer.TOKEN_EOF) {
		p.advance()
		if p.curIs(lexer.TOKEN_RBRACKET) {
			break
		}
		// Skip commas if present (for compatibility)
		if p.curIs(lexer.TOKEN_COMMA) {
			p.advance()
		}
		if p.curIs(lexer.TOKEN_RBRACKET) {
			break
		}
		list.Elements = append(list.Elements, p.parseExpression(precedenceCall+1))
	}

	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}

	return list
}

// parseAttrSet parses attribute set literals.
func (p *Parser) parseAttrSet() types.Expr {
	p.advance() // skip '{'

	attrs := &types.AttrSetExpr{
		Bindings: []types.AttrBinding{},
	}

	// Check for recursive attribute set
	if p.curIs(lexer.TOKEN_REC) {
		attrs.Recursive = true
		p.advance()
	}

	// Empty attribute set
	if p.curIs(lexer.TOKEN_RBRACE) {
		return attrs
	}

	// Parse bindings
	for !p.curIs(lexer.TOKEN_RBRACE) && !p.curIs(lexer.TOKEN_EOF) {
		if p.curIs(lexer.TOKEN_INHERIT) {
			p.parseInherit(attrs)
		} else {
			binding := p.parseBinding()
			if binding != nil {
				attrs.Bindings = append(attrs.Bindings, *binding)
			}
		}

		if p.curIs(lexer.TOKEN_RBRACE) {
			break
		}
	}

	if !p.curIs(lexer.TOKEN_RBRACE) {
		p.errors.Addf(p.cur.Line, p.cur.Column,
			"expected '}', got %v", p.cur.Type)

		return nil
	}

	return attrs
}

// parseBinding parses a single attribute binding.
func (p *Parser) parseBinding() *types.AttrBinding {
	// Parse attribute path
	path := p.parseAttrPath()
	if path == nil {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_ASSIGN) {
		return nil
	}

	p.advance()
	value := p.parseExpression(precedenceLowest)

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	p.advance() // position on next token

	return &types.AttrBinding{
		Path:  path,
		Value: value,
	}
}

// parseAttrPath parses an attribute path.
func (p *Parser) parseAttrPath() []string {
	var path []string

	if !p.curIs(lexer.TOKEN_IDENT) && !p.curIs(lexer.TOKEN_STRING) {
		p.errors.Addf(p.cur.Line, p.cur.Column,
			"expected identifier or string, got %v", p.cur.Type)

		return nil
	}

	path = append(path, p.cur.Literal)

	for p.peekIs(lexer.TOKEN_DOT) {
		p.advance() // consume dot
		p.advance() // get next part

		if !p.curIs(lexer.TOKEN_IDENT) && !p.curIs(lexer.TOKEN_STRING) {
			p.errors.Addf(p.cur.Line, p.cur.Column,
				"expected identifier or string after dot, got %v", p.cur.Type)

			return nil
		}

		path = append(path, p.cur.Literal)
	}

	return path
}

// parseInherit parses inherit statements.
func (p *Parser) parseInherit(attrs *types.AttrSetExpr) {
	p.advance() // skip 'inherit'

	// TODO: Implement full inherit parsing
	// For now, skip to semicolon
	for !p.curIs(lexer.TOKEN_SEMICOLON) && !p.curIs(lexer.TOKEN_EOF) {
		p.advance()
	}

	if p.curIs(lexer.TOKEN_SEMICOLON) {
		p.advance()
	}
}

// parseSelect parses attribute selection.
func (p *Parser) parseSelect(expr types.Expr) types.Expr {
	p.advance() // consume dot

	path := p.parseAttrPath()
	if path == nil {
		return nil
	}

	return &types.SelectExpr{
		Expr:     expr,
		AttrPath: path,
	}
}

// parseHasAttr parses attribute existence test.
func (p *Parser) parseHasAttr(expr types.Expr) types.Expr {
	p.advance() // consume '?'

	path := p.parseAttrPath()
	if path == nil {
		return nil
	}

	return &types.HasAttrExpr{
		Expr:     expr,
		AttrPath: path,
	}
}

// parseOrDefault parses 'or' default expressions.
func (p *Parser) parseOrDefault(expr types.Expr) types.Expr {
	selectExpr, ok := expr.(*types.SelectExpr)
	if !ok {
		p.errors.Addf(p.cur.Line, p.cur.Column,
			"'or' can only be used with attribute selection")

		return nil
	}

	p.advance()
	selectExpr.Default = p.parseExpression(precedenceLowest)

	return selectExpr
}
