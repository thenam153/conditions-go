package token

var (
	version        int              = 1
	defaultVersion int              = 1
	allowVersions  map[int]struct{} = map[int]struct{}{
		0: {},
		1: {},
	}
	openratorLevel []struct {
		Begin Token
		End   Token
		Value int
	} = []struct {
		Begin Token
		End   Token
		Value int
	}{
		{
			Begin: operatorBeginLevel1,
			End:   operatorEndLevel1,
			Value: 1,
		},
		{
			Begin: operatorBeginLevel2,
			End:   operatorEndLevel2,
			Value: 2,
		},
		{
			Begin: operatorBeginLevel3,
			End:   operatorEndLevel3,
			Value: 3,
		},
	}
)

type Token int

const (
	// ILLEGAL token represent invalid token found in statement
	ILLEGAL Token = iota
	// EOF token represent pointer reached end of statement
	EOF

	literalBegin
	IDENT
	NUMBER
	STRING
	ARRAY
	TRUE
	FALSE
	literalEnd

	funcBegin
	JQ
	funcEnd

	// Begin token represent operator
	operatorBegin
	operatorBeginLevel1
	OR
	XOR
	operatorEndLevel1

	operatorBeginLevel2
	AND
	NAND
	operatorEndLevel2

	operatorBeginLevel3
	EQ    // ==
	NEQ   // !=
	LT    // <
	LTE   // <=
	GT    // >
	GTE   // >=
	EREG  // =~
	NEREG // !~
	IN
	NOTIN
	operatorEnd
	operatorEndLevel3
	// End token represent operator

	LPAREN // (
	RPAREN // )
)

var Tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	IDENT:  "IDENT",
	NUMBER: "NUMBER",
	STRING: "STRING",
	ARRAY:  "ARRAY",
	TRUE:   "TRUE",
	FALSE:  "FALSE",

	JQ: "JQ",

	OR:  "OR",
	XOR: "XOR",

	AND:  "AND",
	NAND: "NAND",

	EQ:    "==",
	NEQ:   "!=",
	LT:    "<",
	LTE:   "<=",
	GT:    ">",
	GTE:   ">=",
	EREG:  "=~",
	NEREG: "!~",

	IN:    "IN",
	NOTIN: "NOT IN",

	LPAREN: "(",
	RPAREN: ")",
}

func (tok Token) String() string {
	if tok >= 0 && tok < Token(len(Tokens)) {
		return Tokens[tok]
	}
	return ""
}

func (tok Token) Precedence() int {
	if version == 1 {
		return tok.precedenceV1()
	} else {
		return tok.precedenceV2()
	}
}

func (tok Token) precedenceV1() int {
	switch tok {
	case OR, XOR:
		return 1
	case AND, NAND:
		return 2
	case EQ, NEQ, LT, LTE, GT, GTE, EREG, NEREG, IN, NOTIN:
		return 3
	}
	return 0
}

func (tok Token) precedenceV2() int {
	for _, opl := range openratorLevel {
		if int(tok) > int(opl.Begin) && int(tok) < int(opl.End) {
			return opl.Value
		}
	}
	return 0
}

func (tok Token) IsOperator() bool {
	return tok > operatorBegin && tok < operatorEnd
}

func SetVersion(v int) {
	if _, ok := allowVersions[v]; ok {
		version = v
	} else {
		version = defaultVersion
	}
}
