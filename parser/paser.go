package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/thenam153/conditions-go/ast"
	lerrors "github.com/thenam153/conditions-go/errors"
	"github.com/thenam153/conditions-go/token"

	"github.com/itchyny/gojq"
)

var (
	SCAN_VERSION = 1
)

type ParserInterface interface {
	Parse() (ast.Expr, error)
}

type Parser struct {
	s    scanner.Scanner
	mbuf mbuffer
	buf  buffer
}

// Multi-buffer parser
type mbuffer struct {
	toks []rune
	tts  []string
	fbu  bool // From buffer
}

// Buffer parser
type buffer struct {
	tok rune
	tt  string
	fbu bool // From buffer
}

func NewParser(src io.Reader) ParserInterface {
	p := &Parser{s: scanner.Scanner{}, buf: buffer{}}
	p.s.Mode = scanner.ScanStrings | scanner.ScanFloats | scanner.ScanIdents
	p.s.Init(src)
	return p
}

func (p *Parser) scan() (rune, string) {
	if SCAN_VERSION == 1 {
		return p.scannerScan()
	}
	return p.scannerMScanSingle()
}

func (p *Parser) unscan() {
	if SCAN_VERSION == 1 {
		p.scannerUnScan()
	} else {
		p.scannerMUnScan()
	}
}

func (p *Parser) scannerScan() (rune, string) {
	if !p.buf.fbu {
		p.buf.tok, p.buf.tt = p.s.Scan(), p.s.TokenText()
	} else {
		p.buf.fbu = false
	}
	return p.buf.tok, p.buf.tt
}

func (p *Parser) scannerUnScan() {
	p.buf.fbu = true
}

func (p *Parser) scannerMScan() (rune, string) {
	if p.mbuf.fbu {
		if len(p.mbuf.tts) > 0 {
			t, tt := p.mbuf.toks[0], p.mbuf.tts[0]
			return t, tt
		}
	}
	t, tt := p.s.Scan(), p.s.TokenText()
	p.mbuf.toks, p.mbuf.tts = append(p.mbuf.toks, t), append(p.mbuf.tts, tt)
	return t, tt
}

func (p *Parser) scannerMCommit() {
	if len(p.mbuf.tts) > 0 {
		p.mbuf.toks = p.mbuf.toks[1:]
		p.mbuf.tts = p.mbuf.tts[1:]
	}
}

func (p *Parser) scannerMCommitAll() {
	p.mbuf.toks = make([]rune, 0)
	p.mbuf.tts = make([]string, 0)
}

func (p *Parser) scannerMUnScan() {
	p.mbuf.fbu = true
}

func (p *Parser) scannerMScanSingle() (rune, string) {
	if len(p.mbuf.tts) > 1 {
		p.scannerMCommit()
	}
	t, tt := p.scannerMScan()
	if p.mbuf.fbu {
		p.mbuf.fbu = false
	}
	return t, tt
}

