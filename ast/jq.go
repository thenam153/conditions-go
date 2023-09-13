package ast

import "github.com/itchyny/gojq"

type JQMode int

const (
	JQFirst JQMode = iota
	JQLast
	JQArray
)

var JQModes = map[string]JQMode{
	"first": JQFirst,
	"last":  JQLast,
	"array": JQArray,
}

type JQMsg struct {
	TextToken string
	Mode      string
}

type JQRef struct {
	Value string
	Query *gojq.Query
	Mode  string
}
