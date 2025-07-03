package parser

import (
	"github.com/conneroisu/gix/internal/types"
	"github.com/conneroisu/gix/pkg/lexer"
)

// parseIf parses if-then-else expressions.
func (p *Parser) parseIf() types.Expr {
	p.advance() // skip 'if'

	cond := p.parseExpression(precedenceLowest)

	if !p.expectPeek(lexer.TOKEN_THEN) {
		return nil
	}

	p.advance()
	then := p.parseExpression(precedenceLowest)

	if !p.expectPeek(lexer.TOKEN_ELSE) {
		return nil
	}

	p.advance()
	elseExpr := p.parseExpression(precedenceLowest)

	return &types.IfExpr{
		Cond: cond,
		Then: then,
		Else: elseExpr,
	}
}

// parseLet parses let expressions.
func (p *Parser) parseLet() types.Expr {
	p.advance() // skip 'let'

	let := &types.LetExpr{
		Bindings: []types.Binding{},
	}

	// Parse bindings
	for !p.curIs(lexer.TOKEN_IN) && !p.curIs(lexer.TOKEN_EOF) {
		if !p.curIs(lexer.TOKEN_IDENT) {
			p.errors.Addf(p.cur.Line, p.cur.Column,
				"expected identifier in let binding, got %v", p.cur.Type)

			return nil
		}

		name := p.cur.Literal

		if !p.expectPeek(lexer.TOKEN_ASSIGN) {
			return nil
		}

		p.advance()
		value := p.parseExpression(precedenceLowest)

		let.Bindings = append(let.Bindings, types.Binding{
			Name:  name,
			Value: value,
		})

		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}

		p.advance() // position on next token
	}

	if !p.curIs(lexer.TOKEN_IN) {
		p.errors.Addf(p.cur.Line, p.cur.Column,
			"expected 'in' after let bindings, got %v", p.cur.Type)

		return nil
	}

	p.advance()
	let.Body = p.parseExpression(precedenceLowest)

	return let
}

// parseWith parses with expressions.
func (p *Parser) parseWith() types.Expr {
	p.advance() // skip 'with'

	expr := p.parseExpression(precedenceLowest)

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	p.advance()
	body := p.parseExpression(precedenceLowest)

	return &types.WithExpr{
		Expr: expr,
		Body: body,
	}
}

// parseAssert parses assert expressions.
func (p *Parser) parseAssert() types.Expr {
	p.advance() // skip 'assert'

	cond := p.parseExpression(precedenceLowest)

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	p.advance()
	body := p.parseExpression(precedenceLowest)

	return &types.AssertExpr{
		Cond: cond,
		Body: body,
	}
}