// Scan token from input string reader
func (p *Parser) scanToken() (token.Token, string) {
	var (
		t   rune
		tt  string
		tok token.Token
	)
	// Get token and text token
	t, tt = p.scan()
	switch t {
	case scanner.EOF:
		tok = token.EOF
	case '(':
		tok = token.LPAREN
	case ')':
		tok = token.RPAREN
	case '-':
		t, tt = p.scan()
		if strings.Contains(tt, "-") {
			tok = token.ILLEGAL
			break
		}
		if t == scanner.Float || t == scanner.Int {
			tok = token.NUMBER
			tt = "-" + tt
			break
		}
		tok = token.ILLEGAL
	case scanner.Float, scanner.Int:
		tok = token.NUMBER
	case '$':
		var (
			err  error
			mode string
		)
		tok = token.ILLEGAL
		t, tt = p.scan()
		if t == scanner.Int {
			tt = "$" + tt
			tok = token.IDENT
			break
		}
		if t == scanner.Ident && strings.ToUpper(tt) == token.Tokens[token.JQ] {
			tt, mode, err = p.scanJQ()
			if err == nil {
				tok = token.JQ
			}
			buildJQMsg := func(tt, mode string) string {
				jsMsg := ast.JQMsg{
					TextToken: tt,
					Mode:      mode,
				}
				bytes, _ := json.Marshal(jsMsg)
				return string(bytes)
			}
			tt = buildJQMsg(tt, mode)
		}
	case '!':
		t, tt = p.scan()
		switch t {
		case '=':
			tt = "!="
			tok = token.NEQ
		case '~':
			tt = "!~"
			tok = token.NEREG
		default:
			tok = token.ILLEGAL
		}
	case '>':
		t, tt = p.scan()
		if t == '=' {
			tt = ">" + tt
			tok = token.GTE
		} else {
			tt = ">"
			tok = token.GT
			p.unscan()
		}
	case '<':
		t, tt = p.scan()
		if t == '=' {
			tt = "<" + tt
			tok = token.LTE
		} else {
			tt = "<"
			tok = token.LT
			p.unscan()
		}
	case '=':
		t, tt = p.scan()
		switch t {
		case '=':
			tt = "=="
			tok = token.EQ
		case '~':
			tt = "=~"
			tok = token.EREG
		default:
			tok = token.ILLEGAL
		}
	case '/':
		for {
			_t, _tt := p.scan()
			tt += _tt
			if _t == '/' {
				tok = token.STRING
				break
			}
			if _t == scanner.EOF {
				tok = token.ILLEGAL
				break
			}
		}
	case scanner.String:
		tok = token.STRING
	case scanner.Ident:
		ttU := strings.ToUpper(tt)
		switch ttU {
		case "AND":
			tok = token.AND
		case "NAND":
			tok = token.NAND
		case "OR":
			tok = token.OR
		case "XOR":
			tok = token.XOR
		case "NOT":
			_, _tt := p.scan()
			switch strings.ToUpper(_tt) {
			case "IN":
				tt = fmt.Sprintf("%v %v", tt, _tt)
				tok = token.NOTIN
			case "TRUE":
				tt = fmt.Sprintf("%v %v", tt, _tt)
				tok = token.FALSE
			case "FALSE":
				tt = fmt.Sprintf("%v %v", tt, _tt)
				tok = token.TRUE
			default:
				p.unscan()
				tok = token.ILLEGAL
			}
		case "IN":
			tok = token.IN
		case "TRUE":
			tok = token.TRUE
		case "FALSE":
			tok = token.FALSE
		default:
			tok = token.ILLEGAL
		}
	case '[':
		tok = token.ILLEGAL
		var (
			err error
		)
		tt, err = p.scanArgs()
		if err == nil {
			tok = token.IDENT
			break
		}
		_tt, err := p.scanArray()
		if err == nil {
			tok = token.ARRAY
		}
		tt += _tt
	default:
		tok = token.STRING
	}
	return tok, tt
}

func (p *Parser) scanArgs() (string, error) {
	var (
		tt  string
		sep string
	)
	// Example: [foo][bar] => foo.bar
	for {
		_, _tt := p.scan()
		tt += sep + _tt
		if tt == "@" {
			continue
		}
		_t, _ := p.scan()
		if _t != ']' {
			p.unscan()
			return tt, lerrors.New("Unexpected character, missing ']'")
		}
		__t, _ := p.scan()
		if __t != '[' {
			p.unscan()
			return tt, nil
		}
		sep = "."
	}
}

func (p *Parser) scanArray() (string, error) {
	var (
		tt string
	)
	for {
		_t, _tt := p.scan()
		if _t == ']' {
			return tt, nil
		}
		if _t == scanner.EOF {
			p.unscan()
			return "", lerrors.New("Unexpected character, missing ']'")
		}
		tt += _tt
	}
}

// Extract JQ (goJQ) query from string input
//
//	"$jq[first](.request.number)"
//
// # JQ query from string input
//
// ".request.number": JQ query
// "first": JQ mode
func (p *Parser) scanJQ() (string, string, error) {
	var (
		tt            string
		m             int    = 1
		oldWhitespace uint64 = p.s.Whitespace
	)
	scanJQMode := func() (string, error) {
		t, _ := p.scan()
		if t != '[' {
			p.unscan()
			return "", nil
		}
		t, mode := p.scan()
		if t == ']' {
			return "", nil
		}
		t, _ = p.scan()
		if t != ']' {
			return "", lerrors.New("Unexpected character, missing ']'")
		}
		return mode, nil

	}
	mode, err := scanJQMode()
	if err != nil {
		return "", "", lerrors.NewWrap("Cannot scan JQ mode", err)
	}
	// Reset whitespace scanner to original
	resetWhitespace := func() {
		p.s.Whitespace = oldWhitespace
	}
	// Remove all config whitespace of scanner to scan all characters from this scan
	p.s.Whitespace = 0
	defer resetWhitespace()
	t, _ := p.scan()
	if t != '(' {
		p.unscan()
		return "", mode, lerrors.New("Unexpected character, missing '('")
	}
	// Extract content jq query from input
	for {
		_t, _tt := p.scan()
		if _t == ')' {
			m--
		} else if _t == '(' {
			m++
		} else if _t == scanner.EOF {
			return "", mode, lerrors.New("End of input")
		}
		if m == 0 {
			return tt, mode, nil
		}
		tt += _tt
	}
}

