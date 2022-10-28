package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/siadat/well/syntax/ast"
	"github.com/siadat/well/syntax/scanner"
	strs_parser "github.com/siadat/well/syntax/strs/parser"
	"github.com/siadat/well/syntax/token"
)

type Parser struct {
	src     io.Reader
	root    *ast.Root
	scanner *scanner.Scanner
	debug   bool
}

type ParseError struct {
	err error
}

func (e ParseError) Error() string {
	return e.err.Error()
}

func (p *Parser) proceed() scanner.Token {
	var t, err = p.scanner.NextToken()
	p.checkErr(err)
	return t
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) init(src io.Reader) error {
	p.scanner = scanner.NewScanner(src)
	p.scanner.SetSkipWhitespace(true)
	p.scanner.SetSkipComment(true)
	p.scanner.SetDebug(p.debug)

	var _, err = p.scanner.NextToken()
	return err
}

func (p *Parser) SetDebug(debug bool) {
	p.debug = debug
	if p.scanner != nil {
		p.scanner.SetDebug(debug)
	}
}

func (p *Parser) Parse(src io.Reader) (retRoot *ast.Root, retErr error) {
	if err := p.init(src); err != nil {
		return nil, err
	}

	defer func() {
		var err = recover()
		if err, ok := err.(ParseError); ok {
			// if p.debug { debug.PrintStack() }
			retErr = err
		}
	}()

	retRoot = &ast.Root{Decls: p.parseDecls()}
	return
}

func (p *Parser) ParseExpr(src io.Reader) (retExpr ast.Expr, retErr error) {
	if err := p.init(src); err != nil {
		return nil, err
	}

	defer func() {
		var err = recover()
		if err, ok := err.(ParseError); ok {
			// if p.debug { debug.PrintStack() }
			retErr = err
		}
	}()
	retExpr = p.parseExpr(nil, token.LowestPrecedence)
	return
}

func (p *Parser) parseDecls() []ast.Decl {
	var nodes []ast.Decl
	for {
		var node = p.parseDecl()
		if node != nil {
			nodes = append(nodes, node)
		}
		if p.scanner.Eof() {
			return nodes
		}
	}
}

func (p *Parser) parseDecl() ast.Decl {
	var t = p.scanner.CurrToken()

	switch t.Lit {
	case "let":
		return p.parseLetDecl()
	case "function":
		return p.parseFuncDecl()
	}

	switch t.Typ {
	case token.NEWLINE:
		p.proceed()
		return nil
	case token.EOF:
		return nil
	default:
		panic(ParseError{fmt.Errorf("failed to parse a decl, unexpected %s", t)})
	}
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	switch t := p.scanner.CurrToken(); t.Typ {
	case token.STRING:
		p.proceed()

		return ast.String{Root: MustParseStr(t.Lit)}
	case token.IDENTIFIER:
		p.proceed()

		return ast.Ident{Name: t.Lit, Pos: t.Pos}
	case token.ADD, token.SUB:
		var op = t.Typ
		// signed expression, e.g. -1 or +value
		p.proceed()

		switch t := p.scanner.CurrToken(); t.Typ {
		case token.INTEGER,
			token.FLOAT,
			token.IDENTIFIER,
			token.LPAREN:
			return ast.UnaryExpr{
				X:  p.parsePrimaryExpr(),
				Op: op,
			}
		default:
			panic(ParseError{fmt.Errorf("expected integer or float, got %s", t)})
		}
	case token.INTEGER:
		p.proceed()

		var d, err = strconv.ParseInt(t.Lit, 10, 64)
		p.checkErr(err)
		return ast.Integer{Value: int(d)}
	case token.FLOAT:
		p.proceed()

		var f, err = strconv.ParseFloat(t.Lit, 64)
		p.checkErr(err)
		return ast.Float{Value: f}
	case token.LPAREN:
		return p.parseParenExpr()
	default:
		panic(ParseError{fmt.Errorf("failed to parse primary expression, got %s", t)})
	}
}

func (p *Parser) parseParenExpr() ast.ParenExpr {
	p.expect(token.LPAREN, "(")
	p.proceed()
	var expr = p.parseExpr(nil, token.LowestPrecedence)

	p.expect(token.RPAREN, ")")
	p.proceed()

	return ast.ParenExpr{X: expr}
}

func (p *Parser) checkErr(err error) {
	if err != nil {
		panic(ParseError{err})
	}
}

func (p *Parser) parseExpr(lhs ast.Expr, minPrec token.Precedence) ast.Expr {
	// TODO: support parsing CallExpr
	if lhs == nil {
		lhs = p.parsePrimaryExpr()
	}

	for {
		var tk = p.scanner.CurrToken()
		var prec, isStillExpr = token.Precedences[tk.Typ]
		if !isStillExpr {
			return lhs
		}
		if tk.Typ == token.EOF {
			return lhs
		}
		if prec < minPrec {
			return lhs
		}

		switch tk.Typ {
		case token.LPAREN:
			// TODO: allow multiple args
			var arg = p.parseParenExpr()
			lhs = ast.CallExpr{
				Fun: lhs,
				Arg: arg,
			}
			fmt.Println(lhs)
		default:
			// other kinds
			p.proceed()

			var rhs = p.parseExpr(nil, prec)
			lhs = ast.BinaryExpr{
				X:  lhs,
				Y:  rhs,
				Op: tk.Typ,
			}
		}
	}
}

