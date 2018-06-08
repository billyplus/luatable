package lexer

import (
	"fmt"

	"github.com/billyplus/luasheet/token"
)

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

const eof = -1

type luaToken struct {
	typ  token.Type
	pos  Pos
	val  string
	line int
}

func (t *luaToken) String() string {
	switch t.typ {
	case token.EOF:
		return "EOF"
	case token.Error:
		return "Error"
	}
	return t.val
}

func (t *luaToken) Value() string {
	return t.val
}

func (t *luaToken) Position() string {
	return fmt.Sprintf("第%v行第%v个字符", t.line, t.pos)
}

func (t *luaToken) Type() token.Type {
	return t.typ
}