// Extract unary expression from string input
//
//	"\"foo\" in [\"bar\", \"baz\"]"
//
// Unary expression will be:
//
//	"foo": StringLiteral
//	"in": Operator
//	["bar", "baz"]:  SliceStringLiteral
func (p *Parser) parseUnaryExpr() (ast.Expr, error) {
	tok, lit := p.scanToken()
	if tok == token.LPAREN {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, lerrors.NewWrap("Cannot scan paren expression", err)
		}
		if tok, _ := p.scanToken(); tok != token.RPAREN {
			return nil, fmt.Errorf("unexpected character, missing )")
		}
		return &ast.ParenExpr{
			Expr: expr,
		}, nil
	}
	switch tok {
	case token.IDENT:
		return &ast.VarRef{
			Value: lit,
		}, nil
	case token.STRING:
		return &ast.StringLiteral{Value: lit[1 : len(lit)-1]}, nil
	case token.NUMBER:
		v, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			return nil, lerrors.NewWrap("Cannot convert string to number", err)
		}
		return &ast.NumberLiteral{Value: v}, nil
	case token.TRUE, token.FALSE:
		return &ast.BooleanLiteral{Value: tok == token.TRUE}, nil
	case token.ARRAY:
		arrayValue := []any{}
		if err := json.Unmarshal([]byte("["+lit+"]"), &arrayValue); err != nil {
			return nil, lerrors.NewWrap("Cannot unmarshal string to array", err)
		}
		if len(arrayValue) == 0 {
			return nil, lerrors.New("Length of array must be greater than 0")
		}
		// Get type of first element from array
		switch t := arrayValue[0].(type) {
		case string:
			arrString := make([]string, 0)
			for _, v := range arrayValue {
				_v, ok := v.(string)
				if !ok {
					continue
				}
				arrString = append(arrString, _v)
			}
			return &ast.SliceStringLiteral{Value: arrString}, nil
		case float64:
			arrNumber := make([]float64, 0)
			for _, v := range arrayValue {
				_v, ok := v.(float64)
				if !ok {
					continue
				}
				arrNumber = append(arrNumber, _v)
			}
			return &ast.SliceNumberLiteral{Value: arrNumber}, nil
		default:
			return nil, lerrors.Newf("unknown type %s %T", t, t)
		}
	case token.JQ:
		extractJQMsg := func(msg string) (string, string, error) {
			jqMsg := ast.JQMsg{}
			if err := json.Unmarshal([]byte(msg), &jqMsg); err != nil {
				return "", "", lerrors.NewWrap("Cannot unmarshal JQMsg", err)
			}
			return jqMsg.TextToken, jqMsg.Mode, nil
		}
		qs, mode, err := extractJQMsg(lit)
		if err != nil {
			return nil, lerrors.NewWrap("Cannot extract query string, mode from JQMsg", err)
		}
		query, err := gojq.Parse(qs)
		if err != nil {
			return nil, lerrors.NewWrap("Cannot parse string to jq query", err)
		}
		return &ast.JQRef{
			Value: lit,
			Query: query,
			Mode:  mode,
		}, nil
	default:
		return nil, lerrors.Newf("Unknown token type %s, value: %v", tok, lit)
	}
}

// Parse expression to get ast.Expr
func (p *Parser) parseExpr() (ast.Expr, error) {
	expr, err := p.parseUnaryExpr()
	if err != nil {
		return nil, lerrors.NewWrap("Cannot parse unary expression", err)
	}
	for {
		op, tt := p.scanToken()
		if op == token.ILLEGAL {
			return nil, lerrors.Newf("Must be Operator expression, got: ILLEGAL")
		}
		if op == token.EOF || op == token.LPAREN || op == token.RPAREN {
			p.unscan()
			return expr, nil
		}
		if !op.IsOperator() {
			return expr, lerrors.Newf("Must be Operator expression, got: %v", tt)
		}
		rhs, err := p.parseUnaryExpr()
		if err != nil {
			return nil, lerrors.NewWrap("Cannot get unary expression for RHS", err)
		}
		expr = insertNode(expr, rhs, op)
	}
}

// Compare priority of operator to insert node into ast
func insertNode(l, r ast.Expr, op token.Token) ast.Expr {
	var (
		expr ast.Expr
	)
	lhs, ok := l.(*ast.BinaryExpr)
	if ok {
		if lhs.OP.Precedence() < op.Precedence() {
			return &ast.BinaryExpr{
				LHS: lhs.LHS,
				RHS: insertNode(lhs.RHS, r, op),
				OP:  lhs.OP,
			}
		}
	}
	expr = &ast.BinaryExpr{
		LHS: l,
		OP:  op,
		RHS: r,
	}
	return expr
}

func (p *Parser) Parse() (ast.Expr, error) {
	return p.parseExpr()
}