// parseAssignExpr returns an expr that might be an AssignExpr
func (p *Parser) parseAssignExpr() ast.Expr {
	switch firstToken := p.scanner.CurrToken(); firstToken.Typ {
	case token.IDENTIFIER:
		var firstIdent = ast.Ident{Name: firstToken.Lit, Pos: firstToken.Pos}
		p.proceed()
		if t := p.scanner.CurrToken(); t.Typ == token.ASSIGN && t.Lit == "=" {
			p.proceed()
			return ast.AssignExpr{
				Name: firstIdent.Name,
				Expr: p.parseExpr(nil, token.LowestPrecedence),
			}
		} else {
			return p.parseExpr(firstIdent, token.LowestPrecedence)
		}

	default:
		return p.parseExpr(nil, token.LowestPrecedence)
	}
}

func (p *Parser) skipOptionalNewlines() error {
	for {
		var t = p.scanner.CurrToken()
		if t.Typ == token.NEWLINE {
			p.proceed()
		} else {
			return nil
		}
	}
}

func (p *Parser) expect(typ token.Token, lit string) {
	var t = p.scanner.CurrToken()
	if t.Typ == typ && t.Lit == lit {
		return
	}
	panic(ParseError{fmt.Errorf("expected %q, got %s", lit, t)})
}

func (p *Parser) expectType(typ token.Token) scanner.Token {
	var t = p.scanner.CurrToken()
	if t.Typ == typ {
		return t
	}
	panic(ParseError{fmt.Errorf("expected %s, got %s", typ, t)})
}

func (p *Parser) parseExprList() ast.ExprList {
	var list ast.ExprList

	switch t1 := p.scanner.CurrToken(); t1.Typ {
	case token.LBRACK:
		p.proceed()
		for {
			p.checkErr(p.skipOptionalNewlines())
			switch tk := p.scanner.CurrToken(); tk.Typ {
			case token.RBRACK:
				p.proceed()
				return list
			case token.EOF:
				return list
			default:
				var assign = p.parseAssignExpr()
				list.Items = append(list.Items, assign)

				switch tk := p.scanner.CurrToken(); tk.Typ {
				case token.COMMA:
					p.proceed()
				case token.NEWLINE:
					p.checkErr(p.skipOptionalNewlines())
				case token.RBRACK:
					p.proceed()
				default:
					panic(ParseError{fmt.Errorf("unexpected token %s", tk)})
				}
			}
		}
	default:
		var assign = p.parseAssignExpr()
		list.Items = append(list.Items, assign)
		return list
	}
}

func (p *Parser) parseFuncDecl() ast.Decl {
	p.expect(token.IDENTIFIER, "function")
	p.proceed()

	var name = p.expectType(token.IDENTIFIER)
	p.proceed()

	var signature = p.parseFuncSignature()

	p.expect(token.LBRACE, "{")
	p.proceed()

	p.expectType(token.NEWLINE)
	p.proceed()

	// function body
	var stmts []ast.Stmt
	for {
		var t = p.scanner.CurrToken()
		if t.Typ == token.RBRACE {
			break
		}
		if t.Typ == token.EOF {
			break
		}
		if t.Typ == token.NEWLINE {
			p.proceed()
			continue
		}
		var stmt = p.parseStmt()
		stmts = append(stmts, stmt)
	}

	p.expect(token.RBRACE, "}")
	p.proceed()

	return ast.FuncDecl{
		Name:       name.Lit,
		Signature:  signature,
		Statements: stmts,
	}
}

func (p *Parser) parseStmt() ast.Stmt {
	var t = p.scanner.CurrToken()

	switch t.Lit {
	case "let":
		return p.parseLetDecl()
	}

	switch t.Typ {
	case token.IDENTIFIER:
		return ast.ExprStmt{X: p.parseExpr(nil, token.LowestPrecedence)}
	default:
		panic(ParseError{fmt.Errorf("failed to parse a stmt, unexpected %s", t)})
	}
}

func (p *Parser) parseFuncSignature() ast.FuncSignature {
	p.expect(token.LPAREN, "(")
	p.proceed()

	var argNames []string
	var argTypes []string
	var retTypes []string
	for {
		var t = p.scanner.CurrToken()
		if t.Typ == token.RPAREN && t.Lit == ")" {
			break
		}
		// TODO: parse argNames and argTypes
	}

	p.expect(token.RPAREN, ")")
	p.proceed()

	// TODO: parse retTypes

	return ast.FuncSignature{
		ArgNames: argNames,
		ArgTypes: argTypes,
		RetTypes: retTypes,
	}
}

func MustParseStr(s string) *strs_parser.Root {
	var p = strs_parser.NewParser()
	var root, err = p.Parse(strings.NewReader(s))
	if err != nil {
		panic(ParseError{fmt.Errorf("failed to parse str: %w", err)})
	}
	return root
}

func (p *Parser) parseLetDecl() ast.Decl {
	p.expect(token.IDENTIFIER, "let")
	p.proceed()

	var name = p.expectType(token.IDENTIFIER)
	p.proceed()

	p.expect(token.ASSIGN, "=")
	p.proceed()

	var rhs = p.parseExpr(nil, token.LowestPrecedence)

	return ast.LetDecl{
		Name: name.Lit,
		Rhs:  rhs,
	}
}
